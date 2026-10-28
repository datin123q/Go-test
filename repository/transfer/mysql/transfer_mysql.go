package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type transferMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.TransferRepository {
	return &transferMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (r transferMysqlRepository) GetAll(ctx context.Context, options ...cdomain.TransferQueryOption) (int64, []domain.Transfer, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.TransferQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	transfers := make([]domain.Transfer, 0)

	// MIR-1113: 以前は無条件でテーブル名+"."+SortBy() と連結しており、
	// SortBy() が空のとき「transfer.」という不正な列名になっていた
	// (列名検証の追加後はこれが弾かれ、並び替えが無言で効かなくなる)。
	// SortOrderForTable は QualifySortColumns 経由で空の sortStr をそのまま
	// 返すため、空のときは何も付けない。
	query := orm.Table(db, domain.Transfer{}).
		Equal(domain.Transfer{}.TableName()+".request_id", queryOpts.RequestId()).
		SortOrderForTable(domain.Transfer{}.TableName(), queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset())

	if queryOpts.WithdrawalId() != nil {
		query = query.
			InnerJoin(domain.WithdrawalTransferMapping{}, domain.WithdrawalTransferMapping{}.TableName()+".transaction_id = "+domain.Transfer{}.TableName()+".transaction_id").
			Equal(domain.WithdrawalTransferMapping{}.TableName()+".withdrawal_id", queryOpts.WithdrawalId())
	}

	totalCount, err := query.FindAndCount(&transfers)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err find and count")
	}

	return totalCount, transfers, nil
}
