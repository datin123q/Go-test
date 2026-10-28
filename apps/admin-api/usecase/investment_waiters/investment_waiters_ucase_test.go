package investment_waiters

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"

	"github.com/pkg/errors"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/illuminarean/energy-share-core-api/controllers"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_investmentWaiters_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.InvestmentWaitersGetAllRequest
	}

	tests := []struct {
		name                        string
		args                        args
		investmentWaitersRepository func() *mocks.InvestmentWaitersRepository
		want                        adomain.InvestmentWaitersGetAllResponse
		wantErr                     bool
	}{
		{
			name: "投資待機者リストを照会します: SkipCount(0), MaxResultCount(2) => 正常にリストが照会されます。",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersGetAllRequest{
					SearchInput: controllers.SearchInput{
						SkipCount:      0,
						MaxResultCount: 2,
					},
				},
			},

			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetAll", context.Background(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.MatchedBy(func(field func(*cdomain.InvestmentWaitersQueryOptionFields)) bool {
						opt := cdomain.InvestmentWaitersQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}), // SkipCount
					mock.MatchedBy(func(field func(*cdomain.InvestmentWaitersQueryOptionFields)) bool {
						opt := cdomain.InvestmentWaitersQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 2
					}), // MaxResultCount
				).Return(int64(2), []domain.InvestmentWaiters{
					mockInvestmentWaiters1,
					mockInvestmentWaiters2,
				}, nil)

				return repo
			},
			want: adomain.InvestmentWaitersGetAllResponse{
				TotalCount: 2,
				Items: []adomain.InvestmentWaitersResponse{
					mockInvestmentWaitersResponse1,
					mockInvestmentWaitersResponse2,
				},
			},
			wantErr: false,
		},
		{
			name: "待機者リストを特定プロジェクトでフィルタリングします: ProjectId(2000) => 正常にリストが照会されます。",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersGetAllRequest{
					ProjectId: lo.ToPtr(int64(2000)),
				},
			},

			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.InvestmentWaitersQueryOptionFields)) bool {
						opt := cdomain.InvestmentWaitersQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.ProjectId()) == "2000"
					}), // InvestmentWaitId,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(int64(3), []domain.InvestmentWaiters{
					mockInvestmentWaiters1,
					mockInvestmentWaiters2,
					mockInvestmentWaiters3,
				}, nil)

				return repo
			},
			want: adomain.InvestmentWaitersGetAllResponse{
				TotalCount: 3,
				Items: []adomain.InvestmentWaitersResponse{
					mockInvestmentWaitersResponse1,
					mockInvestmentWaitersResponse2,
					mockInvestmentWaitersResponse3,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := investmentWaitersUsecase{
				investmentWaitersRepository: tt.investmentWaitersRepository(),
			}

			got, err := usecase.GetAll(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "investmentWaitersUsecase.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "investmentWaitersUsecase.GetAll() = %v, want %v", got, tt.want)
		})
	}
}

func Test_investmentWaiters_GetOne(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.InvestmentWaitersGetOneRequest
	}

	tests := []struct {
		name                        string
		args                        args
		investmentWaitersRepository func() *mocks.InvestmentWaitersRepository
		projectRepository           func() *mocks.ProjectRepository
		want                        adomain.InvestmentWaitersGetOneResponse
		wantErr                     bool
	}{
		{
			name: "投資待機者を照会します: InvestmentWaitersId(1000) => 正常に照会されます。",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersGetOneRequest{
					InvestmentWaitersId: lo.ToPtr(int64(1000)),
				},
			},

			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", context.Background(), int64(1000)).Return(mockInvestmentWaiters1, nil)
				repo.On("GetWinningsByWaiterId", context.Background(), mockInvestmentWaiters1.Id).Return(
					[]domain.InvestmentWaitersWinningItem{mockInvestmentWaitersWinningItem1}, nil,
				)

				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				repo.On("GetById", context.Background(), mockInvestmentWaiters1.ProjectId).Return(mockProject1, nil)

				return repo
			},

			// Winningsの期待値はテスト対象と同じFromInvestmentWaitersWinningItemを経由せず、
			// リテラルで書く。同じ変換関数を経由すると、フィールドの取り違え(例えば
			// PriceとAmountの入れ替え)があってもテストが必ず一致してしまい、気づけない
			// (②の敵対的検証で指摘)。
			want: adomain.InvestmentWaitersGetOneResponse{
				InvestmentWaitersResponse: mockInvestmentWaitersResponse1,
				InvestmentProposalMaxQty:  5,
				InvestmentProposalMinQty:  2,
				IsInvestmentAmountLimit:   true,
				InvestmentAmountLimit:     5_000_000,
				PricePerQty:               10_000,
				Winnings: []adomain.InvestmentWaitersWinningItem{
					{
						ShareId:            "SHARE_ID1",
						Qty:                3,
						UnitPrice:          10_000,
						Amount:             30_000,
						ShareStatus:        domain.SUCCESS_INVEST,
						DepositCheckStatus: domain.OK_DEPOSIT,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "投資待機者を照会します: 残口数<申込口数 => 最大口数は残口数になります。",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersGetOneRequest{
					InvestmentWaitersId: lo.ToPtr(int64(1000)),
				},
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", context.Background(), int64(1000)).Return(mockInvestmentWaiters1, nil)
				// このケースは残口数によるキャップの検証が目的のため、繰上げ履歴は
				// 空のまま(MIR-1068のWinningsとは無関係にしておく)。
				repo.On("GetWinningsByWaiterId", context.Background(), mockInvestmentWaiters1.Id).Return(
					[]domain.InvestmentWaitersWinningItem{}, nil,
				)
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				project := mockProject1
				project.RemainingQty = 3
				repo.On("GetById", context.Background(), mockInvestmentWaiters1.ProjectId).Return(project, nil)
				return repo
			},
			want: adomain.InvestmentWaitersGetOneResponse{
				InvestmentWaitersResponse: mockInvestmentWaitersResponse1,
				InvestmentProposalMaxQty:  3,
				InvestmentProposalMinQty:  2,
				IsInvestmentAmountLimit:   true,
				InvestmentAmountLimit:     5_000_000,
				PricePerQty:               10_000,
				Winnings:                  []adomain.InvestmentWaitersWinningItem{},
			},
			wantErr: false,
		},
		{
			name: "投資待機者を照会します: InvestmentWaitersId(2000) => 照会されません。",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersGetOneRequest{
					InvestmentWaitersId: lo.ToPtr(int64(2000)),
				},
			},

			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", context.Background(), int64(2000)).Return(domain.InvestmentWaiters{}, errors.New("not found"))

				return repo
			},

			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				return repo
			},

			want:    adomain.InvestmentWaitersGetOneResponse{},
			wantErr: true,
		},
		{
			name: "繰上げ履歴の取得に失敗した場合 => error",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersGetOneRequest{
					InvestmentWaitersId: lo.ToPtr(mockInvestmentWaiters1.Id),
				},
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", context.Background(), mockInvestmentWaiters1.Id).Return(mockInvestmentWaiters1, nil)
				repo.On("GetWinningsByWaiterId", context.Background(), mockInvestmentWaiters1.Id).Return(
					nil, errors.New("db error"),
				)

				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				repo.On("GetById", context.Background(), mockInvestmentWaiters1.ProjectId).Return(mockProject1, nil)

				return repo
			},
			want:    adomain.InvestmentWaitersGetOneResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := investmentWaitersUsecase{
				investmentWaitersRepository: tt.investmentWaitersRepository(),
				projectRepository:           tt.projectRepository(),
			}

			got, err := usecase.GetOne(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "investmentWaitersUsecase.GetOne() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "investmentWaitersUsecase.GetOne() = %v, want %v", got, tt.want)
		})
	}
}

