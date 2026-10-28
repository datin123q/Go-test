package lottery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/actorcontext"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_lotteryUsecase_Application(t *testing.T) {
	ctx := context.Background()

	baseUser := domain.User{
		Id:    1,
		Name:  "Lottery User",
		Email: "lottery@example.com",
	}

	baseProject := domain.Project{
		Id:            10,
		Name:          "Lottery Fund",
		ProjectStatus: domain.OPEN_FUNDING,
		SelectionType: domain.PROJECT_SELECTION_LOTTERY,
		ProjectType:   domain.LOAN,
		SharePrice:    10000,
	}

	baseConfig := &config.Config{
		CommonConfig: config.CommonConfig{
			ServiceConfig: struct {
				Mail serviceutils.MailConfig
				Url  serviceutils.UrlConfig
			}{
				Mail: serviceutils.MailConfig{
					Info: "info@example.com",
				},
				Url: serviceutils.UrlConfig{
					UserServiceDomain: "https://user.example.com",
				},
			},
		},
	}

	type fields struct {
		cfg                      *config.Config
		txManager                func(t *testing.T) *mocks.TxManager
		lotteryValidationService func(t *testing.T) *mocks.LotteryValidationService
		lotteryRepository        func(t *testing.T) *mocks.LotteryApplicationRepository
		eventMessageService      func(t *testing.T) *mocks.EventMessageService
		mailService              func(t *testing.T) *mocks.MailService
		userRepository           func(t *testing.T) *mocks.UserRepository
		virtualAccountRepository func(t *testing.T) *mocks.UserVirtualAccountRepository
		projectRepository        func(t *testing.T) *mocks.ProjectRepository
	}

	type args struct {
		ctx     context.Context
		request udomain.LotteryApplicationRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    udomain.LotteryApplicationResponse
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "抽選ファンドの場合は申込可能",
			fields: fields{
				cfg: baseConfig,
				txManager: func(t *testing.T) *mocks.TxManager {
					txManager := mocks.NewTxManager(t)
					txManager.On("WithTx", ctx, mock.Anything).Return(func(_ context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
					return txManager
				},
				lotteryValidationService: func(t *testing.T) *mocks.LotteryValidationService {
					service := mocks.NewLotteryValidationService(t)
					service.On("CheckUserOverAge", ctx, mock.MatchedBy(func(user *domain.User) bool {
						return user != nil && user.Id == baseUser.Id
					})).Return(nil)
					return service
				},
				lotteryRepository: func(t *testing.T) *mocks.LotteryApplicationRepository {
					repo := mocks.NewLotteryApplicationRepository(t)
					repo.On("IsAlreadyApplied", ctx, baseProject.Id, baseUser.Id).Return(false, nil)
					repo.On("CreateIfNotExists", ctx, mock.Anything).Return("lottery-application-id", true, nil)
					return repo
				},
				eventMessageService: func(t *testing.T) *mocks.EventMessageService {
					service := mocks.NewEventMessageService(t)
					service.On("PublishEvent", ctx, domain.EVENT_LOTTERY_REQUEST, mock.MatchedBy(func(dto domain.LotteryRequestDto) bool {
						return dto.UserId == baseUser.Id &&
							dto.UserName == baseUser.Name &&
							dto.ProductName == baseProject.Name &&
							dto.ProductId == baseProject.Id &&
							dto.Recruitment == baseProject.OfferingType.ToRecruitment() &&
							dto.OrderQuantity == 2 &&
							dto.OrderUnitPrice == baseProject.SharePrice &&
							dto.OrderAmount == 20000 &&
							!dto.OrderDateTime.IsZero()
					})).Return(nil)
					return service
				},
				mailService: func(t *testing.T) *mocks.MailService {
					service := mocks.NewMailService(t)
					service.On("SendMail", ctx, mock.MatchedBy(func(request cdomain.SendMailRequest) bool {
						requestDate, ok := request.Details["RequestDate"].(string)
						return request.TemplateType == domain.LOTTERY_APPLICATION_TEMPLATE &&
							len(request.Receivers) == 1 &&
							request.Receivers[0] == baseUser.Email &&
							request.Details["Receiver"] == baseUser.Name &&
							request.Details["ServiceUrl"] == baseConfig.ServiceConfig.Url.UserServiceDomain &&
							request.Details["ProjectName"] == baseProject.Name &&
							ok &&
							requestDate != "" &&
							request.Details["SharesCount"] == "2" &&
							request.Details["ApplicationAmount"] == "20,000"
					})).Return(nil)
					return service
				},
				userRepository: func(t *testing.T) *mocks.UserRepository {
					repo := mocks.NewUserRepository(t)
					repo.On("GetById", ctx, baseUser.Id).Return(baseUser, nil)
					return repo
				},
				virtualAccountRepository: func(t *testing.T) *mocks.UserVirtualAccountRepository {
					repo := mocks.NewUserVirtualAccountRepository(t)
					repo.On("GetByUserIdAndLinkedStatus", ctx, baseUser.Id, baseProject.ProjectType, domain.ON_VA_ACCOUNT).Return(domain.UserVirtualAccount{
						Id:               1,
						VaBankName:       "Test Bank",
						VaBranchName:     "Main Branch",
						VaAccountNumber:  "1234567",
						VaHolderNameKana: "LOTTERY USER",
					}, nil)
					return repo
				},
				projectRepository: func(t *testing.T) *mocks.ProjectRepository {
					repo := mocks.NewProjectRepository(t)
					repo.On("GetById", ctx, baseProject.Id).Return(baseProject, nil)
					repo.On("AddFinalFundingAmount", ctx, baseProject.Id, int64(20000)).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: ctx,
				request: udomain.LotteryApplicationRequest{
					ProjectId: baseProject.Id,
					UserID:    baseUser.Id,
					Qty:       2,
				},
			},
			want: udomain.LotteryApplicationResponse{
				ProjectName:       baseProject.Name,
				BankName:          "Test Bank",
				BranchName:        "Main Branch",
				BankAccountNumber: "1234567",
				Depositor:         "LOTTERY USER",
				Price:             baseProject.SharePrice,
				Amount:            20000,
				Qty:               2,
			},
			wantErr: assert.NoError,
		},
		{
			name: "先着ファンドの場合は申込不可",
			fields: fields{
				cfg: baseConfig,
				txManager: func(t *testing.T) *mocks.TxManager {
					return mocks.NewTxManager(t)
				},
				lotteryValidationService: func(t *testing.T) *mocks.LotteryValidationService {
					return mocks.NewLotteryValidationService(t)
				},
				lotteryRepository: func(t *testing.T) *mocks.LotteryApplicationRepository {
					return mocks.NewLotteryApplicationRepository(t)
				},
				eventMessageService: func(t *testing.T) *mocks.EventMessageService {
					return mocks.NewEventMessageService(t)
				},
				mailService: func(t *testing.T) *mocks.MailService {
					return mocks.NewMailService(t)
				},
				userRepository: func(t *testing.T) *mocks.UserRepository {
					repo := mocks.NewUserRepository(t)
					repo.On("GetById", ctx, baseUser.Id).Return(baseUser, nil)
					return repo
				},
				virtualAccountRepository: func(t *testing.T) *mocks.UserVirtualAccountRepository {
					return mocks.NewUserVirtualAccountRepository(t)
				},
				projectRepository: func(t *testing.T) *mocks.ProjectRepository {
					repo := mocks.NewProjectRepository(t)
					project := baseProject
					project.SelectionType = domain.PROJECT_SELECTION_FIRST_COME_FIRST_SERVED
					repo.On("GetById", ctx, baseProject.Id).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx: ctx,
				request: udomain.LotteryApplicationRequest{
					ProjectId: baseProject.Id,
					UserID:    baseUser.Id,
					Qty:       1,
				},
			},
			want: udomain.LotteryApplicationResponse{},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(tt, err, cdomain.ErrBadRequest, i...)
			},
		},
		{
			name: "個人ユーザーでマイナンバー未登録の場合は申込不可",
			fields: fields{
				cfg: baseConfig,
				txManager: func(t *testing.T) *mocks.TxManager {
					return mocks.NewTxManager(t)
				},
				lotteryValidationService: func(t *testing.T) *mocks.LotteryValidationService {
					service := mocks.NewLotteryValidationService(t)
					service.On("CheckMyNumberStatusByUserID", ctx, baseUser.Id).Return(false, nil)
					return service
				},
				lotteryRepository: func(t *testing.T) *mocks.LotteryApplicationRepository {
					return mocks.NewLotteryApplicationRepository(t)
				},
				eventMessageService: func(t *testing.T) *mocks.EventMessageService {
					return mocks.NewEventMessageService(t)
				},
				mailService: func(t *testing.T) *mocks.MailService {
					return mocks.NewMailService(t)
				},
				userRepository: func(t *testing.T) *mocks.UserRepository {
					repo := mocks.NewUserRepository(t)
					user := baseUser
					user.UserType = domain.PRIVATE
					repo.On("GetById", ctx, baseUser.Id).Return(user, nil)
					return repo
				},
				virtualAccountRepository: func(t *testing.T) *mocks.UserVirtualAccountRepository {
					return mocks.NewUserVirtualAccountRepository(t)
				},
				projectRepository: func(t *testing.T) *mocks.ProjectRepository {
					repo := mocks.NewProjectRepository(t)
					repo.On("GetById", ctx, baseProject.Id).Return(baseProject, nil)
					return repo
				},
			},
			args: args{
				ctx: ctx,
				request: udomain.LotteryApplicationRequest{
					ProjectId: baseProject.Id,
					UserID:    baseUser.Id,
					Qty:       1,
				},
			},
			want: udomain.LotteryApplicationResponse{},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				var factoryErr *factory.Error
				if !assert.ErrorAs(tt, err, &factoryErr, i...) {
					return false
				}
				return assert.Equal(tt, factory.ErrorMyNumberNotRegistered.Code, factoryErr.Code)
			},
		},
		{
			name: "投資制限ユーザーの場合は申込不可",
			fields: fields{
				cfg: baseConfig,
				txManager: func(t *testing.T) *mocks.TxManager {
					return mocks.NewTxManager(t)
				},
				lotteryValidationService: func(t *testing.T) *mocks.LotteryValidationService {
					return mocks.NewLotteryValidationService(t)
				},
				lotteryRepository: func(t *testing.T) *mocks.LotteryApplicationRepository {
					return mocks.NewLotteryApplicationRepository(t)
				},
				eventMessageService: func(t *testing.T) *mocks.EventMessageService {
					return mocks.NewEventMessageService(t)
				},
				mailService: func(t *testing.T) *mocks.MailService {
					return mocks.NewMailService(t)
				},
				userRepository: func(t *testing.T) *mocks.UserRepository {
					repo := mocks.NewUserRepository(t)
					user := baseUser
					user.BlockInvest = true
					repo.On("GetById", ctx, baseUser.Id).Return(user, nil)
					return repo
				},
				virtualAccountRepository: func(t *testing.T) *mocks.UserVirtualAccountRepository {
					return mocks.NewUserVirtualAccountRepository(t)
				},
				projectRepository: func(t *testing.T) *mocks.ProjectRepository {
					repo := mocks.NewProjectRepository(t)
					repo.On("GetById", ctx, baseProject.Id).Return(baseProject, nil)
					return repo
				},
			},
			args: args{
				ctx: ctx,
				request: udomain.LotteryApplicationRequest{
					ProjectId: baseProject.Id,
					UserID:    baseUser.Id,
					Qty:       1,
				},
			},
			want: udomain.LotteryApplicationResponse{},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				var factoryErr *factory.Error
				if !assert.ErrorAs(tt, err, &factoryErr, i...) {
					return false
				}
				return assert.Equal(tt, factory.ErrorBlockInvest.Code, factoryErr.Code)
			},
		},
		{
			name: "年齢制限に抵触する場合は申込不可",
			fields: fields{
				cfg: baseConfig,
				txManager: func(t *testing.T) *mocks.TxManager {
					return mocks.NewTxManager(t)
				},
				lotteryValidationService: func(t *testing.T) *mocks.LotteryValidationService {
					service := mocks.NewLotteryValidationService(t)
					service.On("CheckUserOverAge", ctx, mock.MatchedBy(func(user *domain.User) bool {
						return user != nil && user.Id == baseUser.Id
					})).Return(factory.ErrorOverAgeInvestor.New(errors.New("over age")))
					return service
				},
				lotteryRepository: func(t *testing.T) *mocks.LotteryApplicationRepository {
					return mocks.NewLotteryApplicationRepository(t)
				},
				eventMessageService: func(t *testing.T) *mocks.EventMessageService {
					return mocks.NewEventMessageService(t)
				},
				mailService: func(t *testing.T) *mocks.MailService {
					return mocks.NewMailService(t)
				},
				userRepository: func(t *testing.T) *mocks.UserRepository {
					repo := mocks.NewUserRepository(t)
					repo.On("GetById", ctx, baseUser.Id).Return(baseUser, nil)
					return repo
				},
				virtualAccountRepository: func(t *testing.T) *mocks.UserVirtualAccountRepository {
					return mocks.NewUserVirtualAccountRepository(t)
				},
				projectRepository: func(t *testing.T) *mocks.ProjectRepository {
					repo := mocks.NewProjectRepository(t)
					repo.On("GetById", ctx, baseProject.Id).Return(baseProject, nil)
					return repo
				},
			},
			args: args{
				ctx: ctx,
				request: udomain.LotteryApplicationRequest{
					ProjectId: baseProject.Id,
					UserID:    baseUser.Id,
					Qty:       1,
				},
			},
			want: udomain.LotteryApplicationResponse{},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				var factoryErr *factory.Error
				if !assert.ErrorAs(tt, err, &factoryErr, i...) {
					return false
				}
				return assert.Equal(tt, factory.ErrorOverAgeInvestor.Code, factoryErr.Code)
			},
		},
		{
			name: "投資家数上限に達している場合は申込不可",
			fields: fields{
				cfg: baseConfig,
				txManager: func(t *testing.T) *mocks.TxManager {
					return mocks.NewTxManager(t)
				},
				lotteryValidationService: func(t *testing.T) *mocks.LotteryValidationService {
					service := mocks.NewLotteryValidationService(t)
					service.On("CheckUserOverAge", ctx, mock.MatchedBy(func(user *domain.User) bool {
						return user != nil && user.Id == baseUser.Id
					})).Return(nil)
					service.On("GetSuccessInvestorCountByProjectID", ctx, baseProject.Id).Return(int(domain.TargetInvestorCount), nil)
					return service
				},
				lotteryRepository: func(t *testing.T) *mocks.LotteryApplicationRepository {
					return mocks.NewLotteryApplicationRepository(t)
				},
				eventMessageService: func(t *testing.T) *mocks.EventMessageService {
					return mocks.NewEventMessageService(t)
				},
				mailService: func(t *testing.T) *mocks.MailService {
					return mocks.NewMailService(t)
				},
				userRepository: func(t *testing.T) *mocks.UserRepository {
					repo := mocks.NewUserRepository(t)
					repo.On("GetById", ctx, baseUser.Id).Return(baseUser, nil)
					return repo
				},
				virtualAccountRepository: func(t *testing.T) *mocks.UserVirtualAccountRepository {
					return mocks.NewUserVirtualAccountRepository(t)
				},
				projectRepository: func(t *testing.T) *mocks.ProjectRepository {
					repo := mocks.NewProjectRepository(t)
					project := baseProject
					project.IsInvestorLimit = true
					repo.On("GetById", ctx, baseProject.Id).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx: ctx,
				request: udomain.LotteryApplicationRequest{
					ProjectId: baseProject.Id,
					UserID:    baseUser.Id,
					Qty:       1,
				},
			},
			want: udomain.LotteryApplicationResponse{},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				var factoryErr *factory.Error
				if !assert.ErrorAs(tt, err, &factoryErr, i...) {
					return false
				}
				return assert.Equal(tt, factory.ErrorOverInvestorCount.Code, factoryErr.Code)
			},
		},
		{
			name: "投資金額上限を超える場合は申込不可",
			fields: fields{
				cfg: baseConfig,
				txManager: func(t *testing.T) *mocks.TxManager {
					return mocks.NewTxManager(t)
				},
				lotteryValidationService: func(t *testing.T) *mocks.LotteryValidationService {
					service := mocks.NewLotteryValidationService(t)
					service.On("CheckUserOverAge", ctx, mock.MatchedBy(func(user *domain.User) bool {
						return user != nil && user.Id == baseUser.Id
					})).Return(nil)
					service.On("GetInvestedAmountByUserID", ctx, baseProject.Id, baseUser.Id).Return(int64(4990000), nil)
					service.On("GetWaitingAmountByUserID", ctx, baseProject.Id, baseUser.Id).Return(int64(0), nil)
					return service
				},
				lotteryRepository: func(t *testing.T) *mocks.LotteryApplicationRepository {
					return mocks.NewLotteryApplicationRepository(t)
				},
				eventMessageService: func(t *testing.T) *mocks.EventMessageService {
					return mocks.NewEventMessageService(t)
				},
				mailService: func(t *testing.T) *mocks.MailService {
					return mocks.NewMailService(t)
				},
				userRepository: func(t *testing.T) *mocks.UserRepository {
					repo := mocks.NewUserRepository(t)
					repo.On("GetById", ctx, baseUser.Id).Return(baseUser, nil)
					return repo
				},
				virtualAccountRepository: func(t *testing.T) *mocks.UserVirtualAccountRepository {
					return mocks.NewUserVirtualAccountRepository(t)
				},
				projectRepository: func(t *testing.T) *mocks.ProjectRepository {
					repo := mocks.NewProjectRepository(t)
					project := baseProject
					project.IsAmountLimit = true
					repo.On("GetById", ctx, baseProject.Id).Return(project, nil)
					return repo
				},
			},
			args: args{
				ctx: ctx,
				request: udomain.LotteryApplicationRequest{
					ProjectId: baseProject.Id,
					UserID:    baseUser.Id,
					Qty:       2,
				},
			},
			want: udomain.LotteryApplicationResponse{},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				var factoryErr *factory.Error
				if !assert.ErrorAs(tt, err, &factoryErr, i...) {
					return false
				}
				return assert.Equal(tt, factory.ErrorOverLimitedAmount.Code, factoryErr.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := lotteryUsecase{
				cfg:                      tt.fields.cfg,
				txManager:                tt.fields.txManager(t),
				lotteryValidationService: tt.fields.lotteryValidationService(t),
				lotteryRepository:        tt.fields.lotteryRepository(t),
				eventMessageService:      tt.fields.eventMessageService(t),
				mailService:              tt.fields.mailService(t),
				userRepository:           tt.fields.userRepository(t),
				virtualAccountRepository: tt.fields.virtualAccountRepository(t),
				projectRepository:        tt.fields.projectRepository(t),
			}

			got, err := u.Application(tt.args.ctx, tt.args.request)

			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func newLotteryCancelTxManagerMock() *mocks.TxManager {
	txManager := &mocks.TxManager{}
	txManager.On("WithTx", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
	return txManager
}

func TestLotteryUsecase_Cancel(t *testing.T) {
	tests := []struct {
		name               string
		ctx                context.Context
		request            udomain.LotteryCancelRequest
		lotteryRepo        func() *mocks.LotteryApplicationRepository
		projectRepo        func() *mocks.ProjectRepository
		mailService        func(t *testing.T) *mocks.MailService
		userRepository     func(t *testing.T) *mocks.UserRepository
		lotteryHistoryRepo func() *mocks.LotteryHistoryRepository
		want               udomain.LotteryCancelResponse
		wantErr            error
	}{
		{
			name: "(成功) 抽選申込キャンセル",
			ctx:  actorcontext.WithActor(context.Background(), jwthelper.User{Id: 100}),
			request: udomain.LotteryCancelRequest{
				JWTUser:              jwthelper.User{Id: 100},
				LotteryApplicationID: "app-001",
			},
			lotteryRepo: func() *mocks.LotteryApplicationRepository {
				repo := &mocks.LotteryApplicationRepository{}
				repo.On("LockById", mock.Anything, "app-001").
					Return(domain.LotteryApplication{
						Id:        "app-001",
						ProjectId: 1,
						UserId:    100,
						Amount:    10000,
					}, nil)
				repo.On("MarkAsCanceled", mock.Anything, "app-001").
					Return(true, nil)
				repo.On("GetById", mock.Anything, "app-001").
					Return(domain.LotteryApplication{
						Id:         "app-001",
						ProjectId:  1,
						UserId:     100,
						Amount:     10000,
						CanceledAt: jsonutils.JSONTime(time.Date(2026, 4, 27, 12, 34, 56, 0, time.UTC)),
					}, nil)
				return repo
			},
			projectRepo: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("LockById", mock.Anything, int64(1)).
					Return(domain.Project{Id: 1, ProjectStatus: domain.OPEN_FUNDING}, nil)
				repo.On("ReduceFinalFundingAmount", mock.Anything, int64(1), int64(10000)).
					Return(nil)
				repo.On("GetById", mock.Anything, int64(1)).
					Return(domain.Project{Id: 1, Name: "Lottery Fund", ProjectStatus: domain.OPEN_FUNDING}, nil)
				return repo
			},
			mailService: func(t *testing.T) *mocks.MailService {
				service := mocks.NewMailService(t)
				service.On("SendMail", mock.Anything, mock.MatchedBy(func(request cdomain.SendMailRequest) bool {
					return cmp.Equal(request, cdomain.SendMailRequest{
						TemplateType: domain.LOTTERY_APPLICATION_CANCEL_TEMPLATE,
						MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
						Sender:       "info@example.com",
						Receivers:    []string{"lottery@example.com"},
						Details: map[string]interface{}{
							"ProjectName":       "Lottery Fund",
							"Receiver":          "Lottery User",
							"ApplicationAmount": "10,000",
							"CancelTime":        "2026-04-27 21:34:56",
							"ServiceUrl":        "https://user.example.com",
						},
					})
				})).Return(nil)
				return service
			},
			userRepository: func(t *testing.T) *mocks.UserRepository {
				repo := mocks.NewUserRepository(t)
				repo.On("GetById", mock.Anything, int64(100)).Return(domain.User{
					Id:    100,
					Name:  "Lottery User",
					Email: "lottery@example.com",
				}, nil)
				return repo
			},
			lotteryHistoryRepo: func() *mocks.LotteryHistoryRepository {
				repo := &mocks.LotteryHistoryRepository{}
				repo.On("Create", mock.Anything, domain.LotteryHistory{
					ApplicationId: "app-001",
					PrevStatus:    domain.LOTTERY_STATUS_APPLIED,
					NewStatus:     domain.LOTTERY_STATUS_CANCELED,
					ActorId:       100,
				}).Return(nil)
				return repo
			},
			want: udomain.LotteryCancelResponse{},
		},
		{
			name: "(失敗) ファンドのステータスが募集終了の場合はキャンセルできないのでエラー",
			ctx:  actorcontext.WithActor(context.Background(), jwthelper.User{Id: 100}),
			request: udomain.LotteryCancelRequest{
				JWTUser:              jwthelper.User{Id: 100},
				LotteryApplicationID: "app-003",
			},
			lotteryRepo: func() *mocks.LotteryApplicationRepository {
				repo := &mocks.LotteryApplicationRepository{}
				repo.On("LockById", mock.Anything, "app-003").
					Return(domain.LotteryApplication{
						Id:        "app-003",
						ProjectId: 1,
						UserId:    100,
						Amount:    10000,
					}, nil)
				return repo
			},
			projectRepo: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("LockById", mock.Anything, int64(1)).
					Return(domain.Project{Id: 1, ProjectStatus: domain.END_FUNDING}, nil)
				return repo
			},
			mailService: func(t *testing.T) *mocks.MailService {
				return mocks.NewMailService(t)
			},
			userRepository: func(t *testing.T) *mocks.UserRepository {
				return mocks.NewUserRepository(t)
			},
			lotteryHistoryRepo: func() *mocks.LotteryHistoryRepository {
				return &mocks.LotteryHistoryRepository{}
			},
			want:    udomain.LotteryCancelResponse{},
			wantErr: cdomain.ErrBadRequest,
		},
		{
			name: "(失敗) 他人の申込をキャンセルしようとした場合は not found エラー",
			ctx:  actorcontext.WithActor(context.Background(), jwthelper.User{Id: 100}),
			request: udomain.LotteryCancelRequest{
				JWTUser:              jwthelper.User{Id: 100},
				LotteryApplicationID: "app-002",
			},
			lotteryRepo: func() *mocks.LotteryApplicationRepository {
				repo := &mocks.LotteryApplicationRepository{}
				repo.On("LockById", mock.Anything, "app-002").
					Return(domain.LotteryApplication{
						Id:        "app-002",
						ProjectId: 1,
						UserId:    999,
						Amount:    10000,
					}, nil)
				return repo
			},
			projectRepo: func() *mocks.ProjectRepository {
				return &mocks.ProjectRepository{}
			},
			mailService: func(t *testing.T) *mocks.MailService {
				return mocks.NewMailService(t)
			},
			userRepository: func(t *testing.T) *mocks.UserRepository {
				return mocks.NewUserRepository(t)
			},
			lotteryHistoryRepo: func() *mocks.LotteryHistoryRepository {
				return &mocks.LotteryHistoryRepository{}
			},
			want:    udomain.LotteryCancelResponse{},
			wantErr: cdomain.ErrItemNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := lotteryUsecase{
				lotteryRepository:        tt.lotteryRepo(),
				lotteryHistoryRepository: tt.lotteryHistoryRepo(),
				projectRepository:        tt.projectRepo(),
				txManager:                newLotteryCancelTxManagerMock(),
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						ServiceConfig: struct {
							Mail serviceutils.MailConfig
							Url  serviceutils.UrlConfig
						}{
							Mail: serviceutils.MailConfig{Info: "info@example.com"},
							Url:  serviceutils.UrlConfig{UserServiceDomain: "https://user.example.com"},
						},
					},
				},
				mailService:              tt.mailService(t),
				userRepository:           tt.userRepository(t),
				virtualAccountRepository: nil,
			}

			got, err := u.Cancel(tt.ctx, tt.request)

			assert.Equal(t, tt.want, got)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func Test_lotteryUsecase_GetApplications(t *testing.T) {
	appliedAt1 := jsonutils.JSONTime(time.Date(2025, 2, 10, 12, 34, 56, 0, time.FixedZone("JST", 9*60*60)))
	appliedAt2 := jsonutils.JSONTime(time.Date(2025, 2, 1, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60)))
	repo := mocks.NewLotteryApplicationRepository(t)
	repo.On("GetAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, options ...func(*cdomain.LotteryApplicationQueryOptionFields)) int64 {
			queryOpts := cdomain.LotteryApplicationQueryOptionFunc()
			for _, option := range options {
				option(&queryOpts)
			}

			require.Equal(t, int64(1000), *queryOpts.UserId())
			require.Equal(t, 2, queryOpts.Limit())
			require.Equal(t, 2, queryOpts.Offset())
			// 同時刻の申込があるため id で一意に定める。項目数も揃える。
			require.Equal(t, "lottery_application.applied_at,lottery_application.id", queryOpts.SortBy())
			require.Equal(t, "desc,desc", queryOpts.Order())

			return 24
		},
		func(ctx context.Context, options ...func(*cdomain.LotteryApplicationQueryOptionFields)) []domain.LotteryApplication {
			return []domain.LotteryApplication{
				{
					Id:            "application-1",
					ProjectId:     501,
					ProjectName:   "九州太陽光ファンド第1号",
					ProjectStatus: domain.OPEN_FUNDING,
					AppliedAt:     appliedAt1,
					Status:        domain.LOTTERY_STATUS_APPLIED,
					Amount:        300000,
					Winnings: []domain.Share{
						{Amount: 100000, InvestStatus: domain.REQUEST_INVEST},
						{Amount: 50000, InvestStatus: domain.CANCEL_INVEST},
					},
				},
				{
					Id:            "application-2",
					ProjectId:     502,
					ProjectName:   "大阪蓄電池ファンド第3号",
					ProjectStatus: domain.END_FUNDING,
					AppliedAt:     appliedAt2,
					Status:        domain.LOTTERY_STATUS_WINNING,
					Amount:        500000,
					Winnings: []domain.Share{
						{Amount: 150000, InvestStatus: domain.SUCCESS_INVEST},
						{Amount: 50000, InvestStatus: domain.REQUEST_INVEST},
						{Amount: 70000, InvestStatus: domain.OPERATION_INVEST},
					},
				},
			}
		},
		nil,
	)

	u := lotteryUsecase{
		lotteryRepository: repo,
	}

	response, err := u.GetApplications(context.Background(), udomain.LotteryGetApplicationsRequest{
		UserID: 1000,
		Limit:  2,
		Offset: 2,
	})

	require.NoError(t, err)
	require.Len(t, response.Items, 2)

	assert.Equal(t, int64(24), response.TotalCount)
	assert.Equal(t, 2, response.Limit)
	assert.Equal(t, 2, response.Offset)

	assert.Equal(t, udomain.LotteryGetApplicationsItem{
		Id:                "application-1",
		ProjectId:         501,
		ProjectName:       "九州太陽光ファンド第1号",
		ApplicationDate:   appliedAt1,
		LotteryStatus:     domain.LOTTERY_STATUS_APPLIED,
		ApplicationAmount: 300000,
		WinningAmount:     150000,
		Cancelable:        true,
	}, response.Items[0])

	assert.Equal(t, udomain.LotteryGetApplicationsItem{
		Id:                "application-2",
		ProjectId:         502,
		ProjectName:       "大阪蓄電池ファンド第3号",
		ApplicationDate:   appliedAt2,
		LotteryStatus:     domain.LOTTERY_STATUS_WINNING,
		ApplicationAmount: 500000,
		WinningAmount:     270000,
		Cancelable:        false,
	}, response.Items[1])
}
