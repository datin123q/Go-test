package withdrawal_close

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_withdrawalClose_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.WithdrawalCloseGetAllRequest
	}

	tests := []struct {
		name                      string
		args                      args
		withdrawalCloseRepository func() *mocks.WithdrawalCloseRepository
		want                      adomain.WithdrawalCloseGetAllResponse
		wantErr                   bool
	}{
		{
			name: "(成功) 締め切られた出金依頼照会",
			args: args{
				ctx: context.Background(),
				request: adomain.WithdrawalCloseGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    mockAdmin.Id,
						Email: mockAdmin.Email,
						Name:  mockAdmin.Name,
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PAYMENT": 1},
					},
					SearchInput: cdomain.SearchInput{
						SkipCount:      0,
						MaxResultCount: 10,
						Order:          "asc",
						Sortby:         "id",
					},
				},
			},
			withdrawalCloseRepository: func() *mocks.WithdrawalCloseRepository {
				repo := &mocks.WithdrawalCloseRepository{}
				repo.On("GetAll", mock.Anything,
					mock.MatchedBy(func(field func(*cdomain.WithdrawalCloseQueryOptionFields)) bool {
						opt := cdomain.WithdrawalCloseQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}), // SortBy
					mock.MatchedBy(func(field func(*cdomain.WithdrawalCloseQueryOptionFields)) bool {
						opt := cdomain.WithdrawalCloseQueryOptionFunc()
						field(&opt)
						return opt.Order() == "asc"
					}), // Order
					mock.MatchedBy(func(field func(*cdomain.WithdrawalCloseQueryOptionFields)) bool {
						opt := cdomain.WithdrawalCloseQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}), // SkipCount
					mock.MatchedBy(func(field func(*cdomain.WithdrawalCloseQueryOptionFields)) bool {
						opt := cdomain.WithdrawalCloseQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 10
					}), // MaxResultCount
				).Return(int64(1), []domain.WithdrawalCloseDetail{
					mockWithdrawalCloseDetail1,
				}, nil)
				return repo
			},
			want: adomain.WithdrawalCloseGetAllResponse{
				TotalCount: int64(1),
				Items: []adomain.WithdrawalCloseGetAllItem{
					{
						Id:           mockWithdrawalCloseDetail1.Id,
						Total:        mockWithdrawalCloseDetail1.Total,
						Amount:       mockWithdrawalCloseDetail1.Amount,
						Fee:          mockWithdrawalCloseDetail1.Fee,
						CloserId:     mockWithdrawalCloseDetail1.CloserId,
						CloserName:   mockWithdrawalCloseDetail1.CloserName,
						CloseDate:    mockWithdrawalCloseDetail1.CloseDate,
						TransferDate: mockWithdrawalCloseDetail1.TransferDate,
						Status:       mockWithdrawalCloseDetail1.Status,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &withdrawalCloseUsecase{
				withdrawalCloseRepository: tt.withdrawalCloseRepository(),
			}

			got, err := u.GetAll(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetAll() got = %v, want %v", got, tt.want)
		})
	}
}