func Test_investmentWaitersUsecase_Download(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.InvestmentWaitersDownloadRequest
	}

	tests := []struct {
		name                        string
		args                        args
		investmentWaitersRepository func() *mocks.InvestmentWaitersRepository
		investmentWaitersService    func() *mocks.InvestmentWaitersService
		want                        adomain.InvestmentWaitersDownloadResponse
		wantErr                     bool
	}{
		{
			name: "投資待機者リストをダウンロードします: 正常にダウンロードされます。",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersDownloadRequest{
					SearchInput: controllers.SearchInput{
						SkipCount:      0,
						MaxResultCount: 2,
					},
				},
			},

			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetAll", context.Background(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.MatchedBy(func(field func(*cdomain.InvestmentWaitersQueryOptionFields)) bool {
						opt := cdomain.InvestmentWaitersQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}), // SkipCount
					mock.MatchedBy(func(field func(*cdomain.InvestmentWaitersQueryOptionFields)) bool {
						opt := cdomain.InvestmentWaitersQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 2
					}), // MaxResultCount
				).Return(int64(2), []domain.InvestmentWaiters{
					mockInvestmentWaiters1,
					mockInvestmentWaiters2,
				}, nil)

				return repo
			},
			investmentWaitersService: func() *mocks.InvestmentWaitersService {
				service := new(mocks.InvestmentWaitersService)
				service.On("ExportCSV", []domain.InvestmentWaiters{
					mockInvestmentWaiters1,
					mockInvestmentWaiters2,
				}).Return(
					cdomain.CSVData{
						[]string{"待機番号", "投資会員区分", "メールアドレス(ID)", "投資家名", "電話番号", "投資口数", "取引金額", "ステータス", "申請日時"},
						[]string{"1", "個人特定投資家", "gildong@email.com", "고길동", "010-1234-1234", "1", "1", "キャンセル待ち", "2024-01-01 00:00:00"},
						[]string{"2", "個人特定投資家", "dool2@email.com", "둘리", "010-1234-1234", "2", "2", "キャンセル待ち", "2024-01-01 00:00:00"},
					},
				)

				return service
			},
			want: adomain.InvestmentWaitersDownloadResponse{
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
					0x2e, 0x63, 0x6f, 0x6d, 0x2c, 0xea, 0xb3, 0xa0, 0xea, 0xb8, 0xb8, 0xeb, 0x8f, 0x99, 0x2c, 0x30,
					0x31, 0x30, 0x2d, 0x31, 0x32, 0x33, 0x34, 0x2d, 0x31, 0x32, 0x33, 0x34, 0x2c, 0x31, 0x2c, 0x31,
					0x2c, 0xe3, 0x82, 0xad, 0xe3, 0x83, 0xa3, 0xe3, 0x83, 0xb3, 0xe3, 0x82, 0xbb, 0xe3, 0x83, 0xab,
					0xe5, 0xbe, 0x85, 0xe3, 0x81, 0xa1,
					0x2c, 0x32, 0x30, 0x32, 0x34, 0x2d, 0x30, 0x31, 0x2d, 0x30, 0x31, 0x20, 0x30, 0x30, 0x3a, 0x30,
					0x30, 0x3a, 0x30, 0x30, 0x0a, 0x32, 0x2c, 0xe5, 0x80, 0x8b, 0xe4, 0xba, 0xba, 0xe7, 0x89, 0xb9,
					0xe5, 0xae, 0x9a, 0xe6, 0x8a, 0x95, 0xe8, 0xb3, 0x87, 0xe5, 0xae, 0xb6, 0x2c, 0x64, 0x6f, 0x6f,
					0x6c, 0x32, 0x40, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x2e, 0x63, 0x6f, 0x6d, 0x2c, 0xeb, 0x91, 0x98,
					0xeb, 0xa6, 0xac, 0x2c, 0x30, 0x31, 0x30, 0x2d, 0x31, 0x32, 0x33, 0x34, 0x2d, 0x31, 0x32, 0x33,
					0x34, 0x2c, 0x32, 0x2c, 0x32, 0x2c, 0xe3, 0x82, 0xad, 0xe3, 0x83, 0xa3, 0xe3, 0x83, 0xb3, 0xe3, 0x82,
					0xbb, 0xe3, 0x83, 0xab, 0xe5, 0xbe, 0x85, 0xe3, 0x81, 0xa1, 0x2c, 0x32, 0x30, 0x32, 0x34, 0x2d, 0x30,
					0x31, 0x2d, 0x30, 0x31,
					0x20, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x0a,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := investmentWaitersUsecase{
				investmentWaitersRepository: tt.investmentWaitersRepository(),
				investmentWaitersService:    tt.investmentWaitersService(),
			}

			got, err := usecase.Download(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "investmentWaitersUsecase.Download() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "investmentWaitersUsecase.Download() = %v, want %v", got, tt.want)
		})
	}
}

