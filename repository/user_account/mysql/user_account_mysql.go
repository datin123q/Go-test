package mysql

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type userAccountMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.UserAccountRepository {
	return &userAccountMysqlRepository{
		dbFactory: dbFactory,
	}
}
func (u userAccountMysqlRepository) Create(ctx context.Context, userAccount domain.UserAccount) error {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err = db.Insert(userAccount); err != nil {
		return errors.Wrap(err, "failed to insert user account")
	}

	return nil
}

// MIR-1247: userAccountはversion列を持つ楽観ロック対象だが、以前は
// db.Update()の影響行数(第1戻り値)を捨てていたため、version不一致で
// 0行更新になってもerrがnilで返り、呼び出し元(DepositCheckProcess)は
// 残高が減っていないのに成功したものとして先へ進んでいた。domain側の
// WithdrawBalance/DepositBalanceで同じ理由の不具合を直した(applyBalanceDelta)
// のと同一の原因なので、ここでも影響行数を検査してerrorを返す。0行更新は
// 同じリポジトリ層の他の実装(user_login_lock・role・refresh_token・
// corporate_shareholderの各mysqlリポジトリ)と同様にcdomain.ErrItemNotAffected
// を返し、errors.Isで判別できる形に揃える。userAccountは値渡しでこの関数の
// ローカルコピーにすぎないため、呼び出し元の変数を汚さないための復元処理は
// 不要。
func (u userAccountMysqlRepository) Update(ctx context.Context, userAccount domain.UserAccount) error {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	idx, err := db.ID(userAccount.Id).UseBool().Update(userAccount)
	if err != nil {
		return errors.Wrap(err, "failed to update user account")
	}
	if idx == 0 {
		return errors.Wrapf(cdomain.ErrItemNotAffected, "account_id=%d, version=%d", userAccount.Id, userAccount.Version)
	}

	return nil
}

func (u userAccountMysqlRepository) GetByUserId(ctx context.Context, userId int64) (domain.UserAccount, error) {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.UserAccount{}, errors.Wrap(err, "failed to get db")
	}

	userAccount := domain.UserAccount{}

	has, err := orm.Table(db, domain.User{}).
		Select(domain.USER_ACCOUNT_COLUMN).
		InnerJoin(domain.UserAccount{}).
		Equal(domain.User{}.TableName()+".id", lo.ToPtr(fmt.Sprintf("%v", userId))).
		Get(&userAccount)
	if err != nil {
		return domain.UserAccount{}, errors.Wrap(err, "failed to get user account")
	}

	if !has {
		return domain.UserAccount{}, errors.New("user account not found")
	}

	return userAccount, nil
}
