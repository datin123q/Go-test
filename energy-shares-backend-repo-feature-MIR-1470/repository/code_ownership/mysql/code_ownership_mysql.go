package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/pkg/errors"
)

type codeOwnershipMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.CodeOwnershipRepository {
	return &codeOwnershipMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (r codeOwnershipMysqlRepository) Create(ctx context.Context, codeOwnership domain.CodeOwnership) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.InsertOne(&codeOwnership); err != nil {
		return errors.Wrap(err, "err insert code ownership")
	}

	return nil
}

func (r codeOwnershipMysqlRepository) GetByCode(ctx context.Context, codeType domain.CodeType, code string) (*domain.CodeOwnership, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "err get db")
	}

	var codeOwnership domain.CodeOwnership
	has, err := db.Where("code_type = ? AND code = ?", codeType, code).Get(&codeOwnership)
	if err != nil {
		return nil, errors.Wrap(err, "err get code ownership")
	}
	if !has {
		return nil, nil
	}

	return &codeOwnership, nil
}
