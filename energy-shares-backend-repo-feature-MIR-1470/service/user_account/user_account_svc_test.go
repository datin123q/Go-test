package user_account

import (
	"context"
	"testing"

	"github.com/samber/lo"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

func Test_userAccountService_DepositCheckProcess(t *testing.T) {
	type args struct {
		ctx            context.Context
		depositRequest domain.DepositRequest
	}

	tests := []struct {
		name                                string
		args                                args
		userAccountRepository               func() *mocks.UserAccountRepository
		userRepository                      func() *mocks.UserRepository
		depositRequestRepository            func() *mocks.DepositRequestRepository
		projectRepository                   func() *mocks.ProjectRepository
		shareRepository                     func() *mocks.ShareRepository
		userVirtualAccountRepository        func() *mocks.UserVirtualAccountRepository
		depositRequestLackHistoryRepository func() *mocks.DepositRequestLackHistoryRepository
		depositRequestService               func() *mocks.DepositRequestService
		shareService                        func() *mocks.ShareService
		mailService                         func() *mocks.MailService
		userTransactionRepository           func() *mocks.UserTransactionRepository
		wantErr                             bool
	}{
		{
			name: "잔액이 충분할 때 수행이 정상적으로 됩니다. (잔액 : 100000)",
			args: args{
				ctx: context.Background(),
				depositRequest: domain.DepositRequest{
					Id:                   1000,
					UserId:               2000,
					ShareId:              "SHARE_ID1",
					ProjectId:            3000,
					ProjectType:          domain.EQUITY,
					RequestAmount:        30000,
					DepositAmount:        10000,
					DepositCheckStatus:   domain.NO_DEPOSIT,
					DepositAccountNumber: "123-123-123",
				},
			},
			userAccountRepository: func() *mocks.UserAccountRepository {
				repo := new(mocks.UserAccountRepository)
				repo.On("GetByUserId", context.Background(), mockUserId1).Return(domain.UserAccount{
					UserId:  mockUserId1,
					Balance: lo.ToPtr(int64(100000)),
				}, nil)

				repo.On("Update", context.Background(), domain.UserAccount{
					UserId:  mockUserId1,
					Balance: lo.ToPtr(int64(80000)),
				}).Return(nil)
				return repo
			},

			userRepository: func() *mocks.UserRepository {
				repo := new(mocks.UserRepository)
				repo.On("GetById", context.Background(), mockUserId1).Return(mockUser1, nil)
				return repo
			},

			depositRequestService: func() *mocks.DepositRequestService {
				svc := new(mocks.DepositRequestService)
				svc.On("OkDeposit", context.Background(), mockDepositRequest1).Return(nil)
				return svc
			},

			shareService: func() *mocks.ShareService {
				svc := new(mocks.ShareService)
				svc.On("SuccessInvest", context.Background(), mockShareId1, int64(1)).Return(nil)
				return svc
			},
			userTransactionRepository: func() *mocks.UserTransactionRepository { return new(mocks.UserTransactionRepository) },
			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				return repo
			},

			depositRequestRepository: func() *mocks.DepositRequestRepository {
				repo := new(mocks.DepositRequestRepository)
				return repo
			},

			shareRepository: func() *mocks.ShareRepository {
				repo := new(mocks.ShareRepository)
				return repo
			},

			userVirtualAccountRepository: func() *mocks.UserVirtualAccountRepository {
				repo := new(mocks.UserVirtualAccountRepository)
				return repo
			},
			mailService: func() *mocks.MailService {
				svc := new(mocks.MailService)
				return svc
			},

			depositRequestLackHistoryRepository: func() *mocks.DepositRequestLackHistoryRepository {
				repo := new(mocks.DepositRequestLackHistoryRepository)
				return repo
			},
		},
		{
			// MIR-1247: userAccountRepository.Updateは影響行数を検査して
			// version不一致(または行が存在しない)ならerrorを返すように
			// なった。ここではUpdateがエラーを返す場合、SuccessInvest
			// (投資成立)が既に確定した後でも、DepositCheckProcess自体は
			// エラーとして呼び出し元へ伝播することを固定する。以前は
			// Updateが常にnilを返していたため、この経路はwantErr=falseの
			// まま素通りしていた。
			name: "잔액 갱신이 버전 불일치로 실패하면 에러를 반환합니다.",
			args: args{
				ctx: context.Background(),
				depositRequest: domain.DepositRequest{
					Id:                   1000,
					UserId:               2000,
					ShareId:              "SHARE_ID1",
					ProjectId:            3000,
					ProjectType:          domain.EQUITY,
					RequestAmount:        30000,
					DepositAmount:        10000,
					DepositCheckStatus:   domain.NO_DEPOSIT,
					DepositAccountNumber: "123-123-123",
				},
			},
			userAccountRepository: func() *mocks.UserAccountRepository {
				repo := new(mocks.UserAccountRepository)
				repo.On("GetByUserId", context.Background(), mockUserId1).Return(domain.UserAccount{
					UserId:  mockUserId1,
					Balance: lo.ToPtr(int64(100000)),
				}, nil)

				repo.On("Update", context.Background(), domain.UserAccount{
					UserId:  mockUserId1,
					Balance: lo.ToPtr(int64(80000)),
				}).Return(cdomain.ErrItemNotAffected)
				return repo
			},

			userRepository: func() *mocks.UserRepository {
				repo := new(mocks.UserRepository)
				repo.On("GetById", context.Background(), mockUserId1).Return(mockUser1, nil)
				return repo
			},

			depositRequestService: func() *mocks.DepositRequestService {
				svc := new(mocks.DepositRequestService)
				svc.On("OkDeposit", context.Background(), mockDepositRequest1).Return(nil)
				return svc
			},

			shareService: func() *mocks.ShareService {
				svc := new(mocks.ShareService)
				svc.On("SuccessInvest", context.Background(), mockShareId1, int64(1)).Return(nil)
				return svc
			},
			userTransactionRepository: func() *mocks.UserTransactionRepository { return new(mocks.UserTransactionRepository) },
			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				return repo
			},

			depositRequestRepository: func() *mocks.DepositRequestRepository {
				repo := new(mocks.DepositRequestRepository)
				return repo
			},

			shareRepository: func() *mocks.ShareRepository {
				repo := new(mocks.ShareRepository)
				return repo
			},

			userVirtualAccountRepository: func() *mocks.UserVirtualAccountRepository {
				repo := new(mocks.UserVirtualAccountRepository)
				return repo
			},
			mailService: func() *mocks.MailService {
				svc := new(mocks.MailService)
				return svc
			},

			depositRequestLackHistoryRepository: func() *mocks.DepositRequestLackHistoryRepository {
				repo := new(mocks.DepositRequestLackHistoryRepository)
				return repo
			},
			wantErr: true,
		},
		{
			name: "잔액이 부족할 때 수행이 정상적으로 됩니다. (잔액 : 10000) => 잔액부족분 메일전송",
			args: args{
				ctx: context.Background(),
				depositRequest: domain.DepositRequest{
					Id:                   1000,
					UserId:               2000,
					ShareId:              "SHARE_ID1",
					ProjectId:            3000,
					ProjectType:          domain.EQUITY,
					RequestAmount:        30000,
					DepositAmount:        10000,
					DepositCheckStatus:   domain.NO_DEPOSIT,
					DepositAccountNumber: "123-123-123",
				},
			},
			userAccountRepository: func() *mocks.UserAccountRepository {
				repo := new(mocks.UserAccountRepository)
				repo.On("GetByUserId", context.Background(), mockUserId1).Return(domain.UserAccount{
					UserId:  mockUserId1,
					Balance: lo.ToPtr(int64(10000)),
				}, nil)

				repo.On("Update", context.Background(), domain.UserAccount{
					UserId:  mockUserId1,
					Balance: lo.ToPtr(int64(0)),
				}).Return(nil)
				return repo
			},

			userRepository: func() *mocks.UserRepository {
				repo := new(mocks.UserRepository)
				repo.On("GetById", context.Background(), mockUserId1).Return(mockUser1, nil)
				return repo
			},

			depositRequestRepository: func() *mocks.DepositRequestRepository {
				repo := new(mocks.DepositRequestRepository)
				repo.On("Update", context.Background(), mock.MatchedBy(func(depositRequest domain.DepositRequest) bool {
					return depositRequest.DepositAmount == 20000 && depositRequest.DepositCheckStatus == domain.NO_DEPOSIT
				})).Return(nil)
				return repo
			},

			depositRequestService: func() *mocks.DepositRequestService {
				svc := new(mocks.DepositRequestService)
				return svc
			},

			shareService: func() *mocks.ShareService {
				svc := new(mocks.ShareService)
				svc.On("SuccessInvest", context.Background(), mockShareId1, int64(0)).Return(nil)
				return svc
			},
			userTransactionRepository: func() *mocks.UserTransactionRepository { return new(mocks.UserTransactionRepository) },

			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				repo.On("GetById", context.Background(), mockProjectId1).Return(mockProject1, nil)
				return repo
			},

			shareRepository: func() *mocks.ShareRepository {
				repo := new(mocks.ShareRepository)
				repo.On("GetById", context.Background(), mockShareId1).Return(mockShare1, nil)
				return repo
			},

			userVirtualAccountRepository: func() *mocks.UserVirtualAccountRepository {
				repo := new(mocks.UserVirtualAccountRepository)
				repo.On("GetByUserIdAndLinkedStatus", context.Background(), mockUserId1, mockProject1.ProjectType, domain.ON_VA_ACCOUNT).Return(mockUserVirtualAccount1, nil)
				return repo
			},

			mailService: func() *mocks.MailService {
				svc := &mocks.MailService{}
				svc.On("SendMail", mock.Anything, cdomain.SendMailRequest{
					TemplateType: domain.INVEST_INSUFFICIENT_BALANCE_TEMPLATE,
					Sender:       "info@energyshares.jp",
					Receivers:    []string{mockUser1.Email},
					MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
					Details: map[string]interface{}{
						"Receiver":         mockUser1.Name,
						"ServiceUrl":       "http://energyshares.jp",
						"LackAmount":       "10,000",
						"ProductName":      mockProject1.Name,
						"TransactionDate":  jsonutils.JSONTime.ToString(mockShare1.InvestDate),
						"BankName":         mockDepositRequest1.DepositBankName,
						"BranchName":       mockDepositRequest1.DepositBranchName,
						"AccountNumber":    mockDepositRequest1.DepositAccountNumber,
						"Depositor":        mockUserVirtualAccount1.VaHolderNameKana,
						"InvestmentAmount": "30,000",
						"PaymentEnd":       jsonutils.JSONTime.ToString(mockDepositRequest1.ExpireDepositDate),
					},
				}).Return(nil)
				return svc
			},
			depositRequestLackHistoryRepository: func() *mocks.DepositRequestLackHistoryRepository {
				repo := new(mocks.DepositRequestLackHistoryRepository)
				repo.On("Create", context.Background(), mock.MatchedBy(func(depositRequestLackHistory domain.DepositRequestLackHistory) bool {
					ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

					return cmp.Equal(domain.DepositRequestLackHistory{
						UserId:           mockUserId1,
						DepositRequestId: mockDepositRequestId1,
						Amount:           10000,
						AccountNumber:    mockDepositRequest1.DepositAccountNumber,
						ProjectType:      mockProject1.ProjectType,
						ProjectId:        mockProjectId1,
						ShareId:          mockShareId1,
					}, depositRequestLackHistory, ignoreTypes)

				})).Return(nil)
				return repo
			},
		},
		{
			name: "잔액이 부족할 때 수행이 정상적으로 됩니다. (잔액 : 0) => 잔액부족분 메일전송, 유저트랜잭션은 발생하지 않음",
			args: args{
				ctx: context.Background(),
				depositRequest: domain.DepositRequest{
					Id:                   1000,
					UserId:               2000,
					ShareId:              "SHARE_ID1",
					ProjectId:            3000,
					ProjectType:          domain.EQUITY,
					RequestAmount:        30000,
					DepositAmount:        10000,
					DepositCheckStatus:   domain.NO_DEPOSIT,
					DepositAccountNumber: "123-123-123",
				},
			},
			userAccountRepository: func() *mocks.UserAccountRepository {
				repo := new(mocks.UserAccountRepository)
				repo.On("GetByUserId", context.Background(), mockUserId1).Return(domain.UserAccount{
					UserId:  mockUserId1,
					Balance: lo.ToPtr(int64(0)),
				}, nil)

				repo.On("Update", context.Background(), domain.UserAccount{
					UserId:  mockUserId1,
					Balance: lo.ToPtr(int64(0)),
				}).Return(nil)
				return repo
			},

			userRepository: func() *mocks.UserRepository {
				repo := new(mocks.UserRepository)
				repo.On("GetById", context.Background(), mockUserId1).Return(mockUser1, nil)
				return repo
			},

			depositRequestRepository: func() *mocks.DepositRequestRepository {
				repo := new(mocks.DepositRequestRepository)
				repo.On("Update", context.Background(), mock.MatchedBy(func(depositRequest domain.DepositRequest) bool {
					return depositRequest.DepositAmount == 10000 && depositRequest.DepositCheckStatus == domain.NO_DEPOSIT
				})).Return(nil)
				return repo
			},

			depositRequestService: func() *mocks.DepositRequestService {
				svc := new(mocks.DepositRequestService)
				return svc
			},

			shareService: func() *mocks.ShareService {
				svc := new(mocks.ShareService)
				return svc
			},
			userTransactionRepository: func() *mocks.UserTransactionRepository {
				repo := new(mocks.UserTransactionRepository)
				return repo
			},

			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				repo.On("GetById", context.Background(), mockProjectId1).Return(mockProject1, nil)
				return repo
			},

			shareRepository: func() *mocks.ShareRepository {
				repo := new(mocks.ShareRepository)
				repo.On("GetById", context.Background(), mockShareId1).Return(mockShare1, nil)
				return repo
			},

			userVirtualAccountRepository: func() *mocks.UserVirtualAccountRepository {
				repo := new(mocks.UserVirtualAccountRepository)
				repo.On("GetByUserIdAndLinkedStatus", context.Background(), mockUserId1, mockProject1.ProjectType, domain.ON_VA_ACCOUNT).Return(mockUserVirtualAccount1, nil)
				return repo
			},

			mailService: func() *mocks.MailService {
				svc := &mocks.MailService{}
				svc.On("SendMail", mock.Anything, cdomain.SendMailRequest{
					TemplateType: domain.INVEST_INSUFFICIENT_BALANCE_TEMPLATE,
					Sender:       "info@energyshares.jp",
					Receivers:    []string{mockUser1.Email},
					MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
					Details: map[string]interface{}{
						"Receiver":         mockUser1.Name,
						"ServiceUrl":       "http://energyshares.jp",
						"LackAmount":       "20,000",
						"ProductName":      mockProject1.Name,
						"TransactionDate":  jsonutils.JSONTime.ToString(mockShare1.InvestDate),
						"BankName":         mockDepositRequest1.DepositBankName,
						"BranchName":       mockDepositRequest1.DepositBranchName,
						"AccountNumber":    mockDepositRequest1.DepositAccountNumber,
						"Depositor":        mockUserVirtualAccount1.VaHolderNameKana,
						"InvestmentAmount": "30,000",
						"PaymentEnd":       jsonutils.JSONTime.ToString(mockDepositRequest1.ExpireDepositDate),
					},
				}).Return(nil)
				return svc
			},
			depositRequestLackHistoryRepository: func() *mocks.DepositRequestLackHistoryRepository {
				repo := new(mocks.DepositRequestLackHistoryRepository)
				repo.On("Create", context.Background(), mock.MatchedBy(func(depositRequestLackHistory domain.DepositRequestLackHistory) bool {
					ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

					return cmp.Equal(domain.DepositRequestLackHistory{
						UserId:           mockUserId1,
						DepositRequestId: mockDepositRequestId1,
						Amount:           20000,
						AccountNumber:    mockDepositRequest1.DepositAccountNumber,
						ProjectType:      mockProject1.ProjectType,
						ProjectId:        mockProjectId1,
						ShareId:          mockShareId1,
					}, depositRequestLackHistory, ignoreTypes)

				})).Return(nil)
				return repo
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := userAccountService{
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
								UserServiceDomain: "http://energyshares.jp",
							},
						},
					},
				},
				logger:                              factory.NewLoggerFactory(),
				userAccountRepository:               tt.userAccountRepository(),
				userRepository:                      tt.userRepository(),
				depositRequestRepository:            tt.depositRequestRepository(),
				projectRepository:                   tt.projectRepository(),
				shareRepository:                     tt.shareRepository(),
				userVirtualAccountRepository:        tt.userVirtualAccountRepository(),
				depositRequestService:               tt.depositRequestService(),
				shareService:                        tt.shareService(),
				userTransactionRepository:           tt.userTransactionRepository(),
				depositRequestLackHistoryRepository: tt.depositRequestLackHistoryRepository(),
				mailService:                         tt.mailService(),
			}
			err := svc.DepositCheckProcess(tt.args.ctx, tt.args.depositRequest)
			assert.Equalf(t, tt.wantErr, err != nil, "DepositCheckProcess() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
