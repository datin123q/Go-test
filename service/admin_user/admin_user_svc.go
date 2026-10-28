package admin_user

import (
	"context"
	"time"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pwhelper"
)

var adminUserCSVColumns = []string{
	"model.adminstrator.name",         // 名前
	"model.adminstrator.email",        // メール
	"model.adminstrator.phone",        // 電話番号
	"model.adminstrator.sign_up_date", // 最終アクセス日
	"model.adminstrator.update_date",  // 修正日時
}

type adminUserService struct {
	translator             i18nutils.Translator
	pwhelper               pwhelper.PasswordHelper
	adminUserRepository    cdomain.AdminUserRepository
	refreshTokenRepository cdomain.RefreshTokenRepository
}

func New(
	translator i18nutils.Translator,
	pwHelper pwhelper.PasswordHelper,
	adminUserRepository cdomain.AdminUserRepository,
	refreshTokenRepository cdomain.RefreshTokenRepository,
) cdomain.AdminUserService {
	return &adminUserService{
		translator:             translator,
		pwhelper:               pwHelper,
		adminUserRepository:    adminUserRepository,
		refreshTokenRepository: refreshTokenRepository,
	}
}

func (a adminUserService) GetSuperAdmin(ctx context.Context) (domain.AdminUser, error) {
	opts := cdomain.AdminUserQueryOptionFunc()
	totalCount, adminUsers, err := a.adminUserRepository.GetAll(ctx,
		opts.WithLimit(1),
		opts.WithRoleCode(string(domain.ROLE_SUPER_ADMIN)),
	)
	if err != nil {
		return domain.AdminUser{}, errors.Wrap(err, "failed to get super admin")
	}
	if totalCount == 0 {
		return domain.AdminUser{}, cdomain.ErrItemNotFound
	}

	return adminUsers[0], nil
}

func (a adminUserService) ExportCSV(ctx context.Context, items []domain.AdminUser) cdomain.CSVData {
	csvHeaders := a.translator.Translates(adminUserCSVColumns)

	records := [][]string{csvHeaders}

	for _, item := range items {
		record := []string{
			item.Name,
			item.Email,
			item.MobilePhone,
			item.LastLoginDate.ToFormat(time.DateTime),
			item.Updated.ToFormat(time.DateTime),
		}
		records = append(records, record)
	}

	return records
}

func (a adminUserService) IsValidPeriod(ctx context.Context, adminUser domain.AdminUser) (bool, error) {
	if adminUser.IsSuperAdmin {
		return true, nil
	}

	now := time.Now()
	startTime := adminUser.ValidityStartDate.Time()
	endTime := adminUser.ValidityEndDate.Time()
	if startTime.Before(now) && endTime.After(now) {
		return true, nil
	}

	return false, nil
}
