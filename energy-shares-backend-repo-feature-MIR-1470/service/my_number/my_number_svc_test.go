package my_number

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/cipher"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/idgenerator"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_MyNumberService_RegisterWithFiles(t *testing.T) {
	uid, _ := uuid.FromString(mockMyNumber2.CommonId)

	type args struct {
		ctx     context.Context
		request cdomain.MyNumberRegisterInput
	}

	tests := []struct {
		name                     string
		args                     args
		logger                   factory.LoggerFactory
		crypto                   func() cipher.Crypto
		idGenerator              func() idgenerator.IdGenerator
		myNumberRepository       func() cdomain.MyNumberRepository
		lookUpMyNumberRepository func() cdomain.LookUpMyNumberRepository
		userFileService          func() cdomain.UserFileService
		want                     domain.MyNumber
		wantErr                  bool
	}{
		{
			name: "(성공) 마이넘버 등록",
			args: args{
				ctx: context.Background(),
				request: cdomain.MyNumberRegisterInput{
					UserId:        mockMyNumber2.UserId,
					MyNumberValue: mockMyNumber2.MyNumber,
					FrontFileUrl:  mockMyNumber2Front.FileUrl,
					BackFileUrl:   mockMyNumber2Back.FileUrl,
					MyNumber:      domain.MyNumber{Status: domain.UNREGISTERED},
				},
			},
			crypto: func() cipher.Crypto {
				crypto := &mocks.Crypto{}
				crypto.On("Encrypt", mockMyNumber2.MyNumber).Return("encrypted my number", nil)
				return crypto
			},
			idGenerator: func() idgenerator.IdGenerator {
				idGenerator := &mocks.IdGenerator{}
				idGenerator.On("GenUuidV4").Return(uid)
				return idGenerator
			},
			myNumberRepository: func() cdomain.MyNumberRepository {
				repo := &mocks.MyNumberRepository{}
				repo.On("Create", context.Background(), domain.MyNumber{
					UserId:   mockMyNumber2.UserId,
					Status:   domain.WAIT_APPROVE,
					CommonId: uid.String(),
				}).Return(mockMyNumber2.Id, nil)
				return repo
			},
			lookUpMyNumberRepository: func() cdomain.LookUpMyNumberRepository {
				repo := &mocks.LookUpMyNumberRepository{}
				repo.On("GetByMyNumber", context.Background(), "encrypted my number").Return(domain.LookUpMyNumber{}, cdomain.ErrItemNotFound)
				repo.On("Create", context.Background(), domain.LookUpMyNumber{
					MyNumber:   "encrypted my number",
					MyNumberId: mockMyNumber2.Id,
				}).Return(int64(1), nil)
				return repo
			},
			userFileService: func() cdomain.UserFileService {
				service := &mocks.UserFileService{}
				f := mockMyNumber2Front
				f.Id = 0
				b := mockMyNumber2Back
				b.Id = 0
				service.On("CreateUserFile", context.Background(), f).Return(mockMyNumber2Front, nil)
				service.On("CreateUserFile", context.Background(), b).Return(mockMyNumber2Back, nil)
				return service
			},
			want: domain.MyNumber{
				Id:        mockMyNumber2.Id,
				MyNumber:  mockMyNumber2.MyNumber,
				Status:    mockMyNumber2.Status,
				FrontFile: &mockMyNumber2Front,
				BackFile:  &mockMyNumber2Back,
			},
			wantErr: false,
		},
		{
			name: "(실패) 마이넘버 등록 - 마이넘버가 이미 등록된 상태인 경우",
			args: args{
				ctx: context.Background(),
				request: cdomain.MyNumberRegisterInput{
					UserId:        mockMyNumber2.UserId,
					MyNumberValue: mockMyNumber2.MyNumber,
					FrontFileUrl:  mockMyNumber2Front.FileUrl,
					BackFileUrl:   mockMyNumber2Back.FileUrl,
				},
			},
			crypto: func() cipher.Crypto {
				crypto := &mocks.Crypto{}
				crypto.On("Encrypt", mockMyNumber2.MyNumber).Return("encrypted my number", nil)
				return crypto
			},
			idGenerator: func() idgenerator.IdGenerator {
				idGenerator := &mocks.IdGenerator{}
				return idGenerator
			},
			myNumberRepository: func() cdomain.MyNumberRepository {
				repo := &mocks.MyNumberRepository{}
				repo.On("GetById", context.Background(), mockMyNumber2.Id).Return(domain.MyNumber{
					Id:     int64(1),
					Status: domain.APPROVE,
				}, nil)
				return repo
			},
			lookUpMyNumberRepository: func() cdomain.LookUpMyNumberRepository {
				repo := &mocks.LookUpMyNumberRepository{}
				repo.On("GetByMyNumber", context.Background(), "encrypted my number").Return(domain.LookUpMyNumber{MyNumberId: int64(1)}, nil)
				return repo
			},
			userFileService: func() cdomain.UserFileService {
				service := &mocks.UserFileService{}
				return service
			},
			want:    domain.MyNumber{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := myNumberService{
				crypto:                   tt.crypto(),
				idGenerator:              tt.idGenerator(),
				myNumberRepository:       tt.myNumberRepository(),
				lookUpMyNumberRepository: tt.lookUpMyNumberRepository(),
				userFileService:          tt.userFileService(),
			}

			ignoreFields := cmpopts.IgnoreFields(domain.MyNumber{}, "Id", "FrontFile", "BackFile")
			ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := service.RegisterWithFiles(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.wantErr, err != nil, "RegisterWithFiles() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(tt.want.MyNumber, got.MyNumber, ignoreTypes, ignoreFields), true, "RegisterWithFiles() = %+v, want %+v", got, tt.want)
			if tt.want.FrontFile != nil {
				assert.Equalf(t, cmp.Equal(*tt.want.FrontFile, *got.FrontFile, ignoreTypes), true, "RegisterWithFiles() = %+v, want %+v", *got.FrontFile, *tt.want.FrontFile)
			}
			if tt.want.BackFile != nil {
				assert.Equalf(t, cmp.Equal(*tt.want.BackFile, *got.BackFile, ignoreTypes), true, "RegisterWithFiles() = %+v, want %+v", *got.BackFile, *tt.want.BackFile)
			}
		})
	}
}

func Test_MyNumberService_GetByUserId(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name                  string
		args                  args
		myNumberRepository    func() *mocks.MyNumberRepository
		lookUpMyNumberService func() *mocks.LookUpMyNumberService
		userFileRepository    func() *mocks.UserFileRepository
		want                  domain.MyNumber
		wantErr               error
	}{
		{
			name: "(성공) MyNumber 조회 (Id:1000)",
			args: args{
				ctx:    context.Background(),
				userId: mockMyNumber1.UserId,
			},
			myNumberRepository: func() *mocks.MyNumberRepository {
				myNumberRepository := &mocks.MyNumberRepository{}
				myNumberRepository.On("GetByUserId", context.Background(),
					mockMyNumber1.UserId).
					Return(mockMyNumberFinal, nil)
				return myNumberRepository
			},
			lookUpMyNumberService: func() *mocks.LookUpMyNumberService {
				lookUpMyNumberService := &mocks.LookUpMyNumberService{}
				lookUpMyNumberService.On("GetDecryptedMyNumberByMyNumberId",
					context.Background(), mockMyNumber1.Id).
					Return(decryptedMyNumber, nil)
				return lookUpMyNumberService
			},
			userFileRepository: func() *mocks.UserFileRepository {
				userFileRepository := &mocks.UserFileRepository{}
				userFileRepository.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)")).
					Return(int64(1), domain.UserFiles{mockUserFile3}, nil)
				return userFileRepository
			},
			want:    mockMyNumberFinal,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := myNumberService{
				myNumberRepository:    tt.myNumberRepository(),
				lookUpMyNumberService: tt.lookUpMyNumberService(),
				userFileRepository:    tt.userFileRepository(),
			}

			got, err := service.GetByUserId(tt.args.ctx, tt.args.userId)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("MyNumberService.GetByUserId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equalf(t, tt.want, got, "MyNumberService.GetByUserId() = %v, want %v", got, tt.want)
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			ignoreFields := cmpopts.IgnoreFields(domain.MyNumber{}, "FrontFile", "BackFile", "FrontFile.FileUrl", "BackFile.FileUrl")
			if diff := cmp.Diff(tt.want, got, ignoreTime, ignoreFields); diff != "" {
				t.Errorf("MyNumberService.GetByUserId() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_MyNumberService_CheckApproveMyNumberExists(t *testing.T) {
	type args struct {
		ctx      context.Context
		myNumber string
	}

	tests := []struct {
		name                     string
		args                     args
		myNumberRepository       func() *mocks.MyNumberRepository
		lookUpMyNumberRepository func() *mocks.LookUpMyNumberRepository
		want                     bool
		wantErr                  bool
	}{
		{
			name: "(성공) APPROVE 상태인 동일한 마이 넘버의 존재 여부 확인 - 존재함",
			args: args{
				ctx:      context.Background(),
				myNumber: mockApproveMyNumber.MyNumber,
			},
			myNumberRepository: func() *mocks.MyNumberRepository {
				repo := &mocks.MyNumberRepository{}
				repo.On("GetById", context.Background(), mockApproveMyNumber.Id).Return(mockApproveMyNumber, nil)
				return repo
			},
			lookUpMyNumberRepository: func() *mocks.LookUpMyNumberRepository {
				repo := &mocks.LookUpMyNumberRepository{}
				repo.On("GetByMyNumber", context.Background(), mockApproveMyNumber.MyNumber).Return(domain.LookUpMyNumber{
					Id:         1000,
					MyNumberId: mockApproveMyNumber.Id,
				}, nil)
				return repo
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "(성공) APPROVE 상태인 동일한 마이 넘버의 존재 여부 확인 - DISAPPROVED 상태",
			args: args{
				ctx:      context.Background(),
				myNumber: mockDisapprovedMyNumber.MyNumber,
			},
			myNumberRepository: func() *mocks.MyNumberRepository {
				repo := &mocks.MyNumberRepository{}
				repo.On("GetById", context.Background(), mockDisapprovedMyNumber.Id).Return(mockDisapprovedMyNumber, nil)
				return repo
			},
			lookUpMyNumberRepository: func() *mocks.LookUpMyNumberRepository {
				repo := &mocks.LookUpMyNumberRepository{}
				repo.On("GetByMyNumber", context.Background(), mockDisapprovedMyNumber.MyNumber).Return(domain.LookUpMyNumber{
					Id:         1000,
					MyNumberId: mockDisapprovedMyNumber.Id,
				}, nil)
				return repo
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "(성공) APPROVE 상태인 동일한 마이 넘버의 존재 여부 확인 - WAIT_APPROVE 상태",
			args: args{
				ctx:      context.Background(),
				myNumber: mockWaitApproveMyNumber.MyNumber,
			},
			myNumberRepository: func() *mocks.MyNumberRepository {
				repo := &mocks.MyNumberRepository{}
				repo.On("GetById", context.Background(), mockWaitApproveMyNumber.Id).Return(mockWaitApproveMyNumber, nil)
				return repo
			},
			lookUpMyNumberRepository: func() *mocks.LookUpMyNumberRepository {
				repo := &mocks.LookUpMyNumberRepository{}
				repo.On("GetByMyNumber", context.Background(), mockWaitApproveMyNumber.MyNumber).Return(domain.LookUpMyNumber{
					Id:         1000,
					MyNumberId: mockWaitApproveMyNumber.Id,
				}, nil)
				return repo
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := myNumberService{
				myNumberRepository:       tt.myNumberRepository(),
				lookUpMyNumberRepository: tt.lookUpMyNumberRepository(),
			}

			got, err := service.CheckApproveMyNumberExists(tt.args.ctx, tt.args.myNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf("MyNumberService.CheckApproveMyNumberExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equalf(t, tt.want, got, "MyNumberService.CheckApproveMyNumberExists() = %v, want %v", got, tt.want)
		})
	}
}
