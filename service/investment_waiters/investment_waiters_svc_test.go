package investment_waiters

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

var testAssetPath = "../../assets"

func TestInvestmentWaitingStatusTranslations(t *testing.T) {
	translator := i18nutils.New(testAssetPath)
	tests := []struct {
		status domain.INVESTMENT_WAITING_STATUS
		want   map[string]string
	}{
		{domain.INVESTMENT_WAITING_STATUS_WAITING, map[string]string{"ja": "キャンセル待ち", "en": "Cancellation Waitlist", "ko": "투자 취소 대기"}},
		{domain.INVESTMENT_WAITING_STATUS_END_FUNDING, map[string]string{"ja": "募集終了", "en": "Funding Closed", "ko": "모집 종료"}},
		{domain.INVESTMENT_WAITING_STATUS_CANCELLED, map[string]string{"ja": "キャンセル", "en": "Cancelled", "ko": "취소"}},
		{domain.INVESTMENT_WAITING_STATUS_WINNING, map[string]string{"ja": "当選（未入金）", "en": "Selected (Awaiting Deposit)", "ko": "당첨(미입금)"}},
		{domain.INVESTMENT_WAITING_STATUS_WINNING_AFTER_SUCCESS, map[string]string{"ja": "当選（確定）", "en": "Selected (Confirmed)", "ko": "당첨(확정)"}},
		{domain.INVESTMENT_WAITING_STATUS_WINNING_AFTER_CANCELED, map[string]string{"ja": "当選（キャンセル）", "en": "Selected (Cancelled)", "ko": "당첨(취소)"}},
		{domain.INVESTMENT_WAITING_STATUS_WINNING_AFTER_FAILED, map[string]string{"ja": "当選（不成立）", "en": "Selected (Unsuccessful)", "ko": "당첨(불성립)"}},
		{domain.INVESTMENT_WAITING_STATUS_PROPOSAL, map[string]string{"ja": "投資提案中", "en": "Investment Proposal Pending", "ko": "투자 제안 중"}},
	}

	for _, tt := range tests {
		for locale, want := range tt.want {
			t.Run(string(tt.status)+"/"+locale, func(t *testing.T) {
				key := "commoncode.INVESTMENT_WAITING_STATUS." + string(tt.status)
				assert.Equal(t, want, translator.Translate(key, locale))
			})
		}
	}
}

func Test_investmentWaitersService_ExportCSV(t *testing.T) {
	tests := []struct {
		name          string
		status        domain.INVESTMENT_WAITING_TYPE
		projectStatus *domain.PROJECT_STATUS
		investStatus  *domain.INVEST_STATUS
		wantStatus    string
	}{
		{name: "キャンセル待ち", status: domain.WAITING, projectStatus: investmentWaitersTestPtr(domain.OPEN_FUNDING), wantStatus: "キャンセル待ち"},
		{name: "募集終了_ファンド成立後", status: domain.WAITING, projectStatus: investmentWaitersTestPtr(domain.SUCCESS_FUNDING), wantStatus: "募集終了"},
		{name: "キャンセル", status: domain.CANCELLED, wantStatus: "キャンセル"},
		{name: "募集終了_Shareなし", status: domain.WAITING_END, wantStatus: "募集終了"},
		{name: "当選_未入金", status: domain.WAITING_END, investStatus: investmentWaitersTestPtr(domain.REQUEST_INVEST), wantStatus: "当選（未入金）"},
		{name: "当選_未入金_入金要求", status: domain.DEPOSIT_REQUEST, wantStatus: "当選（未入金）"},
		{name: "当選_確定", status: domain.WAITING_END, investStatus: investmentWaitersTestPtr(domain.SUCCESS_INVEST), wantStatus: "当選（確定）"},
		{name: "当選_確定_運用中", status: domain.WAITING_END, investStatus: investmentWaitersTestPtr(domain.OPERATION_INVEST), wantStatus: "当選（確定）"},
		{name: "当選_確定_運用終了", status: domain.WAITING_END, investStatus: investmentWaitersTestPtr(domain.REFUND_INVEST), wantStatus: "当選（確定）"},
		{name: "当選_キャンセル", status: domain.WAITING_END, investStatus: investmentWaitersTestPtr(domain.CANCEL_INVEST), wantStatus: "当選（キャンセル）"},
		{name: "当選_不成立", status: domain.WAITING_END, investStatus: investmentWaitersTestPtr(domain.FAIL_INVEST), wantStatus: "当選（不成立）"},
		{name: "投資提案中", status: domain.PROPOSAL, wantStatus: "投資提案中"},
	}

	svc := investmentWaitersService{
		translator: i18nutils.New(testAssetPath),
	}
	wantHeader := []string{"待機番号", "メールアドレス(ID)", "投資家名", "電話番号", "投資キャンセル待ち申請口数", "申込金額", "ステータス", "申請日時"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := domain.InvestmentWaiters{
				WaitingTurn:      1,
				UserInvestorType: domain.PRIVATE_PRO_INVESTOR,
				UserEmail:        "gildong@email.com",
				UserName:         "고길동",
				UserMobilePhone:  "010-1234-1234",
				Qty:              1,
				Amount:           10000,
				Status:           tt.status,
				ProjectStatus:    tt.projectStatus,
				InvestStatus:     tt.investStatus,
				WaitingDate:      jsonutils.JSONTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			}
			got := svc.ExportCSV([]domain.InvestmentWaiters{item})

			require.Len(t, got, 2)
			assert.Equal(t, wantHeader, got[0])
			assert.Equal(t, []string{"1", "gildong@email.com", "고길동", "010-1234-1234", "1", "10000", tt.wantStatus, "2024-01-01 09:00:00"}, got[1])
		})
	}
}

