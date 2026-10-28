package my_number

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type myNumberUsecase struct {
	myNumberService cdomain.MyNumberService
	userRepository  cdomain.UserRepository
}

func New(
	myNumberService cdomain.MyNumberService,
	userRepository cdomain.UserRepository,
) udomain.MyNumberUsecase {
	return &myNumberUsecase{
		myNumberService: myNumberService,
		userRepository:  userRepository,
	}
}

func (m myNumberUsecase) Create(ctx context.Context, request udomain.MyNumberCreateRequest) (udomain.MyNumberCreateResponse, error) {
	// TODO: マイナンバー登録復活までエラーとします。
	return udomain.MyNumberCreateResponse{}, errors.Wrap(cdomain.ErrServiceUnavailable, "Cannot register my number.")

	// // check my number exists
	// myNumber, err := m.myNumberService.GetByUserId(ctx, request.JWTUser.Id)
	// if err != nil {
	// 	return udomain.MyNumberCreateResponse{}, errors.Wrap(err, "failed to get my number")
	// }
	// if myNumber.IsEmpty() {
	// 	myNumber.Status = domain.UNREGISTERED
	// }

	// if myNumber.Status == domain.APPROVE || myNumber.Status == domain.WAIT_APPROVE {
	// 	// 이미 승인된 마이넘버가 있는 경우 또는 승인 대기중인 경우 에러 반환
	// 	return udomain.MyNumberCreateResponse{}, errors.Wrap(cdomain.ErrAlreadyExists, "my number already exists")
	// }

	// // register my number
	// registeredMyNumber, err := m.myNumberService.RegisterWithFiles(ctx,
	// 	cdomain.MyNumberRegisterInput{
	// 		UserId:        request.JWTUser.Id,
	// 		MyNumberValue: request.MyNumber,
	// 		FrontFileUrl:  request.FrontFileUrl,
	// 		BackFileUrl:   request.BackFileUrl,
	// 		MyNumber:      myNumber,
	// 	})
	// if err != nil {
	// 	return udomain.MyNumberCreateResponse{}, errors.Wrap(err, "failed to register my number")
	// }

	// // update audit status
	// user, err := m.userRepository.GetById(ctx, request.JWTUser.Id)
	// if err != nil {
	// 	return udomain.MyNumberCreateResponse{}, errors.Wrap(err, "failed to get user for update audit status")
	// }
	// user.AuditStatus = domain.CHECK_MY_NUMBER_AUDIT

	// if err := m.userRepository.Update(ctx, user); err != nil {
	// 	return udomain.MyNumberCreateResponse{}, errors.Wrap(err, "failed to update user")
	// }

	// // return response
	// return udomain.MyNumberCreateResponse{
	// 	MyNumberBody: *udomain.MyNumberBody{}.FromMyNumber(registeredMyNumber),
	// }, nil
}

func (m myNumberUsecase) Get(ctx context.Context, request udomain.MyNumberGetRequest) (udomain.MyNumberGetResponse, error) {

	myNumber, err := m.myNumberService.GetByUserId(ctx, request.JWTUser.Id)
	if err != nil {
		return udomain.MyNumberGetResponse{}, errors.Wrap(err, "failed to get my number")
	}

	return udomain.MyNumberGetResponse{
		MyNumberBody: udomain.MyNumberBody{}.FromMyNumber(myNumber),
	}, nil
}
