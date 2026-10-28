package mysql

import (
	"context"
	"fmt"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type withdrawalCloseMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.WithdrawalCloseRepository {
	return &withdrawalCloseMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (r withdrawalCloseMysqlRepository) Create(ctx context.Context, withdrawalClose domain.WithdrawalClose) (int64, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(&withdrawalClose); err != nil {
		return 0, errors.Wrap(err, "err insert withdrawal close")
	}

	return withdrawalClose.Id, nil
}

func (r withdrawalCloseMysqlRepository) Update(ctx context.Context, withdrawalClose domain.WithdrawalClose) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.Update(&withdrawalClose); err != nil {
		return errors.Wrap(err, "err update withdrawal close")
	}

	return nil
}

func (r withdrawalCloseMysqlRepository) UpdateMany(ctx context.Context, withdrawalClose domain.WithdrawalClose, options ...cdomain.WithdrawalCloseQueryOption) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if len(options) == 0 {
		return errors.New("no query options")
	}

	queryOpts := cdomain.WithdrawalCloseQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	if len(queryOpts.Ids()) > 0 {
		db = db.In(domain.WithdrawalClose{}.TableName()+".id", queryOpts.Ids())
	}

	if _, err = db.Update(&withdrawalClose); err != nil {
		return errors.Wrap(err, "err update withdrawal close status")
	}

	return nil
}

// GetById TC呼び出し用
func (r withdrawalCloseMysqlRepository) GetById(ctx context.Context, id int64) (domain.WithdrawalClose, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.WithdrawalClose{}, errors.Wrap(err, "err get db")
	}

	var withdrawalClose domain.WithdrawalClose
	has, err := db.ID(id).Get(&withdrawalClose)
	if err != nil {
		return domain.WithdrawalClose{}, errors.Wrap(err, "err get")
	}
	if !has {
		return domain.WithdrawalClose{}, cdomain.ErrItemNotFound
	}

	return withdrawalClose, nil
}

func (r withdrawalCloseMysqlRepository) GetAll(ctx context.Context, options ...cdomain.WithdrawalCloseQueryOption) (int64, []domain.WithdrawalCloseDetail, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.WithdrawalCloseQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	closedWithdrawals := make([]domain.WithdrawalCloseDetail, 0)

	// MIR-1113: 以前は無条件でテーブル名+"."+SortBy() と連結しており、
	// SortBy() が空のとき「withdrawal_close.」という不正な列名になっていた
	// (列名検証の追加後はこれが弾かれ、並び替えが無言で効かなくなる)。
	// SortOrderForTable は QualifySortColumns 経由で空の sortStr をそのまま
	// 返すため、空のときは何も付けない。
	sql := orm.Table(db, domain.WithdrawalClose{}).
		Select(domain.WithdrawalClose{}.SelectAll(), "admin_user.name as closer_name").
		LeftJoin(domain.AdminUser{}, fmt.Sprintf("%s.id = %s.closer_id", domain.AdminUser{}.TableName(), domain.WithdrawalClose{}.TableName())).
		SortOrderForTable(domain.WithdrawalClose{}.TableName(), queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset())

	totalCount, err := sql.FindAndCount(&closedWithdrawals)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err find closed withdrawals")
	}

	return totalCount, closedWithdrawals, nil
}
