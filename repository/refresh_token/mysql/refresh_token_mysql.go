package mysql

import (
	"context"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type refreshTokenMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.RefreshTokenRepository {
	return &refreshTokenMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (r refreshTokenMysqlRepository) Create(ctx context.Context, refreshToken domain.RefreshToken) (domain.RefreshToken, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.RefreshToken{}, errors.Wrap(err, "failed to get db")
	}

	if _, err = db.Insert(&refreshToken); err != nil {
		return domain.RefreshToken{}, errors.Wrap(err, "failed to insert refresh token")
	}

	return refreshToken, nil
}

func (r refreshTokenMysqlRepository) GetByToken(ctx context.Context, token string) (domain.RefreshToken, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.RefreshToken{}, errors.Wrap(err, "failed to get db")
	}

	refreshToken := domain.RefreshToken{}

	has, err := orm.Table(db, domain.RefreshToken{}).
		Select(domain.RefreshToken{}.SelectAll()).
		Equal(domain.RefreshToken{}.TableName()+".token", token).
		Get(&refreshToken)
	if err != nil {
		return domain.RefreshToken{}, errors.Wrap(err, "failed to get refresh token")
	}

	if !has {
		return domain.RefreshToken{}, cdomain.ErrItemNotFound
	}

	return refreshToken, nil
}

func (r refreshTokenMysqlRepository) Update(ctx context.Context, refreshToken domain.RefreshToken) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	affected, err := db.ID(refreshToken.Id).UseBool().Update(refreshToken)
	if err != nil {
		return errors.Wrap(err, "failed to update refresh token")
	}
	if affected == 0 {
		return cdomain.ErrItemNotAffected
	}

	return nil
}
