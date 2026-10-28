package investment

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	adomain "github.com/illuminarean/energy-share-core-api/domain"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"

	"github.com/illuminarean/energy-share-core-api/controllers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
)

func Test_investmentUsecase_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.InvestmentGetAllRequest
	}
	tests := []struct {
		name                      string
		args                      args
		investmentRepository      func() *mocks.InvestmentRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		want                      adomain.InvestmentGetAllResponse
		wantErr                   bool
	}{
		{
			name: "管理者アドミン投資リスト照会",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					SearchInput: controllers.SearchInput{
						SkipCount:      0,
						MaxResultCount: 10,
					},
				},
			},
			investmentRepository: func() *mocks.InvestmentRepository {
				repo := new(mocks.InvestmentRepository)
				repo.On("GetAll", context.Background(),
					mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything,
				).Return(int64(2), []domain.Investment{
					{
						Share: domain.Share{
							Id:           "SHARE-1",
							OwnerId:      1000,
							InvestStatus: domain.REQUEST_INVEST,
							ProjectType:  domain.EQUITY,
							Amount:       100_000,
						},
						Project: domain.Project{
							Name:  "プロジェクト1",
							SpvId: 2000,
						},
						SpvName:  "ドゥリ株式会社",
						UserName: "コ・ギルドン",
					},
					{
						Share: domain.Share{
							Id:           "SHARE-2",
							OwnerId:      1001,
							InvestStatus: domain.REQUEST_INVEST,
							ProjectType:  domain.EQUITY,
							Amount:       200_000,
						},
						Project: domain.Project{
							Name:  "プロジェクト2",
							SpvId: 2000,
						},
						SpvName:  "ドゥリ株式会社",
						UserName: "ドゥリ",
					},
				}, nil)
				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := new(mocks.OperatorHistoryRepository)
				repo.On("Create", context.Background(), domain.OperatorHistory{
					UserId: 1000,
					Email:  "gildong@email.com",
					Menu:   "menu.manage_accounting.payment_list",
					Ip:     "127.0.0.1",
					State:  domain.LOOKUP,
				}).Return(nil)
				return repo
			},
			want: adomain.InvestmentGetAllResponse{
				TotalCount: 2,
				Items: []adomain.InvestmentItemResponse{
					{
						Id:           "SHARE-1",
						ShareId:      "SHARE-1",
						SpvId:        2000,
						SpvName:      "ドゥリ株式会社",
						UserName:     "コ・ギルドン",
						InvestStatus: domain.REQUEST_INVEST,
						ProjectType:  domain.EQUITY,
						ProjectName:  "プロジェクト1",
						Amount:       100_000,
						OwnerId:      1000,
					},
					{
						Id:           "SHARE-2",
						ShareId:      "SHARE-2",
						SpvId:        2000,
						SpvName:      "ドゥリ株式会社",
						UserName:     "ドゥリ",
						InvestStatus: domain.REQUEST_INVEST,
						ProjectType:  domain.EQUITY,
						ProjectName:  "プロジェクト2",
						Amount:       200_000,
						OwnerId:      1001,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := investmentUsecase{
				investmentRepository:      tt.investmentRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}
			got, err := usecase.GetAll(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.wantErr, err != nil, "investmentUsecase.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "investmentUsecase.GetAll() = %v, want %v", got, tt.want)
		})
	}
}

func Test_investmentUsecase_GetOne(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.InvestmentGetOneRequest
	}
	tests := []struct {
		name                      string
		args                      args
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		investmentService         func() *mocks.InvestmentService
		want                      adomain.InvestmentGetOneResponse
		wantErr                   bool
	}{
		{
			name: "投資詳細照会：",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentGetOneRequest{
					InvestmentId: "SHARE-1",
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
				},
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := new(mocks.OperatorHistoryRepository)
				repo.On("Create", context.Background(), domain.OperatorHistory{
					UserId: 1000,
					Email:  "gildong@email.com",
					Menu:   "menu.manage_accounting.payment_list",
					Ip:     "127.0.0.1",
					State:  domain.LOOKUP,
				}).Return(nil)
				return repo
			},
			investmentService: func() *mocks.InvestmentService {
				service := new(mocks.InvestmentService)
				service.On("GetDetail", context.Background(), "SHARE-1").Return(domain.InvestmentDetail{
					Investment: domain.Investment{
						InvestmentId: "SHARE-1",
						Share: domain.Share{
							Id:                     "SHARE-1",
							ProjectReportCheckDate: jsonutils.JSONTime(time.Unix(100, 0)),
						},
						Project: domain.Project{
							ProjectReportDate: jsonutils.JSONTime(time.Unix(200, 0)),
						},
						UserType:       string(domain.PRIVATE),
						UserName:       "コ・ギルドン",
						FamilyName:     "コ",
						FirstName:      "ギルドン",
						FamilyNameKana: "コ",
						FirstNameKana:  "キルドン",
					},
				}, nil)
				return service
			},
			want: adomain.InvestmentGetOneResponse{
				InvestmentItemDetailResponse: adomain.InvestmentItemDetailResponse{
					Id:                     "SHARE-1",
					ShareId:                "SHARE-1",
					FirstName:              "ギルドン",
					FamilyName:             "コ",
					FirstNameKana:          "キルドン",
					FamilyNameKana:         "コ",
					Name:                   "コ ギルドン",
					NameKana:               "コ キルドン",
					ProjectReportCheckDate: jsonutils.JSONTime{},
					ContractStatus:         domain.CONTRACT_NONE,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := investmentUsecase{
				operatorHistoryRepository: tt.operatorHistoryRepository(),
				investmentService:         tt.investmentService(),
			}
			got, err := usecase.GetOne(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "investmentUsecase.GetOne() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "investmentUsecase.GetOne() = %v, want %v", got, tt.want)
		})
	}
}

