package user_file

import (
	"context"
	"net/url"
	"path"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/filehelper"
	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

type userFileService struct {
	cfg                *config.Config
	s3pkg              s3utils.S3
	fileHelper         filehelper.FileHelper
	userFileRepository cdomain.UserFileRepository
}

func New(
	cfg *config.Config,
	s3pkg s3utils.S3,
	fileHelper filehelper.FileHelper,
	userFileRepository cdomain.UserFileRepository,
) cdomain.UserFileService {
	return &userFileService{
		cfg:                cfg,
		s3pkg:              s3pkg,
		fileHelper:         fileHelper,
		userFileRepository: userFileRepository,
	}
}

func (svc userFileService) CreateUserFile(ctx context.Context, userFile domain.UserFile) (domain.UserFile, error) {

	// ファイルパスとファイル名の抽出
	filePath, err := svc.fileHelper.ExtractFilePath(userFile.FileUrl)
	if err != nil {
		return domain.UserFile{}, errors.Wrap(err, "failed to extract file path")
	}
	userFile.FilePath = path.Dir(filePath) + "/"
	userFile.FileExtension = path.Ext(filePath)
	userFile.FileName = svc.fileHelper.ExtractFileName(filePath)

	// S3 からファイルメタ情報を取得
	header, err := svc.s3pkg.GetHeadObject(svc.cfg.AWS.S3_PRIVATE_BUCKET, userFile.GetKey())
	if err != nil {
		return domain.UserFile{}, errors.Wrap(err, "failed to get head object")
	}

	decodedFileName, err := url.QueryUnescape(header.OriginalFileName)
	if err != nil {
		return domain.UserFile{}, errors.Wrap(err, "failed to decode original file name")
	}

	userFile.OriginalName = decodedFileName
	userFile.FileSize = header.FileSize

	// ファイルメタ情報DB保存
	if err := svc.userFileRepository.Create(ctx, userFile); err != nil {
		return domain.UserFile{}, errors.Wrap(err, "failed to create user file")
	}

	return userFile, nil
}

func (svc userFileService) GetPrivateUserIdentificationFiles(ctx context.Context, userId int64) (cdomain.PrivateUserIdentificationFiles, error) {

	var identificationFiles cdomain.PrivateUserIdentificationFiles
	var err error

	total, front1, err := svc.userFileRepository.GetAll(ctx, cdomain.UserFileQueryOptionFunc().WithUserId(userId),
		cdomain.UserFileQueryOptionFunc().WithFileType(domain.IDENTIFICATION1_FRONT),
		cdomain.UserFileQueryOptionFunc().WithSortBy("created"),
		cdomain.UserFileQueryOptionFunc().WithOrder("desc"),
		cdomain.UserFileQueryOptionFunc().WithLimit(1),
	)
	if err != nil || total == 0 {
		return cdomain.PrivateUserIdentificationFiles{}, errors.Wrap(err, "failed to get IDENTIFICATION1_FRONT")
	}

	total, back1, err := svc.userFileRepository.GetAll(ctx, cdomain.UserFileQueryOptionFunc().WithUserId(userId),
		cdomain.UserFileQueryOptionFunc().WithFileType(domain.IDENTIFICATION1_BACK),
		cdomain.UserFileQueryOptionFunc().WithSortBy("created"),
		cdomain.UserFileQueryOptionFunc().WithOrder("desc"),
		cdomain.UserFileQueryOptionFunc().WithLimit(1),
	)
	if err != nil || total == 0 {
		return cdomain.PrivateUserIdentificationFiles{}, errors.Wrap(err, "failed to get IDENTIFICATION1_BACK")
	}

	total, front2, err := svc.userFileRepository.GetAll(ctx, cdomain.UserFileQueryOptionFunc().WithUserId(userId),
		cdomain.UserFileQueryOptionFunc().WithFileType(domain.IDENTIFICATION2_FRONT),
		cdomain.UserFileQueryOptionFunc().WithSortBy("created"),
		cdomain.UserFileQueryOptionFunc().WithOrder("desc"),
		cdomain.UserFileQueryOptionFunc().WithLimit(1),
	)
	if err != nil || total == 0 {
		return cdomain.PrivateUserIdentificationFiles{}, errors.Wrap(err, "failed to get IDENTIFICATION2_FRONT")
	}

	total, back2, err := svc.userFileRepository.GetAll(ctx, cdomain.UserFileQueryOptionFunc().WithUserId(userId),
		cdomain.UserFileQueryOptionFunc().WithFileType(domain.IDENTIFICATION2_BACK),
		cdomain.UserFileQueryOptionFunc().WithSortBy("created"),
		cdomain.UserFileQueryOptionFunc().WithOrder("desc"),
		cdomain.UserFileQueryOptionFunc().WithLimit(1),
	)
	if err != nil || total == 0 {
		return cdomain.PrivateUserIdentificationFiles{}, errors.Wrap(err, "failed to get IDENTIFICATION2_BACK")
	}

	front1[0].FileUrl = front1[0].GetCDNSignUrl()
	back1[0].FileUrl = back1[0].GetCDNSignUrl()
	front2[0].FileUrl = front2[0].GetCDNSignUrl()
	back2[0].FileUrl = back2[0].GetCDNSignUrl()

	identificationFiles.Front1 = front1[0]
	identificationFiles.Back1 = back1[0]
	identificationFiles.Front2 = front2[0]
	identificationFiles.Back2 = back2[0]

	return identificationFiles, nil
}

func (svc userFileService) GetCorporateUserIdentificationFiles(ctx context.Context, userId int64) (cdomain.CorporateUserIdentificationFiles, error) {

	var identificationFiles cdomain.CorporateUserIdentificationFiles

	_, front1, err := svc.userFileRepository.GetAll(ctx, cdomain.UserFileQueryOptionFunc().WithUserId(userId),
		cdomain.UserFileQueryOptionFunc().WithFileType(domain.IDENTIFICATION1_FRONT),
		cdomain.UserFileQueryOptionFunc().WithSortBy("created"),
		cdomain.UserFileQueryOptionFunc().WithOrder("desc"),
		cdomain.UserFileQueryOptionFunc().WithLimit(1),
	)
	if err != nil {
		return cdomain.CorporateUserIdentificationFiles{}, errors.Wrap(err, "failed to get IDENTIFICATION1_FRONT")
	}

	_, back1, err := svc.userFileRepository.GetAll(ctx, cdomain.UserFileQueryOptionFunc().WithUserId(userId),
		cdomain.UserFileQueryOptionFunc().WithFileType(domain.IDENTIFICATION1_BACK),
		cdomain.UserFileQueryOptionFunc().WithSortBy("created"),
		cdomain.UserFileQueryOptionFunc().WithOrder("desc"),
		cdomain.UserFileQueryOptionFunc().WithLimit(1),
	)
	if err != nil {
		return cdomain.CorporateUserIdentificationFiles{}, errors.Wrap(err, "failed to get IDENTIFICATION1_BACK")
	}

	_, front2, err := svc.userFileRepository.GetAll(ctx, cdomain.UserFileQueryOptionFunc().WithUserId(userId),
		cdomain.UserFileQueryOptionFunc().WithFileType(domain.IDENTIFICATION2_FRONT),
		cdomain.UserFileQueryOptionFunc().WithSortBy("created"),
		cdomain.UserFileQueryOptionFunc().WithOrder("desc"),
		cdomain.UserFileQueryOptionFunc().WithLimit(1),
	)
	if err != nil {
		return cdomain.CorporateUserIdentificationFiles{}, errors.Wrap(err, "failed to get IDENTIFICATION2_FRONT")
	}

	_, back2, err := svc.userFileRepository.GetAll(ctx, cdomain.UserFileQueryOptionFunc().WithUserId(userId),
		cdomain.UserFileQueryOptionFunc().WithFileType(domain.IDENTIFICATION2_BACK),
		cdomain.UserFileQueryOptionFunc().WithSortBy("created"),
		cdomain.UserFileQueryOptionFunc().WithOrder("desc"),
		cdomain.UserFileQueryOptionFunc().WithLimit(1),
	)
	if err != nil {
		return cdomain.CorporateUserIdentificationFiles{}, errors.Wrap(err, "failed to get IDENTIFICATION2_BACK")
	}

	_, certified, err := svc.userFileRepository.GetAll(ctx, cdomain.UserFileQueryOptionFunc().WithUserId(userId),
		cdomain.UserFileQueryOptionFunc().WithFileType(domain.CERTIFIED_COPY),
		cdomain.UserFileQueryOptionFunc().WithSortBy("created"),
		cdomain.UserFileQueryOptionFunc().WithOrder("desc"),
		cdomain.UserFileQueryOptionFunc().WithLimit(1),
	)
	if err != nil {
		return cdomain.CorporateUserIdentificationFiles{}, errors.Wrap(err, "failed to get CERTIFIED_COPY")
	}

	_, poa, err := svc.userFileRepository.GetAll(ctx, cdomain.UserFileQueryOptionFunc().WithUserId(userId),
		cdomain.UserFileQueryOptionFunc().WithFileType(domain.POA),
		cdomain.UserFileQueryOptionFunc().WithSortBy("created"),
		cdomain.UserFileQueryOptionFunc().WithOrder("desc"),
		cdomain.UserFileQueryOptionFunc().WithLimit(1),
	)
	if err != nil {
		return cdomain.CorporateUserIdentificationFiles{}, errors.Wrap(err, "failed to get POA")
	}

	_, aoi, err := svc.userFileRepository.GetAll(ctx, cdomain.UserFileQueryOptionFunc().WithUserId(userId),
		cdomain.UserFileQueryOptionFunc().WithFileType(domain.AOI),
		cdomain.UserFileQueryOptionFunc().WithSortBy("created"),
		cdomain.UserFileQueryOptionFunc().WithOrder("desc"),
		cdomain.UserFileQueryOptionFunc().WithLimit(1),
	)
	if err != nil {
		return cdomain.CorporateUserIdentificationFiles{}, errors.Wrap(err, "failed to get POA")
	}

	_, shareholders, err := svc.userFileRepository.GetAll(ctx, cdomain.UserFileQueryOptionFunc().WithUserId(userId),
		cdomain.UserFileQueryOptionFunc().WithFileType(domain.SHAREHOLDERS),
		cdomain.UserFileQueryOptionFunc().WithSortBy("created"),
		cdomain.UserFileQueryOptionFunc().WithOrder("desc"),
		cdomain.UserFileQueryOptionFunc().WithLimit(1),
	)
	if err != nil {
		return cdomain.CorporateUserIdentificationFiles{}, errors.Wrap(err, "failed to get POA")
	}

	_, others, err := svc.userFileRepository.GetAll(ctx, cdomain.UserFileQueryOptionFunc().WithUserId(userId),
		cdomain.UserFileQueryOptionFunc().WithFileType(domain.USER_FILE_OTHERS),
		cdomain.UserFileQueryOptionFunc().WithSortBy("created"),
		cdomain.UserFileQueryOptionFunc().WithOrder("desc"),
		cdomain.UserFileQueryOptionFunc().WithLimit(1),
	)
	if err != nil {
		return cdomain.CorporateUserIdentificationFiles{}, errors.Wrap(err, "failed to get POA")
	}

	if len(front1) > 0 {
		front1[0].FileUrl = front1[0].GetCDNSignUrl()
		identificationFiles.Front1 = &front1[0]
	}

	if len(back1) > 0 {
		back1[0].FileUrl = back1[0].GetCDNSignUrl()
		identificationFiles.Back1 = &back1[0]
	}

	if len(front2) > 0 {
		front2[0].FileUrl = front2[0].GetCDNSignUrl()
		identificationFiles.Front2 = &front2[0]
	}

	if len(back2) > 0 {
		back2[0].FileUrl = back2[0].GetCDNSignUrl()
		identificationFiles.Back2 = &back2[0]
	}

	if len(certified) > 0 {
		certified[0].FileUrl = certified[0].GetCDNSignUrl()
		identificationFiles.CertifiedCopy = &certified[0]
	}

	if len(poa) > 0 {
		poa[0].FileUrl = poa[0].GetCDNSignUrl()
		identificationFiles.POA = &poa[0]
	}

	if len(aoi) > 0 {
		aoi[0].FileUrl = aoi[0].GetCDNSignUrl()
		identificationFiles.AOI = &aoi[0]
	}

	if len(shareholders) > 0 {
		shareholders[0].FileUrl = shareholders[0].GetCDNSignUrl()
		identificationFiles.Shareholders = &shareholders[0]
	}

	if len(others) > 0 {
		others[0].FileUrl = others[0].GetCDNSignUrl()
		identificationFiles.Others = &others[0]
	}

	return identificationFiles, nil
}