func Test_investmentWaitersUsecase_StatusUpdate(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.InvestmentWaitersStatusUpdateRequest
	}

	tests := []struct {
		name                        string
		investmentWaitersRepository func() *mocks.InvestmentWaitersRepository
		investmentWaitersService    func() *mocks.InvestmentWaitersService
		args                        args
		want                        adomain.InvestmentWaitersStatusUpdateResponse
		wantErr                     bool
		wantErrIs                   error
	}{
		{
			name: "待機者ステータス更新 : DEPOSIT_REQUEST => 正常に更新されます。",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersStatusUpdateRequest{
					InvestmentWaitersId: lo.ToPtr(int64(1000)),
					InvestmentWaitersStatusUpdateBody: adomain.InvestmentWaitersStatusUpdateBody{
						Status: domain.DEPOSIT_REQUEST,
					},
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
				},
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaiters1.Id).Return(mockInvestmentWaiters1, nil)
				return repo
			},
			investmentWaitersService: func() *mocks.InvestmentWaitersService {
				service := new(mocks.InvestmentWaitersService)
				service.On("ChangeDepositRequestStatus", mock.Anything, mockInvestmentWaiters1, mockUser1.Id).Return(nil)
				return service
			},
			wantErr: false,
			want: adomain.InvestmentWaitersStatusUpdateResponse{
				InvestmentWaitersResponse: mockInvestmentWaitersResponse1,
			},
		},
		{
			name: "待機者ステータス更新 : WAITING_END => 正常に更新されます。",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersStatusUpdateRequest{
					InvestmentWaitersId: lo.ToPtr(int64(1000)),
					InvestmentWaitersStatusUpdateBody: adomain.InvestmentWaitersStatusUpdateBody{
						Status: domain.WAITING_END,
					},
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
				},
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaiters1.Id).Return(mockInvestmentWaiters1, nil)
				return repo
			},
			investmentWaitersService: func() *mocks.InvestmentWaitersService {
				service := new(mocks.InvestmentWaitersService)
				service.On("ChangeWaitingEndStatus", mock.Anything, mockInvestmentWaiters1, mockUser1.Id).Return(nil)
				return service
			},
			wantErr: false,
			want: adomain.InvestmentWaitersStatusUpdateResponse{
				InvestmentWaitersResponse: mockInvestmentWaitersResponse1,
			},
		},
		{
			name: "대기자 상태 업데이트 : PROPOSAL => 정상적으로 업데이트 됩니다.",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersStatusUpdateRequest{
					InvestmentWaitersId: lo.ToPtr(int64(1000)),
					InvestmentWaitersStatusUpdateBody: adomain.InvestmentWaitersStatusUpdateBody{
						Status: domain.PROPOSAL,
					},
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
				},
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaiters1.Id).Return(mockInvestmentWaiters1, nil)
				return repo
			},
			investmentWaitersService: func() *mocks.InvestmentWaitersService {
				service := new(mocks.InvestmentWaitersService)
				service.On("ChangeProposalStatus", mock.Anything, mockInvestmentWaiters1, mockUser1.Id).Return(nil)
				return service
			},
			wantErr: false,
			want: adomain.InvestmentWaitersStatusUpdateResponse{
				InvestmentWaitersResponse: mockInvestmentWaitersResponse1,
			},
		},
		{
			name: "待機者ステータス更新 : CANCELLED(未繰上げ) => 正常に更新されます。",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersStatusUpdateRequest{
					InvestmentWaitersId: lo.ToPtr(mockInvestmentWaitersWaiting.Id),
					InvestmentWaitersStatusUpdateBody: adomain.InvestmentWaitersStatusUpdateBody{
						Status: domain.CANCELLED,
					},
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
				},
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaitersWaiting.Id).Return(mockInvestmentWaitersWaiting, nil)
				repo.On("LockAndSumAdvancedQty", mock.Anything, mockInvestmentWaitersWaiting.Id).Return(int64(0), nil)
				return repo
			},
			investmentWaitersService: func() *mocks.InvestmentWaitersService {
				service := new(mocks.InvestmentWaitersService)
				service.On("ChangeCancelledStatus", mock.Anything, mockInvestmentWaitersWaiting, mockUser1.Id).Return(nil)
				return service
			},
			wantErr: false,
			want: adomain.InvestmentWaitersStatusUpdateResponse{
				InvestmentWaitersResponse: adomain.InvestmentWaitersResponse{}.FromInvestmentWaiters(mockInvestmentWaitersWaiting),
			},
		},
		{
			// MIR-1068: 繰上げ後もStatusがWAITINGのまま残るようになったため、
			// Status != WAITINGだけでは弾けない。advancedQty > 0 なら弾くこと
			// (②の敵対的検証で発見した対応漏れの回帰テスト)。
			name: "待機者ステータス更新 : CANCELLED(一部繰上げ済み) => ErrBadRequest",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersStatusUpdateRequest{
					InvestmentWaitersId: lo.ToPtr(mockInvestmentWaitersWaiting.Id),
					InvestmentWaitersStatusUpdateBody: adomain.InvestmentWaitersStatusUpdateBody{
						Status: domain.CANCELLED,
					},
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
				},
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaitersWaiting.Id).Return(mockInvestmentWaitersWaiting, nil)
				repo.On("LockAndSumAdvancedQty", mock.Anything, mockInvestmentWaitersWaiting.Id).Return(int64(4), nil)
				return repo
			},
			investmentWaitersService: func() *mocks.InvestmentWaitersService {
				return new(mocks.InvestmentWaitersService)
			},
			wantErr:   true,
			wantErrIs: cdomain.ErrBadRequest,
			want:      adomain.InvestmentWaitersStatusUpdateResponse{},
		},
		{
			name: "待機者ステータス更新 : CANCELLED(Statusが別状態) => ErrBadRequest",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersStatusUpdateRequest{
					InvestmentWaitersId: lo.ToPtr(mockInvestmentWaiters1.Id),
					InvestmentWaitersStatusUpdateBody: adomain.InvestmentWaitersStatusUpdateBody{
						Status: domain.CANCELLED,
					},
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
				},
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaiters1.Id).Return(mockInvestmentWaiters1, nil)
				return repo
			},
			investmentWaitersService: func() *mocks.InvestmentWaitersService {
				return new(mocks.InvestmentWaitersService)
			},
			wantErr:   true,
			wantErrIs: cdomain.ErrBadRequest,
			want:      adomain.InvestmentWaitersStatusUpdateResponse{},
		},
		{
			name: "待機者ステータス更新 : WAITING => 未知の状態でエラーが発生します。",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersStatusUpdateRequest{
					InvestmentWaitersId: lo.ToPtr(int64(1000)),
					InvestmentWaitersStatusUpdateBody: adomain.InvestmentWaitersStatusUpdateBody{
						Status: domain.WAITING,
					},
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
				},
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaiters1.Id).Return(mockInvestmentWaiters1, nil)
				return repo
			},
			investmentWaitersService: func() *mocks.InvestmentWaitersService {
				service := new(mocks.InvestmentWaitersService)
				return service
			},
			wantErr: true,
			want:    adomain.InvestmentWaitersStatusUpdateResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			investmentWaitersRepository := tt.investmentWaitersRepository()
			investmentWaitersService := tt.investmentWaitersService()
			usecase := investmentWaitersUsecase{
				investmentWaitersRepository: investmentWaitersRepository,
				investmentWaitersService:    investmentWaitersService,
			}

			got, err := usecase.StatusUpdate(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "investmentWaitersUsecase.StatusUpdate() error = %v, wantErr %v", err, tt.wantErr)
			if tt.wantErrIs != nil {
				assert.ErrorIs(t, err, tt.wantErrIs)
			}
			assert.Equalf(t, tt.want, got, "investmentWaitersUsecase.StatusUpdate() = %v, want %v", got, tt.want)
			investmentWaitersRepository.AssertExpectations(t)
			investmentWaitersService.AssertExpectations(t)
		})
	}
}