func Test_investmentUsecase_Download(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.InvestmentDownloadHistoriesRequest
	}

	tests := []struct {
		name                 string
		args                 args
		investmentRepository func() *mocks.InvestmentRepository
		investmentService    func() *mocks.InvestmentService
		want                 adomain.InvestmentDownloadHistoriesResponse
		wantErr              bool
	}{
		{
			name: "投資リストをダウンロード：正常にダウンロードされます。",
			args: args{
				ctx:     context.Background(),
				request: adomain.InvestmentDownloadHistoriesRequest{},
			},

			investmentRepository: func() *mocks.InvestmentRepository {
				repo := new(mocks.InvestmentRepository)
				repo.On("GetAll", context.Background(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(int64(0), []domain.Investment{}, nil)

				return repo
			},
			investmentService: func() *mocks.InvestmentService {
				service := new(mocks.InvestmentService)
				service.On("ExportCSV", []domain.Investment{}).Return(
					cdomain.CSVData{
						[]string{"待機番号", "投資会員区分", "メールアドレス(ID)", "投資家名", "電話番号", "投資口数", "取引金額", "ステータス", "申請日時"},
						[]string{"1", "個人特定投資家", "gildong@email.com", "コ・ギルドン", "010-1234-1234", "1", "1", "投資待機中", "2024-01-01 00:00:00"},
						[]string{"2", "個人特定投資家", "dool2@email.com", "ドゥリ", "010-1234-1234", "2", "2", "投資待機中", "2024-01-01 00:00:00"},
					},
				)

				return service
			},
			want: adomain.InvestmentDownloadHistoriesResponse{
				Data: []byte{
					0xe5, 0xbe, 0x85, 0xe6, 0xa9, 0x9f, 0xe7, 0x95, 0xaa, 0xe5, 0x8f, 0xb7, 0x2c, 0xe6, 0x8a, 0x95,
					0xe8, 0xb3, 0x87, 0xe4, 0xbc, 0x9a, 0xe5, 0x93, 0xa1, 0xe5, 0x8c, 0xba, 0xe5, 0x88, 0x86, 0x2c,
					0xe3, 0x83, 0xa1, 0xe3, 0x83, 0xbc, 0xe3, 0x83, 0xab, 0xe3, 0x82, 0xa2, 0xe3, 0x83, 0x89, 0xe3,
					0x83, 0xac, 0xe3, 0x82, 0xb9, 0x28, 0x49, 0x44, 0x29, 0x2c, 0xe6, 0x8a, 0x95, 0xe8, 0xb3, 0x87,
					0xe5, 0xae, 0xb6, 0xe5, 0x90, 0x8d, 0x2c, 0xe9, 0x9b, 0xbb, 0xe8, 0xa9, 0xb1, 0xe7, 0x95, 0xaa,
					0xe5, 0x8f, 0xb7, 0x2c, 0xe6, 0x8a, 0x95, 0xe8, 0xb3, 0x87, 0xe5, 0x8f, 0xa3, 0xe6, 0x95, 0xb0,
					0x2c, 0xe5, 0x8f, 0x96, 0xe5, 0xbc, 0x95, 0xe9, 0x87, 0x91, 0xe9, 0xa1, 0x8d, 0x2c, 0xe3, 0x82,
					0xb9, 0xe3, 0x83, 0x86, 0xe3, 0x83, 0xbc, 0xe3, 0x82, 0xbf, 0xe3, 0x82, 0xb9, 0x2c, 0xe7, 0x94,
					0xb3, 0xe8, 0xab, 0x8b, 0xe6, 0x97, 0xa5, 0xe6, 0x99, 0x82, 0x0a, 0x31, 0x2c, 0xe5, 0x80, 0x8b,
					0xe4, 0xba, 0xba, 0xe7, 0x89, 0xb9, 0xe5, 0xae, 0x9a, 0xe6, 0x8a, 0x95, 0xe8, 0xb3, 0x87, 0xe5,
					0xae, 0xb6, 0x2c, 0x67, 0x69, 0x6c, 0x64, 0x6f, 0x6e, 0x67, 0x40, 0x65, 0x6d, 0x61, 0x69, 0x6c,
					0x2e, 0x63, 0x6f, 0x6d, 0x2c, 0xe3, 0x82, 0xb3, 0xe3, 0x83, 0xbb, 0xe3, 0x82, 0xae, 0xe3, 0x83,
					0xab, 0xe3, 0x83, 0x89, 0xe3, 0x83, 0xb3, 0x2c, 0x30, 0x31, 0x30, 0x2d, 0x31, 0x32, 0x33, 0x34,
					0x2d, 0x31, 0x32, 0x33, 0x34, 0x2c, 0x31, 0x2c, 0x31, 0x2c, 0xe6, 0x8a, 0x95, 0xe8, 0xb3, 0x87,
					0xe5, 0xbe, 0x85, 0xe6, 0xa9, 0x9f, 0xe4, 0xb8, 0xad, 0x2c, 0x32, 0x30, 0x32, 0x34, 0x2d, 0x30,
					0x31, 0x2d, 0x30, 0x31, 0x20, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x0a, 0x32, 0x2c,
					0xe5, 0x80, 0x8b, 0xe4, 0xba, 0xba, 0xe7, 0x89, 0xb9, 0xe5, 0xae, 0x9a, 0xe6, 0x8a, 0x95, 0xe8,
					0xb3, 0x87, 0xe5, 0xae, 0xb6, 0x2c, 0x64, 0x6f, 0x6f, 0x6c, 0x32, 0x40, 0x65, 0x6d, 0x61, 0x69,
					0x6c, 0x2e, 0x63, 0x6f, 0x6d, 0x2c, 0xe3, 0x83, 0x89, 0xe3, 0x82, 0xa5, 0xe3, 0x83, 0xaa, 0x2c,
					0x30, 0x31, 0x30, 0x2d, 0x31, 0x32, 0x33, 0x34, 0x2d, 0x31, 0x32, 0x33, 0x34, 0x2c, 0x32, 0x2c,
					0x32, 0x2c, 0xe6, 0x8a, 0x95, 0xe8, 0xb3, 0x87, 0xe5, 0xbe, 0x85, 0xe6, 0xa9, 0x9f, 0xe4, 0xb8,
					0xad, 0x2c, 0x32, 0x30, 0x32, 0x34, 0x2d, 0x30, 0x31, 0x2d, 0x30, 0x31, 0x20, 0x30, 0x30, 0x3a,
					0x30, 0x30, 0x3a, 0x30, 0x30, 0x0a,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := investmentUsecase{
				investmentRepository: tt.investmentRepository(),
				investmentService:    tt.investmentService(),
			}

			got, err := usecase.DownloadHistories(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "investmentUsecase.Download() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "investmentUsecase.Download() = %v, want %v", got, tt.want)
		})
	}
}

func Test_investmentUsecase_UpdateStatus(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.InvestmentUpdateStatusRequest
	}
	tests := []struct {
		name                               string
		args                               args
		shareRepository                    func() *mocks.ShareRepository
		shareService                       func() *mocks.ShareService
		investmentRepository               func() *mocks.InvestmentRepository
		transactionInfoHistoryRepository   func() *mocks.TransactionInfoHistoryRepository
		userAccountService                 func() *mocks.UserAccountService
		depositRequestRepository           func() *mocks.DepositRequestRepository
		projectRepository                  func() *mocks.ProjectRepository
		projectInvestmentExecuteRepository func() *mocks.ProjectInvestmentExecuteRepository
		want                               adomain.InvestmentUpdateStatusResponse
		wantErr                            bool
	}{
		{
			// MIR-1295: 手動での状態変更は申込中→取消・成立→取消の2つだけを
			// 許すため、それ以外の遷移先(ここではREQUEST_INVEST)は遷移元に
			// 関わらずすべて拒否される。旧仕様(正常に更新される)からの回帰。
			name: "MIR-1295: 成立から申込中への変更は拒否される",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentUpdateStatusRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Name:  "コ・ギルドン",
						Email: "gildong@email.com",
					},
					InvestmentId: "SHARE-1",
					InvestmentUpdateStatusBody: adomain.InvestmentUpdateStatusBody{
						InvestStatus: domain.REQUEST_INVEST,
					},
				},
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := new(mocks.ShareRepository)
				repo.On("GetById", mock.Anything, "SHARE-1").Return(domain.Share{Id: "SHARE-1", InvestStatus: domain.SUCCESS_INVEST}, nil)
				return repo
			},
			shareService: func() *mocks.ShareService {
				return new(mocks.ShareService)
			},
			investmentRepository: func() *mocks.InvestmentRepository {
				return new(mocks.InvestmentRepository)
			},
			transactionInfoHistoryRepository: func() *mocks.TransactionInfoHistoryRepository {
				return new(mocks.TransactionInfoHistoryRepository)
			},
			depositRequestRepository: func() *mocks.DepositRequestRepository {
				return new(mocks.DepositRequestRepository)
			},
			userAccountService: func() *mocks.UserAccountService {
				return new(mocks.UserAccountService)
			},
			projectRepository: func() *mocks.ProjectRepository {
				return new(mocks.ProjectRepository)
			},
			projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
				return new(mocks.ProjectInvestmentExecuteRepository)
			},
			want:    adomain.InvestmentUpdateStatusResponse{},
			wantErr: true,
		},
		{
			// MIR-1295: 申込中→運用終了は許可表に無いため拒否される。
			name: "MIR-1295: 申込中から運用終了への変更は拒否される",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentUpdateStatusRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Name:  "コ・ギルドン",
						Email: "gildong@email.com",
					},
					InvestmentId: "SHARE-1",
					InvestmentUpdateStatusBody: adomain.InvestmentUpdateStatusBody{
						InvestStatus: domain.REFUND_INVEST,
					},
				},
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := new(mocks.ShareRepository)
				repo.On("GetById", mock.Anything, "SHARE-1").Return(domain.Share{Id: "SHARE-1", InvestStatus: domain.REQUEST_INVEST}, nil)
				return repo
			},
			shareService: func() *mocks.ShareService {
				return new(mocks.ShareService)
			},
			investmentRepository: func() *mocks.InvestmentRepository {
				return new(mocks.InvestmentRepository)
			},
			transactionInfoHistoryRepository: func() *mocks.TransactionInfoHistoryRepository {
				return new(mocks.TransactionInfoHistoryRepository)
			},
			depositRequestRepository: func() *mocks.DepositRequestRepository {
				return new(mocks.DepositRequestRepository)
			},
			userAccountService: func() *mocks.UserAccountService {
				return new(mocks.UserAccountService)
			},
			projectRepository: func() *mocks.ProjectRepository {
				return new(mocks.ProjectRepository)
			},
			projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
				return new(mocks.ProjectInvestmentExecuteRepository)
			},
			want:    adomain.InvestmentUpdateStatusResponse{},
			wantErr: true,
		},
		{
			// MIR-1295: 申込中→運用中は許可表に無いため拒否される。
			name: "MIR-1295: 申込中から運用中への変更は拒否される",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentUpdateStatusRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Name:  "コ・ギルドン",
						Email: "gildong@email.com",
					},
					InvestmentId: "SHARE-1",
					InvestmentUpdateStatusBody: adomain.InvestmentUpdateStatusBody{
						InvestStatus: domain.OPERATION_INVEST,
					},
				},
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := new(mocks.ShareRepository)
				repo.On("GetById", mock.Anything, "SHARE-1").Return(domain.Share{Id: "SHARE-1", InvestStatus: domain.REQUEST_INVEST}, nil)
				return repo
			},
			shareService: func() *mocks.ShareService {
				return new(mocks.ShareService)
			},
			investmentRepository: func() *mocks.InvestmentRepository {
				return new(mocks.InvestmentRepository)
			},
			transactionInfoHistoryRepository: func() *mocks.TransactionInfoHistoryRepository {
				return new(mocks.TransactionInfoHistoryRepository)
			},
			depositRequestRepository: func() *mocks.DepositRequestRepository {
				return new(mocks.DepositRequestRepository)
			},
			userAccountService: func() *mocks.UserAccountService {
				return new(mocks.UserAccountService)
			},
			projectRepository: func() *mocks.ProjectRepository {
				return new(mocks.ProjectRepository)
			},
			projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
				return new(mocks.ProjectInvestmentExecuteRepository)
			},
			want:    adomain.InvestmentUpdateStatusResponse{},
			wantErr: true,
		},
		{
			// MIR-1295: 申込中→成立は許可表に無いため拒否される(チケットの
			// 「見つかった経路」そのもの。入金前に成立へ変えると、後続の
			// 取消で払っていない申込額が残高へ戻ってしまう)。
			name: "MIR-1295: 申込中から成立への変更は拒否される(回帰: MIR-1295で見つかった経路)",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentUpdateStatusRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Name:  "コ・ギルドン",
						Email: "gildong@email.com",
					},
					InvestmentId: "SHARE-1",
					InvestmentUpdateStatusBody: adomain.InvestmentUpdateStatusBody{
						InvestStatus: domain.SUCCESS_INVEST,
					},
				},
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := new(mocks.ShareRepository)
				repo.On("GetById", mock.Anything, "SHARE-1").Return(domain.Share{
					Id:           "SHARE-1",
					InvestStatus: domain.REQUEST_INVEST,
					OwnerId:      1000,
				}, nil)
				return repo
			},
			shareService: func() *mocks.ShareService {
				return new(mocks.ShareService)
			},
			investmentRepository: func() *mocks.InvestmentRepository {
				return new(mocks.InvestmentRepository)
			},
			transactionInfoHistoryRepository: func() *mocks.TransactionInfoHistoryRepository {
				return new(mocks.TransactionInfoHistoryRepository)
			},
			depositRequestRepository: func() *mocks.DepositRequestRepository {
				return new(mocks.DepositRequestRepository)
			},
			userAccountService: func() *mocks.UserAccountService {
				return new(mocks.UserAccountService)
			},
			projectRepository: func() *mocks.ProjectRepository {
				return new(mocks.ProjectRepository)
			},
			projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
				return new(mocks.ProjectInvestmentExecuteRepository)
			},
			want:    adomain.InvestmentUpdateStatusResponse{},
			wantErr: true,
		},
		{
			// MIR-1295: 成立→取消は出資実行の記録が無ければ許可する。
			// project→shareの順にロックしてから判定する新しい経路の検証。
			name: "MIR-1295: 成立から取消への変更は出資実行の記録が無ければ成功する",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentUpdateStatusRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Name:  "コ・ギルドン",
						Email: "gildong@email.com",
					},
					InvestmentId: "SHARE-1",
					InvestmentUpdateStatusBody: adomain.InvestmentUpdateStatusBody{
						InvestStatus: domain.CANCEL_INVEST,
					},
				},
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := new(mocks.ShareRepository)
				// project_id解決用(ロック無し)。判定用の値は使わない。
				repo.On("GetById", mock.Anything, "SHARE-1").Return(domain.Share{
					Id:           "SHARE-1",
					ProjectId:    2000,
					InvestStatus: domain.REQUEST_INVEST, // 判定には使われない値
				}, nil)
				repo.On("LockById", mock.Anything, "SHARE-1").Return(domain.Share{
					Id:           "SHARE-1",
					ProjectId:    2000,
					InvestStatus: domain.SUCCESS_INVEST,
					OwnerId:      1000,
				}, nil)
				return repo
			},
			shareService: func() *mocks.ShareService {
				service := new(mocks.ShareService)
				service.On("CancelInvest", mock.Anything, "SHARE-1", int64(1000)).Return(nil)
				return service
			},
			investmentRepository: func() *mocks.InvestmentRepository {
				repo := new(mocks.InvestmentRepository)
				repo.On("GetById", mock.Anything, "SHARE-1").Return(domain.Investment{InvestmentId: "SHARE-1"}, nil)
				return repo
			},

			transactionInfoHistoryRepository: func() *mocks.TransactionInfoHistoryRepository {
				repo := new(mocks.TransactionInfoHistoryRepository)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(transactionInfoHistory domain.TransactionInfoHistory) bool {
					ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

					return cmp.Equal(domain.TransactionInfoHistory{
						ShareId:       "SHARE-1",
						SubstanceType: string(domain.TRANSACTION_HISTORY_TRANSACTION),
						ProcessMethod: domain.USER_HISTORY_PROCESS_OPERATOR,
						ContentType:   string(domain.TRANSACTION_CHANGED_CONTENT_TRANSACTION_STATUS),
						PrevContent:   string(domain.SUCCESS_INVEST),
						CurrContent:   string(domain.CANCEL_INVEST),
						Email:         "gildong@email.com",
						Name:          "コ・ギルドン",
					}, transactionInfoHistory, ignoreTypes)

				})).Return(nil)

				return repo
			},

			depositRequestRepository: func() *mocks.DepositRequestRepository {
				repo := new(mocks.DepositRequestRepository)
				repo.On("GetAll", mock.Anything,
					mock.MatchedBy(func(field func(fields *cdomain.DepositRequestQueryOptionFields)) bool {
						opt := cdomain.DepositRequestQueryOptionFunc()
						field(&opt)

						return *opt.UserId() == fmt.Sprintf("%v", int64(1000))
					}),
					mock.MatchedBy(func(field func(fields *cdomain.DepositRequestQueryOptionFields)) bool {
						opt := cdomain.DepositRequestQueryOptionFunc()
						field(&opt)

						return opt.ExpireDepositDate() != ""
					}),
				).Return(
					[]domain.DepositRequest{
						{
							Id:                 3000,
							ShareId:            "SHARE-1",
							DepositCheckStatus: domain.CANCEL_DEPOSIT,
						},
						{
							Id:                 3001,
							ShareId:            "SHARE-2",
							DepositCheckStatus: domain.NO_DEPOSIT,
						},
						{
							Id:                 3002,
							ShareId:            "SHARE-3",
							DepositCheckStatus: domain.NO_DEPOSIT,
						},
					}, nil)
				return repo
			},
			userAccountService: func() *mocks.UserAccountService {
				svc := &mocks.UserAccountService{}
				svc.On("DepositCheckProcess", mock.Anything,
					domain.DepositRequest{Id: 3001, ShareId: "SHARE-2", DepositCheckStatus: domain.NO_DEPOSIT}).Return(nil)
				svc.On("DepositCheckProcess", mock.Anything,
					domain.DepositRequest{Id: 3002, ShareId: "SHARE-3", DepositCheckStatus: domain.NO_DEPOSIT}).Return(nil)
				return svc
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				repo.On("LockById", mock.Anything, int64(2000)).Return(domain.Project{Id: 2000}, nil)
				return repo
			},
			projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
				repo := new(mocks.ProjectInvestmentExecuteRepository)
				repo.On("GetAllByProjectId", mock.Anything, int64(2000)).Return([]domain.ProjectInvestmentExecute{}, nil)
				return repo
			},
			want: adomain.InvestmentUpdateStatusResponse{
				Investment: domain.Investment{InvestmentId: "SHARE-1"},
			},
			wantErr: false,
		},
		{
			// MIR-1295: 成立→取消は出資実行が完了していれば許可する。
			name: "MIR-1295: 成立から取消への変更は出資実行が完了していれば成功する",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentUpdateStatusRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Name:  "コ・ギルドン",
						Email: "gildong@email.com",
					},
					InvestmentId: "SHARE-1",
					InvestmentUpdateStatusBody: adomain.InvestmentUpdateStatusBody{
						InvestStatus: domain.CANCEL_INVEST,
					},
				},
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := new(mocks.ShareRepository)
				repo.On("GetById", mock.Anything, "SHARE-1").Return(domain.Share{
					Id:           "SHARE-1",
					ProjectId:    2000,
					InvestStatus: domain.REQUEST_INVEST,
				}, nil)
				repo.On("LockById", mock.Anything, "SHARE-1").Return(domain.Share{
					Id:           "SHARE-1",
					ProjectId:    2000,
					InvestStatus: domain.SUCCESS_INVEST,
					OwnerId:      1000,
				}, nil)
				return repo
			},
			shareService: func() *mocks.ShareService {
				service := new(mocks.ShareService)
				service.On("CancelInvest", mock.Anything, "SHARE-1", int64(1000)).Return(nil)
				return service
			},
			investmentRepository: func() *mocks.InvestmentRepository {
				repo := new(mocks.InvestmentRepository)
				repo.On("GetById", mock.Anything, "SHARE-1").Return(domain.Investment{InvestmentId: "SHARE-1"}, nil)
				return repo
			},
			transactionInfoHistoryRepository: func() *mocks.TransactionInfoHistoryRepository {
				repo := new(mocks.TransactionInfoHistoryRepository)
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)
				return repo
			},
			depositRequestRepository: func() *mocks.DepositRequestRepository {
				repo := new(mocks.DepositRequestRepository)
				repo.On("GetAll", mock.Anything, mock.Anything, mock.Anything).Return([]domain.DepositRequest{}, nil)
				return repo
			},
			userAccountService: func() *mocks.UserAccountService {
				return new(mocks.UserAccountService)
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				repo.On("LockById", mock.Anything, int64(2000)).Return(domain.Project{Id: 2000}, nil)
				return repo
			},
			projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
				repo := new(mocks.ProjectInvestmentExecuteRepository)
				repo.On("GetAllByProjectId", mock.Anything, int64(2000)).Return([]domain.ProjectInvestmentExecute{
					{Id: 9000, ProjectId: 2000, ExecuteStatus: domain.COMPLETED},
				}, nil)
				return repo
			},
			want: adomain.InvestmentUpdateStatusResponse{
				Investment: domain.Investment{InvestmentId: "SHARE-1"},
			},
			wantErr: false,
		},
		{
			// MIR-1295: 成立→取消は出資実行が途中(分割実行の途中)なら拒否する。
			// 許すと募集額が減り、残りの「運用開始」が10106で必ず落ちるため。
			name: "MIR-1295: 成立から取消への変更は出資実行が途中なら拒否される",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentUpdateStatusRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Name:  "コ・ギルドン",
						Email: "gildong@email.com",
					},
					InvestmentId: "SHARE-1",
					InvestmentUpdateStatusBody: adomain.InvestmentUpdateStatusBody{
						InvestStatus: domain.CANCEL_INVEST,
					},
				},
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := new(mocks.ShareRepository)
				repo.On("GetById", mock.Anything, "SHARE-1").Return(domain.Share{
					Id:           "SHARE-1",
					ProjectId:    2000,
					InvestStatus: domain.REQUEST_INVEST,
				}, nil)
				repo.On("LockById", mock.Anything, "SHARE-1").Return(domain.Share{
					Id:           "SHARE-1",
					ProjectId:    2000,
					InvestStatus: domain.SUCCESS_INVEST,
					OwnerId:      1000,
				}, nil)
				return repo
			},
			shareService: func() *mocks.ShareService {
				return new(mocks.ShareService)
			},
			investmentRepository: func() *mocks.InvestmentRepository {
				return new(mocks.InvestmentRepository)
			},
			transactionInfoHistoryRepository: func() *mocks.TransactionInfoHistoryRepository {
				return new(mocks.TransactionInfoHistoryRepository)
			},
			depositRequestRepository: func() *mocks.DepositRequestRepository {
				return new(mocks.DepositRequestRepository)
			},
			userAccountService: func() *mocks.UserAccountService {
				return new(mocks.UserAccountService)
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				repo.On("LockById", mock.Anything, int64(2000)).Return(domain.Project{Id: 2000}, nil)
				return repo
			},
			projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
				repo := new(mocks.ProjectInvestmentExecuteRepository)
				repo.On("GetAllByProjectId", mock.Anything, int64(2000)).Return([]domain.ProjectInvestmentExecute{
					{Id: 9000, ProjectId: 2000, ExecuteStatus: domain.SCHEDULED},
				}, nil)
				return repo
			},
			want:    adomain.InvestmentUpdateStatusResponse{},
			wantErr: true,
		},
		{
			// MIR-1295: 申込中→投資不成立は許可表に無いため拒否される。
			name: "MIR-1295: 申込中から投資不成立への変更は拒否される",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentUpdateStatusRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Name:  "コ・ギルドン",
						Email: "gildong@email.com",
					},
					InvestmentId: "SHARE-1",
					InvestmentUpdateStatusBody: adomain.InvestmentUpdateStatusBody{
						InvestStatus: domain.FAIL_INVEST,
					},
				},
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := new(mocks.ShareRepository)
				repo.On("GetById", mock.Anything, "SHARE-1").Return(domain.Share{
					Id:           "SHARE-1",
					InvestStatus: domain.REQUEST_INVEST,
					OwnerId:      1000,
				}, nil)
				return repo
			},
			shareService: func() *mocks.ShareService {
				return new(mocks.ShareService)
			},
			investmentRepository: func() *mocks.InvestmentRepository {
				return new(mocks.InvestmentRepository)
			},
			transactionInfoHistoryRepository: func() *mocks.TransactionInfoHistoryRepository {
				return new(mocks.TransactionInfoHistoryRepository)
			},
			depositRequestRepository: func() *mocks.DepositRequestRepository {
				return new(mocks.DepositRequestRepository)
			},
			userAccountService: func() *mocks.UserAccountService {
				return new(mocks.UserAccountService)
			},
			projectRepository: func() *mocks.ProjectRepository {
				return new(mocks.ProjectRepository)
			},
			projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
				return new(mocks.ProjectInvestmentExecuteRepository)
			},
			want:    adomain.InvestmentUpdateStatusResponse{},
			wantErr: true,
		},
		{
			name: "投資ステータス更新：InvestStatus(REQUEST_INVEST)、Share状態(REQUEST_INVEST)=> エラーを返します。",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentUpdateStatusRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "gildong@email.com",
					},
					InvestmentId: "SHARE-1",
					InvestmentUpdateStatusBody: adomain.InvestmentUpdateStatusBody{
						InvestStatus: domain.REQUEST_INVEST,
					},
				},
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := new(mocks.ShareRepository)
				repo.On("GetById", mock.Anything, "SHARE-1").Return(domain.Share{Id: "SHARE-1", InvestStatus: domain.REQUEST_INVEST}, nil)
				return repo
			},
			shareService: func() *mocks.ShareService {
				service := new(mocks.ShareService)
				return service
			},
			investmentRepository: func() *mocks.InvestmentRepository {
				repo := new(mocks.InvestmentRepository)
				return repo
			},

			transactionInfoHistoryRepository: func() *mocks.TransactionInfoHistoryRepository {
				repo := new(mocks.TransactionInfoHistoryRepository)
				return repo
			},
			depositRequestRepository: func() *mocks.DepositRequestRepository {
				repo := new(mocks.DepositRequestRepository)
				return repo
			},

			userAccountService: func() *mocks.UserAccountService {
				service := new(mocks.UserAccountService)
				return service
			},
			projectRepository: func() *mocks.ProjectRepository {
				return new(mocks.ProjectRepository)
			},
			projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
				return new(mocks.ProjectInvestmentExecuteRepository)
			},
			want:    adomain.InvestmentUpdateStatusResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txManager := newInvestmentTxManagerMock()
			usecase := investmentUsecase{
				txManager:                          txManager,
				shareRepository:                    tt.shareRepository(),
				shareService:                       tt.shareService(),
				investmentRepository:               tt.investmentRepository(),
				depositRequestRepository:           tt.depositRequestRepository(),
				userAccountService:                 tt.userAccountService(),
				transactionInfoHistoryRepository:   tt.transactionInfoHistoryRepository(),
				projectRepository:                  tt.projectRepository(),
				projectInvestmentExecuteRepository: tt.projectInvestmentExecuteRepository(),
			}

			got, err := usecase.UpdateStatus(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "investmentUsecase.UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "investmentUsecase.UpdateStatus() = %v, want %v", got, tt.want)
			if tt.args.request.InvestStatus == domain.CANCEL_INVEST {
				txManager.AssertExpectations(t)
			}
		})
	}
}

