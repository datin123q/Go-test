package look_up_my_number

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/cipher"
)

type lookUpMyNumberService struct {
	crypto                   cipher.Crypto
	lookUpMyNumberRepository cdomain.LookUpMyNumberRepository
}

func New(
	crypto cipher.Crypto,
	lookUpMyNumberRepository cdomain.LookUpMyNumberRepository,
) cdomain.LookUpMyNumberService {
	return &lookUpMyNumberService{
		crypto:                   crypto,
		lookUpMyNumberRepository: lookUpMyNumberRepository,
	}
}

// GetDecryptedMyNumberByMyNumberId my_number_idに対応するmy_numberを取得します。
func (svc lookUpMyNumberService) GetDecryptedMyNumberByMyNumberId(ctx context.Context, myNumberId int64) (string, error) {

	lookUpMyNumber, err := svc.lookUpMyNumberRepository.GetByMyNumberIdIncludeDeleted(ctx, myNumberId)
	if err != nil {
		return "", err
	}
	if lookUpMyNumber.Id == 0 {
		return "", cdomain.ErrItemNotFound
	}

	decryptedMyNumber, err := svc.crypto.Decrypt(lookUpMyNumber.MyNumber)
	if err != nil {
		return "", errors.Wrap(err, "failed decrypt MyNumber")
	}

	return decryptedMyNumber, nil
}
