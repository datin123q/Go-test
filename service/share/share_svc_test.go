package share

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

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
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

func Test_shareService_RequestInvest(t *testing.T) {
	type args struct {
		ctx      context.Context
		sharedId string
		userId   int64
	}
	tests := []struct {
		name                         string
		args                         args
		shareRepository              func() *mocks.ShareRepository
		userRepository               func() *mocks.UserRepository
		userAccountRepository        func() *mocks.UserAccountRepository
		userVirtualAccountRepository func() *mocks.UserVirtualAccountRepository
		projectRepository            func() *mocks.ProjectRepository
		depositRequestRepository     func() *mocks.DepositRequestRepository
		depositRequestService        func() *mocks.DepositRequestService
		projectService               func() *mocks.ProjectService
		mailService                  func() *mocks.MailService
		eventMessageService          func() *mocks.EventMessageService
		wantErr                      bool
	}{
		{
			name: "投資要求が正常に処理されます : 残高が十分でない場合 => 残高不足メールを送信して正常に終了します。",
			args: args{
				ctx:      context.Background(),
				sharedId: mockShare1Id,
				userId:   mockActor1Id,
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := &mocks.ShareRepository{}
				repo.On("GetById", mock.Anything, mockShare1Id).Return(mockShare1, nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(share domain.Share) bool {
					return share.InvestStatus == domain.REQUEST_INVEST
				})).Return(nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", mock.Anything, mockShare1.OwnerId).Return(mockUser1, nil)
				return repo
			},
			userAccountRepository: func() *mocks.UserAccountRepository {
				repo := &mocks.UserAccountRepository{}
				repo.On("GetByUserId", mock.Anything, mockUser1Id).Return(mockUserAccount1, nil)
				return repo
			},
			userVirtualAccountRepository: func() *mocks.UserVirtualAccountRepository {
				repo := &mocks.UserVirtualAccountRepository{}
				repo.On("GetByUserIdAndLinkedStatus", mock.Anything, mockUser1Id, mockProject1.ProjectType, domain.ON_VA_ACCOUNT).Return(mockUserVirtualAccount1, nil)
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("GetById", mock.Anything, mockProject1Id).Return(mockProject1, nil)
				return repo
			},
			depositRequestRepository: func() *mocks.DepositRequestRepository {
				repo := &mocks.DepositRequestRepository{}
				repo.On("GetDepositRequestByShareId", mock.Anything, mockShare1Id).Return(mockDepositRequest1, nil)
				return repo
			},
			depositRequestService: func() *mocks.DepositRequestService {
				service := &mocks.DepositRequestService{}
				return service
			},
			projectService: func() *mocks.ProjectService { return &mocks.ProjectService{} },
			mailService: func() *mocks.MailService {
				service := &mocks.MailService{}
				return service
			},
			eventMessageService: func() *mocks.EventMessageService {
				service := &mocks.EventMessageService{}
				service.On("PublishEvent", mock.Anything, domain.EVENT_INVESTMENT_REQUEST, domain.InvestmentRequestDto{
					UserId:         mockUser1.Id,
					UserName:       mockUser1.Name,
					InvestmentId:   mockShare1.Id,
					ProductName:    mockProject1.Name,
					ProductId:      mockProject1.Id,
					Recruitment:    mockProject1.OfferingType.ToRecruitment(),
					OrderDateTime:  mockShare1.InvestDate.Time(),
					OrderQuantity:  mockShare1.Qty,
					OrderUnitPrice: mockShare1.Price,
					OrderAmount:    mockShare1.Amount,
				}).Return(nil)
				return service
			},
			wantErr: false,
		},
		{
			name: "投資要求が正常に処理されます : 投資金額と人数が満たされていない場合 => プロジェクトファンディングを終了せずに正常に終了します。",
			args: args{
				ctx:      context.Background(),
				sharedId: mockShare1Id,
				userId:   mockActor1Id,
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := &mocks.ShareRepository{}
				repo.On("GetById", mock.Anything, mockShare1Id).Return(mockShare1, nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(share domain.Share) bool {
					return share.InvestStatus == domain.REQUEST_INVEST
				})).Return(nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", mock.Anything, mockShare1.OwnerId).Return(mockUser1, nil)
				return repo
			},
			userAccountRepository: func() *mocks.UserAccountRepository {
				repo := &mocks.UserAccountRepository{}
				repo.On("GetByUserId", mock.Anything, mockUser1Id).Return(mockUserAccount1, nil)
				return repo
			},
			userVirtualAccountRepository: func() *mocks.UserVirtualAccountRepository {
				repo := &mocks.UserVirtualAccountRepository{}
				repo.On("GetByUserIdAndLinkedStatus", mock.Anything, mockUser1Id, mockProject1.ProjectType, domain.ON_VA_ACCOUNT).Return(mockUserVirtualAccount1, nil)
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("GetById", mock.Anything, mockProject1Id).Return(mockProject1, nil)
				return repo
			},
			depositRequestRepository: func() *mocks.DepositRequestRepository {
				repo := &mocks.DepositRequestRepository{}
				insufficientDepositRequest := mockDepositRequest1

				// Need Amount 가 600,000으로 나오도록 설정
				insufficientDepositRequest.RequestAmount = 700_000

				repo.On("GetDepositRequestByShareId", mock.Anything, mockShare1Id).Return(insufficientDepositRequest, nil)
				return repo
			},
			depositRequestService: func() *mocks.DepositRequestService {
				service := &mocks.DepositRequestService{}
				return service
			},
			projectService: func() *mocks.ProjectService { return &mocks.ProjectService{} },
			mailService: func() *mocks.MailService {
				service := &mocks.MailService{}
				service.On("SendMail", mock.Anything, cdomain.SendMailRequest{
					TemplateType: domain.INVEST_INSUFFICIENT_BALANCE_TEMPLATE,
					Sender:       "info@energyshares.jp",
					Receivers:    []string{mockUser1.Email},
					MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
					Details: map[string]interface{}{
						"Receiver":         mockUser1.Name,
						"ServiceUrl":       "http://dev.energyshares.jp",
						"LackAmount":       "9,400,000",
						"ProductName":      mockProject1.Name,
						"TransactionDate":  jsonutils.JSONTime.ToString(mockShare1.InvestDate),
						"BankName":         mockDepositRequest1.DepositBankName,
						"BranchName":       mockDepositRequest1.DepositBranchName,
						"AccountNumber":    mockDepositRequest1.DepositAccountNumber,
						"Depositor":        mockUserVirtualAccount1.VaHolderNameKana,
						"InvestmentAmount": "10,000,000",
						"PaymentEnd":       jsonutils.JSONTime.ToString(mockDepositRequest1.ExpireDepositDate),
					},
				}).Return(nil)
				return service
			},
			eventMessageService: func() *mocks.EventMessageService {
				service := &mocks.EventMessageService{}
				service.On("PublishEvent", mock.Anything, domain.EVENT_INVESTMENT_REQUEST, domain.InvestmentRequestDto{
					UserId:         mockUser1.Id,
					UserName:       mockUser1.Name,
					InvestmentId:   mockShare1.Id,
					ProductName:    mockProject1.Name,
					ProductId:      mockProject1.Id,
					Recruitment:    mockProject1.OfferingType.ToRecruitment(),
					OrderDateTime:  mockShare1.InvestDate.Time(),
					OrderQuantity:  mockShare1.Qty,
					OrderUnitPrice: mockShare1.Price,
					OrderAmount:    mockShare1.Amount,
				}).Return(nil)
				return service
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := shareService{
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
				shareRepository:              tt.shareRepository(),
				projectRepository:            tt.projectRepository(),
				userRepository:               tt.userRepository(),
				userAccountRepository:        tt.userAccountRepository(),
				userVirtualAccountRepository: tt.userVirtualAccountRepository(),
				depositRequestRepository:     tt.depositRequestRepository(),
				depositRequestService:        tt.depositRequestService(),
				projectService:               tt.projectService(),
				mailService:                  tt.mailService(),
				eventMessageService:          tt.eventMessageService(),
			}

			err := service.RequestInvest(tt.args.ctx, tt.args.sharedId, tt.args.userId)
			assert.Equalf(t, tt.wantErr, err != nil, "shareService.RequestInvest() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_shareService_SuccessInvest(t *testing.T) {
	type args struct {
		ctx      context.Context
		sharedId string
		userId   int64
	}
	tests := []struct {
		name                      string
		args                      args
		shareRepository           func() *mocks.ShareRepository
		adminUserRepository       func() *mocks.AdminUserRepository
		userRepository            func() *mocks.UserRepository
		projectRepository         func() *mocks.ProjectRepository
		contractRepository        func() *mocks.ContractRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		depositRequestRepository  func() *mocks.DepositRequestRepository
		spvRepository             func() *mocks.SpvRepository
		depositRequestService     func() *mocks.DepositRequestService
		projectService            func() *mocks.ProjectService
		mailService               func() *mocks.MailService
		eventMessageService       func() *mocks.EventMessageService
		wantErr                   bool
	}{
		{
			name: "投資成功として正常に処理されます : 投資金額と人数が満たされていない場合 => プロジェクトファンディングを終了せずに正常に終了します。",
			args: args{
				ctx:      context.Background(),
				sharedId: mockShare1Id,
				userId:   mockActor1Id,
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := &mocks.ShareRepository{}

				repo.On("GetById", mock.Anything, mockShare1Id).Return(mockShare1, nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(share domain.Share) bool {
					return share.InvestStatus == domain.SUCCESS_INVEST &&
						share.UpdateId == mockActor1Id
				})).Return(nil)
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("GetById", mock.Anything, mockProject1Id).Return(mockProject1, nil)
				return repo
			},
			adminUserRepository: func() *mocks.AdminUserRepository {
				repo := &mocks.AdminUserRepository{}
				repo.On("GetById", mock.Anything, mockActor1Id).Return(mockActor1, nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", mock.Anything, mockUser1Id).Return(mockUser1, nil)
				return repo
			},
			contractRepository: func() *mocks.ContractRepository {
				repo := &mocks.ContractRepository{}
				repo.On("Create", mock.Anything, domain.Contract{
					ShareId:     mockShare1Id,
					ShareQty:    mockShare1.Qty,
					ShareAmount: mockShare1.Amount,
					Status:      domain.CONTRACT_NONE,
					ProjectId:   mockProject1Id,
					UserName:    mockUser1.Name,
				}).Return(int64(0), nil)
				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := &mocks.OperatorHistoryRepository{}
				repo.On("Create", mock.Anything, domain.OperatorHistory{
					UserId: mockActor1Id,
					Email:  mockActor1.Email,
					Menu:   "commoncode.INVEST_STATUS.SUCCESS_INVEST",
					Ip:     "::1",
					State:  domain.CHANGE,
				}).Return(nil)
				return repo
			},
			projectService: func() *mocks.ProjectService {
				return &mocks.ProjectService{}
			},

			depositRequestRepository: func() *mocks.DepositRequestRepository {
				repo := &mocks.DepositRequestRepository{}
				repo.On("GetDepositRequestByShareId", mock.Anything, mockShare1Id).Return(mockDepositRequest1, nil)
				return repo
			},
			depositRequestService: func() *mocks.DepositRequestService {
				service := &mocks.DepositRequestService{}
				service.On("OkDeposit", mock.Anything, mockDepositRequest1).Return(nil)
				return service
			},

			spvRepository: func() *mocks.SpvRepository {
				repo := &mocks.SpvRepository{}
				repo.On("GetById", mock.Anything, mockProject1.SpvId).Return(mockSpv1, nil)
				return repo
			},
			mailService: func() *mocks.MailService {
				service := &mocks.MailService{}
				service.On("SendMail", mock.Anything, cdomain.SendMailRequest{
					TemplateType: domain.INVEST_SUCCESS_TEMPLATE,
					Sender:       "info@energyshares.jp",
					Receivers:    []string{mockUser1.Email},
					MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
					Details: map[string]interface{}{
						"Receiver":      mockUser1.Name,
						"ServiceUrl":    "http://dev.energyshares.jp",
						"ProjectName":   mockProject1.Name,
						"TradingTime":   mockShare1.Updated.Time().Format(time.DateTime),
						"DepositDate":   mockDepositRequest1.DepositDate.ToString(),
						"DepositAmount": strconv.FormatInt(mockDepositRequest1.DepositAmount, 10),
					},
				}).Return(nil)
				return service
			},
			eventMessageService: func() *mocks.EventMessageService {
				service := &mocks.EventMessageService{}
				service.On("PublishEvent", mock.Anything, domain.EVENT_INVESTMENT_COMPLETED, domain.InvestmentCompletedDto{
					UserId:          mockShare1.OwnerId,
					UserName:        mockUser1.Name,
					InvestmentId:    mockShare1.Id,
					CompanyName:     mockProject1.Spv.Name,
					ProductName:     mockProject1.Name,
					ProductId:       mockProject1.Id,
					Recruitment:     mockProject1.OfferingType.ToRecruitment(),
					CommitDateTime:  mockDepositRequest1.DepositDate.Time(),
					CommitQuantity:  mockShare1.Qty,
					CommitUnitPrice: mockShare1.Price,
					CommitAmount:    mockShare1.Amount,
				}).Return(nil)
				return service
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := shareService{
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
				shareRepository:           tt.shareRepository(),
				adminUserRepository:       tt.adminUserRepository(),
				userRepository:            tt.userRepository(),
				projectRepository:         tt.projectRepository(),
				contractRepository:        tt.contractRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
				depositRequestRepository:  tt.depositRequestRepository(),
				spvRepository:             tt.spvRepository(),
				projectService:            tt.projectService(),
				depositRequestService:     tt.depositRequestService(),
				mailService:               tt.mailService(),
				eventMessageService:       tt.eventMessageService(),
			}
			err := service.SuccessInvest(tt.args.ctx, tt.args.sharedId, tt.args.userId)
			assert.Equalf(t, tt.wantErr, err != nil, "shareService.SuccessInvest() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_shareService_CancelInvest(t *testing.T) {
	type args struct {
		ctx      context.Context
		sharedId string
		userId   int64
	}
	tests := []struct {
		name                      string
		args                      args
		shareRepository           func() *mocks.ShareRepository
		projectRepository         func() *mocks.ProjectRepository
		adminUserRepository       func() *mocks.AdminUserRepository
		userRepository            func() *mocks.UserRepository
		depositRequestRepository  func() *mocks.DepositRequestRepository
		depositRequestService     func() *mocks.DepositRequestService
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		mailService               func() *mocks.MailService
		eventMessageService       func() *mocks.EventMessageService
		wantErr                   bool
	}{
		{
			name: "投資完了のビジネス型プロジェクトをキャンセルし、プロジェクトも更新する",
			args: args{
				ctx:      context.Background(),
				sharedId: mockShare1Id,
				userId:   mockActor1Id,
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := &mocks.ShareRepository{}
				manualShare := mockShare1
				manualShare.ShareType = domain.MANUAL_SHARE
				manualShare.InvestStatus = domain.SUCCESS_INVEST

				repo.On("GetById", mock.Anything, mockShare1Id).Return(manualShare, nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(share domain.Share) bool {
					return share.InvestStatus == domain.CANCEL_INVEST &&
						share.UpdateId == mockActor1Id
				})).Return(nil)
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("GetById", mock.Anything, mockProject1Id).Return(mockProject1, nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(project domain.Project) bool {
					return project.IsManualInvest == false
				})).Return(nil)
				return repo
			},

			depositRequestRepository: func() *mocks.DepositRequestRepository {
				repo := &mocks.DepositRequestRepository{}
				repo.On("GetDepositRequestByShareId", mock.Anything, mockShare1Id).
					Return(mockDepositRequest1, nil)

				repo.On("GetAll", mock.Anything, mock.MatchedBy(func(field func(fields *cdomain.DepositRequestQueryOptionFields)) bool {
					opt := cdomain.DepositRequestQueryOptionFunc()
					field(&opt)

					return *opt.UserId() == fmt.Sprintf("%v", mockUser1Id)
				})).Return(
					[]domain.DepositRequest{
						mockDepositRequest1,
						{
							Id: 3001,
						},
						{
							Id: 3002,
						},
					}, nil)

				return repo
			},

			adminUserRepository: func() *mocks.AdminUserRepository {
				repo := &mocks.AdminUserRepository{}
				repo.On("GetById", mock.Anything, mockActor1Id).Return(mockActor1, nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", mock.Anything, mockUser1Id).Return(mockUser1, nil)
				return repo
			},

			depositRequestService: func() *mocks.DepositRequestService {
				service := &mocks.DepositRequestService{}
				service.On("CancelDeposit", mock.Anything, mockDepositRequest1).
					Return(nil)
				return service
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := &mocks.OperatorHistoryRepository{}
				repo.On("Create", mock.Anything, domain.OperatorHistory{
					UserId: mockActor1Id,
					Email:  mockActor1.Email,
					Menu:   "commoncode.INVEST_STATUS.CANCEL_INVEST",
					Ip:     "::1",
					State:  domain.CHANGE,
				}).Return(nil)
				return repo
			},
			mailService: func() *mocks.MailService {
				service := &mocks.MailService{}
				service.On("SendMail", mock.Anything, mock.MatchedBy(func(request cdomain.SendMailRequest) bool {
					ignoreFields := cmpopts.IgnoreFields(cdomain.SendMailRequest{}, "Details")

					return cmp.Equal(request, cdomain.SendMailRequest{
						TemplateType: domain.INVEST_CANCEL_TEMPLATE,
						Sender:       "info@energyshares.jp",
						Receivers:    []string{mockUser1.Email},
						MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
					}, ignoreFields)
				})).Return(nil)
				return service
			},

			eventMessageService: func() *mocks.EventMessageService {
				service := &mocks.EventMessageService{}
				service.On("PublishEvent", mock.Anything, domain.EVENT_INVESTMENT_CANCELED, mock.Anything).Return(nil)
				return service
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := shareService{
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
				shareRepository:           tt.shareRepository(),
				adminUserRepository:       tt.adminUserRepository(),
				userRepository:            tt.userRepository(),
				projectRepository:         tt.projectRepository(),
				depositRequestRepository:  tt.depositRequestRepository(),
				depositRequestService:     tt.depositRequestService(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
				mailService:               tt.mailService(),
				eventMessageService:       tt.eventMessageService(),
			}
			err := service.CancelInvest(tt.args.ctx, tt.args.sharedId, tt.args.userId)
			assert.Equalf(t, tt.wantErr, err != nil, "shareService.CancelInvest() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_shareService_CancelInvest_NotifiesAfterCommit(t *testing.T) {
	share := mockShare1
	share.InvestStatus = domain.REQUEST_INVEST
	project := mockProject1
	depositRequest := mockDepositRequest1
	depositRequest.DepositCheckStatus = domain.NO_DEPOSIT

	shareRepository := new(mocks.ShareRepository)
	shareRepository.On("GetById", mock.Anything, share.Id).Return(share, nil)
	shareRepository.On("Update", mock.Anything, mock.MatchedBy(func(updated domain.Share) bool {
		return updated.Id == share.Id && updated.InvestStatus == domain.CANCEL_INVEST
	})).Return(nil)
	projectRepository := new(mocks.ProjectRepository)
	projectRepository.On("GetById", mock.Anything, project.Id).Return(project, nil)
	depositRequestRepository := new(mocks.DepositRequestRepository)
	depositRequestRepository.On("GetDepositRequestByShareId", mock.Anything, share.Id).Return(depositRequest, nil)
	depositRequestService := new(mocks.DepositRequestService)
	depositRequestService.On("CancelDeposit", mock.Anything, depositRequest).Return(nil)
	adminUserRepository := new(mocks.AdminUserRepository)
	adminUserRepository.On("GetById", mock.Anything, mockActor1.Id).Return(mockActor1, nil)
	operatorHistoryRepository := new(mocks.OperatorHistoryRepository)
	operatorHistoryRepository.On("Create", mock.Anything, mock.AnythingOfType("domain.OperatorHistory")).Return(nil)
	userRepository := new(mocks.UserRepository)
	userRepository.On("GetById", mock.Anything, share.OwnerId).Return(mockUser1, nil)
	mailService := new(mocks.MailService)
	mailService.On("SendMail", mock.Anything, mock.AnythingOfType("clean_arch_domain.SendMailRequest")).Return(nil)
	eventMessageService := new(mocks.EventMessageService)
	eventMessageService.On("PublishEvent", mock.Anything, domain.EVENT_INVESTMENT_VOIDED, mock.Anything).Return(nil)

	service := shareService{
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
		shareRepository:           shareRepository,
		adminUserRepository:       adminUserRepository,
		userRepository:            userRepository,
		projectRepository:         projectRepository,
		operatorHistoryRepository: operatorHistoryRepository,
		depositRequestRepository:  depositRequestRepository,
		depositRequestService:     depositRequestService,
		eventMessageService:       eventMessageService,
		mailService:               mailService,
	}

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	err = factory.NewTxManager(engine).WithTx(context.Background(), func(txCtx context.Context) error {
		if err := service.CancelInvest(txCtx, share.Id, mockActor1.Id); err != nil {
			return err
		}
		userRepository.AssertNotCalled(t, "GetById", mock.Anything, mock.Anything)
		mailService.AssertNotCalled(t, "SendMail", mock.Anything, mock.Anything)
		eventMessageService.AssertNotCalled(t, "PublishEvent", mock.Anything, mock.Anything, mock.Anything)
		return nil
	})

	require.NoError(t, err)
	userRepository.AssertExpectations(t)
	mailService.AssertExpectations(t)
	eventMessageService.AssertExpectations(t)
}

func Test_shareService_CancelInvest_InvalidStatus(t *testing.T) {
	tests := []struct {
		name         string
		investStatus domain.INVEST_STATUS
	}{
		{name: "投資不成立はキャンセルできない", investStatus: domain.FAIL_INVEST},
		{name: "キャンセル済みは再度キャンセルできない", investStatus: domain.CANCEL_INVEST},
		{name: "運用中はキャンセルできない", investStatus: domain.OPERATION_INVEST},
		{name: "運用終了はキャンセルできない", investStatus: domain.REFUND_INVEST},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shareRepository := new(mocks.ShareRepository)
			shareRepository.On("GetById", mock.Anything, mockShare1Id).Return(domain.Share{
				Id:           mockShare1Id,
				InvestStatus: tt.investStatus,
			}, nil)

			service := shareService{shareRepository: shareRepository}

			err := service.CancelInvest(context.Background(), mockShare1Id, mockActor1Id)

			assert.ErrorIs(t, err, cdomain.ErrBadRequest)
			shareRepository.AssertExpectations(t)
		})
	}
}

func Test_shareService_FailInvest(t *testing.T) {
	type args struct {
		ctx      context.Context
		sharedId string
		userId   int64
	}
	tests := []struct {
		name                      string
		args                      args
		shareRepository           func() *mocks.ShareRepository
		adminUserRepository       func() *mocks.AdminUserRepository
		userRepository            func() *mocks.UserRepository
		projectRepository         func() *mocks.ProjectRepository
		depositRequestRepository  func() *mocks.DepositRequestRepository
		depositRequestService     func() *mocks.DepositRequestService
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		mailService               func() *mocks.MailService
		eventMessageService       func() *mocks.EventMessageService
		wantErr                   bool
	}{
		{
			name: "投資が失敗し、投資状態が失敗に変更されます。 => 正常に実行されることを期待します。",
			args: args{
				ctx:      context.Background(),
				sharedId: mockShare1Id,
				userId:   mockActor1Id,
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := &mocks.ShareRepository{}
				repo.On("GetById", mock.Anything, mockShare1Id).Return(mockShare1, nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(share domain.Share) bool {
					return share.InvestStatus == domain.FAIL_INVEST &&
						share.UpdateId == mockActor1Id
				})).Return(nil)
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("GetById", mock.Anything, mockProject1Id).Return(mockProject1, nil)
				return repo
			},
			depositRequestRepository: func() *mocks.DepositRequestRepository {
				repo := &mocks.DepositRequestRepository{}
				repo.On("GetDepositRequestByShareId", mock.Anything, mockShare1Id).
					Return(mockDepositRequest1, nil)
				repo.On("GetAll", mock.Anything, mock.MatchedBy(func(field func(fields *cdomain.DepositRequestQueryOptionFields)) bool {
					opt := cdomain.DepositRequestQueryOptionFunc()
					field(&opt)

					return *opt.UserId() == fmt.Sprintf("%v", mockUser1Id)
				})).Return(
					[]domain.DepositRequest{
						mockDepositRequest1,
						{
							Id: 3001,
						},
						{
							Id: 3002,
						},
					}, nil)

				return repo
			},
			depositRequestService: func() *mocks.DepositRequestService {
				service := &mocks.DepositRequestService{}
				service.On("CancelDeposit", mock.Anything, mockDepositRequest1).Return(nil)
				return service
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := &mocks.OperatorHistoryRepository{}
				repo.On("Create", mock.Anything, domain.OperatorHistory{
					UserId: mockActor1Id,
					Email:  mockActor1.Email,
					Menu:   "commoncode.INVEST_STATUS.FAIL_INVEST",
					Ip:     "::1",
					State:  domain.CHANGE,
				}).Return(nil)
				return repo
			},
			adminUserRepository: func() *mocks.AdminUserRepository {
				repo := &mocks.AdminUserRepository{}
				repo.On("GetById", mock.Anything, mockActor1Id).Return(mockActor1, nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", mock.Anything, mockUser1Id).Return(mockUser1, nil)
				return repo
			},
			eventMessageService: func() *mocks.EventMessageService {
				service := &mocks.EventMessageService{}
				service.On("PublishEvent", mock.Anything, domain.EVENT_INVESTMENT_CANCELED, mock.Anything).Return(nil)
				return service
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := shareService{
				shareRepository:           tt.shareRepository(),
				adminUserRepository:       tt.adminUserRepository(),
				userRepository:            tt.userRepository(),
				projectRepository:         tt.projectRepository(),
				depositRequestRepository:  tt.depositRequestRepository(),
				depositRequestService:     tt.depositRequestService(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
				eventMessageService:       tt.eventMessageService(),
			}
			err := service.FailInvest(tt.args.ctx, tt.args.sharedId, tt.args.userId)
			assert.Equalf(t, tt.wantErr, err != nil, "shareService.FailInvest() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