func Test_investmentUsecase_UpdateStatus_CancelTransition(t *testing.T) {
	tests := []struct {
		name         string
		investStatus domain.INVEST_STATUS
		wantErr      error
	}{
		{name: "投資不成立からキャンセルへ変更できない", investStatus: domain.FAIL_INVEST, wantErr: cdomain.ErrCannotManuallyChangeInvestStatus},
		{name: "運用中からキャンセルへ変更できない", investStatus: domain.OPERATION_INVEST, wantErr: cdomain.ErrCannotManuallyChangeInvestStatus},
		{name: "運用終了からキャンセルへ変更できない", investStatus: domain.REFUND_INVEST, wantErr: cdomain.ErrCannotManuallyChangeInvestStatus},
		{name: "キャンセル済みから再度キャンセルできない", investStatus: domain.CANCEL_INVEST, wantErr: cdomain.ErrItemConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shareRepository := new(mocks.ShareRepository)
			shareRepository.On("GetById", mock.Anything, "SHARE-1").Return(domain.Share{
				Id:        "SHARE-1",
				ProjectId: 2000,
			}, nil)
			shareRepository.On("LockById", mock.Anything, "SHARE-1").Return(domain.Share{
				Id:           "SHARE-1",
				ProjectId:    2000,
				InvestStatus: tt.investStatus,
			}, nil)
			shareService := new(mocks.ShareService)

			projectRepository := new(mocks.ProjectRepository)
			projectRepository.On("LockById", mock.Anything, int64(2000)).Return(domain.Project{Id: 2000}, nil)

			projectInvestmentExecuteRepository := new(mocks.ProjectInvestmentExecuteRepository)
			// キャンセル済みから再度キャンセルする場合はfrom==toのチェックで
			// 弾かれるため、GetAllByProjectIdまで到達しない。呼ばれなくても
			// mockery製のstrictモックはOnしていない呼び出しでpanicするため、
			// 到達すれば必ずここで検知できる。
			if tt.investStatus != domain.CANCEL_INVEST {
				projectInvestmentExecuteRepository.On("GetAllByProjectId", mock.Anything, int64(2000)).
					Return([]domain.ProjectInvestmentExecute{}, nil)
			}

			txManager := newInvestmentTxManagerMock()
			usecase := investmentUsecase{
				txManager:                          txManager,
				shareRepository:                    shareRepository,
				shareService:                       shareService,
				projectRepository:                  projectRepository,
				projectInvestmentExecuteRepository: projectInvestmentExecuteRepository,
			}

			_, err := usecase.UpdateStatus(context.Background(), adomain.InvestmentUpdateStatusRequest{
				JWTUser:      jwthelper.User{Id: 1000},
				InvestmentId: "SHARE-1",
				InvestmentUpdateStatusBody: adomain.InvestmentUpdateStatusBody{
					InvestStatus: domain.CANCEL_INVEST,
				},
			})

			assert.ErrorIs(t, err, tt.wantErr)
			shareService.AssertNotCalled(t, "CancelInvest", mock.Anything, mock.Anything, mock.Anything)
			txManager.AssertExpectations(t)
		})
	}
}

