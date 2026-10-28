package investment

import (
	"context"
	"testing"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"

	"github.com/samber/lo"

	"github.com/stretchr/testify/mock"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var testAssetPath = "../../assets"

func Test_investmentService_GetDetail(t *testing.T) {
	// 事前に秘密鍵を設定しておく
	serviceutils.SetEnv("local")
	serviceutils.SetPrivateKeyPath(testAssetPath + "/key/local/private_key.pem")
	serviceutils.SetPrivateKey("")
	if err := serviceutils.PreloadPrivateKey("local", ""); err != nil {
		t.Fatalf("failed to preload private key: %v", err)
	}

	type args struct {
		ctx          context.Context
		investmentId string
	}
	tests := []struct {
		name                         string
		args                         args
		s3pkg                        s3utils.S3
		investmentRepository         func() *mocks.InvestmentRepository
		userRepository               func() *mocks.UserRepository
		contractRepository           func() *mocks.ContractRepository
		userVirtualAccountRepository func() *mocks.UserVirtualAccountRepository
		depositRequestRepository     func() *mocks.DepositRequestRepository
		repaymentDividendSchedule    func() *mocks.RepaymentDividendScheduleRepository
		want                         domain.InvestmentDetail
		wantErr                      bool
	}{
		{
			name: "정상적으로 투자 상세정보를 조회합니다.",
			args: args{
				ctx:          context.Background(),
				investmentId: "SHARE-1",
			},
			investmentRepository: func() *mocks.InvestmentRepository {
				investmentRepository := &mocks.InvestmentRepository{}
				investmentRepository.On("GetById", context.Background(), "SHARE-1").Return(mockInvestment1, nil)
				return investmentRepository
			},

			userRepository: func() *mocks.UserRepository {
				userRepository := &mocks.UserRepository{}
				userRepository.On("GetById", context.Background(), int64(1000)).Return(mockUser1, nil)
				return userRepository
			},

			contractRepository: func() *mocks.ContractRepository {
				contractRepository := &mocks.ContractRepository{}
				contractRepository.On("GetAll", context.Background(), mock.MatchedBy(func(field func(*cdomain.ContractQueryOptionFields)) bool {
					opt := cdomain.ContractQueryOptionFunc()
					field(&opt)
					return lo.FromPtr(opt.ShareId()) == "SHARE-1"
				})).Return(domain.Contracts{mockContract1}, nil)
				return contractRepository
			},

			userVirtualAccountRepository: func() *mocks.UserVirtualAccountRepository {
				userVirtualAccountRepository := &mocks.UserVirtualAccountRepository{}
				userVirtualAccountRepository.On("GetByUserIdAndLinkedStatus", context.Background(), int64(1000), domain.EQUITY, domain.ON_VA_ACCOUNT).Return(mockUserVirtualAccount1, nil)
				return userVirtualAccountRepository
			},
			depositRequestRepository: func() *mocks.DepositRequestRepository {
				depositRequestRepository := &mocks.DepositRequestRepository{}
				depositRequestRepository.On("GetDepositRequestByShareId", context.Background(), "SHARE-1").Return(mockDepositRequest1, nil)
				return depositRequestRepository
			},
			repaymentDividendSchedule: func() *mocks.RepaymentDividendScheduleRepository {
				repaymentDividendScheduleRepository := &mocks.RepaymentDividendScheduleRepository{}
				repaymentDividendScheduleRepository.On("GetAll", context.Background(),
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(int64(1), domain.RepaymentDividendSchedules{mockRepaymentDividendSchedule1}, nil)
				return repaymentDividendScheduleRepository
			},

			want: domain.InvestmentDetail{
				Investment:        mockInvestment1,
				DepositRequest:    mockDepositRequest1,
				VaAccount:         mockUserVirtualAccount1,
				LastRepaymentDate: lo.ToPtr(jsonutils.JSONTime(time.Unix(0, 0)).ToString()),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := investmentService{
				s3pkg: s3utils.New(&config.Config{
					CommonConfig: config.CommonConfig{
						AWS: struct {
							Region                  string
							SES_Region              string
							SNS_Region              string
							CDN                     string
							S3_PRIVATE_BUCKET       string
							S3_PUBLIC_BUCKET        string
							S3_EKYC_IMAGES_BUCKET   string
							SES_ARN                 string
							SNS_ARN                 string
							PUBLIC_KEY_ID           string
							PRIVATE_KEY_PATH        string
							BASE_ENDPOINT           string
							INVEST_WORKER_QUEUE_URL string
						}{
							PUBLIC_KEY_ID:    "public-key-id",
							PRIVATE_KEY_PATH: testAssetPath + "/key/local/private_key.pem",
						},
					},
				}),
				investmentRepository:         tt.investmentRepository(),
				userRepository:               tt.userRepository(),
				contractRepository:           tt.contractRepository(),
				userVirtualAccountRepository: tt.userVirtualAccountRepository(),
				depositRequestRepository:     tt.depositRequestRepository(),
				repaymentDividendSchedule:    tt.repaymentDividendSchedule(),
			}

			got, err := service.GetDetail(tt.args.ctx, tt.args.investmentId)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			ignoreFields := cmpopts.IgnoreFields(domain.InvestmentDetail{}, "ContractAfterPageUrl")

			assert.Equalf(t, tt.wantErr, err != nil, "GetDetail() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(got, tt.want, ignoreTime, ignoreFields), "GetDetail() = %v, want %v", got, tt.want)
		})
	}
}

func Test_investmentService_ExportCSV(t *testing.T) {
	type args struct {
		items []domain.Investment
	}
	tests := []struct {
		name string
		args args
		want cdomain.CSVData
	}{
		{
			name: "CSV 형태로 데이터가 반환되는지 확인 => 번역이 제대로 동작하고 CSV 형태로 1건 반환됨",
			args: args{
				items: []domain.Investment{
					{
						Share: domain.Share{
							Id:           "SHARE-1",
							InvestStatus: domain.SUCCESS_INVEST,
							ProjectType:  domain.EQUITY,
							Qty:          1000,
							Price:        500,
							InvestDate:   jsonutils.JSONTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
						},
						Project: domain.Project{
							Name: "낙동강댐 프로젝트",
						},
						UserName: "고길동",
						SpvName:  "수자원공사",
					},
				},
			},
			want: cdomain.CSVData{
				[]string{"取引番号", "ステータス", "投資家名", "組成会社", "タイプ", "ファンド名", "取引金額", "取引日時"},
				[]string{"SHARE-1", "投資完了", "고길동", "수자원공사", "事業型", "낙동강댐 프로젝트", "500000", "2024-01-01 09:00:00"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := investmentService{
				translator: i18nutils.New(testAssetPath),
			}
			got := svc.ExportCSV(tt.args.items)

			assert.Equalf(t, tt.want, got, "ExportCSV() = %v, want %v", got, tt.want)
		})
	}
}
