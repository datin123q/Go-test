package service

import (
	"context"
	"fmt"

	"github.com/illuminarean/energy-shares-backend-repo/service/dto"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type ICorporateAdminService interface {
	GetAll(ctx context.Context, id, name, email, investorType, auditStatus *string, sortby, order string, offset, limit int, startDate, endDate string, profileCond domain.ProfileSearchCondition) (int64, []*dto.CorporateUserItem, error)
	GetById(ctx context.Context, id int64) (*dto.CorporateUserAdminDetail, error)
}

var _ ICorporateAdminService = (*CorporateAdminService)(nil)

type CorporateAdminService struct{}

func NewCorporateUserAdminService() ICorporateAdminService {
	return &CorporateAdminService{}
}

func (s *CorporateAdminService) GetAll(ctx context.Context, id, name, email, investorType, auditStatus *string, sortby, order string, offset, limit int, startDate, endDate string, profileCond domain.ProfileSearchCondition) (int64, []*dto.CorporateUserItem, error) {
	var users []domain.CorporateUser
	totalCount, users, err := domain.CorporateUser{}.GetAll(ctx, id, name, email, investorType, auditStatus, sortby, order, offset, limit, startDate, endDate, profileCond)
	if err != nil {
		return 0, nil, err
	}
	items := make([]*dto.CorporateUserItem, 0)
	for i := range users {
		items = append(items, dto.CorporateUserItem{}.New(users[i]))
	}
	return totalCount, items, nil
}

func FindCorporateUserIdentificationFiles(ctx context.Context, user *dto.CorporateUserAdminDetail) (err error) {
	var cu domain.CorporateUser
	cu.User.Id = user.Id
	if err := cu.GetIdentificationFiles(ctx); err != nil {
		return err
	}
	user.Identification1FrontFile = cu.Identification1FrontFile
	user.Identification1BackFile = cu.Identification1BackFile
	user.Identification2FrontFile = cu.Identification2FrontFile
	user.Identification2BackFile = cu.Identification2BackFile
	user.CertifiedCopyFiles = cu.CertifiedCopyFiles
	user.PoaFiles = cu.PoaFiles
	user.AoiFiles = cu.AoiFiles
	user.ShareholdersFiles = cu.ShareholdersFiles
	user.OthersFiles = cu.OthersFiles
	return err
}

func (s *CorporateAdminService) GetById(ctx context.Context, id int64) (*dto.CorporateUserAdminDetail, error) {
	var user *dto.CorporateUserAdminDetail
	var err error = nil
	item, err := domain.CorporateUser{}.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("user id %d not found", id)
	}
	equity, err := domain.UserVirtualAccount{}.GetVaByUserId(ctx, id, domain.EQUITY)
	if err != nil {
		return nil, err
	}
	antisocialCheck, err := domain.UserAntisocialCheck{}.GetByCheckUserId(ctx, id)
	if err != nil {
		return nil, err
	}
	account, err := domain.UserAccount{}.GetByUserId(ctx, item.User.Id)
	if err != nil {
		return nil, err
	}
	user = dto.CorporateUserAdminDetail{}.New(*item)
	err = FindCorporateUserIdentificationFiles(ctx, user)
	if err != nil {
		return nil, err
	}
	user.Roles = []string{string(domain.ROLE_USER)}
	user.EquityVaAccount = equity
	user.AntisocialCheck = antisocialCheck
	user.UserAccount = account
	return user, nil
}