func Test_investmentUsecase_UpdateStatus_CancelInfrastructureErrors(t *testing.T) {
	tests := []struct {
		name            string
		txErr           error
		shareGetErr     error
		projectLockErr  error
		lockErr         error
		executesErr     error
		cancelInvestErr error
	}{
		{name: "トランザクション開始失敗を返す", txErr: errors.New("failed to begin transaction")},
		{name: "project_id解決用のShare取得失敗を返す", shareGetErr: errors.New("failed to get share")},
		{name: "projectの行のロック失敗を返す", projectLockErr: errors.New("failed to lock project")},
		{name: "Share行のロック失敗を返す", lockErr: errors.New("failed to lock share")},
		{name: "出資実行記録の取得失敗を返す", executesErr: errors.New("failed to get project investment executes")},
		{name: "投資キャンセル処理の失敗を返す", cancelInvestErr: errors.New("failed to cancel investment")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txManager := new(mocks.TxManager)
			shareRepository := new(mocks.ShareRepository)
			shareService := new(mocks.ShareService)
			projectRepository := new(mocks.ProjectRepository)
			projectInvestmentExecuteRepository := new(mocks.ProjectInvestmentExecuteRepository)

			// MIR-1295: project_id解決用のGetByIdはWithTxの外(トランザクション
			// 開始前)で呼ばれるため、txErrの有無に関わらず最初に評価する。
			if tt.shareGetErr != nil {
				shareRepository.On("GetById", mock.Anything, "SHARE-1").Return(domain.Share{}, tt.shareGetErr)
			} else {
				shareRepository.On("GetById", mock.Anything, "SHARE-1").Return(domain.Share{
					Id:        "SHARE-1",
					ProjectId: 2000,
				}, nil)

				if tt.txErr != nil {
					txManager.On("WithTx", mock.Anything, mock.Anything).Return(tt.txErr)
				} else {
					txManager.On("WithTx", mock.Anything, mock.Anything).
						Return(func(ctx context.Context, fn func(context.Context) error) error {
							return fn(ctx)
						})

					if tt.projectLockErr != nil {
						projectRepository.On("LockById", mock.Anything, int64(2000)).Return(domain.Project{}, tt.projectLockErr)
					} else {
						projectRepository.On("LockById", mock.Anything, int64(2000)).Return(domain.Project{Id: 2000}, nil)

						if tt.lockErr != nil {
							shareRepository.On("LockById", mock.Anything, "SHARE-1").
								Return(domain.Share{}, tt.lockErr)
						} else {
							shareRepository.On("LockById", mock.Anything, "SHARE-1").Return(domain.Share{
								Id:           "SHARE-1",
								ProjectId:    2000,
								OwnerId:      1000,
								InvestStatus: domain.REQUEST_INVEST,
							}, nil)

							if tt.executesErr != nil {
								projectInvestmentExecuteRepository.On("GetAllByProjectId", mock.Anything, int64(2000)).
									Return(nil, tt.executesErr)
							} else {
								projectInvestmentExecuteRepository.On("GetAllByProjectId", mock.Anything, int64(2000)).
									Return([]domain.ProjectInvestmentExecute{}, nil)
								shareService.On("CancelInvest", mock.Anything, "SHARE-1", int64(1000)).
									Return(tt.cancelInvestErr)
							}
						}
					}
				}
			}

			usecase := investmentUsecase{
				txManager:                          txManager,
				shareRepository:                    shareRepository,
				shareService:                       shareService,
				projectRepository:                  projectRepository,
				projectInvestmentExecuteRepository: projectInvestmentExecuteRepository,
			}

			_, err := usecase.UpdateStatus(context.Background(), adomain.InvestmentUpdateStatusRequest{
				JWTUser:      jwthelper.User{Id: 1000},
				InvestmentId: "SHARE-1",
				InvestmentUpdateStatusBody: adomain.InvestmentUpdateStatusBody{
					InvestStatus: domain.CANCEL_INVEST,
				},
			})

			wantErr := tt.shareGetErr
			if wantErr == nil {
				wantErr = tt.txErr
			}
			if wantErr == nil {
				wantErr = tt.projectLockErr
			}
			if wantErr == nil {
				wantErr = tt.lockErr
			}
			if wantErr == nil {
				wantErr = tt.executesErr
			}
			if wantErr == nil {
				wantErr = tt.cancelInvestErr
			}
			assert.ErrorIs(t, err, wantErr)
			txManager.AssertExpectations(t)
			shareRepository.AssertExpectations(t)
			shareService.AssertExpectations(t)
			projectRepository.AssertExpectations(t)
			projectInvestmentExecuteRepository.AssertExpectations(t)
		})
	}
}

