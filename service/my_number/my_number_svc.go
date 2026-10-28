package my_number

import (
	"context"
	"strconv"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/cipher"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/idgenerator"
	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type myNumberService struct {
	idGenerator              idgenerator.IdGenerator
	crypto                   cipher.Crypto
	lookUpMyNumberRepository cdomain.LookUpMyNumberRepository
	myNumberRepository       cdomain.MyNumberRepository
	userFileRepository       cdomain.UserFileRepository
	lookUpMyNumberService    cdomain.LookUpMyNumberService
	userFileService          cdomain.UserFileService
}

func New(
	idGenerator idgenerator.IdGenerator,
	crypto cipher.Crypto,
	lookUpMyNumberRepository cdomain.LookUpMyNumberRepository,
	myNumberRepository cdomain.MyNumberRepository,
	userFileRepository cdomain.UserFileRepository,
	lookUpMyNumberService cdomain.LookUpMyNumberService,
	userFileService cdomain.UserFileService,
) cdomain.MyNumberService {
	return &myNumberService{
		idGenerator:              idGenerator,
		crypto:                   crypto,
		lookUpMyNumberRepository: lookUpMyNumberRepository,
		myNumberRepository:       myNumberRepository,
		userFileRepository:       userFileRepository,
		lookUpMyNumberService:    lookUpMyNumberService,
		userFileService:          userFileService,
	}
}

// RegisterWithFiles my_numberをファイルと一緒に登録します。
func (svc myNumberService) RegisterWithFiles(ctx context.Context, input cdomain.MyNumberRegisterInput) (domain.MyNumber, error) {

	// encrypt
	encryptedMyNumberValue, err := svc.crypto.Encrypt(input.MyNumberValue)
	if err != nil {
		return domain.MyNumber{}, errors.Wrap(err, "failed to encrypt my number")
	}

	// check my number validation
	alreadyExists, err := svc.CheckApproveMyNumberExists(ctx, encryptedMyNumberValue)
	if err != nil {
		return domain.MyNumber{}, errors.Wrap(err, "failed to check approve my number exists")
	}
	if alreadyExists {
		return domain.MyNumber{}, errors.Wrap(cdomain.ErrAlreadyExists, "my number already exists")
	}

	// 作成前にmyNumberを削除
	if !input.MyNumber.IsEmpty() {
		if err := svc.myNumberRepository.Delete(ctx, input.MyNumber.Id); err != nil {
			return domain.MyNumber{}, errors.Wrap(err, "failed to delete my number")
		}
		if err := svc.lookUpMyNumberRepository.DeleteByMyNumberId(ctx, input.MyNumber.Id); err != nil {
			return domain.MyNumber{}, errors.Wrap(err, "failed to delete look up my number")
		}
	}

	// MyNumber作成
	myNumberInfo := domain.MyNumber{
		UserId:   input.UserId,
		Status:   domain.WAIT_APPROVE,
		CommonId: svc.idGenerator.GenUuidV4().String(),
	}
	myNumberInfo.Id, err = svc.myNumberRepository.Create(ctx, myNumberInfo)
	if err != nil {
		return domain.MyNumber{}, errors.Wrap(err, "failed to create my number")
	}

	// LookUpMyNumber作成
	if _, err := svc.lookUpMyNumberRepository.Create(ctx, domain.LookUpMyNumber{
		MyNumber:   encryptedMyNumberValue,
		MyNumberId: myNumberInfo.Id,
	}); err != nil {
		return domain.MyNumber{}, errors.Wrap(err, "failed to create look up my number")
	}

	// my numberファイル保存
	var frontFile, backFile domain.UserFile
	frontFile, err = svc.userFileService.CreateUserFile(ctx, domain.UserFile{
		UserId:   input.UserId,
		OwnerId:  input.UserId,
		UpdateId: input.UserId,
		UserType: domain.PRIVATE,
		FileType: domain.MY_NUMBER_FRONT,
		FileUrl:  input.FrontFileUrl,
		Etc:      strconv.Itoa(int(myNumberInfo.Id)),
	})
	if err != nil {
		return domain.MyNumber{}, errors.Wrap(err, "failed to upload front file")
	}

	backFile, err = svc.userFileService.CreateUserFile(ctx, domain.UserFile{
		UserId:   input.UserId,
		OwnerId:  input.UserId,
		UpdateId: input.UserId,
		UserType: domain.PRIVATE,
		FileType: domain.MY_NUMBER_BACK,
		FileUrl:  input.BackFileUrl,
		Etc:      strconv.Itoa(int(myNumberInfo.Id)),
	})
	if err != nil {
		return domain.MyNumber{}, errors.Wrap(err, "failed to upload back file")
	}

	myNumberInfo.FrontFile = &frontFile
	myNumberInfo.BackFile = &backFile
	myNumberInfo.MyNumber = input.MyNumberValue
	return myNumberInfo, nil
}

// GetByUserIdはuser_idに対応するmy_numberを取得します。
func (svc myNumberService) GetByUserId(ctx context.Context, userId int64) (domain.MyNumber, error) {

	myNumber, err := svc.myNumberRepository.GetByUserId(ctx, userId)
	if err != nil {
		if errors.Is(err, cdomain.ErrItemNotFound) {
			return domain.MyNumber{}, nil
		}
		return domain.MyNumber{}, err
	}

	decryptedMyNumber, err := svc.lookUpMyNumberService.GetDecryptedMyNumberByMyNumberId(ctx, myNumber.Id)
	if err != nil {
		return domain.MyNumber{}, err
	}
	myNumber.MyNumber = decryptedMyNumber

	total, frontFile, err := svc.userFileRepository.GetAll(ctx,
		cdomain.UserFileQueryOptionFunc().WithUserId(userId),
		cdomain.UserFileQueryOptionFunc().WithFileType(domain.MY_NUMBER_FRONT),
		cdomain.UserFileQueryOptionFunc().WithEtc(strconv.Itoa(int(myNumber.Id))),
		cdomain.UserFileQueryOptionFunc().WithSortBy("created"),
		cdomain.UserFileQueryOptionFunc().WithOrder("desc"),
		cdomain.UserFileQueryOptionFunc().WithLimit(1),
	)
	if err != nil || total == 0 {
		return domain.MyNumber{}, err
	}
	total, backFile, err := svc.userFileRepository.GetAll(ctx,
		cdomain.UserFileQueryOptionFunc().WithUserId(userId),
		cdomain.UserFileQueryOptionFunc().WithFileType(domain.MY_NUMBER_BACK),
		cdomain.UserFileQueryOptionFunc().WithEtc(strconv.Itoa(int(myNumber.Id))),
		cdomain.UserFileQueryOptionFunc().WithSortBy("created"),
		cdomain.UserFileQueryOptionFunc().WithOrder("desc"),
		cdomain.UserFileQueryOptionFunc().WithLimit(1),
	)
	if err != nil || total == 0 {
		return domain.MyNumber{}, err
	}

	frontFile[0].FileUrl = frontFile[0].GetCDNSignUrl()
	backFile[0].FileUrl = backFile[0].GetCDNSignUrl()

	myNumber.FrontFile = &frontFile[0]
	myNumber.BackFile = &backFile[0]

	return myNumber, nil
}

// CheckApproveMyNumberExists APPROVE状態の同一のマイナンバーの存在確認
func (svc myNumberService) CheckApproveMyNumberExists(ctx context.Context, myNumber string) (bool, error) {
	// MyNumber DBに登録されているかどうかを照会
	lookUpMyNumber, err := svc.lookUpMyNumberRepository.GetByMyNumber(ctx, myNumber)
	if err != nil {
		if errors.Is(err, cdomain.ErrItemNotFound) { // MyNumber DBにない場合: false
			return false, nil
		}
		return false, errors.Wrap(err, "failed to get look up my number")
	}

	// ESDBからmy numberを照会
	myNumberInfo, err := svc.myNumberRepository.GetById(ctx, lookUpMyNumber.MyNumberId)
	if err != nil {
		return false, errors.Wrap(err, "failed to get my number")
	}

	// MyNumber DBに既に登録されている場合: true
	if myNumberInfo.Status == domain.APPROVE {
		return true, nil
	}

	// MyNumber DBにあるが承認されていない場合: false
	return false, nil
}
