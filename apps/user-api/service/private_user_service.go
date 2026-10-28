package service

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/service"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energyshares-user-api/service/dto"
)

type PrivateUserService struct {
	userInfoHistoryService service.IUserInfoHistoryService
}

func NewPrivateUserService() *PrivateUserService {
	return &PrivateUserService{
		userInfoHistoryService: service.NewUserInfoHistoryService(),
	}
}

func (s *PrivateUserService) GetEkycRequestLog(ctx context.Context, ekycId string) (*domain.EkycRequestLog, error) {
	ekycReqLog := domain.EkycRequestLog{}

	item, err := ekycReqLog.GetByEkycId(ctx, ekycId)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}

	return item, nil
}

func (s *PrivateUserService) GetById(ctx context.Context, id int64) (*dto.PrivateUserDetail, error) {

	var err error = nil

	item, err := domain.PrivateUser{}.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	equity, err := domain.UserVirtualAccount{}.GetVaByUserId(ctx, id, domain.EQUITY)
	if err != nil {
		return nil, err
	}

	antisocialCheck, err := domain.UserAntisocialCheck{}.GetByCheckUserId(ctx, id)
	if err != nil {
		return nil, err
	}

	account, err := domain.UserAccount{}.GetByUserId(ctx, id)
	if err != nil {
		return nil, err
	}

	myNumber, err := domain.MyNumber{}.GetByUserId(ctx, id)
	if err != nil {
		return nil, err
	}

	user := *dto.PrivateUserDetail{}.New(item, equity, antisocialCheck, account, myNumber)

	err = FindPrivateUserIdentificationFiles(ctx, &user)
	if err != nil {
		return nil, err
	}

	user.Roles = []string{string(domain.ROLE_USER)}

	return &user, nil
}
func FindPrivateUserIdentificationFiles(ctx context.Context, user *dto.PrivateUserDetail) (err error) {
	user.Identification1FrontFile, err = domain.UserFile{}.FindOneUserIdByAndFileType(ctx, user.UserId, domain.IDENTIFICATION1_FRONT)
	if err != nil {
		return err
	}
	user.Identification1BackFile, err = domain.UserFile{}.FindOneUserIdByAndFileType(ctx, user.UserId, domain.IDENTIFICATION1_BACK)
	if err != nil {
		return err
	}

	user.Identification2FrontFile, err = domain.UserFile{}.FindOneUserIdByAndFileType(ctx, user.UserId, domain.IDENTIFICATION2_FRONT)
	if err != nil {
		return err
	}
	user.Identification2BackFile, err = domain.UserFile{}.FindOneUserIdByAndFileType(ctx, user.UserId, domain.IDENTIFICATION2_BACK)
	if err != nil {
		return err
	}

	return err
}