func newInvestmentTxManagerMock() *mocks.TxManager {
	txManager := new(mocks.TxManager)
	txManager.On("WithTx", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
	return txManager
}

func Test_investmentUsecase_GetAllTransactionInfoHistories(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.InvestmentGetAllTransactionInfoHistoriesRequest
	}

	tests := []struct {
		name                             string
		args                             args
		transactionInfoHistoryRepository func() *mocks.TransactionInfoHistoryRepository
		want                             adomain.InvestmentGetAllTransactionInfoHistoriesResponse
		wantErr                          bool
	}{
		{
			name: "投資IDで履歴を照会：InvestmentId(SHARE-1)=> リストが正常に照会されます。",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentGetAllTransactionInfoHistoriesRequest{
					InvestmentId: "SHARE-1",
				},
			},

			transactionInfoHistoryRepository: func() *mocks.TransactionInfoHistoryRepository {
				repo := new(mocks.TransactionInfoHistoryRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.TransactionInfoHistoryQueryOptionFields)) bool {
						opt := cdomain.TransactionInfoHistoryQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.ShareId()) == "SHARE-1"
					}), // InvestmentId,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(int64(3), []domain.TransactionInfoHistory{
					mockTransactionInfoHistory1,
					mockTransactionInfoHistory2,
					mockTransactionInfoHistory3,
				}, nil)

				return repo
			},
			want: adomain.InvestmentGetAllTransactionInfoHistoriesResponse{
				TotalCount: 3,
				Items: []domain.TransactionInfoHistory{
					mockTransactionInfoHistory1,
					mockTransactionInfoHistory2,
					mockTransactionInfoHistory3,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := investmentUsecase{
				transactionInfoHistoryRepository: tt.transactionInfoHistoryRepository(),
			}

			got, err := usecase.GetAllTransactionInfoHistories(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "investmentWaitHistoryUsecase.GetAllTransactionInfoHistories() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "investmentWaitHistoryUsecase.GetAllTransactionInfoHistories() = %v, want %v", got, tt.want)
		})
	}
}