func TestInvestmentWaitersUsecase_Cancel_Status(t *testing.T) {
	requestInvestStatus := domain.REQUEST_INVEST
	successInvestStatus := domain.SUCCESS_INVEST
	cancelErr := errors.New("failed to cancel investment waiter")
	jwtUser := jwthelper.User{Id: mockUser1.Id, Email: mockUser1.Email, Name: mockUser1.Name}
	tests := []struct {
		name           string
		waiter         domain.InvestmentWaiters
		advancedQty    int64
		wantCancel     bool
		cancelErr      error
		wantBadRequest bool
	}{
		{
			name: "キャンセル待ち申込は取消処理を実行する",
			waiter: domain.InvestmentWaiters{
				Id:     1000,
				Status: domain.WAITING,
			},
			advancedQty: 0,
			wantCancel:  true,
		},
		{
			name: "キャンセル待ち申込の取消失敗を返す",
			waiter: domain.InvestmentWaiters{
				Id:     1000,
				Status: domain.WAITING,
			},
			advancedQty: 0,
			wantCancel:  true,
			cancelErr:   cancelErr,
		},
		{
			// MIR-1068: 繰上げ後もStatusがWAITINGのまま残るようになったため、
			// Status != WAITINGだけでは弾けない。advancedQty > 0 なら弾くこと(①の
			// 敵対的検証で発見した対応漏れの回帰テスト)。
			name: "一部繰上げ済み(WAITINGのまま)はキャンセルできない",
			waiter: domain.InvestmentWaiters{
				Id:     1000,
				Status: domain.WAITING,
			},
			advancedQty:    4,
			wantBadRequest: true,
		},
		{
			name: "当選後かつ投資申請中でもキャンセルできない",
			waiter: domain.InvestmentWaiters{
				Id:           1000,
				UserId:       3000,
				ProjectId:    2000,
				Status:       domain.WAITING_END,
				ShareId:      "share-1",
				InvestStatus: &requestInvestStatus,
			},
			wantBadRequest: true,
		},
		{
			name: "当選後かつ投資完了でもキャンセルできない",
			waiter: domain.InvestmentWaiters{
				Id:           1000,
				UserId:       3000,
				ProjectId:    2000,
				Status:       domain.WAITING_END,
				ShareId:      "share-1",
				InvestStatus: &successInvestStatus,
			},
			wantBadRequest: true,
		},
		{
			name:           "入金要求中はキャンセルできない",
			waiter:         domain.InvestmentWaiters{Id: 1000, Status: domain.DEPOSIT_REQUEST},
			wantBadRequest: true,
		},
		{
			name:           "投資提案中はキャンセルできない",
			waiter:         domain.InvestmentWaiters{Id: 1000, Status: domain.PROPOSAL},
			wantBadRequest: true,
		},
		{
			name: "キャンセル済みの申込は再度キャンセルできない",
			waiter: domain.InvestmentWaiters{
				Id:     1000,
				Status: domain.CANCELLED,
			},
			wantBadRequest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := new(mocks.InvestmentWaitersRepository)
			repository.On("LockById", mock.Anything, tt.waiter.Id).Return(tt.waiter, nil)
			// Status != WAITINGなら早期リターンするため、Status==WAITINGのケースだけ
			// LockAndSumAdvancedQtyが呼ばれる。
			if tt.waiter.Status == domain.WAITING {
				repository.On("LockAndSumAdvancedQty", mock.Anything, tt.waiter.Id).Return(tt.advancedQty, nil)
			}

			service := new(mocks.InvestmentWaitersService)
			if tt.wantCancel {
				service.On("ChangeCancelledStatus", mock.Anything, tt.waiter, mockUser1.Id).Return(tt.cancelErr)
			}

			txManager := newInvestmentWaitersTxManagerMock()
			usecase := investmentWaitersUsecase{
				txManager:                   txManager,
				investmentWaitersRepository: repository,
				investmentWaitersService:    service,
			}
			_, err := usecase.Cancel(context.Background(), adomain.InvestmentWaitersCancelRequest{
				JWTUser:             jwtUser,
				InvestmentWaitersId: lo.ToPtr(tt.waiter.Id),
			})

			switch {
			case tt.wantBadRequest:
				assert.ErrorIs(t, err, cdomain.ErrBadRequest)
			case tt.cancelErr != nil:
				assert.ErrorIs(t, err, tt.cancelErr)
			default:
				assert.NoError(t, err)
			}
			service.AssertExpectations(t)
			repository.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}

func TestInvestmentWaitersUsecase_Cancel_RechecksLockedStatus(t *testing.T) {
	waiter := domain.InvestmentWaiters{Id: 1000, Status: domain.WAITING}
	lockedWaiter := waiter
	lockedWaiter.Status = domain.CANCELLED
	repository := new(mocks.InvestmentWaitersRepository)
	repository.On("LockById", mock.Anything, waiter.Id).Return(lockedWaiter, nil)
	txManager := newInvestmentWaitersTxManagerMock()
	service := new(mocks.InvestmentWaitersService)

	usecase := investmentWaitersUsecase{
		txManager:                   txManager,
		investmentWaitersRepository: repository,
		investmentWaitersService:    service,
	}

	_, err := usecase.Cancel(context.Background(), adomain.InvestmentWaitersCancelRequest{
		JWTUser:             jwthelper.User{Id: mockUser1.Id},
		InvestmentWaitersId: lo.ToPtr(waiter.Id),
	})

	assert.ErrorIs(t, err, cdomain.ErrBadRequest)
	service.AssertNotCalled(t, "ChangeCancelledStatus", mock.Anything, mock.Anything, mock.Anything)
	repository.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

func TestInvestmentWaitersUsecase_Cancel_InfrastructureErrors(t *testing.T) {
	waiter := domain.InvestmentWaiters{Id: 1000, Status: domain.WAITING}
	tests := []struct {
		name           string
		txErr          error
		lockErr        error
		advancedQtyErr error
	}{
		{name: "トランザクション開始失敗を返す", txErr: errors.New("failed to begin transaction")},
		{name: "キャンセル待ちが存在しない場合はNotFoundを返す", lockErr: cdomain.ErrItemNotFound},
		{name: "キャンセル待ち行のロック失敗を返す", lockErr: errors.New("failed to lock investment waiter")},
		{name: "繰上げ済み累計の取得失敗を返す", advancedQtyErr: errors.New("failed to resolve advanced qty")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := new(mocks.InvestmentWaitersRepository)

			txManager := new(mocks.TxManager)
			switch {
			case tt.txErr != nil:
				txManager.On("WithTx", mock.Anything, mock.Anything).Return(tt.txErr)
			case tt.lockErr != nil:
				repository.On("LockById", mock.Anything, waiter.Id).
					Return(domain.InvestmentWaiters{}, tt.lockErr)
				txManager.On("WithTx", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
			case tt.advancedQtyErr != nil:
				repository.On("LockById", mock.Anything, waiter.Id).Return(waiter, nil)
				repository.On("LockAndSumAdvancedQty", mock.Anything, waiter.Id).
					Return(int64(0), tt.advancedQtyErr)
				txManager.On("WithTx", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
			}

			usecase := investmentWaitersUsecase{
				txManager:                   txManager,
				investmentWaitersRepository: repository,
				investmentWaitersService:    new(mocks.InvestmentWaitersService),
			}

			_, err := usecase.Cancel(context.Background(), adomain.InvestmentWaitersCancelRequest{
				JWTUser:             jwthelper.User{Id: mockUser1.Id},
				InvestmentWaitersId: lo.ToPtr(waiter.Id),
			})

			wantErr := tt.txErr
			if wantErr == nil {
				wantErr = tt.lockErr
			}
			if wantErr == nil {
				wantErr = tt.advancedQtyErr
			}
			assert.ErrorIs(t, err, wantErr)
			repository.AssertExpectations(t)
			txManager.AssertExpectations(t)
		})
	}
}

func newInvestmentWaitersTxManagerMock() *mocks.TxManager {
	txManager := new(mocks.TxManager)
	txManager.On("WithTx", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
	return txManager
}

func Test_investmentWaitersUsecase_InvestmentPropose(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.InvestmentWaitersInvestmentProposeRequest
	}

	tests := []struct {
		name                        string
		args                        args
		investmentWaitersRepository func() *mocks.InvestmentWaitersRepository
		investmentWaitersService    func() *mocks.InvestmentWaitersService
		wantErr                     bool
		want                        adomain.InvestmentWaitersInvestmentProposeResponse
	}{
		{
			name: "待機者ステータス更新 : PROPOSAL => 正常に更新されます。",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersInvestmentProposeRequest{
					InvestmentWaitersId: lo.ToPtr(int64(1000)),
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
				},
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaiters1.Id).Return(mockInvestmentWaiters1, nil)
				return repo
			},
			investmentWaitersService: func() *mocks.InvestmentWaitersService {
				service := new(mocks.InvestmentWaitersService)
				service.On("ChangeProposalStatus", mock.Anything, mockInvestmentWaiters1, mockUser1.Id).Return(nil)
				return service
			},
			wantErr: false,
			want: adomain.InvestmentWaitersInvestmentProposeResponse{
				InvestmentWaitersResponse: mockInvestmentWaitersResponse1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := investmentWaitersUsecase{
				investmentWaitersRepository: tt.investmentWaitersRepository(),
				investmentWaitersService:    tt.investmentWaitersService(),
			}

			got, err := usecase.InvestmentPropose(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "investmentWaitersUsecase.InvestmentPropose() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "investmentWaitersUsecase.InvestmentPropose() = %v, want %v", got, tt.want)
		})
	}
}

func Test_investmentWaitersUsecase_InvestmentRequest(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.InvestmentWaitersInvestmentRequestRequest
	}

	tests := []struct {
		name                        string
		args                        args
		investmentWaitersRepository func() *mocks.InvestmentWaitersRepository
		userRepository              func() *mocks.UserRepository
		wantErr                     bool
		wantErrIs                   error
		wantErrIsNot                error
		want                        adomain.InvestmentWaitersInvestmentRequestResponse
	}{
		{
			name: "ステータスがWAITINGでない場合 (PROPOSAL) => ErrBadRequest",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersInvestmentRequestRequest{
					InvestmentWaitersId: lo.ToPtr(mockInvestmentWaiters1.Id),
					JWTUser:             jwthelper.User{Id: mockUser1.Id},
					InvestmentWaitersInvestmentRequestBody: adomain.InvestmentWaitersInvestmentRequestBody{
						Qty: 5,
					},
				},
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaiters1.Id).Return(mockInvestmentWaiters1, nil) // Status=PROPOSAL
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				return new(mocks.UserRepository)
			},
			wantErr:   true,
			wantErrIs: cdomain.ErrBadRequest,
			want:      adomain.InvestmentWaitersInvestmentRequestResponse{},
		},
		{
			name: "申請口数が申込口数を超える場合 => ErrBadRequest",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersInvestmentRequestRequest{
					InvestmentWaitersId: lo.ToPtr(mockInvestmentWaitersWaiting.Id),
					JWTUser:             jwthelper.User{Id: mockUser1.Id},
					InvestmentWaitersInvestmentRequestBody: adomain.InvestmentWaitersInvestmentRequestBody{
						Qty: mockInvestmentWaitersWaiting.Qty + 1,
					},
				},
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaitersWaiting.Id).Return(mockInvestmentWaitersWaiting, nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				return new(mocks.UserRepository)
			},
			wantErr:   true,
			wantErrIs: cdomain.ErrBadRequest,
			want:      adomain.InvestmentWaitersInvestmentRequestResponse{},
		},
		{
			name: "投資待機者の取得に失敗した場合 => error",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersInvestmentRequestRequest{
					InvestmentWaitersId: lo.ToPtr(int64(999)),
					JWTUser:             jwthelper.User{Id: mockUser1.Id},
					InvestmentWaitersInvestmentRequestBody: adomain.InvestmentWaitersInvestmentRequestBody{
						Qty: 5,
					},
				},
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", mock.Anything, int64(999)).Return(domain.InvestmentWaiters{}, errors.New("not found"))
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				return new(mocks.UserRepository)
			},
			wantErr: true,
			want:    adomain.InvestmentWaitersInvestmentRequestResponse{},
		},
		{
			name: "ユーザーの取得に失敗した場合 => error",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersInvestmentRequestRequest{
					InvestmentWaitersId: lo.ToPtr(mockInvestmentWaitersWaiting.Id),
					JWTUser:             jwthelper.User{Id: mockUser1.Id},
					InvestmentWaitersInvestmentRequestBody: adomain.InvestmentWaitersInvestmentRequestBody{
						Qty: 5,
					},
				},
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaitersWaiting.Id).Return(mockInvestmentWaitersWaiting, nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				repo := new(mocks.UserRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaitersWaiting.UserId).Return(domain.User{}, errors.New("user not found"))
				return repo
			},
			wantErr: true,
			want:    adomain.InvestmentWaitersInvestmentRequestResponse{},
		},
		{
			// MIR-1068: WAITING_ENDは「繰上げ完了」と「打ち切り」の両方を意味するため、
			// ここでは弾かない。繰上げ実績の有無による可否判定はWaitInvest(LockById配下の
			// InvestmentWaiterCanAdvance)に委ねる(設計書の「既存データの再繰上げは手動操作に
			// 限る」の実現経路)。ここではステータスチェックを通過し、ユーザー取得まで
			// 到達することだけを確認する。wantErrIsNotでErrBadRequest(ステータスチェックで
			// 弾かれた場合のエラー)ではないことまで確認しないと、修正前(弾かれてエラー)でも
			// 修正後(ユーザー取得エラー)でも「エラーかつゼロ値」で一致してしまい、この
			// テストが退行を検出できない(②の敵対的検証で指摘)。
			name: "ステータスがWAITING_ENDの場合はステータスチェックでは弾かれない",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersInvestmentRequestRequest{
					InvestmentWaitersId: lo.ToPtr(mockInvestmentWaiters2.Id),
					JWTUser:             jwthelper.User{Id: mockUser1.Id},
					InvestmentWaitersInvestmentRequestBody: adomain.InvestmentWaitersInvestmentRequestBody{
						Qty: 5,
					},
				},
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				// mockInvestmentWaiters2はQty未設定(ゼロ値)のため、申請口数(request.Qty)が
				// 申込口数(Qty)を超えるチェック(investment_waiters_ucase.goのInvestmentRequest、
				// Qty<request.Qty)に先に弾かれてしまう。このテストの関心事はステータス
				// チェックの通過だけなので、Qtyだけをローカルに上書きする。
				waiter := mockInvestmentWaiters2
				waiter.Qty = 5
				repo.On("GetById", mock.Anything, mockInvestmentWaiters2.Id).Return(waiter, nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				repo := new(mocks.UserRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaiters2.UserId).Return(domain.User{}, errors.New("user not found"))
				return repo
			},
			wantErr:      true,
			wantErrIsNot: cdomain.ErrBadRequest,
			want:         adomain.InvestmentWaitersInvestmentRequestResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			investmentWaitersRepository := tt.investmentWaitersRepository()
			userRepository := tt.userRepository()
			usecase := investmentWaitersUsecase{
				investmentWaitersRepository: investmentWaitersRepository,
				userRepository:              userRepository,
			}

			got, err := usecase.InvestmentRequest(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "InvestmentRequest() error = %v, wantErr %v", err, tt.wantErr)
			if tt.wantErrIs != nil {
				assert.ErrorIs(t, err, tt.wantErrIs)
			}
			if tt.wantErrIsNot != nil {
				assert.NotErrorIs(t, err, tt.wantErrIsNot)
			}
			assert.Equalf(t, tt.want, got, "InvestmentRequest() = %v, want %v", got, tt.want)
			investmentWaitersRepository.AssertExpectations(t)
			userRepository.AssertExpectations(t)
		})
	}
}

