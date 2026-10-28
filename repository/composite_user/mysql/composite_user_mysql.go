package mysql

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type compositeUserRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.CompositeUserRepository {
	return &compositeUserRepository{
		dbFactory: dbFactory,
	}
}

func (r compositeUserRepository) GetAll(ctx context.Context, options ...cdomain.CompositeUserQueryOption) (int64, []cdomain.CompositeUser, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.CompositeUserQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	// 最終的に抽出するデータフィールド一覧
	selectFields := []string{
		"user.id AS id",
		"user.email AS email",
		"user.name AS name",
		"user.audit_status AS audit_status",
		"user.user_type AS user_type",
		"private_profile.nationality AS nationality",
		"private_profile.gender AS gender",
		"private_profile.job AS job",
		"private_profile.asset AS asset",
		"private_profile.annual_income AS annual_income",
		"user.is_newsletter AS allows_newsletter",
		"my_number.status AS my_number_status",
		"private_profile.birthday AS birthday",
		"user.created AS sign_up_request_date",
		"user.sign_up_approval_date AS sign_up_approval_date",
		"user.last_login_date AS last_login_date",
		"user_asset_view.total_asset AS total_asset",
		"user_asset_view.invest_operation_amount AS invest_operation_amount",
		"user_asset_view.deposit AS deposit",
		"user_asset_view.invest_success_amount AS invest_success_amount",
		"user_asset_view.invest_request_amount AS invest_request_amount",
		"user_asset_view.unpaid_amount AS unpaid_amount",
	}

	// 時間がかかっても良いクエリなので全てjoinしてからフィルタリング実行
	query := orm.Table(db, domain.User{}).
		Select(strings.Join(selectFields, ", ")).
		LeftJoin(domain.PrivateProfile{}).
		LeftJoin(domain.UserAssetView{}).
		LeftJoin(domain.MyNumber{},
			fmt.Sprintf("%s.id = %s.user_id AND (%s.deleted IS NULL OR %s.deleted = '0001-01-01 00:00:00')",
				domain.User{}.TableName(), domain.MyNumber{}.TableName(), domain.MyNumber{}.TableName(), domain.MyNumber{}.TableName()),
		)

	query.In(domain.User{}.TableName()+".email", queryOpts.Emails()...).
		In(domain.User{}.TableName()+".audit_status", queryOpts.AuditStatuses()...). // AuditStatuses、NoAuditStatusはいずれか一方のみ適用される(相互排他的)。
		NotEqual(domain.User{}.TableName()+".audit_status", queryOpts.NoAuditStatus()).
		In(domain.User{}.TableName()+".user_type", queryOpts.UserTypes()...).
		In(domain.PrivateProfile{}.TableName()+".nationality", queryOpts.Nationalities()...).
		In(domain.PrivateProfile{}.TableName()+".gender", queryOpts.Genders()...).
		In(domain.PrivateProfile{}.TableName()+".job", queryOpts.Jobs()...).
		In(domain.PrivateProfile{}.TableName()+".asset", queryOpts.Assets()...).
		In(domain.PrivateProfile{}.TableName()+".annual_income", queryOpts.AnnualIncomes()...).
		Equal(domain.User{}.TableName()+".is_newsletter", queryOpts.AllowsNewsletter()).
		Between(domain.User{}.TableName()+".created", queryOpts.MinSignUpRequestDate(), queryOpts.MaxSignUpRequestDate()).
		Between(domain.User{}.TableName()+".sign_up_approval_date", queryOpts.MinSignUpApprovalDate(), queryOpts.MaxSignUpApprovalDate()).
		Between(domain.User{}.TableName()+".last_login_date", queryOpts.MinLastLoginDate(), queryOpts.MaxLastLoginDate()).
		Between(domain.PrivateProfile{}.TableName()+".birthday", queryOpts.MinBirthday(), queryOpts.MaxBirthday()).
		Between(fmt.Sprintf("COALESCE(%s.deposit, 0)", domain.UserAssetView{}.TableName()), queryOpts.MinDeposit(), queryOpts.MaxDeposit()).
		Between(fmt.Sprintf("COALESCE(%s.invest_operation_amount, 0)", domain.UserAssetView{}.TableName()), queryOpts.MinInvestOperationAmount(), queryOpts.MaxInvestOperationAmount()).
		Between(fmt.Sprintf("COALESCE(%s.invest_success_amount, 0)", domain.UserAssetView{}.TableName()), queryOpts.MinInvestSuccessAmount(), queryOpts.MaxInvestSuccessAmount()).
		Between(fmt.Sprintf("COALESCE(%s.invest_request_amount, 0)", domain.UserAssetView{}.TableName()), queryOpts.MaxInvestRequestAmount(), queryOpts.MaxInvestRequestAmount()).
		Between(fmt.Sprintf("COALESCE(%s.unpaid_amount, 0)", domain.UserAssetView{}.TableName()), queryOpts.MinUnpaidAmount(), queryOpts.MaxUnpaidAmount()).
		Between(fmt.Sprintf("COALESCE(%s.total_asset, 0)", domain.UserAssetView{}.TableName()), queryOpts.MinTotalAsset(), queryOpts.MaxTotalAsset()).
		NotEqual(domain.User{}.TableName()+".user_type", string(domain.ADMIN)) // TODO: adminをuserテーブルから全て削除してadmin_userテーブルに分離し、検査後に削除しても良いコード

	if queryOpts.HasMyNumber() != nil {
		if *queryOpts.HasMyNumber() {
			query.Equal(domain.MyNumber{}.TableName()+".status", string(domain.APPROVE))
		} else {
			query.NotEqual(domain.MyNumber{}.TableName()+".status", string(domain.APPROVE))
		}
	}

	compositeUsers := make([]cdomain.CompositeUser, 0)
	totalCount, err := query.FindAndCount(&compositeUsers)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err find")
	}

	return totalCount, compositeUsers, nil
}