func investmentWaitersTestPtr[T any](value T) *T {
	return &value
}

func Test_investmentWaitersService_ChangeDepositRequestStatus(t *testing.T) {
	type args struct {
		ctx               context.Context
		investmentWaiters domain.InvestmentWaiters
		userId            int64
	}
	tests := []struct {
		name                         string
		args                         args
		investmentWaitersRepository  func() *mocks.InvestmentWaitersRepository
		investmentWaitHistoryService func() *mocks.InvestmentWaitHistoryService
		wantErr                      bool
	}{
		{
			name: "투자 대기자 입금요청 상태 변경 성공 => 상태가 변경되고 에러가 없음",
			args: args{
				ctx: context.Background(),
				investmentWaiters: domain.InvestmentWaiters{
					Id:        2000,
					ProjectId: 3000,
					UserId:    1000,
				},
				userId: 1000,
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := &mocks.InvestmentWaitersRepository{}
				repo.On("Update", mock.Anything, domain.InvestmentWaiters{
					Id:        2000,
					ProjectId: 3000,
					UserId:    1000,
					Status:    domain.DEPOSIT_REQUEST,
				}).Return(int64(0), nil)

				return repo
			},
			investmentWaitHistoryService: func() *mocks.InvestmentWaitHistoryService {
				service := &mocks.InvestmentWaitHistoryService{}
				service.On("CreateStatusHistory",
					mock.Anything,
					domain.InvestmentWaitersComparer{
						Before: domain.InvestmentWaiters{
							Id:        2000,
							ProjectId: 3000,
							UserId:    1000,
						},
						After: domain.InvestmentWaiters{
							Id:        2000,
							ProjectId: 3000,
							UserId:    1000,
							Status:    domain.DEPOSIT_REQUEST,
						},
					},
					domain.USER_HISTORY_PROCESS_OPERATOR,
					int64(1000),
				).Return(nil)

				return service
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := investmentWaitersService{
				investmentWaitersRepository:  tt.investmentWaitersRepository(),
				investmentWaitHistoryService: tt.investmentWaitHistoryService(),
			}

			err := svc.ChangeDepositRequestStatus(tt.args.ctx, tt.args.investmentWaiters, tt.args.userId)
			assert.Equalf(t, tt.wantErr, err != nil, "ChangeDepositRequestStatus() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_investmentWaitersService_ChangeWaitingEndStatus(t *testing.T) {
	type args struct {
		ctx               context.Context
		investmentWaiters domain.InvestmentWaiters
		userId            int64
	}
	tests := []struct {
		name                         string
		args                         args
		investmentWaitersRepository  func() *mocks.InvestmentWaitersRepository
		investmentRequestRepository  func() *mocks.InvestmentRequestRepository
		investmentWaitHistoryService func() *mocks.InvestmentWaitHistoryService
		wantErr                      bool
	}{
		{
			name: "투자 대기자 대기종료 상태 변경 성공 => 상태가 변경되고 에러가 없음",
			args: args{
				ctx: context.Background(),
				investmentWaiters: domain.InvestmentWaiters{
					Id:        2000,
					ProjectId: 3000,
					UserId:    1000,
				},
				userId: 1000,
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := &mocks.InvestmentWaitersRepository{}
				repo.On("Update", mock.Anything, domain.InvestmentWaiters{
					Id:        2000,
					ProjectId: 3000,
					UserId:    1000,
					Status:    domain.WAITING_END,
				}).Return(int64(1), nil)

				return repo
			},

			investmentRequestRepository: func() *mocks.InvestmentRequestRepository {
				repo := &mocks.InvestmentRequestRepository{}
				repo.On("GetByInvestmentWaitersId", mock.Anything, int64(2000)).Return(
					domain.InvestmentRequest{
						Id:                  4000,
						InvestmentWaitersId: 2000,
						ProjectId:           3000,
					}, nil)
				repo.On("Update", mock.Anything, domain.InvestmentRequest{
					Id:                  4000,
					InvestmentWaitersId: 2000,
					ProjectId:           3000,
					IsExpire:            true,
				}).Return(nil)
				return repo
			},

			investmentWaitHistoryService: func() *mocks.InvestmentWaitHistoryService {
				service := &mocks.InvestmentWaitHistoryService{}
				service.On("CreateStatusHistory",
					mock.Anything,
					domain.InvestmentWaitersComparer{
						Before: domain.InvestmentWaiters{
							Id:        2000,
							ProjectId: 3000,
							UserId:    1000,
						},
						After: domain.InvestmentWaiters{
							Id:        2000,
							ProjectId: 3000,
							UserId:    1000,
							Status:    domain.WAITING_END,
						},
					},
					domain.USER_HISTORY_PROCESS_OPERATOR,
					int64(1000),
				).Return(nil)

				return service
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := investmentWaitersService{
				investmentWaitersRepository:  tt.investmentWaitersRepository(),
				investmentRequestRepository:  tt.investmentRequestRepository(),
				investmentWaitHistoryService: tt.investmentWaitHistoryService(),
			}

			err := svc.ChangeWaitingEndStatus(tt.args.ctx, tt.args.investmentWaiters, tt.args.userId)
			assert.Equalf(t, tt.wantErr, err != nil, "ChangeWaitingEndStatus() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestInvestmentWaitersService_ChangeProposalStatus(t *testing.T) {
	uid, _ := uuid.FromString("bddad349-d0cc-4a1d-8843-431d54fcd2c7")

	type args struct {
		ctx               context.Context
		investmentWaiters domain.InvestmentWaiters
		userId            int64
	}
	tests := []struct {
		name                            string
		args                            args
		idGenerator                     func() *mocks.IdGenerator
		userRepository                  func() *mocks.UserRepository
		projectRepository               func() *mocks.ProjectRepository
		investmentWaitersRepository     func() *mocks.InvestmentWaitersRepository
		investmentRequestRepository     func() *mocks.InvestmentRequestRepository
		investmentWaitHistoryRepository func() *mocks.InvestmentWaitHistoryRepository
		investmentWaitHistoryService    func() *mocks.InvestmentWaitHistoryService
		mailService                     func() *mocks.MailService
		wantErr                         bool
	}{
		{
			name: "투자 대기자 제안으로 상태변경 성공 => 상태가 변경되고 에러가 없음",
			args: args{
				ctx: context.Background(),
				investmentWaiters: domain.InvestmentWaiters{
					Id:        2000,
					ProjectId: 3000,
					UserId:    1000,
					Qty:       1000,
				},
				userId: 1000,
			},
			idGenerator: func() *mocks.IdGenerator {
				gen := new(mocks.IdGenerator)
				gen.On("GenUuidV4").Return(uid)
				return gen
			},

			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", mock.Anything, int64(1000)).Return(domain.User{
					Id:    1000,
					Email: "gildong@email.com",
					Name:  "고길동",
				}, nil)
				return repo
			},

			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("GetById", mock.Anything, int64(3000)).Return(domain.Project{
					Id:           3000,
					RemainingQty: 1000,
				}, nil)
				return repo
			},

			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := &mocks.InvestmentWaitersRepository{}
				repo.On("Update", mock.Anything, domain.InvestmentWaiters{
					Id:        2000,
					ProjectId: 3000,
					UserId:    1000,
					Qty:       1000,
					Status:    domain.PROPOSAL,
				}).Return(int64(0), nil)

				return repo
			},

			investmentRequestRepository: func() *mocks.InvestmentRequestRepository {
				repo := &mocks.InvestmentRequestRepository{}
				repo.On("Create", mock.Anything, mock.MatchedBy(func(investmentRequest domain.InvestmentRequest) bool {
					ignoreFields := cmpopts.IgnoreFields(domain.InvestmentRequest{}, "RequestId", "Url")
					ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

					return cmp.Equal(investmentRequest, domain.InvestmentRequest{
						InvestmentWaitersId: 2000,
						ProjectId:           3000,
						IsExpire:            false,
						Email:               "gildong@email.com",
						Qty:                 1000,
					}, ignoreFields, ignoreTypes)
				})).Return(int64(0), nil)
				return repo
			},

			investmentWaitHistoryService: func() *mocks.InvestmentWaitHistoryService {
				service := &mocks.InvestmentWaitHistoryService{}
				service.On("CreateStatusHistory",
					mock.Anything,
					domain.InvestmentWaitersComparer{
						Before: domain.InvestmentWaiters{
							Id:        2000,
							ProjectId: 3000,
							UserId:    1000,
							Qty:       1000,
						},
						After: domain.InvestmentWaiters{
							Id:        2000,
							ProjectId: 3000,
							UserId:    1000,
							Qty:       1000,
							Status:    domain.PROPOSAL,
						},
					},
					domain.USER_HISTORY_PROCESS_OPERATOR,
					int64(1000),
				).Return(nil)

				return service
			},

			investmentWaitHistoryRepository: func() *mocks.InvestmentWaitHistoryRepository {
				repo := &mocks.InvestmentWaitHistoryRepository{}
				repo.On("Create", mock.Anything, mock.MatchedBy(func(investmentHistory domain.InvestmentWaitHistory) bool {
					ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

					return cmp.Equal(investmentHistory, domain.InvestmentWaitHistory{
						InvestmentWaitId: 2000,
						SubstanceType:    string(domain.INVESTMENT_WAIT_HISTORY_MAIL),
						Name:             "고길동",
						Email:            "gildong@email.com",
						ProcessMethod:    domain.USER_HISTORY_PROCESS_OPERATOR,
					}, ignoreTime)
				})).Return(nil)
				return repo
			},
			mailService: func() *mocks.MailService {
				service := &mocks.MailService{}
				service.On("SendMail", mock.Anything, mock.MatchedBy(func(request cdomain.SendMailRequest) bool {
					ignoreFields := cmpopts.IgnoreFields(cdomain.SendMailRequest{}, "Details")

					return cmp.Equal(request, cdomain.SendMailRequest{
						TemplateType: domain.INVEST_OFFER_TEMPLATE,
						Sender:       "info@energyshares.jp",
						Receivers:    []string{"gildong@email.com"},
						MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
					}, ignoreFields) &&
						request.Details["InvestedQty"] == "1,000" &&
						request.Details["InvestableQty"] == "1,000"
				})).Return(nil)
				return service
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := investmentWaitersService{
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						ServiceConfig: struct {
							Mail serviceutils.MailConfig
							Url  serviceutils.UrlConfig
						}{
							Mail: serviceutils.MailConfig{
								Info: "info@energyshares.jp",
							},
							Url: serviceutils.UrlConfig{
								UserServiceDomain: "http://dev.energyshares.jp",
							},
						},
					},
				},
				idGenerator:                     tt.idGenerator(),
				userRepository:                  tt.userRepository(),
				projectRepository:               tt.projectRepository(),
				investmentWaitersRepository:     tt.investmentWaitersRepository(),
				investmentWaitHistoryRepository: tt.investmentWaitHistoryRepository(),
				investmentRequestRepository:     tt.investmentRequestRepository(),
				investmentWaitHistoryService:    tt.investmentWaitHistoryService(),
				mailService:                     tt.mailService(),
			}

			err := svc.ChangeProposalStatus(tt.args.ctx, tt.args.investmentWaiters, tt.args.userId)
			assert.Equalf(t, tt.wantErr, err != nil, "ChangeProposalStatus() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestInvestmentWaitersService_ChangeCancelledStatus_NotifiesAfterCommit(t *testing.T) {
	waiter := domain.InvestmentWaiters{
		Id:        2000,
		ProjectId: 3000,
		UserId:    1000,
		Amount:    100_000,
		Status:    domain.WAITING,
	}
	updatedWaiter := waiter
	updatedWaiter.Status = domain.CANCELLED
	user := domain.User{Id: waiter.UserId, Name: "キャンセル待ちユーザー", Email: "waiter@example.com"}
	project := domain.Project{Id: waiter.ProjectId, Name: "テストファンド"}

	investmentWaitersRepository := new(mocks.InvestmentWaitersRepository)
	investmentWaitersRepository.On("Update", mock.Anything, updatedWaiter).Return(int64(1), nil)
	investmentWaitHistoryService := new(mocks.InvestmentWaitHistoryService)
	investmentWaitHistoryService.On(
		"CreateStatusHistory",
		mock.Anything,
		domain.InvestmentWaitersComparer{Before: waiter, After: updatedWaiter},
		domain.USER_HISTORY_PROCESS_OPERATOR,
		int64(9999),
	).Return(nil)
	userRepository := new(mocks.UserRepository)
	userRepository.On("GetById", mock.Anything, waiter.UserId).Return(user, nil)
	projectRepository := new(mocks.ProjectRepository)
	projectRepository.On("GetById", mock.Anything, waiter.ProjectId).Return(project, nil)
	mailService := new(mocks.MailService)
	mailService.On("SendMail", mock.Anything, mock.AnythingOfType("clean_arch_domain.SendMailRequest")).Return(nil)
	investmentWaitHistoryRepository := new(mocks.InvestmentWaitHistoryRepository)
	investmentWaitHistoryRepository.On(
		"Create",
		mock.Anything,
		mock.MatchedBy(func(history domain.InvestmentWaitHistory) bool {
			return history.InvestmentWaitId == waiter.Id &&
				history.SubstanceType == string(domain.INVESTMENT_WAIT_HISTORY_MAIL) &&
				history.Email == user.Email
		}),
	).Return(nil)

	service := investmentWaitersService{
		cfg: &config.Config{
			CommonConfig: config.CommonConfig{
				ServiceConfig: struct {
					Mail serviceutils.MailConfig
					Url  serviceutils.UrlConfig
				}{
					Mail: serviceutils.MailConfig{Info: "info@energyshares.jp"},
					Url:  serviceutils.UrlConfig{UserServiceDomain: "http://dev.energyshares.jp"},
				},
			},
		},
		userRepository:                  userRepository,
		projectRepository:               projectRepository,
		investmentWaitersRepository:     investmentWaitersRepository,
		investmentWaitHistoryRepository: investmentWaitHistoryRepository,
		investmentWaitHistoryService:    investmentWaitHistoryService,
		mailService:                     mailService,
	}

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	err = factory.NewTxManager(engine).WithTx(context.Background(), func(txCtx context.Context) error {
		if err := service.ChangeCancelledStatus(txCtx, waiter, 9999); err != nil {
			return err
		}
		userRepository.AssertNotCalled(t, "GetById", mock.Anything, mock.Anything)
		projectRepository.AssertNotCalled(t, "GetById", mock.Anything, mock.Anything)
		mailService.AssertNotCalled(t, "SendMail", mock.Anything, mock.Anything)
		investmentWaitHistoryRepository.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		return nil
	})

	require.NoError(t, err)
	userRepository.AssertExpectations(t)
	projectRepository.AssertExpectations(t)
	mailService.AssertExpectations(t)
	investmentWaitHistoryRepository.AssertExpectations(t)
}
