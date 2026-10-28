package withdrawal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_withdrawal_Create(t *testing.T) {
	type fields struct {
		cfg                       *config.Config
		txManager                 func(*testing.T) *mocks.TxManager
		mailService               func(*testing.T) *mocks.MailService
		userRepository            func(*testing.T) *mocks.UserRepository
		userAccountRepository     func(*testing.T) *mocks.UserAccountRepository
		operatorHistoryRepository func(*testing.T) *mocks.OperatorHistoryRepository
		withdrawalRepository      func(*testing.T) *mocks.WithdrawalRepository
		withdrawBalance           func(context.Context, *domain.UserAccount, int64, domain.USER_TRANSACTION_DETAIL_TYPE) (int64, error)
		requestFee                func(context.Context, domain.User, int64) (int64, error)
		now                       func() time.Time
	}
	type args struct {
		ctx     context.Context
		request adomain.WithdrawalCreateRequest
	}

	fixedTime, _ := time.ParseInLocation(time.DateTime, "2026-07-13 12:34:56", time.Local)
	feeRequestErr := errors.New("fee request failed")
	mailSendErr := errors.New("mail send failed")
	baseConfig := &config.Config{}
	baseConfig.CommonConfig.ServiceConfig.Mail.Info = "info@energyshares.jp"
	baseConfig.CommonConfig.ServiceConfig.Url.UserServiceDomain = "https://user.energyshares.jp"

	validUser := domain.User{
		Id:          1001,
		Name:        "山田太郎",
		Email:       "taro@example.com",
		UserType:    domain.PRIVATE,
		AuditStatus: domain.COMPLETED_AUDIT,
	}
	validAccount := domain.UserAccount{
		Id:      501,
		UserId:  validUser.Id,
		Balance: lo.ToPtr(int64(50000)),
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    adomain.WithdrawalCreateResponse
		wantErr error
	}{
		{
			name: "（成功）管理画面から出金依頼を作成できる",
			fields: fields{
				cfg: baseConfig,
				txManager: func(t *testing.T) *mocks.TxManager {
					txManager := mocks.NewTxManager(t)
					txManager.On("WithTx", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
					return txManager
				},
				mailService: func(t *testing.T) *mocks.MailService {
					mailService := mocks.NewMailService(t)
					mailService.On("SendMail", mock.Anything, mock.MatchedBy(func(request cdomain.SendMailRequest) bool {
						return request.TemplateType == domain.WITHDRAWAL_REQUEST_TEMPLATE &&
							request.MailType == domain.MAIL_SEND_PRIORITY_NORMAL &&
							request.Sender == "info@energyshares.jp" &&
							cmp.Equal(request.Receivers, []string{"taro@example.com"}) &&
							request.Details["Receiver"] == "山田太郎" &&
							request.Details["ServiceUrl"] == "https://user.energyshares.jp" &&
							request.Details["RequestDate"] == "2026-07-13" &&
							request.Details["Amount"] == "12000"
					})).Return(nil)
					return mailService
				},
				userRepository: func(t *testing.T) *mocks.UserRepository {
					repo := mocks.NewUserRepository(t)
					repo.On("GetById", mock.Anything, validUser.Id).Return(validUser, nil)
					return repo
				},
				userAccountRepository: func(t *testing.T) *mocks.UserAccountRepository {
					repo := mocks.NewUserAccountRepository(t)
					repo.On("GetByUserId", mock.Anything, validUser.Id).Return(validAccount, nil)
					return repo
				},
				operatorHistoryRepository: func(t *testing.T) *mocks.OperatorHistoryRepository {
					repo := mocks.NewOperatorHistoryRepository(t)
					repo.On("Create", mock.Anything, mock.MatchedBy(func(history domain.OperatorHistory) bool {
						return history.UserId == mockAdmin.Id &&
							history.Name == mockAdmin.Name &&
							history.Email == mockAdmin.Email &&
							history.Menu == "menu.manage_accounting.withdrawal_request" &&
							history.Ip == "127.0.0.1" &&
							history.State == domain.REGISTER
					})).Return(nil)
					return repo
				},
				withdrawalRepository: func(t *testing.T) *mocks.WithdrawalRepository {
					repo := mocks.NewWithdrawalRepository(t)
					repo.On("Create", mock.Anything, mock.MatchedBy(func(withdrawal domain.Withdrawal) bool {
						return withdrawal.InvestorId == "1001" &&
							withdrawal.InvestorName == "山田太郎" &&
							withdrawal.Amount == 12000 &&
							withdrawal.Fee == 330 &&
							withdrawal.RequestDate.Time().Equal(fixedTime) &&
							withdrawal.Status == domain.WITHDRAWAL_STATUS_INCOMPLETE &&
							withdrawal.IsManual
					})).Return(int64(9001), nil)
					return repo
				},
				withdrawBalance: func(_ context.Context, account *domain.UserAccount, amount int64, transactionType domain.USER_TRANSACTION_DETAIL_TYPE) (int64, error) {
					if transactionType != domain.WITHDRAWAL_INVESTOR {
						return 0, errors.New("unexpected transaction type")
					}
					account.Balance = lo.ToPtr(lo.FromPtr(account.Balance) - amount)
					return 1, nil
				},
				requestFee: func(context.Context, domain.User, int64) (int64, error) {
					return 330, nil
				},
				now: func() time.Time {
					return fixedTime
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.WithdrawalCreateRequest{
					JWTUser: jwthelper.User{
						Id:    mockAdmin.Id,
						Email: mockAdmin.Email,
						Name:  mockAdmin.Name,
					},
					ClientIp: "127.0.0.1",
					UserId:   validUser.Id,
					Amount:   12000,
				},
			},
			want: adomain.WithdrawalCreateResponse{
				Id:           9001,
				InvestorId:   "1001",
				InvestorName: "山田太郎",
				Amount:       12000,
				Fee:          330,
				RequestDate:  jsonutils.JSONTime(fixedTime),
				Status:       domain.WITHDRAWAL_STATUS_INCOMPLETE,
				Created:      jsonutils.JSONTime(fixedTime),
				Updated:      jsonutils.JSONTime(fixedTime),
			},
		},
		{
			name: "（成功）メール送信に失敗しても出金依頼作成は成功扱いにする",
			fields: fields{
				cfg: baseConfig,
				txManager: func(t *testing.T) *mocks.TxManager {
					txManager := mocks.NewTxManager(t)
					txManager.On("WithTx", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
					return txManager
				},
				mailService: func(t *testing.T) *mocks.MailService {
					mailService := mocks.NewMailService(t)
					mailService.On("SendMail", mock.Anything, mock.MatchedBy(func(request cdomain.SendMailRequest) bool {
						return request.TemplateType == domain.WITHDRAWAL_REQUEST_TEMPLATE &&
							request.MailType == domain.MAIL_SEND_PRIORITY_NORMAL &&
							request.Sender == "info@energyshares.jp" &&
							cmp.Equal(request.Receivers, []string{"taro@example.com"})
					})).Return(mailSendErr)
					return mailService
				},
				userRepository: func(t *testing.T) *mocks.UserRepository {
					repo := mocks.NewUserRepository(t)
					repo.On("GetById", mock.Anything, validUser.Id).Return(validUser, nil)
					return repo
				},
				userAccountRepository: func(t *testing.T) *mocks.UserAccountRepository {
					repo := mocks.NewUserAccountRepository(t)
					repo.On("GetByUserId", mock.Anything, validUser.Id).Return(validAccount, nil)
					return repo
				},
				operatorHistoryRepository: func(t *testing.T) *mocks.OperatorHistoryRepository {
					repo := mocks.NewOperatorHistoryRepository(t)
					repo.On("Create", mock.Anything, mock.MatchedBy(func(history domain.OperatorHistory) bool {
						return history.UserId == mockAdmin.Id &&
							history.Name == mockAdmin.Name &&
							history.Email == mockAdmin.Email &&
							history.Menu == "menu.manage_accounting.withdrawal_request" &&
							history.Ip == "127.0.0.1" &&
							history.State == domain.REGISTER
					})).Return(nil)
					return repo
				},
				withdrawalRepository: func(t *testing.T) *mocks.WithdrawalRepository {
					repo := mocks.NewWithdrawalRepository(t)
					repo.On("Create", mock.Anything, mock.MatchedBy(func(withdrawal domain.Withdrawal) bool {
						return withdrawal.InvestorId == "1001" &&
							withdrawal.InvestorName == "山田太郎" &&
							withdrawal.Amount == 12000 &&
							withdrawal.Fee == 330 &&
							withdrawal.RequestDate.Time().Equal(fixedTime) &&
							withdrawal.Status == domain.WITHDRAWAL_STATUS_INCOMPLETE &&
							withdrawal.IsManual
					})).Return(int64(9002), nil)
					return repo
				},
				withdrawBalance: func(_ context.Context, account *domain.UserAccount, amount int64, transactionType domain.USER_TRANSACTION_DETAIL_TYPE) (int64, error) {
					if transactionType != domain.WITHDRAWAL_INVESTOR {
						return 0, errors.New("unexpected transaction type")
					}
					account.Balance = lo.ToPtr(lo.FromPtr(account.Balance) - amount)
					return 1, nil
				},
				requestFee: func(context.Context, domain.User, int64) (int64, error) {
					return 330, nil
				},
				now: func() time.Time {
					return fixedTime
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.WithdrawalCreateRequest{
					JWTUser: jwthelper.User{
						Id:    mockAdmin.Id,
						Email: mockAdmin.Email,
						Name:  mockAdmin.Name,
					},
					ClientIp: "127.0.0.1",
					UserId:   validUser.Id,
					Amount:   12000,
				},
			},
			want: adomain.WithdrawalCreateResponse{
				Id:           9002,
				InvestorId:   "1001",
				InvestorName: "山田太郎",
				Amount:       12000,
				Fee:          330,
				RequestDate:  jsonutils.JSONTime(fixedTime),
				Status:       domain.WITHDRAWAL_STATUS_INCOMPLETE,
				Created:      jsonutils.JSONTime(fixedTime),
				Updated:      jsonutils.JSONTime(fixedTime),
			},
		},
		{
			name: "（失敗）対象ユーザーが存在しない",
			fields: fields{
				cfg: baseConfig,
				txManager: func(t *testing.T) *mocks.TxManager {
					return mocks.NewTxManager(t)
				},
				mailService: func(t *testing.T) *mocks.MailService {
					return mocks.NewMailService(t)
				},
				userRepository: func(t *testing.T) *mocks.UserRepository {
					repo := mocks.NewUserRepository(t)
					repo.On("GetById", mock.Anything, int64(9999)).Return(domain.User{}, cdomain.ErrItemNotFound)
					return repo
				},
				userAccountRepository: func(t *testing.T) *mocks.UserAccountRepository {
					return mocks.NewUserAccountRepository(t)
				},
				operatorHistoryRepository: func(t *testing.T) *mocks.OperatorHistoryRepository {
					return mocks.NewOperatorHistoryRepository(t)
				},
				withdrawalRepository: func(t *testing.T) *mocks.WithdrawalRepository {
					return mocks.NewWithdrawalRepository(t)
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.WithdrawalCreateRequest{
					UserId: 9999,
					Amount: 1000,
				},
			},
			wantErr: cdomain.ErrItemNotFound,
		},
		{
			name: "（失敗）本人確認完了前のユーザーは出金依頼を作成できない",
			fields: fields{
				cfg: baseConfig,
				txManager: func(t *testing.T) *mocks.TxManager {
					return mocks.NewTxManager(t)
				},
				mailService: func(t *testing.T) *mocks.MailService {
					return mocks.NewMailService(t)
				},
				userRepository: func(t *testing.T) *mocks.UserRepository {
					repo := mocks.NewUserRepository(t)
					invalidUser := validUser
					invalidUser.AuditStatus = domain.WAIT_AUDIT
					repo.On("GetById", mock.Anything, validUser.Id).Return(invalidUser, nil)
					return repo
				},
				userAccountRepository: func(t *testing.T) *mocks.UserAccountRepository {
					return mocks.NewUserAccountRepository(t)
				},
				operatorHistoryRepository: func(t *testing.T) *mocks.OperatorHistoryRepository {
					return mocks.NewOperatorHistoryRepository(t)
				},
				withdrawalRepository: func(t *testing.T) *mocks.WithdrawalRepository {
					return mocks.NewWithdrawalRepository(t)
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.WithdrawalCreateRequest{
					UserId: validUser.Id,
					Amount: 1000,
				},
			},
			wantErr: cdomain.ErrBadRequest,
		},
		{
			name: "（失敗）ADMINユーザーは出金依頼を作成できない",
			fields: fields{
				cfg: baseConfig,
				txManager: func(t *testing.T) *mocks.TxManager {
					return mocks.NewTxManager(t)
				},
				mailService: func(t *testing.T) *mocks.MailService {
					return mocks.NewMailService(t)
				},
				userRepository: func(t *testing.T) *mocks.UserRepository {
					repo := mocks.NewUserRepository(t)
					adminUser := validUser
					adminUser.UserType = domain.ADMIN
					repo.On("GetById", mock.Anything, validUser.Id).Return(adminUser, nil)
					return repo
				},
				userAccountRepository: func(t *testing.T) *mocks.UserAccountRepository {
					return mocks.NewUserAccountRepository(t)
				},
				operatorHistoryRepository: func(t *testing.T) *mocks.OperatorHistoryRepository {
					return mocks.NewOperatorHistoryRepository(t)
				},
				withdrawalRepository: func(t *testing.T) *mocks.WithdrawalRepository {
					return mocks.NewWithdrawalRepository(t)
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.WithdrawalCreateRequest{
					UserId: validUser.Id,
					Amount: 1000,
				},
			},
			wantErr: cdomain.ErrBadRequest,
		},
		{
			name: "（失敗）amountが0以下の場合は出金依頼を作成できない",
			fields: fields{
				cfg: baseConfig,
				txManager: func(t *testing.T) *mocks.TxManager {
					return mocks.NewTxManager(t)
				},
				mailService: func(t *testing.T) *mocks.MailService {
					return mocks.NewMailService(t)
				},
				userRepository: func(t *testing.T) *mocks.UserRepository {
					return mocks.NewUserRepository(t)
				},
				userAccountRepository: func(t *testing.T) *mocks.UserAccountRepository {
					return mocks.NewUserAccountRepository(t)
				},
				operatorHistoryRepository: func(t *testing.T) *mocks.OperatorHistoryRepository {
					return mocks.NewOperatorHistoryRepository(t)
				},
				withdrawalRepository: func(t *testing.T) *mocks.WithdrawalRepository {
					return mocks.NewWithdrawalRepository(t)
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.WithdrawalCreateRequest{
					UserId: validUser.Id,
					Amount: 0,
				},
			},
			wantErr: cdomain.ErrBadRequest,
		},
		{
			name: "（失敗）残高不足の場合は出金依頼を作成できない",
			fields: fields{
				cfg: baseConfig,
				txManager: func(t *testing.T) *mocks.TxManager {
					return mocks.NewTxManager(t)
				},
				mailService: func(t *testing.T) *mocks.MailService {
					return mocks.NewMailService(t)
				},
				userRepository: func(t *testing.T) *mocks.UserRepository {
					repo := mocks.NewUserRepository(t)
					repo.On("GetById", mock.Anything, validUser.Id).Return(validUser, nil)
					return repo
				},
				userAccountRepository: func(t *testing.T) *mocks.UserAccountRepository {
					repo := mocks.NewUserAccountRepository(t)
					lowBalanceAccount := validAccount
					lowBalanceAccount.Balance = lo.ToPtr(int64(100))
					repo.On("GetByUserId", mock.Anything, validUser.Id).Return(lowBalanceAccount, nil)
					return repo
				},
				operatorHistoryRepository: func(t *testing.T) *mocks.OperatorHistoryRepository {
					return mocks.NewOperatorHistoryRepository(t)
				},
				withdrawalRepository: func(t *testing.T) *mocks.WithdrawalRepository {
					return mocks.NewWithdrawalRepository(t)
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.WithdrawalCreateRequest{
					UserId: validUser.Id,
					Amount: 1000,
				},
			},
			wantErr: cdomain.ErrBadRequest,
		},
		{
			name: "（失敗）手数料取得に失敗した場合は出金依頼を作成できない",
			fields: fields{
				cfg: baseConfig,
				txManager: func(t *testing.T) *mocks.TxManager {
					return mocks.NewTxManager(t)
				},
				mailService: func(t *testing.T) *mocks.MailService {
					return mocks.NewMailService(t)
				},
				userRepository: func(t *testing.T) *mocks.UserRepository {
					repo := mocks.NewUserRepository(t)
					repo.On("GetById", mock.Anything, validUser.Id).Return(validUser, nil)
					return repo
				},
				userAccountRepository: func(t *testing.T) *mocks.UserAccountRepository {
					repo := mocks.NewUserAccountRepository(t)
					repo.On("GetByUserId", mock.Anything, validUser.Id).Return(validAccount, nil)
					return repo
				},
				operatorHistoryRepository: func(t *testing.T) *mocks.OperatorHistoryRepository {
					return mocks.NewOperatorHistoryRepository(t)
				},
				withdrawalRepository: func(t *testing.T) *mocks.WithdrawalRepository {
					return mocks.NewWithdrawalRepository(t)
				},
				requestFee: func(context.Context, domain.User, int64) (int64, error) {
					return 0, feeRequestErr
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.WithdrawalCreateRequest{
					UserId: validUser.Id,
					Amount: 1000,
				},
			},
			wantErr: feeRequestErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &withdrawalUsecase{
				cfg:                       tt.fields.cfg,
				txManager:                 tt.fields.txManager(t),
				mailService:               tt.fields.mailService(t),
				userRepository:            tt.fields.userRepository(t),
				userAccountRepository:     tt.fields.userAccountRepository(t),
				operatorHistoryRepository: tt.fields.operatorHistoryRepository(t),
				withdrawalRepository:      tt.fields.withdrawalRepository(t),
				withdrawBalance:           tt.fields.withdrawBalance,
				requestFee:                tt.fields.requestFee,
				now:                       tt.fields.now,
			}

			got, err := u.Create(tt.args.ctx, tt.args.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_withdrawal_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.WithdrawalGetAllRequest
	}

	tests := []struct {
		name                 string
		args                 args
		withdrawalRepository func() *mocks.WithdrawalRepository
		transferRepository   func() *mocks.TransferRepository
		want                 adomain.WithdrawalGetAllResponse
		wantErr              bool
	}{
		{
			name: "(成功) 金額1000以上、RequestStartDate 2024-02-01 00:00:00以降の出金内訳照会",
			args: args{
				ctx: context.Background(),
				request: adomain.WithdrawalGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    mockAdmin.Id,
						Email: mockAdmin.Email,
						Name:  mockAdmin.Name,
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PAYMENT": 1},
					},
					WithdrawalGetAllQueryParam: adomain.WithdrawalGetAllQueryParam{
						SearchInput: cdomain.SearchInput{
							SkipCount:      0,
							MaxResultCount: 10,
							Order:          "asc",
							Sortby:         "id",
						},
						AmountMax:        lo.ToPtr(int64(1000)),
						RequestStartDate: lo.ToPtr("2024-02-01 00:00:00"),
					},
				},
			},
			withdrawalRepository: func() *mocks.WithdrawalRepository {
				repo := &mocks.WithdrawalRepository{}
				repo.On("GetAll", mock.Anything,
					mock.MatchedBy(func(field func(*cdomain.WithdrawalQueryOptionFields)) bool {
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
					mock.Anything, // InvestorId
					mock.Anything, // InvestorName
					mock.Anything, // AmountMin
					mock.MatchedBy(func(field func(*cdomain.WithdrawalQueryOptionFields)) bool {
						opt := cdomain.WithdrawalQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.AmountMax()) == int64(1000)
					}), // AmountMax
					mock.MatchedBy(func(field func(*cdomain.WithdrawalQueryOptionFields)) bool {
						opt := cdomain.WithdrawalQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.RequestStartDate()) == "2024-02-01 00:00:00"
					}), // RequestStartDate
					mock.Anything, // RequestEndDate
					mock.Anything, // CompleteStartDate
					mock.Anything, // CompleteEndDate
					mock.Anything, // Status
				).Return(int64(2), []domain.WithdrawalDetail{
					mockWithdrawalDetail2,
					mockWithdrawalDetail3,
				}, nil)
				repo.On("GetAll", mock.Anything,
					mock.MatchedBy(func(field func(*cdomain.WithdrawalQueryOptionFields)) bool {
						opt := cdomain.WithdrawalQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}), // SkipCount
					mock.MatchedBy(func(field func(*cdomain.WithdrawalQueryOptionFields)) bool {
						opt := cdomain.WithdrawalQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 99999
					}), // MaxResultCount
					mock.MatchedBy(func(field func(*cdomain.WithdrawalQueryOptionFields)) bool {
						opt := cdomain.WithdrawalQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.Status()) == domain.WITHDRAWAL_STATUS_INCOMPLETE
					}), // Status
				).Return(int64(1),
					[]domain.WithdrawalDetail{
						mockWithdrawalDetail1,
					},
					nil,
				)
				return repo
			},
			transferRepository: func() *mocks.TransferRepository {
				repo := &mocks.TransferRepository{}
				repo.On("GetAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(0), []domain.Transfer{}, nil)
				return repo
			},
			want: adomain.WithdrawalGetAllResponse{
				TotalCount:           int64(2),
				TotalIncompleteCount: int64(1),
				Items: []adomain.WithdrawalItem{
					mockWithdrawalItem2,
					mockWithdrawalItem3,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &withdrawalUsecase{
				withdrawalRepository: tt.withdrawalRepository(),
				transferRepository:   tt.transferRepository(),
			}

			got, err := u.GetAll(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetAll() got = %v, want %v", got, tt.want)
		})
	}
}

func Test_withdrawal_GetOne(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.WithdrawalGetOneRequest
	}

	tests := []struct {
		name                 string
		args                 args
		withdrawalRepository func() *mocks.WithdrawalRepository
		transferRepository   func() *mocks.TransferRepository
		want                 adomain.WithdrawalGetOneResponse
		wantErr              bool
	}{
		{
			name: "(成功) 出金内訳詳細照会",
			args: args{
				ctx: context.Background(),
				request: adomain.WithdrawalGetOneRequest{
					JWTUser: jwthelper.User{
						Id:    mockAdmin.Id,
						Email: mockAdmin.Email,
						Name:  mockAdmin.Name,
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PAYMENT": 1},
					},
					Id: mockWithdrawalDetail1.Withdrawal.Id,
				},
			},
			withdrawalRepository: func() *mocks.WithdrawalRepository {
				repo := &mocks.WithdrawalRepository{}
				repo.On("GetById", mock.Anything, mockWithdrawalDetail1.Withdrawal.Id).Return(mockWithdrawalDetail1, nil)
				return repo
			},
			transferRepository: func() *mocks.TransferRepository {
				repo := &mocks.TransferRepository{}
				repo.On("GetAll", mock.Anything,
					mock.MatchedBy(func(field func(fields *cdomain.TransferQueryOptionFields)) bool {
						opt := cdomain.TransferQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.WithdrawalId()) == mockWithdrawalDetail1.Withdrawal.Id
					}), // WithdrawalId
					mock.MatchedBy(func(field func(fields *cdomain.TransferQueryOptionFields)) bool {
						opt := cdomain.TransferQueryOptionFunc()
						field(&opt)
						return opt.Order() == "desc"
					}), // Order
					mock.MatchedBy(func(field func(fields *cdomain.TransferQueryOptionFields)) bool {
						opt := cdomain.TransferQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}), // SortBy
					mock.MatchedBy(func(field func(fields *cdomain.TransferQueryOptionFields)) bool {
						opt := cdomain.TransferQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 1
					}), // Limit 1
				).Return(int64(2), []domain.Transfer{
					{
						Id:             mockTransfer2.Id,
						RequestId:      mockTransfer2.RequestId,
						TransactionId:  mockTransfer2.TransactionId,
						TransferDate:   mockTransfer2.TransferDate,
						TransferAmount: mockTransfer2.TransferAmount,
						TransferFee:    mockTransfer2.TransferFee,
						TransferStatus: mockTransfer2.TransferStatus,
						ErrorCode:      mockTransfer2.ErrorCode,
						ErrorMsg:       mockTransfer2.ErrorMsg,
					},
					{
						Id:             mockTransfer1.Id,
						RequestId:      mockTransfer1.RequestId,
						TransactionId:  mockTransfer1.TransactionId,
						TransferDate:   mockTransfer1.TransferDate,
						TransferAmount: mockTransfer1.TransferAmount,
						TransferFee:    mockTransfer1.TransferFee,
						TransferStatus: mockTransfer1.TransferStatus,
						ErrorCode:      mockTransfer1.ErrorCode,
						ErrorMsg:       mockTransfer1.ErrorMsg,
					},
				}, nil)
				return repo
			},
			want: adomain.WithdrawalGetOneResponse{
				WithdrawalDetailItem: adomain.WithdrawalDetailItem{
					WithdrawalItem: mockWithdrawalItem1,
					BankName:       mockWithdrawalDetail1.BankName,
					BranchName:     mockWithdrawalDetail1.BranchName,
					BranchCode:     mockWithdrawalDetail1.BranchCode,
					DepositType:    mockWithdrawalDetail1.DepositType,
					BankAccount:    mockWithdrawalDetail1.BankAccount,
					UserType:       mockWithdrawalDetail1.UserType,
				},
				TransferLogs: []adomain.WithdrawalTransferItem{
					mockTransfer2,
					mockTransfer1,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &withdrawalUsecase{
				withdrawalRepository: tt.withdrawalRepository(),
				transferRepository:   tt.transferRepository(),
			}

			ignoreType := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := u.GetOne(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreType), true, "GetOne() = %+v, want %+v", got, tt.want)
		})
	}
}