func Test_withdrawalCloseUsecase_GetWithdrawals(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.WithdrawalCloseGetWithdrawalsRequest
	}

	tests := []struct {
		name                 string
		args                 args
		withdrawalRepository func() *mocks.WithdrawalRepository
		want                 adomain.WithdrawalCloseGetWithdrawalsResponse
		wantErr              bool
	}{
		{
			name: "(成功) 締め切られた出金依頼リスト照会",
			args: args{
				ctx: context.Background(),
				request: adomain.WithdrawalCloseGetWithdrawalsRequest{
					JWTUser: jwthelper.User{
						Id:    mockAdmin.Id,
						Email: mockAdmin.Email,
						Name:  mockAdmin.Name,
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PAYMENT": 1},
					},
					Id: int64(1),
					SearchInput: cdomain.SearchInput{
						SkipCount:      0,
						MaxResultCount: 10,
						Order:          "asc",
						Sortby:         "id",
					},
				},
			},
			withdrawalRepository: func() *mocks.WithdrawalRepository {
				repo := &mocks.WithdrawalRepository{}
				repo.On("GetAll", mock.Anything,
					mock.MatchedBy(func(field func(option *cdomain.WithdrawalQueryOptionFields)) bool {
						opt := cdomain.WithdrawalQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}), // SortBy
					mock.MatchedBy(func(field func(*cdomain.WithdrawalQueryOptionFields)) bool {
						opt := cdomain.WithdrawalQueryOptionFunc()
						field(&opt)
						return opt.Order() == "asc"
					}), // Order
					mock.MatchedBy(func(field func(*cdomain.WithdrawalQueryOptionFields)) bool {
						opt := cdomain.WithdrawalQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}), // SkipCount
					mock.MatchedBy(func(field func(*cdomain.WithdrawalQueryOptionFields)) bool {
						opt := cdomain.WithdrawalQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 10
					}), // MaxResultCount
					mock.MatchedBy(func(field func(*cdomain.WithdrawalQueryOptionFields)) bool {
						opt := cdomain.WithdrawalQueryOptionFunc()
						field(&opt)
						return *opt.CloseId() == int64(1)
					}), // CloseId
				).Return(int64(2), []domain.WithdrawalDetail{
					mockWithdrawalDetail1,
					mockWithdrawalDetail2,
				}, nil)
				return repo
			},
			want: adomain.WithdrawalCloseGetWithdrawalsResponse{
				TotalCount: int64(2),
				Items: []adomain.WithdrawalCloseGetWithdrawalsItem{
					{
						Id:           mockWithdrawalDetail1.Withdrawal.Id,
						InvestorId:   mockWithdrawalDetail1.Withdrawal.InvestorId,
						InvestorName: mockWithdrawalDetail1.Withdrawal.InvestorName,
						Amount:       mockWithdrawalDetail1.Withdrawal.Amount,
						Fee:          mockWithdrawalDetail1.Withdrawal.Fee,
						RequestDate:  mockWithdrawalDetail1.Withdrawal.RequestDate,
						CompleteDate: mockWithdrawalDetail1.Withdrawal.CompleteDate,
						Status:       mockWithdrawalDetail1.Withdrawal.Status,
						Created:      mockWithdrawalDetail1.Withdrawal.Created,
						CloseDate:    mockWithdrawalDetail1.Withdrawal.CloseDate,
					},
					{
						Id:           mockWithdrawalDetail2.Withdrawal.Id,
						InvestorId:   mockWithdrawalDetail2.Withdrawal.InvestorId,
						InvestorName: mockWithdrawalDetail2.Withdrawal.InvestorName,
						Amount:       mockWithdrawalDetail2.Withdrawal.Amount,
						Fee:          mockWithdrawalDetail2.Withdrawal.Fee,
						RequestDate:  mockWithdrawalDetail2.Withdrawal.RequestDate,
						CompleteDate: mockWithdrawalDetail2.Withdrawal.CompleteDate,
						Status:       mockWithdrawalDetail2.Withdrawal.Status,
						Created:      mockWithdrawalDetail2.Withdrawal.Created,
						CloseDate:    mockWithdrawalDetail2.Withdrawal.CloseDate,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &withdrawalCloseUsecase{
				withdrawalRepository: tt.withdrawalRepository(),
			}

			_, err := u.GetWithdrawals(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "GetWithdrawals() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, tt.want, "GetWithdrawals() got = %v, want %v", tt.want, tt.want)
		})
	}
}
func Test_withdrawalCloseUsecase_Download(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.WithdrawalCloseDownloadRequest
	}

	tests := []struct {
		name                      string
		args                      args
		withdrawalRepository      func() *mocks.WithdrawalRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		withdrawalService         func() *mocks.WithdrawalService
		want                      adomain.WithdrawalCloseDownloadResponse
		wantErr                   bool
	}{
		{
			name: "(成功) 締め切られた出金依頼リストダウンロード",
			args: args{
				ctx: context.Background(),
				request: adomain.WithdrawalCloseDownloadRequest{
					Id: int64(1),
					JWTUser: jwthelper.User{
						Id:    mockAdmin.Id,
						Email: mockAdmin.Email,
						Name:  mockAdmin.Name,
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PAYMENT": 1},
					},
					ClientIp: "clientIp",
					SearchInput: cdomain.SearchInput{
						SkipCount:      0,
						MaxResultCount: 10,
						Order:          "asc",
						Sortby:         "id",
					},
				},
			},
			withdrawalRepository: func() *mocks.WithdrawalRepository {
				repo := &mocks.WithdrawalRepository{}
				repo.On("GetAll", mock.Anything,
					mock.MatchedBy(func(field func(option *cdomain.WithdrawalQueryOptionFields)) bool {
						opt := cdomain.WithdrawalQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}), // SortBy
					mock.MatchedBy(func(field func(*cdomain.WithdrawalQueryOptionFields)) bool {
						opt := cdomain.WithdrawalQueryOptionFunc()
						field(&opt)
						return opt.Order() == "asc"
					}), // Order
					mock.MatchedBy(func(field func(*cdomain.WithdrawalQueryOptionFields)) bool {
						opt := cdomain.WithdrawalQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}), // SkipCount
					mock.MatchedBy(func(field func(*cdomain.WithdrawalQueryOptionFields)) bool {
						opt := cdomain.WithdrawalQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 10
					}), // MaxResultCount
					mock.MatchedBy(func(field func(*cdomain.WithdrawalQueryOptionFields)) bool {
						opt := cdomain.WithdrawalQueryOptionFunc()
						field(&opt)
						return *opt.CloseId() == int64(1)
					}), // CloseId
				).Return(int64(2), []domain.WithdrawalDetail{
					mockWithdrawalDetail1,
					mockWithdrawalDetail2,
				}, nil)
				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := &mocks.OperatorHistoryRepository{}
				repo.On("Create", mock.Anything, mock.MatchedBy(func(op domain.OperatorHistory) bool {
					ignoreTypes := cmpopts.IgnoreTypes(map[string]int{}, jsonutils.JSONTime{})
					return cmp.Equal(op, domain.OperatorHistory{
						UserId: mockAdmin.Id,
						Name:   mockAdmin.Name,
						Email:  mockAdmin.Email,
						Menu:   MENU_WITHDRAWAL_DOWNLOAD,
						Ip:     "clientIp",
						State:  domain.DOWNLOAD,
					}, ignoreTypes)
				})).Return(nil)
				return repo
			},
			withdrawalService: func() *mocks.WithdrawalService {
				service := &mocks.WithdrawalService{}
				service.On("ExportCSV", mock.Anything, []domain.WithdrawalDetail{
					mockWithdrawalDetail1,
					mockWithdrawalDetail2,
				}).Return(cdomain.CSVData{}, nil)
				return service
			},
			want:    adomain.WithdrawalCloseDownloadResponse{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &withdrawalCloseUsecase{
				withdrawalRepository:      tt.withdrawalRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
				withdrawalService:         tt.withdrawalService(),
			}

			_, err := u.Download(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "Download() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