func Test_investmentWaitersUsecase_Propose(t *testing.T) {
	uid, _ := uuid.FromString("bddad349-d0cc-4a1d-8843-431d54fcd2c7")

	type args struct {
		ctx     context.Context
		request adomain.InvestmentWaitersProposeRequest
	}

	tests := []struct {
		name                            string
		args                            args
		cfg                             *config.Config
		idGenerator                     func() *mocks.IdGenerator
		investmentWaitersRepository     func() *mocks.InvestmentWaitersRepository
		projectRepository               func() *mocks.ProjectRepository
		userRepository                  func() *mocks.UserRepository
		investmentRequestRepository     func() *mocks.InvestmentRequestRepository
		investmentWaitersService        func() *mocks.InvestmentWaitersService
		investmentWaitHistoryService    func() *mocks.InvestmentWaitHistoryService
		investmentWaitHistoryRepository func() *mocks.InvestmentWaitHistoryRepository
		mailService                     func() *mocks.MailService
		wantErr                         bool
		want                            adomain.InvestmentWaitersProposeResponse
	}{
		{
			name: "投資提案 残口数>=最低投資口数",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersProposeRequest{
					InvestmentWaitersId: lo.ToPtr(int64(1000)),
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
					InvestmentWaitersProposeBody: adomain.InvestmentWaitersProposeBody{
						Qty: 2,
					},
				},
			},
			idGenerator: func() *mocks.IdGenerator {
				gen := new(mocks.IdGenerator)
				gen.On("GenUuidV4").Return(uid)
				return gen
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaiters1.Id).Return(mockInvestmentWaiters1, nil)

				updateInvestmentWaiters := mockInvestmentWaiters1
				updateInvestmentWaiters.Status = domain.PROPOSAL

				repo.On("Update", mock.Anything, updateInvestmentWaiters).Return(int64(0), nil)
				return repo
			},

			investmentRequestRepository: func() *mocks.InvestmentRequestRepository {
				repo := new(mocks.InvestmentRequestRepository)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(request domain.InvestmentRequest) bool {
					ignoreFields := cmpopts.IgnoreFields(domain.InvestmentRequest{}, "RequestId", "Url")
					ignoreType := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

					return cmp.Equal(request, domain.InvestmentRequest{
						InvestmentWaitersId: mockInvestmentWaiters1.Id,
						ProjectId:           mockInvestmentWaiters1.ProjectId,
						IsExpire:            false,
						Email:               mockUser1.Email,
						Qty:                 2,
					}, ignoreFields, ignoreType)
				})).Return(int64(0), nil)
				return repo
			},

			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaiters1.ProjectId).Return(mockProject1, nil)
				return repo
			},

			investmentWaitersService: func() *mocks.InvestmentWaitersService {
				service := new(mocks.InvestmentWaitersService)
				service.On("ChangeProposalStatus", mock.Anything, mockInvestmentWaiters1, mockUser1.Id).Return(nil)
				return service
			},
			userRepository: func() *mocks.UserRepository {
				repo := new(mocks.UserRepository)
				repo.On("GetById", mock.Anything, mockUser1.Id).Return(mockUser1, nil)
				return repo
			},

			mailService: func() *mocks.MailService {
				service := new(mocks.MailService)
				service.On("SendMail", mock.Anything, mock.MatchedBy(func(request cdomain.SendMailRequest) bool {
					ignoreFields := cmpopts.IgnoreFields(cdomain.SendMailRequest{}, "Details")

					return cmp.Equal(request, cdomain.SendMailRequest{
						TemplateType: domain.INVEST_OFFER_TEMPLATE,
						Sender:       "info@energyshares.jp",
						Receivers:    []string{"gildong@email.com"},
						MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
					}, ignoreFields) &&
						request.Details["InvestedQty"] == "5" &&
						request.Details["InvestableQty"] == "2" &&
						request.Details["TotalInvestmentAmount"] == "20,000"
				})).Return(nil)
				return service
			},

			investmentWaitHistoryService: func() *mocks.InvestmentWaitHistoryService {
				service := new(mocks.InvestmentWaitHistoryService)
				service.On("CreateStatusHistory",
					mock.Anything,
					domain.InvestmentWaitersComparer{
						Before: mockInvestmentWaiters1,
						After: domain.InvestmentWaiters{
							Id:        mockInvestmentWaiters1.Id,
							ProjectId: mockInvestmentWaiters1.ProjectId,
							UserId:    mockInvestmentWaiters1.UserId,
							Qty:       mockInvestmentWaiters1.Qty,
							Status:    domain.PROPOSAL,
						},
					},
					domain.USER_HISTORY_PROCESS_OPERATOR,
					mockUser1.Id,
				).Return(nil)
				return service
			},

			investmentWaitHistoryRepository: func() *mocks.InvestmentWaitHistoryRepository {
				repo := new(mocks.InvestmentWaitHistoryRepository)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(investmentHistory domain.InvestmentWaitHistory) bool {
					ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

					return cmp.Equal(investmentHistory, domain.InvestmentWaitHistory{
						InvestmentWaitId: mockInvestmentWaiters1.Id,
						SubstanceType:    string(domain.INVESTMENT_WAIT_HISTORY_MAIL),
						Name:             mockUser1.Name,
						Email:            mockUser1.Email,
						ProcessMethod:    domain.USER_HISTORY_PROCESS_OPERATOR,
					}, ignoreTime)
				})).Return(nil)

				return repo
			},
			wantErr: false,
			want: adomain.InvestmentWaitersProposeResponse{
				InvestmentWaitersResponse: mockInvestmentWaitersResponse1,
			},
		},
		{
			name: "投資提案 提案口数>申込口数の場合はエラー",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersProposeRequest{
					InvestmentWaitersId: lo.ToPtr(int64(1000)),
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
					InvestmentWaitersProposeBody: adomain.InvestmentWaitersProposeBody{
						Qty: 6,
					},
				},
			},
			idGenerator: func() *mocks.IdGenerator {
				return new(mocks.IdGenerator)
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaiters1.Id).Return(mockInvestmentWaiters1, nil)
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaiters1.ProjectId).Return(mockProject1, nil)
				return repo
			},
			investmentRequestRepository: func() *mocks.InvestmentRequestRepository {
				return new(mocks.InvestmentRequestRepository)
			},
			investmentWaitersService: func() *mocks.InvestmentWaitersService {
				return new(mocks.InvestmentWaitersService)
			},
			userRepository: func() *mocks.UserRepository {
				return new(mocks.UserRepository)
			},
			mailService: func() *mocks.MailService {
				return new(mocks.MailService)
			},
			investmentWaitHistoryService: func() *mocks.InvestmentWaitHistoryService {
				return new(mocks.InvestmentWaitHistoryService)
			},
			investmentWaitHistoryRepository: func() *mocks.InvestmentWaitHistoryRepository {
				return new(mocks.InvestmentWaitHistoryRepository)
			},
			wantErr: true,
			want:    adomain.InvestmentWaitersProposeResponse{},
		},
		{
			name: "投資提案 残口数<最低投資口数",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitersProposeRequest{
					InvestmentWaitersId: lo.ToPtr(int64(1000)),
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
					InvestmentWaitersProposeBody: adomain.InvestmentWaitersProposeBody{
						Qty: 1,
					},
				},
			},
			idGenerator: func() *mocks.IdGenerator {
				gen := new(mocks.IdGenerator)
				gen.On("GenUuidV4").Return(uid)
				return gen
			},
			investmentWaitersRepository: func() *mocks.InvestmentWaitersRepository {
				repo := new(mocks.InvestmentWaitersRepository)
				repo.On("GetById", mock.Anything, mockInvestmentWaiters1.Id).Return(mockInvestmentWaiters1, nil)

				updateInvestmentWaiters := mockInvestmentWaiters1
				updateInvestmentWaiters.Status = domain.PROPOSAL

				repo.On("Update", mock.Anything, updateInvestmentWaiters).Return(int64(0), nil)
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				project := mockProject1
				project.RemainingQty = 1
				project.MinShareQty = 10
				repo.On("GetById", mock.Anything, mockInvestmentWaiters1.ProjectId).Return(project, nil)
				return repo
			},

			investmentRequestRepository: func() *mocks.InvestmentRequestRepository {
				repo := new(mocks.InvestmentRequestRepository)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(request domain.InvestmentRequest) bool {
					ignoreFields := cmpopts.IgnoreFields(domain.InvestmentRequest{}, "RequestId", "Url")
					ignoreType := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

					return cmp.Equal(request, domain.InvestmentRequest{
						InvestmentWaitersId: mockInvestmentWaiters1.Id,
						ProjectId:           mockInvestmentWaiters1.ProjectId,
						IsExpire:            false,
						Email:               mockUser1.Email,
						Qty:                 1,
					}, ignoreFields, ignoreType)
				})).Return(int64(0), nil)
				return repo
			},

			investmentWaitersService: func() *mocks.InvestmentWaitersService {
				service := new(mocks.InvestmentWaitersService)
				return service
			},

			userRepository: func() *mocks.UserRepository {
				repo := new(mocks.UserRepository)
				repo.On("GetById", mock.Anything, mockUser1.Id).Return(mockUser1, nil)
				return repo
			},

			mailService: func() *mocks.MailService {
				service := new(mocks.MailService)
				service.On("SendMail", mock.Anything, mock.MatchedBy(func(request cdomain.SendMailRequest) bool {
					ignoreFields := cmpopts.IgnoreFields(cdomain.SendMailRequest{}, "Details")

					return cmp.Equal(request, cdomain.SendMailRequest{
						TemplateType: domain.INVEST_OFFER_TEMPLATE,
						Sender:       "info@energyshares.jp",
						Receivers:    []string{"gildong@email.com"},
						MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
					}, ignoreFields) &&
						request.Details["InvestedQty"] == "5" &&
						request.Details["InvestableQty"] == "1" &&
						request.Details["TotalInvestmentAmount"] == "10,000"
				})).Return(nil)
				return service
			},

			investmentWaitHistoryService: func() *mocks.InvestmentWaitHistoryService {
				service := new(mocks.InvestmentWaitHistoryService)
				service.On("CreateStatusHistory",
					mock.Anything,
					domain.InvestmentWaitersComparer{
						Before: mockInvestmentWaiters1,
						After: domain.InvestmentWaiters{
							Id:        mockInvestmentWaiters1.Id,
							ProjectId: mockInvestmentWaiters1.ProjectId,
							UserId:    mockInvestmentWaiters1.UserId,
							Qty:       mockInvestmentWaiters1.Qty,
							Status:    domain.PROPOSAL,
						},
					},
					domain.USER_HISTORY_PROCESS_OPERATOR,
					mockUser1.Id,
				).Return(nil)
				return service
			},

			investmentWaitHistoryRepository: func() *mocks.InvestmentWaitHistoryRepository {
				repo := new(mocks.InvestmentWaitHistoryRepository)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(investmentHistory domain.InvestmentWaitHistory) bool {
					ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

					return cmp.Equal(investmentHistory, domain.InvestmentWaitHistory{
						InvestmentWaitId: mockInvestmentWaiters1.Id,
						SubstanceType:    string(domain.INVESTMENT_WAIT_HISTORY_MAIL),
						Name:             mockUser1.Name,
						Email:            mockUser1.Email,
						ProcessMethod:    domain.USER_HISTORY_PROCESS_OPERATOR,
					}, ignoreTime)
				})).Return(nil)
				return repo
			},

			wantErr: false,
			want: adomain.InvestmentWaitersProposeResponse{
				InvestmentWaitersResponse: mockInvestmentWaitersResponse1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := investmentWaitersUsecase{
				idGenerator:                     tt.idGenerator(),
				investmentWaitersRepository:     tt.investmentWaitersRepository(),
				investmentRequestRepository:     tt.investmentRequestRepository(),
				projectRepository:               tt.projectRepository(),
				userRepository:                  tt.userRepository(),
				investmentWaitersService:        tt.investmentWaitersService(),
				mailService:                     tt.mailService(),
				investmentWaitHistoryService:    tt.investmentWaitHistoryService(),
				investmentWaitHistoryRepository: tt.investmentWaitHistoryRepository(),
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
			}

			got, err := usecase.Propose(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "investmentWaitersUsecase.Propose() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "investmentWaitersUsecase.Propose() = %v, want %v", got, tt.want)
		})
	}
}
