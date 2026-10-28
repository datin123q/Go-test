package my_number

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

func Test_myNumberUsecase_Create(t *testing.T) {
	// マイナンバー登録復活後にスキップ解除
	t.Skip()

	type args struct {
		ctx     context.Context
		request udomain.MyNumberCreateRequest
	}

	tests := []struct {
		name            string
		args            args
		logger          factory.LoggerFactory
		userRepository  func() cdomain.UserRepository
		myNumberService func() cdomain.MyNumberService
		want            udomain.MyNumberCreateResponse
		wantErr         bool
	}{
		{
			name: "(成功) マイナンバー登録",
			args: args{
				ctx: context.Background(),
				request: udomain.MyNumberCreateRequest{
					JWTUser: jwthelper.User{
						Id: mockMyNumber.UserId,
					},
					MyNumber:     mockMyNumber.MyNumber,
					FrontFileUrl: mockMyNumberFront.FileUrl,
					BackFileUrl:  mockMyNumberBack.FileUrl,
				},
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", context.Background(), mockMyNumber.UserId).
					Return(domain.User{
						Id:          mockMyNumber.UserId,
						AuditStatus: domain.CHECK_ANTISOCIAL_AUDIT,
					}, nil)
				repo.On("Update", context.Background(), domain.User{
					Id:          mockMyNumber.UserId,
					AuditStatus: domain.CHECK_ANTISOCIAL_AUDIT, // TODO: マイナンバー登録復活後にdomain.CHECK_MY_NUMBER_AUDITに戻す
				}).Return(nil)
				return repo
			},
			myNumberService: func() cdomain.MyNumberService {
				service := &mocks.MyNumberService{}
				service.On("RegisterWithFiles", context.Background(),
					cdomain.MyNumberRegisterInput{
						UserId:        mockMyNumber.UserId,
						MyNumberValue: mockMyNumber.MyNumber,
						FrontFileUrl:  mockMyNumberFront.FileUrl,
						BackFileUrl:   mockMyNumberBack.FileUrl,
						MyNumber:      domain.MyNumber{Status: domain.UNREGISTERED},
					}).Return(domain.MyNumber{
					Id:        mockMyNumber.Id,
					UserId:    mockMyNumber.UserId,
					MyNumber:  mockMyNumber.MyNumber,
					Status:    domain.WAIT_APPROVE,
					FrontFile: &mockMyNumberFront,
					BackFile:  &mockMyNumberBack,
				}, nil)
				service.On("GetByUserId", context.Background(), mockMyNumber.UserId).Return(domain.MyNumber{}, nil)
				return service
			},
			want: udomain.MyNumberCreateResponse{
				MyNumberBody: udomain.MyNumberBody{
					Id:       mockMyNumber.Id,
					MyNumber: mockMyNumber.MyNumber,
					Status:   mockMyNumber.Status,
					FrontFile: &udomain.UserFileBody{
						Id:            mockMyNumberFront.Id,
						OriginalName:  "",
						FileType:      mockMyNumberFront.FileType,
						FileSize:      0,
						FileExtension: "",
						FilePath:      "",
						FileUrl:       mockMyNumberFront.FileUrl,
						OwnerId:       mockMyNumberFront.OwnerId,
						UpdateId:      mockMyNumberFront.UpdateId,
					},
					BackFile: &udomain.UserFileBody{
						Id:            mockMyNumberBack.Id,
						OriginalName:  "",
						FileType:      mockMyNumberBack.FileType,
						FileSize:      0,
						FileExtension: "",
						FilePath:      "",
						FileUrl:       mockMyNumberBack.FileUrl,
						OwnerId:       mockMyNumberBack.OwnerId,
						UpdateId:      mockMyNumberBack.UpdateId,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "(失敗) マイナンバー登録 - マイナンバーがWAIT_APPROVE状態の場合",
			args: args{
				ctx: context.Background(),
				request: udomain.MyNumberCreateRequest{
					JWTUser: jwthelper.User{
						Id: mockMyNumber.UserId,
					},
					MyNumber:     mockMyNumber.MyNumber,
					FrontFileUrl: mockMyNumberFront.FileUrl,
					BackFileUrl:  mockMyNumberBack.FileUrl,
				},
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				return repo
			},
			myNumberService: func() cdomain.MyNumberService {
				service := &mocks.MyNumberService{}
				service.On("GetByUserId", context.Background(), mockMyNumber.UserId).Return(mockMyNumber, nil)
				return service
			},
			want:    udomain.MyNumberCreateResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := myNumberUsecase{
				userRepository:  tt.userRepository(),
				myNumberService: tt.myNumberService(),
			}
			got, err := usecase.Create(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
			ignoreFields := cmpopts.IgnoreFields(udomain.MyNumberBody{}, "Id", "FrontFile", "BackFile")
			ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			assert.Equalf(t, cmp.Equal(tt.want.MyNumberBody, got.MyNumberBody, ignoreTypes, ignoreFields), true, "Create() = %+v, want %+v", got, tt.want)
			if tt.want.MyNumberBody.FrontFile != nil {
				assert.Equalf(t, cmp.Equal(*tt.want.MyNumberBody.FrontFile, *got.MyNumberBody.FrontFile, ignoreTypes), true, "Create() = %+v, want %+v", *got.MyNumberBody.FrontFile, *tt.want.MyNumberBody.FrontFile)
			}
			if tt.want.MyNumberBody.BackFile != nil {
				assert.Equalf(t, cmp.Equal(*tt.want.MyNumberBody.BackFile, *got.MyNumberBody.BackFile, ignoreTypes), true, "Create() = %+v, want %+v", *got.MyNumberBody.BackFile, *tt.want.MyNumberBody.BackFile)
			}
		})
	}
}

func Test_myNumberUsecase_Get(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.MyNumberGetRequest
	}

	tests := []struct {
		name            string
		args            args
		myNumberService func() cdomain.MyNumberService
		want            udomain.MyNumberGetResponse
		wantErr         bool
	}{
		{
			name: "(成功) マイナンバー照会",
			args: args{
				ctx: context.Background(),
				request: udomain.MyNumberGetRequest{
					JWTUser: jwthelper.User{
						Id: mockMyNumber2.UserId,
					},
				},
			},
			myNumberService: func() cdomain.MyNumberService {
				service := &mocks.MyNumberService{}
				service.On("GetByUserId", context.Background(), mockMyNumber2.UserId).Return(mockMyNumber2, nil)
				return service
			},
			want: udomain.MyNumberGetResponse{
				MyNumberBody: &udomain.MyNumberBody{
					Id:       mockMyNumber2.Id,
					MyNumber: mockMyNumber2.MyNumber,
					Status:   mockMyNumber2.Status,
					FrontFile: &udomain.UserFileBody{
						OriginalName:  "",
						FileType:      mockMyNumber2.FrontFile.FileType,
						FileSize:      0,
						FileExtension: "",
						FilePath:      "",
						FileUrl:       mockMyNumber2.FrontFile.FileUrl,
						OwnerId:       mockMyNumber2.FrontFile.OwnerId,
						UpdateId:      mockMyNumber2.FrontFile.UpdateId,
					},
					BackFile: &udomain.UserFileBody{
						OriginalName:  "",
						FileType:      mockMyNumber2.BackFile.FileType,
						FileSize:      0,
						FileExtension: "",
						FilePath:      "",
						FileUrl:       mockMyNumber2.BackFile.FileUrl,
						OwnerId:       mockMyNumber2.BackFile.OwnerId,
						UpdateId:      mockMyNumber2.BackFile.UpdateId,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := myNumberUsecase{
				myNumberService: tt.myNumberService(),
			}
			got, err := usecase.Get(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.wantErr, err != nil, "Get() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want.MyNumberBody.Status, got.MyNumberBody.Status, "Get().Status error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want.MyNumberBody.MyNumber, got.MyNumberBody.MyNumber, "Get().MyNumber error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
