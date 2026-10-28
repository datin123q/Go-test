package service

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type IMyNumberService interface {
	GetDuplicatedMyNumbersInApproveAndWait(context.Context, int64) ([]domain.MyNumber, error)
	GetMyNumberIdListByMyNumber(context.Context, string) ([]int64, error)
}

var _ IMyNumberService = (*MyNumberService)(nil)

type MyNumberService struct {
}

func NewMyNumberService() IMyNumberService {
	return &MyNumberService{}
}

// GetDuplicatedMyNumbersInApproveAndWait APPROVE または WAIT_APPROVE 状態のマイナンバー照会
func (s *MyNumberService) GetDuplicatedMyNumbersInApproveAndWait(ctx context.Context, myNumberId int64) ([]domain.MyNumber, error) {
	// myNumberId以外に同じMyNumberが登録されているかどうかを照会
	lookUpMyNumbers, err := domain.LookUpMyNumber{}.GetOthersWithSameMyNumber(ctx, myNumberId)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}

	// 重複するMyNumber中で承認中または待機中のものを照会
	var dupes []domain.MyNumber
	for _, lum := range lookUpMyNumbers {
		myNumberInfo, err := domain.MyNumber{}.GetById(ctx, lum.MyNumberId)
		if err != nil {
			return nil, factory.ErrorDB.New(err)
		}
		if myNumberInfo != nil && (myNumberInfo.Status == domain.APPROVE || myNumberInfo.Status == domain.WAIT_APPROVE) {
			dupes = append(dupes, *myNumberInfo)
		}
	}

	return dupes, nil
}

// GetMyNumberIdListByMyNumber MyNumber でMyNumberId照会
func (s *MyNumberService) GetMyNumberIdListByMyNumber(ctx context.Context, myNumber string) ([]int64, error) {
	var myNumberIdList []int64

	// encrypt MyNumber
	m := &domain.MyNumber{MyNumber: myNumber}
	encryptedMyNumber, err := m.EncryptMyNumber()
	if err != nil {
		return nil, factory.ErrorSystem.New(err)
	}

	// get MyNumberId list by encrypted MyNumber
	lookUpMyNumbers, err := domain.LookUpMyNumber{}.GetAllByMyNumberIncludeDeleted(ctx, encryptedMyNumber)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}

	// get MyNumberId list
	for _, lmn := range lookUpMyNumbers {
		myNumberIdList = append(myNumberIdList, lmn.MyNumberId)
	}

	return myNumberIdList, nil
}
