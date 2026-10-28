package repayment_dividend

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/mock"

	"github.com/illuminarean/energy-share-core-api/controllers"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"

	"github.com/stretchr/testify/assert"
)

var repaymentDividendTestJST = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	return loc
}()

func mustParseJSONTimeJST(value string) jsonutils.JSONTime {
	parsed, err := jsonutils.ParseJSONTimeJST(time.DateOnly, value)
	if err != nil {
		panic(err)
	}
	return parsed
}

// hasLimitOption はWithLimit(...)が指定された呼び出しにだけマッチする。
// mock.Anythingは呼び出し側の引数が足りない(可変長オプションの数が少ない)
// 場合でも常にマッチしてしまうため、publishLegalRecordEvent(5引数)と
// countUnprocessedShares(2引数)のどちらの呼び出しかをテストのモックで
// 区別するには使えない。実際に値を検査するマッチャーにすることで、
// 引数の数が足りない呼び出しに対してはpanic経由で不一致として扱われる。
func hasLimitOption() interface{} {
	return mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
		opt := cdomain.ShareQueryOptionFunc()
		field(&opt)
		return opt.Limit() == ExecuteInvestmentBatchSize
	})
}

// hasSuccessInvestOnlyOption は、WithInvestStatuses(SUCCESS_INVEST)だけを
// 指定した呼び出しにマッチする。SUCCESS_INVESTでの絞り込みが外れたり
// 他のステータスが混ざったりしていないことを、値まで見て検証する。
func hasSuccessInvestOnlyOption() interface{} {
	return mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
		opt := cdomain.ShareQueryOptionFunc()
		field(&opt)
		statuses := opt.InvestStatuses()
		return len(statuses) == 1 && statuses[0] == string(domain.SUCCESS_INVEST)
	})
}

func Test_repaymentDividendUsecase_ExecuteInvestment(t *testing.T) {
	type fields struct {
		projectInvestmentExecuteRepository func() *mocks.ProjectInvestmentExecuteRepository
		shareRepository                    func() *mocks.ShareRepository
		userAccountRepository              func() *mocks.UserAccountRepository
		projectRepository                  func() *mocks.ProjectRepository
	}
	type args struct {
		ctx     context.Context
		request adomain.ExecuteInvestmentRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    adomain.ExecuteInvestmentResponse
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "投資実行(初回。実行レコードを新規作成する)",
			fields: fields{
				projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
					repo := new(mocks.ProjectInvestmentExecuteRepository)
					// 初回は実行レコードが無いので、まず not found が返る。
					repo.On("GetByProjectId", mock.Anything, mockProject1.Id).
						Return(domain.ProjectInvestmentExecute{}, errors.Wrap(cdomain.ErrItemNotFound, "not found")).Once()
					// Createはxormが採番した本物のIdを書き戻して返す(MIR-1204の
					// 回帰テスト)。Id=0のまま返すモックだと、直後のUpdateが
					// 「Id!=0」を検査しないかぎり気付けずに素通りしてしまう。
					repo.On("Create", mock.Anything, mock.Anything).
						Return(domain.ProjectInvestmentExecute{Id: 1}, nil)
					// 残数が0になるため、COMPLETEDへの更新が呼ばれる。Id=0(未採番)
					// のまま渡ると本番ではWHERE id=0で無言の0件更新になるため、
					// ここでも検査する。
					repo.On("Update", mock.Anything, mock.MatchedBy(func(e domain.ProjectInvestmentExecute) bool {
						return e.Id != 0 && e.ExecuteStatus == domain.COMPLETED
					})).Return(nil)
					return repo

				},
				shareRepository: func() *mocks.ShareRepository {
					repo := new(mocks.ShareRepository)
					// publishLegalRecordEvent: SUCCESS_INVESTのみ・上限あり(5引数)。
					repo.On("GetAll", context.Background(),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							return *opt.ProjectId() == "2000"
						}),
						hasSuccessInvestOnlyOption(),
						mock.Anything,
						mock.Anything,
						hasLimitOption(),
					).Return(
						[]domain.Share{
							{
								InvestStatus: domain.SUCCESS_INVEST,
							},
						}, nil,
					)
					// countUnprocessedShares: SUCCESS_INVESTのみ・上限なし(2引数)。
					// 処理後は残り0件。
					repo.On("GetAll", context.Background(), mock.Anything, hasSuccessInvestOnlyOption()).
						Return([]domain.Share{}, nil)
					return repo
				},
				userAccountRepository: func() *mocks.UserAccountRepository {
					repo := new(mocks.UserAccountRepository)
					repo.On("GetByUserId", mock.Anything, mock.Anything).Return(domain.UserAccount{}, nil)
					return repo
				},
				projectRepository: func() *mocks.ProjectRepository {
					repo := new(mocks.ProjectRepository)
					repo.On("TryLockById", mock.Anything, mockProject1.Id).Return(mockProject1, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.ExecuteInvestmentRequest{
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
					ProjectId: 2000,
					ExecuteInvestmentBody: adomain.ExecuteInvestmentBody{
						Amount:      1000,
						ExecuteDate: "2021-01-01",
					},
				},
			},
			want: adomain.ExecuteInvestmentResponse{
				ProjectInvestmentExecute: domain.ProjectInvestmentExecute{
					Id:            1,
					ExecuteStatus: domain.COMPLETED,
				},
				ProcessedCount: 1,
				RemainingCount: 0,
			},
			wantErr: assert.NoError,
		},
		{
			// MIR-1204の回帰テスト: 分割実行(同じボタンの複数回押下)で、
			// 既存の実行レコードを再利用し、新規のCreateが呼ばれないこと、
			// 未処理のshareだけが続けて処理されることを確認する。
			name: "投資実行(2回目。既存の実行レコードを再利用し、続きを処理する)",
			fields: fields{
				projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
					repo := new(mocks.ProjectInvestmentExecuteRepository)
					existing := domain.ProjectInvestmentExecute{
						Id:        1,
						ProjectId: mockProject1.Id,
						Amount:    1000,
						// まだ全件処理し終えていないので、1回目の時点ではSCHEDULEDの
						// まま(MIR-1204)。今回の呼び出しで残数が0になるため、
						// COMPLETEDへの更新が呼ばれる。
						ExecuteStatus: domain.SCHEDULED,
						ExecuteDate:   mustParseJSONTimeJST("2021-01-01"),
					}
					repo.On("GetByProjectId", mock.Anything, mockProject1.Id).Return(existing, nil)
					// Createが呼ばれたら(=重複作成)、モックが未設定のため
					// panicしてテストが失敗する。それ自体が回帰の検知になる。
					repo.On("Update", mock.Anything, mock.MatchedBy(func(e domain.ProjectInvestmentExecute) bool {
						return e.ExecuteStatus == domain.COMPLETED
					})).Return(nil)
					return repo
				},
				shareRepository: func() *mocks.ShareRepository {
					repo := new(mocks.ShareRepository)
					repo.On("GetAll", context.Background(),
						mock.Anything, hasSuccessInvestOnlyOption(), mock.Anything, mock.Anything, hasLimitOption(),
					).Return(
						[]domain.Share{{InvestStatus: domain.SUCCESS_INVEST}}, nil,
					)
					repo.On("GetAll", context.Background(), mock.Anything, hasSuccessInvestOnlyOption()).
						Return([]domain.Share{}, nil)
					return repo
				},
				userAccountRepository: func() *mocks.UserAccountRepository {
					repo := new(mocks.UserAccountRepository)
					repo.On("GetByUserId", mock.Anything, mock.Anything).Return(domain.UserAccount{}, nil)
					return repo
				},
				projectRepository: func() *mocks.ProjectRepository {
					repo := new(mocks.ProjectRepository)
					repo.On("TryLockById", mock.Anything, mockProject1.Id).Return(mockProject1, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.ExecuteInvestmentRequest{
					JWTUser:   jwthelper.User{Id: mockUser1.Id},
					ProjectId: 2000,
					ExecuteInvestmentBody: adomain.ExecuteInvestmentBody{
						Amount:      1000,
						ExecuteDate: "2021-01-01",
					},
				},
			},
			want: adomain.ExecuteInvestmentResponse{
				ProjectInvestmentExecute: domain.ProjectInvestmentExecute{
					Id:            1,
					ProjectId:     mockProject1.Id,
					Amount:        1000,
					ExecuteStatus: domain.COMPLETED,
					ExecuteDate:   mustParseJSONTimeJST("2021-01-01"),
				},
				ProcessedCount: 1,
				RemainingCount: 0,
			},
			wantErr: assert.NoError,
		},
		{
			// MIR-1204の回帰テスト: 実行日は法定帳簿へ出資日として送られる
			// ため、既存の実行レコードと異なる日付を黙って使う・上書きする
			// ことはせず、エラーにする。
			name: "投資実行(実行日が既存の実行レコードと異なる場合はエラー)",
			fields: fields{
				projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
					repo := new(mocks.ProjectInvestmentExecuteRepository)
					existing := domain.ProjectInvestmentExecute{
						Id:            1,
						ProjectId:     mockProject1.Id,
						Amount:        1000,
						ExecuteStatus: domain.COMPLETED,
						ExecuteDate:   mustParseJSONTimeJST("2021-01-01"),
					}
					repo.On("GetByProjectId", mock.Anything, mockProject1.Id).Return(existing, nil)
					return repo
				},
				shareRepository: func() *mocks.ShareRepository {
					// publishLegalRecordEventまで到達しないため、モックは
					// 未設定のままでよい(呼ばれたらpanicして検知する)。
					return new(mocks.ShareRepository)
				},
				userAccountRepository: func() *mocks.UserAccountRepository {
					return new(mocks.UserAccountRepository)
				},
				projectRepository: func() *mocks.ProjectRepository {
					repo := new(mocks.ProjectRepository)
					repo.On("TryLockById", mock.Anything, mockProject1.Id).Return(mockProject1, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.ExecuteInvestmentRequest{
					JWTUser:   jwthelper.User{Id: mockUser1.Id},
					ProjectId: 2000,
					ExecuteInvestmentBody: adomain.ExecuteInvestmentBody{
						Amount:      1000,
						ExecuteDate: "2021-01-02", // 既存レコードの2021-01-01と異なる
					},
				},
			},
			want: adomain.ExecuteInvestmentResponse{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if !assert.ErrorIs(t, err, cdomain.ErrExecuteInvestmentDateMismatch, i...) {
					return false
				}
				return assert.Equal(t,
					"このファンドは2021-01-01で出資実行を開始しています。残りを処理するには、同じ日付を指定してください。",
					err.Error())
			},
		},
		{
			// MIR-1204の回帰テスト: Amountも法定帳簿・投資額の判定に使われる
			// ため、既存の実行レコードと異なる金額を黙って使う・上書きする
			// ことはせず、エラーにする(ExecuteDateと同じ扱い)。
			name: "投資実行(金額が既存の実行レコードと異なる場合はエラー)",
			fields: fields{
				projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
					repo := new(mocks.ProjectInvestmentExecuteRepository)
					existing := domain.ProjectInvestmentExecute{
						Id:            1,
						ProjectId:     mockProject1.Id,
						Amount:        1000,
						ExecuteStatus: domain.COMPLETED,
						ExecuteDate:   mustParseJSONTimeJST("2021-01-01"),
					}
					repo.On("GetByProjectId", mock.Anything, mockProject1.Id).Return(existing, nil)
					return repo
				},
				shareRepository: func() *mocks.ShareRepository {
					// publishLegalRecordEventまで到達しないため、モックは
					// 未設定のままでよい(呼ばれたらpanicして検知する)。
					return new(mocks.ShareRepository)
				},
				userAccountRepository: func() *mocks.UserAccountRepository {
					return new(mocks.UserAccountRepository)
				},
				projectRepository: func() *mocks.ProjectRepository {
					repo := new(mocks.ProjectRepository)
					repo.On("TryLockById", mock.Anything, mockProject1.Id).Return(mockProject1, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.ExecuteInvestmentRequest{
					JWTUser:   jwthelper.User{Id: mockUser1.Id},
					ProjectId: 2000,
					ExecuteInvestmentBody: adomain.ExecuteInvestmentBody{
						Amount:      2000, // 既存レコードの1000と異なる
						ExecuteDate: "2021-01-01",
					},
				},
			},
			want: adomain.ExecuteInvestmentResponse{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if !assert.ErrorIs(t, err, cdomain.ErrExecuteInvestmentAmountMismatch, i...) {
					return false
				}
				return assert.Equal(t,
					"このファンドは金額1000円で出資実行を開始しています。残りを処理するには、同じ金額を指定してください。",
					err.Error())
			},
		},
		{
			// MIR-1204の回帰テスト: 上限件数(ここでは2件としてテストする)を
			// 超えるshareがあっても、超過分は処理せず残数として返す。
			name: "投資実行(上限件数までしか処理せず、残数を返す)",
			fields: fields{
				projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
					repo := new(mocks.ProjectInvestmentExecuteRepository)
					repo.On("GetByProjectId", mock.Anything, mockProject1.Id).
						Return(domain.ProjectInvestmentExecute{}, errors.Wrap(cdomain.ErrItemNotFound, "not found")).Once()
					repo.On("Create", mock.Anything, mock.Anything).Return(domain.ProjectInvestmentExecute{}, nil)
					return repo
				},
				shareRepository: func() *mocks.ShareRepository {
					repo := new(mocks.ShareRepository)
					// 上限(5引数の呼び出し)で2件だけ返す想定。
					repo.On("GetAll", context.Background(),
						mock.Anything, hasSuccessInvestOnlyOption(), mock.Anything, mock.Anything, hasLimitOption(),
					).Return(
						[]domain.Share{
							{InvestStatus: domain.SUCCESS_INVEST},
							{InvestStatus: domain.SUCCESS_INVEST},
						}, nil,
					)
					// 処理後もまだ3件残っている想定(2引数の呼び出し)。
					repo.On("GetAll", context.Background(), mock.Anything, hasSuccessInvestOnlyOption()).
						Return([]domain.Share{{}, {}, {}}, nil)
					return repo
				},
				userAccountRepository: func() *mocks.UserAccountRepository {
					repo := new(mocks.UserAccountRepository)
					repo.On("GetByUserId", mock.Anything, mock.Anything).Return(domain.UserAccount{}, nil)
					return repo
				},
				projectRepository: func() *mocks.ProjectRepository {
					repo := new(mocks.ProjectRepository)
					repo.On("TryLockById", mock.Anything, mockProject1.Id).Return(mockProject1, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.ExecuteInvestmentRequest{
					JWTUser:   jwthelper.User{Id: mockUser1.Id},
					ProjectId: 2000,
					ExecuteInvestmentBody: adomain.ExecuteInvestmentBody{
						Amount:      1000,
						ExecuteDate: "2021-01-01",
					},
				},
			},
			want: adomain.ExecuteInvestmentResponse{
				ProcessedCount: 2,
				RemainingCount: 3,
			},
			wantErr: assert.NoError,
		},
		{
			// MIR-1204の回帰テスト: 同じプロジェクトへの実行が他のリクエストで
			// 進行中の場合、ロックを待たずにErrItemConflictを返す(NOWAIT)。
			name: "投資実行(他のリクエストが実行中の場合はロックを待たずにエラー)",
			fields: fields{
				projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
					return new(mocks.ProjectInvestmentExecuteRepository)
				},
				shareRepository: func() *mocks.ShareRepository {
					return new(mocks.ShareRepository)
				},
				userAccountRepository: func() *mocks.UserAccountRepository {
					return new(mocks.UserAccountRepository)
				},
				projectRepository: func() *mocks.ProjectRepository {
					repo := new(mocks.ProjectRepository)
					repo.On("TryLockById", mock.Anything, mockProject1.Id).
						Return(domain.Project{}, errors.Wrap(cdomain.ErrItemConflict, "locked"))
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.ExecuteInvestmentRequest{
					JWTUser:   jwthelper.User{Id: mockUser1.Id},
					ProjectId: 2000,
					ExecuteInvestmentBody: adomain.ExecuteInvestmentBody{
						Amount:      1000,
						ExecuteDate: "2021-01-01",
					},
				},
			},
			want: adomain.ExecuteInvestmentResponse{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, cdomain.ErrItemConflict, i...)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := repaymentDividendUsecase{
				projectInvestmentExecuteRepository: tt.fields.projectInvestmentExecuteRepository(),
				shareRepository:                    tt.fields.shareRepository(),
				userAccountRepository:              tt.fields.userAccountRepository(),
				projectRepository:                  tt.fields.projectRepository(),
				eventMessagesService: func() *mocks.EventMessageService {
					ems := new(mocks.EventMessageService)
					ems.On("RecordEvent", mock.Anything, mock.Anything, mock.Anything).Return(domain.EventMessage{}, nil)
					ems.On("SendRecordedEvent", mock.Anything, mock.Anything).Return(true, nil)
					return ems
				}(),
				calculateCustomerLedgerBalance: func(ctx context.Context, userId int64) (int64, error) {
					return 0, nil
				},
				withdrawalUserTransaction: func(ctx context.Context, userId, balance, amount, projectId int64, projectType domain.PROJECT_TYPE, accountNumber string, transactionDetailType domain.USER_TRANSACTION_DETAIL_TYPE, transactionDate jsonutils.JSONTime) error {
					return nil
				},
				updateShareToOperation: func(ctx context.Context, share *domain.Share, operatorId int64) error {
					share.InvestStatus = domain.OPERATION_INVEST
					return nil
				},
				timeNow: func() time.Time {
					return time.Date(2024, 12, 1, 10, 0, 0, 0, repaymentDividendTestJST)
				},
			}
			got, err := r.ExecuteInvestment(tt.args.ctx, tt.args.request)
			if !tt.wantErr(t, err, fmt.Sprintf("ExecuteInvestment(%v, %v)", tt.args.ctx, tt.args.request)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ExecuteInvestment(%v, %v)", tt.args.ctx, tt.args.request)
		})
	}
}

// MIR-1202: 出資成立時にEVENT_FUND_TRANSFERを送るのを止めた(法定帳簿アプリのいずれも
// このイベント種別を処理しておらず、RecordEvent+SendRecordedEvent1回ぶん(INSERT→
// Kafka送信→UPDATE)の所要時間だけが本番のタイムアウトを招いていたため)。
// EVENT_TRANSACTIONは従来どおり発行され続けることをあわせて確認する(fund_transfers
// の帳簿データはこちら経由で作られるため、片方だけ消えると帳簿側が空になる)。
func Test_repaymentDividendUsecase_ExecuteInvestment_DoesNotPublishEventFundTransfer(t *testing.T) {
	projectInvestmentExecuteRepository := new(mocks.ProjectInvestmentExecuteRepository)
	projectInvestmentExecuteRepository.On("GetByProjectId", mock.Anything, mockProject1.Id).
		Return(domain.ProjectInvestmentExecute{}, errors.Wrap(cdomain.ErrItemNotFound, "not found")).Once()
	projectInvestmentExecuteRepository.On("Create", mock.Anything, mock.Anything).Return(domain.ProjectInvestmentExecute{Id: 1}, nil)
	projectInvestmentExecuteRepository.On("Update", mock.Anything, mock.Anything).Return(nil)

	shareRepository := new(mocks.ShareRepository)
	shareRepository.On("GetAll", context.Background(),
		mock.Anything, hasSuccessInvestOnlyOption(), mock.Anything, mock.Anything, hasLimitOption(),
	).Return(
		[]domain.Share{{InvestStatus: domain.SUCCESS_INVEST, OwnerId: mockUser1.Id}}, nil,
	)
	shareRepository.On("GetAll", context.Background(), mock.Anything, hasSuccessInvestOnlyOption()).
		Return([]domain.Share{}, nil)

	userAccountRepository := new(mocks.UserAccountRepository)
	userAccountRepository.On("GetByUserId", mock.Anything, mock.Anything).Return(domain.UserAccount{}, nil)

	projectRepository := new(mocks.ProjectRepository)
	projectRepository.On("TryLockById", mock.Anything, mockProject1.Id).Return(mockProject1, nil)

	eventMessagesService := new(mocks.EventMessageService)
	eventMessagesService.On("RecordEvent", mock.Anything, domain.EVENT_TRANSACTION, mock.Anything).Return(domain.EventMessage{}, nil)
	eventMessagesService.On("SendRecordedEvent", mock.Anything, mock.Anything).Return(true, nil)

	r := repaymentDividendUsecase{
		projectInvestmentExecuteRepository: projectInvestmentExecuteRepository,
		shareRepository:                    shareRepository,
		userAccountRepository:              userAccountRepository,
		projectRepository:                  projectRepository,
		eventMessagesService:               eventMessagesService,
		calculateCustomerLedgerBalance: func(ctx context.Context, userId int64) (int64, error) {
			return 0, nil
		},
		withdrawalUserTransaction: func(ctx context.Context, userId, balance, amount, projectId int64, projectType domain.PROJECT_TYPE, accountNumber string, transactionDetailType domain.USER_TRANSACTION_DETAIL_TYPE, transactionDate jsonutils.JSONTime) error {
			return nil
		},
		updateShareToOperation: func(ctx context.Context, share *domain.Share, operatorId int64) error {
			return nil
		},
		timeNow: func() time.Time {
			return time.Date(2024, 12, 1, 10, 0, 0, 0, repaymentDividendTestJST)
		},
	}

	_, err := r.ExecuteInvestment(context.Background(), adomain.ExecuteInvestmentRequest{
		JWTUser:   jwthelper.User{Id: mockUser1.Id},
		ProjectId: mockProject1.Id,
		ExecuteInvestmentBody: adomain.ExecuteInvestmentBody{
			Amount:      1000,
			ExecuteDate: "2021-01-01",
		},
	})

	assert.NoError(t, err)
	eventMessagesService.AssertCalled(t, "RecordEvent", mock.Anything, domain.EVENT_TRANSACTION, mock.Anything)
	eventMessagesService.AssertNotCalled(t, "RecordEvent", mock.Anything, domain.EVENT_FUND_TRANSFER, mock.Anything)
	eventMessagesService.AssertNumberOfCalls(t, "RecordEvent", 1)
}

// MIR-1202: calculateCustomerLedgerBalanceで算出したbalanceを、withdrawalUserTransaction
// (WithdrawalUserTransactionWithBalance)へそのまま渡すことを確認する。withdrawalUserTransaction
// 内で同じ値をもう一度算出させると、user_asset_view(本番実測で1回約58ms)への問い合わせが
// 1口ごとに2回になり二重計算になるため、渡された値だけで済んでいることを固定する。
func Test_repaymentDividendUsecase_ExecuteInvestment_PassesCalculatedBalanceToWithdrawalUserTransaction(t *testing.T) {
	projectInvestmentExecuteRepository := new(mocks.ProjectInvestmentExecuteRepository)
	projectInvestmentExecuteRepository.On("GetByProjectId", mock.Anything, mockProject1.Id).
		Return(domain.ProjectInvestmentExecute{}, errors.Wrap(cdomain.ErrItemNotFound, "not found")).Once()
	projectInvestmentExecuteRepository.On("Create", mock.Anything, mock.Anything).Return(domain.ProjectInvestmentExecute{Id: 1}, nil)
	projectInvestmentExecuteRepository.On("Update", mock.Anything, mock.Anything).Return(nil)

	shareRepository := new(mocks.ShareRepository)
	shareRepository.On("GetAll", context.Background(),
		mock.Anything, hasSuccessInvestOnlyOption(), mock.Anything, mock.Anything, hasLimitOption(),
	).Return(
		[]domain.Share{{InvestStatus: domain.SUCCESS_INVEST, OwnerId: mockUser1.Id, Amount: 5000}}, nil,
	)
	shareRepository.On("GetAll", context.Background(), mock.Anything, hasSuccessInvestOnlyOption()).
		Return([]domain.Share{}, nil)

	userAccountRepository := new(mocks.UserAccountRepository)
	userAccountRepository.On("GetByUserId", mock.Anything, mock.Anything).Return(domain.UserAccount{}, nil)

	projectRepository := new(mocks.ProjectRepository)
	projectRepository.On("TryLockById", mock.Anything, mockProject1.Id).Return(mockProject1, nil)

	eventMessagesService := new(mocks.EventMessageService)
	eventMessagesService.On("RecordEvent", mock.Anything, mock.Anything, mock.Anything).Return(domain.EventMessage{}, nil)
	eventMessagesService.On("SendRecordedEvent", mock.Anything, mock.Anything).Return(true, nil)

	const calculatedBalance = int64(123456)
	var calculateCustomerLedgerBalanceCalls int
	var gotBalance int64
	var withdrawalUserTransactionCalls int

	r := repaymentDividendUsecase{
		projectInvestmentExecuteRepository: projectInvestmentExecuteRepository,
		shareRepository:                    shareRepository,
		userAccountRepository:              userAccountRepository,
		projectRepository:                  projectRepository,
		eventMessagesService:               eventMessagesService,
		calculateCustomerLedgerBalance: func(ctx context.Context, userId int64) (int64, error) {
			calculateCustomerLedgerBalanceCalls++
			return calculatedBalance, nil
		},
		withdrawalUserTransaction: func(ctx context.Context, userId, balance, amount, projectId int64, projectType domain.PROJECT_TYPE, accountNumber string, transactionDetailType domain.USER_TRANSACTION_DETAIL_TYPE, transactionDate jsonutils.JSONTime) error {
			withdrawalUserTransactionCalls++
			gotBalance = balance
			return nil
		},
		updateShareToOperation: func(ctx context.Context, share *domain.Share, operatorId int64) error {
			return nil
		},
		timeNow: func() time.Time {
			return time.Date(2024, 12, 1, 10, 0, 0, 0, repaymentDividendTestJST)
		},
	}

	_, err := r.ExecuteInvestment(context.Background(), adomain.ExecuteInvestmentRequest{
		JWTUser:   jwthelper.User{Id: mockUser1.Id},
		ProjectId: mockProject1.Id,
		ExecuteInvestmentBody: adomain.ExecuteInvestmentBody{
			Amount:      1000,
			ExecuteDate: "2021-01-01",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, calculateCustomerLedgerBalanceCalls, "1口あたりbalanceの算出は1回だけであること")
	assert.Equal(t, 1, withdrawalUserTransactionCalls)
	assert.Equal(t, calculatedBalance, gotBalance, "算出済みのbalanceがそのまま渡されること")
}

func Test_repaymentDividendUsecase_CreateRepaymentDividendSchedules(t *testing.T) {
	type fields struct {
		repaymentDividendScheduleRepository func() *mocks.RepaymentDividendScheduleRepository
		calculateRepaymentDividendSchedule  func(ctx context.Context, request adomain.CreateRepaymentDividendRequest) (domain.RepaymentDividendSchedule, int64, error)
		createRepaymentDividendForInvestors func(ctx context.Context, sched domain.RepaymentDividendSchedule) (int64, error)
	}
	type args struct {
		ctx     context.Context
		request adomain.CreateRepaymentDividendRequest
	}
	tests := []struct {
		name                       string
		fields                     fields
		args                       args
		want                       error
		wantErr                    assert.ErrorAssertionFunc
		wantUnmatchedInvestmentCnt int64
	}{
		{
			name: "スケジュール作成",
			fields: fields{
				calculateRepaymentDividendSchedule: func(ctx context.Context, request adomain.CreateRepaymentDividendRequest) (domain.RepaymentDividendSchedule, int64, error) {
					return domain.RepaymentDividendSchedule{
						ProjectId:       2000,
						RepaymentAmount: 1000,
						DividendAmount:  100,
					}, 0, nil
				},

				repaymentDividendScheduleRepository: func() *mocks.RepaymentDividendScheduleRepository {
					repo := new(mocks.RepaymentDividendScheduleRepository)
					repo.On("Create", mock.Anything, mock.Anything).Return(domain.RepaymentDividendSchedule{}, nil)
					repo.On("Update", mock.Anything, mock.Anything).Return(nil)
					return repo
				},

				createRepaymentDividendForInvestors: func(ctx context.Context, sched domain.RepaymentDividendSchedule) (int64, error) {
					return 0, nil
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.CreateRepaymentDividendRequest{
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
					ProjectId: 2000,
					CreateRepaymentDividendBody: adomain.CreateRepaymentDividendBody{
						RepaymentAmount: 1000,
						DividendAmount:  0,
						BrokerFee:       0,
						ReturnRate:      0,
						DecideDate:      "2024-12-01",
						ExecuteDate:     "2024-12-11",
					},
				},
			},
			wantErr:                    assert.NoError,
			wantUnmatchedInvestmentCnt: 0,
		},
		{
			// MIR-1239: calculateRepaymentDividendScheduleが返す紛れ込み件数
			// (第2戻り値)が、そのままレスポンスのUnmatchedInvestmentCountに
			// 反映されることを確認する。
			name: "紛れ込みがあってもスケジュールは作成され、件数がレスポンスに載る",
			fields: fields{
				calculateRepaymentDividendSchedule: func(ctx context.Context, request adomain.CreateRepaymentDividendRequest) (domain.RepaymentDividendSchedule, int64, error) {
					return domain.RepaymentDividendSchedule{
						ProjectId:       2000,
						RepaymentAmount: 1000,
						DividendAmount:  100,
					}, 3, nil
				},

				repaymentDividendScheduleRepository: func() *mocks.RepaymentDividendScheduleRepository {
					repo := new(mocks.RepaymentDividendScheduleRepository)
					repo.On("Create", mock.Anything, mock.Anything).Return(domain.RepaymentDividendSchedule{}, nil)
					repo.On("Update", mock.Anything, mock.Anything).Return(nil)
					return repo
				},

				createRepaymentDividendForInvestors: func(ctx context.Context, sched domain.RepaymentDividendSchedule) (int64, error) {
					return 0, nil
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.CreateRepaymentDividendRequest{
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
					ProjectId: 2000,
					CreateRepaymentDividendBody: adomain.CreateRepaymentDividendBody{
						RepaymentAmount: 1000,
						DividendAmount:  0,
						BrokerFee:       0,
						ReturnRate:      0,
						DecideDate:      "2024-12-01",
						ExecuteDate:     "2024-12-11",
					},
				},
			},
			wantErr:                    assert.NoError,
			wantUnmatchedInvestmentCnt: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := repaymentDividendUsecase{
				calculateRepaymentDividendSchedule:  tt.fields.calculateRepaymentDividendSchedule,
				repaymentDividendScheduleRepository: tt.fields.repaymentDividendScheduleRepository(),
				createRepaymentDividendForInvestors: tt.fields.createRepaymentDividendForInvestors,
			}

			response, err := r.CreateRepaymentDividendSchedules(tt.args.ctx, tt.args.request)
			if !tt.wantErr(t, err, fmt.Sprintf("CreateRepaymentDividendSchedules(%v, %v)", tt.args.ctx, tt.args.request)) {
				return
			}
			assert.Equal(t, tt.wantUnmatchedInvestmentCnt, response.UnmatchedInvestmentCount)
		})
	}
}

func Test_repaymentDividendUsecase_calculateRepaymentDividendSchedule(t *testing.T) {
	type fields struct {
		projectRepository                   func() *mocks.ProjectRepository
		shareRepository                     func() *mocks.ShareRepository
		repaymentDividendScheduleRepository func() *mocks.RepaymentDividendScheduleRepository
		projectInvestmentExecuteRepository  func() *mocks.ProjectInvestmentExecuteRepository
	}
	type args struct {
		ctx     context.Context
		request adomain.CreateRepaymentDividendRequest
	}

	now := time.Now().In(repaymentDividendTestJST)
	today := now.Format(time.DateOnly)
	tomorrow := now.Add(24 * time.Hour).Format(time.DateOnly)

	tests := []struct {
		name                   string
		fields                 fields
		args                   args
		want                   error
		wantErr                assert.ErrorAssertionFunc
		wantProjectGetByIdSkip bool // MIR-1239: ①②で弾かれ、後続のGetByIdやOPERATION_INVESTのGetSumが呼ばれないことを確認する
	}{
		{
			// MIR-1239: ①出資実行の記録が無い場合。projectRepository・
			// shareRepositoryにはOnを設定せず、呼ばれたらモックの
			// "call not expected"パニックでテストが失敗する形で検知する。
			name: "①出資実行の記録が無ければ、後続の集計を呼ぶ前に弾く",
			fields: fields{
				projectRepository: func() *mocks.ProjectRepository {
					return new(mocks.ProjectRepository)
				},
				shareRepository: func() *mocks.ShareRepository {
					return new(mocks.ShareRepository)
				},
				repaymentDividendScheduleRepository: func() *mocks.RepaymentDividendScheduleRepository {
					return new(mocks.RepaymentDividendScheduleRepository)
				},
				projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
					repo := new(mocks.ProjectInvestmentExecuteRepository)
					repo.On("GetByProjectId", context.Background(), mockProject1.Id).
						Return(domain.ProjectInvestmentExecute{}, cdomain.ErrItemNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.CreateRepaymentDividendRequest{
					JWTUser:   jwthelper.User{Id: mockUser1.Id},
					ProjectId: 2000,
					CreateRepaymentDividendBody: adomain.CreateRepaymentDividendBody{
						RepaymentAmount: 1000,
						DecideDate:      today,
						ExecuteDate:     tomorrow,
					},
				},
			},
			want:    domain.ErrInvestmentExecutionIncomplete,
			wantErr: assert.Error,
		},
		{
			// MIR-1239: ②ExecuteStatusがCOMPLETEDでない(分割実行の途中)場合。
			// getUnprocessedSharesは判定のために呼ぶ必要があるためモックするが、
			// projectRepository.GetByIdやOPERATION_INVESTのGetSumは呼ばれない。
			name: "②出資実行が未完了なら、後続の集計を呼ぶ前に弾く",
			fields: fields{
				projectRepository: func() *mocks.ProjectRepository {
					return new(mocks.ProjectRepository)
				},
				shareRepository: func() *mocks.ShareRepository {
					repo := new(mocks.ShareRepository)
					repo.On("GetAll", context.Background(), mock.Anything, hasSuccessInvestOnlyOption()).
						Return([]domain.Share{{InvestStatus: domain.SUCCESS_INVEST}}, nil)
					repo.On("GetSum", context.Background(),
						mock.Anything,
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							statuses := opt.InvestStatuses()
							return len(statuses) == 2 &&
								statuses[0].(string) == string(domain.OPERATION_INVEST) &&
								statuses[1].(string) == string(domain.REFUND_INVEST)
						}),
						mock.Anything,
					).Return(float64(0), nil)
					return repo
				},
				repaymentDividendScheduleRepository: func() *mocks.RepaymentDividendScheduleRepository {
					return new(mocks.RepaymentDividendScheduleRepository)
				},
				projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
					repo := new(mocks.ProjectInvestmentExecuteRepository)
					repo.On("GetByProjectId", context.Background(), mockProject1.Id).
						Return(domain.ProjectInvestmentExecute{
							ProjectId:     mockProject1.Id,
							ExecuteStatus: domain.SCHEDULED,
							Amount:        2000,
						}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.CreateRepaymentDividendRequest{
					JWTUser:   jwthelper.User{Id: mockUser1.Id},
					ProjectId: 2000,
					CreateRepaymentDividendBody: adomain.CreateRepaymentDividendBody{
						RepaymentAmount: 1000,
						DecideDate:      today,
						ExecuteDate:     tomorrow,
					},
				},
			},
			want:                   domain.ErrInvestmentExecutionIncomplete,
			wantErr:                assert.Error,
			wantProjectGetByIdSkip: true,
		},
		{
			// MIR-1239: ③(OPERATION_INVEST+REFUND_INVESTの金額合計)が
			// execute.Amountと一致しない場合、後続のOPERATION_INVESTのみの
			// 集計(projectRepository.GetById)を呼ぶ前に弾く。
			name: "③金額が一致しなければ、後続の集計を呼ぶ前に弾く",
			fields: fields{
				projectRepository: func() *mocks.ProjectRepository {
					return new(mocks.ProjectRepository)
				},
				shareRepository: func() *mocks.ShareRepository {
					repo := new(mocks.ShareRepository)
					repo.On("GetAll", context.Background(), mock.Anything, hasSuccessInvestOnlyOption()).
						Return([]domain.Share{}, nil)
					repo.On("GetSum", context.Background(),
						mock.Anything,
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							statuses := opt.InvestStatuses()
							return len(statuses) == 2 &&
								statuses[0].(string) == string(domain.OPERATION_INVEST) &&
								statuses[1].(string) == string(domain.REFUND_INVEST)
						}),
						mock.Anything,
					// COMPLETEDだが、execute.Amount(2000)と1円ずれた1999を返す。
					).Return(float64(1999), nil)
					return repo
				},
				repaymentDividendScheduleRepository: func() *mocks.RepaymentDividendScheduleRepository {
					return new(mocks.RepaymentDividendScheduleRepository)
				},
				projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
					repo := new(mocks.ProjectInvestmentExecuteRepository)
					repo.On("GetByProjectId", context.Background(), mockProject1.Id).
						Return(domain.ProjectInvestmentExecute{
							ProjectId:     mockProject1.Id,
							ExecuteStatus: domain.COMPLETED,
							Amount:        2000,
						}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.CreateRepaymentDividendRequest{
					JWTUser:   jwthelper.User{Id: mockUser1.Id},
					ProjectId: 2000,
					CreateRepaymentDividendBody: adomain.CreateRepaymentDividendBody{
						RepaymentAmount: 1000,
						DecideDate:      today,
						ExecuteDate:     tomorrow,
					},
				},
			},
			want:                   domain.ErrInvestmentExecutionIncomplete,
			wantErr:                assert.Error,
			wantProjectGetByIdSkip: true,
		},
		{
			name: "スケジュール計算",
			fields: fields{
				projectRepository: func() *mocks.ProjectRepository {
					repo := new(mocks.ProjectRepository)
					repo.On("GetById", mock.Anything, mockProject1.Id).Return(mockProject1, nil)
					return repo
				},
				shareRepository: func() *mocks.ShareRepository {
					repo := new(mocks.ShareRepository)
					// MIR-1239: getUnprocessedShares(SUCCESS_INVESTの残件、②の
					// 未処理件数と紛れ込み警告の両方に使う)。ここでは紛れ込みなし
					// (0件)としておく。
					repo.On("GetAll", context.Background(), mock.Anything, hasSuccessInvestOnlyOption()).
						Return([]domain.Share{}, nil)
					// MIR-1239: ③の判定用。OPERATION_INVESTとREFUND_INVESTの
					// 2ステータスを対象にしたGetSum("amount")で、
					// mockProjectInvestmentExecute1.Amount(2000)と一致させる。
					// InvestStatuses()の長さで、既存の②用(OPERATION_INVESTのみ)
					// と区別する。
					repo.On("GetSum", context.Background(),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							projectIdStr := *opt.ProjectId()
							projectId, _ := strconv.Atoi(projectIdStr)
							return int64(projectId) == mockProject1.Id
						}),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							statuses := opt.InvestStatuses()
							return len(statuses) == 2 &&
								statuses[0].(string) == string(domain.OPERATION_INVEST) &&
								statuses[1].(string) == string(domain.REFUND_INVEST)
						}),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							return *opt.SumColumn() == "amount"
						}),
					).Return(float64(mockProjectInvestmentExecute1.Amount), nil).Once()
					repo.On("GetSum", context.Background(),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							projectIdStr := *opt.ProjectId()
							projectId, _ := strconv.Atoi(projectIdStr)
							return int64(projectId) == mockProject1.Id
						}),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							statuses := opt.InvestStatuses()
							return len(statuses) == 1 && statuses[0].(string) == string(domain.OPERATION_INVEST)
						}),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							return *opt.SumColumn() == "qty"
						}),
					).Return(float64(10), nil).Once()
					repo.On("GetSum", context.Background(),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							projectIdStr := *opt.ProjectId()
							projectId, _ := strconv.Atoi(projectIdStr)
							return int64(projectId) == mockProject1.Id
						}),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							statuses := opt.InvestStatuses()
							return len(statuses) == 1 && statuses[0].(string) == string(domain.OPERATION_INVEST)
						}),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							return *opt.SumColumn() == "amount"
						}),
					).Return(float64(1000), nil).Once()
					repo.On("GetInvestor", context.Background(),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							projectIdStr := *opt.ProjectId()
							projectId, _ := strconv.Atoi(projectIdStr)
							return int64(projectId) == mockProject1.Id
						}),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							statuses := opt.InvestStatuses()
							return len(statuses) == 1 && statuses[0].(string) == string(domain.OPERATION_INVEST)
						}),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							return *opt.GroupBy() == "owner_id"
						}),
					).Return([]domain.User{}, nil)
					repo.On("GetSum", context.Background(),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							projectIdStr := *opt.ProjectId()
							projectId, _ := strconv.Atoi(projectIdStr)
							return int64(projectId) == mockProject1.Id
						}),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							statuses := opt.InvestStatuses()
							return len(statuses) == 1 && statuses[0].(string) == string(domain.OPERATION_INVEST)
						}),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							return *opt.UserId() == "1000"
						}),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							return *opt.SumColumn() == "qty"
						}),
					).Return(float64(1000), nil).Once()
					return repo
				},

				repaymentDividendScheduleRepository: func() *mocks.RepaymentDividendScheduleRepository {
					repo := new(mocks.RepaymentDividendScheduleRepository)
					repo.On("GetAll", context.Background(),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							projectIdStr := *opt.ProjectId()
							projectId, _ := strconv.Atoi(projectIdStr)
							return int64(projectId) == mockProject1.Id
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return opt.SortBy() == "execute_date,id"
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return opt.Order() == "desc"
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return opt.Limit() == 1
						}),
					).Return(int64(0), domain.RepaymentDividendSchedules{}, nil)
					repo.On("GetSum", context.Background(),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							projectIdStr := *opt.ProjectId()
							projectId, _ := strconv.Atoi(projectIdStr)
							return int64(projectId) == mockProject1.Id
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return *opt.SumColumn() == "repayment_amount"
						}),
					).Return(float64(1000), nil)
					return repo
				},

				projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
					repo := new(mocks.ProjectInvestmentExecuteRepository)
					repo.On("GetByProjectId", context.Background(), mockProject1.Id).
						Return(mockProjectInvestmentExecute1, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.CreateRepaymentDividendRequest{
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
					ProjectId: 2000,
					CreateRepaymentDividendBody: adomain.CreateRepaymentDividendBody{
						RepaymentAmount: 1000,
						DividendAmount:  0,
						BrokerFee:       0,
						ReturnRate:      0,
						DecideDate:      today,
						ExecuteDate:     tomorrow,
					},
				},
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRepo := tt.fields.projectRepository()
			r := repaymentDividendUsecase{
				projectRepository:                   projectRepo,
				shareRepository:                     tt.fields.shareRepository(),
				repaymentDividendScheduleRepository: tt.fields.repaymentDividendScheduleRepository(),
				projectInvestmentExecuteRepository:  tt.fields.projectInvestmentExecuteRepository(),
			}
			r.calculateRepaymentDividendSchedule = r._calculateRepaymentDividendSchedule

			_, _, err := r.calculateRepaymentDividendSchedule(tt.args.ctx, tt.args.request)
			if !tt.wantErr(t, err, fmt.Sprintf("calculateRepaymentDividendSchedule(%v, %v)", tt.args.ctx, tt.args.request)) {
				return
			}
			if tt.want != nil {
				assert.ErrorIs(t, err, tt.want)
			}
			// MIR-1239: ①②で弾かれた場合、後続のprojectRepository.GetByIdが
			// 呼ばれていないこと(=OPERATION_INVESTの集計まで進んでいないこと)
			// を確認する。
			if tt.wantProjectGetByIdSkip {
				projectRepo.AssertNotCalled(t, "GetById", mock.Anything, mock.Anything)
			}
		})
	}
}

func Test_repaymentDividendUsecase_createRepaymentDividendForInvestors(t *testing.T) {

	// uid, _ := uuid.FromString("bddad349-d0cc-4a1d-8843-431d54fcd2c7")
	ksuid, _ := ksuid.FromBytes([]byte("1b9d6bcd-bbfd-4b2d-9b5d-ab8dfbbd4bed"))

	type fields struct {
		idGenerator        func() *mocks.IdGenerator
		shareRepository    func() *mocks.ShareRepository
		dividendRepository func() *mocks.DividendRepository
	}
	type args struct {
		ctx      context.Context
		schedule domain.RepaymentDividendSchedule
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    error
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "投資家別配当返済スケジュール作成",
			fields: fields{
				idGenerator: func() *mocks.IdGenerator {
					gen := new(mocks.IdGenerator)
					gen.On("GenKsuid").Return(ksuid)
					return gen
				},
				shareRepository: func() *mocks.ShareRepository {
					repo := new(mocks.ShareRepository)
					repo.On("GetInvestor", context.Background(),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							projectIdStr := *opt.ProjectId()
							projectId, _ := strconv.Atoi(projectIdStr)
							return int64(projectId) == mockProject1.Id
						}),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							return *opt.GroupBy() == "owner_id"
						}),
					).Return([]domain.User{mockUser1}, nil)
					repo.On("GetAll", context.Background(),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							projectIdStr := *opt.ProjectId()
							projectId, _ := strconv.Atoi(projectIdStr)
							return int64(projectId) == mockProject1.Id
						}),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							return *opt.UserId() == "3000"
						}),
						mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
							opt := cdomain.ShareQueryOptionFunc()
							field(&opt)
							investStatus := opt.InvestStatuses()[0].(string)
							return investStatus == string(domain.OPERATION_INVEST)
						}),
					).Return([]domain.Share{mockShare1, mockShare2, mockShare3, mockShare4}, nil)
					return repo
				},

				dividendRepository: func() *mocks.DividendRepository {
					repo := new(mocks.DividendRepository)
					repo.On("CreateMany", context.Background(),
						mock.Anything,
						mock.Anything,
					).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				schedule: mockSchedule1,
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := repaymentDividendUsecase{
				idGenerator:        tt.fields.idGenerator(),
				shareRepository:    tt.fields.shareRepository(),
				dividendRepository: tt.fields.dividendRepository(),
			}
			r.createRepaymentDividendForInvestors = r._createRepaymentDividendForInvestors

			_, err := r.createRepaymentDividendForInvestors(tt.args.ctx, tt.args.schedule)
			if !tt.wantErr(t, err, fmt.Sprintf("createRepaymentDividendForInvestors(%v, %v)", tt.args.ctx, tt.args.schedule)) {
				return
			}
		})
	}
}

func Test_repaymentDividendUsecase_GetInvestmentAndReturnsInfo(t *testing.T) {
	type fields struct {
		projectRepository                   func() *mocks.ProjectRepository
		projectInvestmentExecuteRepository  func() *mocks.ProjectInvestmentExecuteRepository
		repaymentDividendScheduleRepository func() *mocks.RepaymentDividendScheduleRepository
		operatorHistoryRepository           func() *mocks.OperatorHistoryRepository
		shareRepository                     func() *mocks.ShareRepository
	}
	type args struct {
		ctx     context.Context
		request adomain.GetInvestmentAndReturnsInfoRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    adomain.GetInvestmentAndReturnsInfoResponse
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "出資/配当/返済情報取得 - 成功",
			fields: fields{
				projectRepository: func() *mocks.ProjectRepository {
					repo := new(mocks.ProjectRepository)
					repo.On("GetById", mock.Anything, mockProject1.Id).Return(mockProject1, nil)
					return repo
				},

				projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
					repo := new(mocks.ProjectInvestmentExecuteRepository)
					repo.On("GetByProjectId", mock.Anything, mockProject1.Id).Return(domain.ProjectInvestmentExecute{}, nil)
					return repo
				},

				repaymentDividendScheduleRepository: func() *mocks.RepaymentDividendScheduleRepository {
					repo := new(mocks.RepaymentDividendScheduleRepository)
					repo.On("GetAll", mock.Anything,
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return *opt.ProjectId() == fmt.Sprintf("%v", mockProject1.Id)
						}),
						mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), domain.RepaymentDividendSchedules{}, nil)

					columnNames := []string{"repayment_amount", "net_dividend_after_broker_fee", "broker_fee", "withholding_tax"}
					sums := []float64{1000, 2000, 300, 40}
					repo.On("GetSums", mock.Anything,
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return *opt.ProjectId() == fmt.Sprintf("%v", mockProject1.Id)
						}),
						mock.Anything,
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)

							if len(*opt.SumColumns()) != len(columnNames) {
								return false
							}
							for i, v := range *opt.SumColumns() {
								if v != columnNames[i] {
									return false
								}
							}
							return true
						}),
					).Return(sums, nil)
					return repo

				},
				operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
					repo := new(mocks.OperatorHistoryRepository)
					repo.On("Create", mock.Anything, mock.Anything).Return(nil)
					return repo
				},
				shareRepository: func() *mocks.ShareRepository {
					repo := new(mocks.ShareRepository)
					// countUnprocessedShares: SUCCESS_INVESTのみ・上限なし(2引数)。
					// 分割実行の途中で、まだ5件未処理として残っている状態。
					repo.On("GetAll", context.Background(), mock.Anything, hasSuccessInvestOnlyOption()).
						Return([]domain.Share{
							{InvestStatus: domain.SUCCESS_INVEST},
							{InvestStatus: domain.SUCCESS_INVEST},
							{InvestStatus: domain.SUCCESS_INVEST},
							{InvestStatus: domain.SUCCESS_INVEST},
							{InvestStatus: domain.SUCCESS_INVEST},
						}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.GetInvestmentAndReturnsInfoRequest{
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
					ProjectId: 2000,
					ClientIp:  "127.0.0.1",
				},
			},
			want: adomain.GetInvestmentAndReturnsInfoResponse{
				ProjectAccount:                 "",
				ProjectDepositor:               "",
				ServiceAccount:                 "000000",
				TargetShareAmount:              0,
				RaisedAmount:                   10000,
				InvestedAmount:                 0,
				InvestmentBalance:              0,
				TotalRepayment:                 1000,
				RepaymentReserve:               0,
				TotalNetDividendAfterBrokerFee: 2000,
				TotalBrokerFee:                 300,
				TotalWithholdingTax:            40,
				UnprocessedShareCount:          5,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := repaymentDividendUsecase{
				projectRepository:                   tt.fields.projectRepository(),
				projectInvestmentExecuteRepository:  tt.fields.projectInvestmentExecuteRepository(),
				repaymentDividendScheduleRepository: tt.fields.repaymentDividendScheduleRepository(),
				operatorHistoryRepository:           tt.fields.operatorHistoryRepository(),
				shareRepository:                     tt.fields.shareRepository(),
			}
			got, err := r.GetInvestmentAndReturnsInfo(tt.args.ctx, tt.args.request)
			if !tt.wantErr(t, err, fmt.Sprintf("GetInvestmentAndReturnsInfo(%v, %v)", tt.args.ctx, tt.args.request)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetInvestmentAndReturnsInfo(%v, %v)", tt.args.ctx, tt.args.request)
		})
	}
}

// MIR-1209: 分割実行の途中でも、テーブル1行目(出資情報レコード)の元金残額
// (InvestmentBalance)が募集総額のまま表示され、処理済みの金額を反映しない
// 不具合を修正したことを確認する。
// hasProjectIdOption は、WithProjectId(projectId)が指定された呼び出しに
// だけマッチする。値まで見ないと、projectIdの絞り込みが外れていたり
// 別プロジェクトのshareを集計してしまう不具合をテストが検知できない。
func hasProjectIdOption(projectId int64) interface{} {
	return mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
		opt := cdomain.ShareQueryOptionFunc()
		field(&opt)
		want := fmt.Sprintf("%d", projectId)
		return opt.ProjectId() != nil && *opt.ProjectId() == want
	})
}

// hasSumColumnOption は、WithSumColumn(column)が指定された呼び出しにだけ
// マッチする。件数用の集計(qty等)と取り違えて金額(amount)以外を合計して
// いないかを、値まで見て検証する。
func hasSumColumnOption(column string) interface{} {
	return mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
		opt := cdomain.ShareQueryOptionFunc()
		field(&opt)
		return opt.SumColumn() != nil && *opt.SumColumn() == column
	})
}

func Test_repaymentDividendUsecase_GetRepaymentDividendSchedulesTable_InvestmentBalance(t *testing.T) {
	tests := []struct {
		name                  string
		unprocessedAmount     float64
		executeAmount         int64
		wantInvestmentBalance int64
	}{
		{
			name:                  "未処理shareが無ければ募集総額のまま",
			unprocessedAmount:     0,
			executeAmount:         10000,
			wantInvestmentBalance: 10000,
		},
		{
			name:                  "分割実行の途中(未処理shareが残っている)なら、その金額分を差し引く",
			unprocessedAmount:     4000, // 3000+1000
			executeAmount:         10000,
			wantInvestmentBalance: 6000, // 10000 - 4000
		},
		{
			// execute.Amountは生SQLの3ステータス合算(論理削除を除外しない)、
			// unprocessedAmountはXORM経由(論理削除を除外する)で、母集団が
			// 完全には一致しない。開発時点で実際に確認できている乖離要因は
			// これだけだが、他の要因(想定外のデータ)で理論上マイナスになる
			// 入力を与えても、画面にマイナスの残額を出さないことを確認する。
			name:                  "未処理share合計が募集総額を上回っても、残額は0未満にならない",
			unprocessedAmount:     8000,
			executeAmount:         5000,
			wantInvestmentBalance: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executeRepo := new(mocks.ProjectInvestmentExecuteRepository)
			executeRepo.On("GetByProjectId", mock.Anything, int64(2000)).
				Return(domain.ProjectInvestmentExecute{
					ProjectId: 2000,
					Amount:    tt.executeAmount,
				}, nil)

			shareRepo := new(mocks.ShareRepository)
			shareRepo.On("GetSum", context.Background(),
				hasProjectIdOption(2000), hasSuccessInvestOnlyOption(), hasSumColumnOption("amount")).
				Return(tt.unprocessedAmount, nil)

			r := repaymentDividendUsecase{
				projectInvestmentExecuteRepository: executeRepo,
				shareRepository:                    shareRepo,
			}

			// MaxResultCountを1にして、出資情報レコードを積んだ直後の
			// 早期returnだけを通す(分配/償還スケジュール側のリポジトリ呼び出し
			// を避け、InvestmentBalanceの算出だけを見る)。
			got, err := r.GetRepaymentDividendSchedulesTable(context.Background(), adomain.GetRepaymentDividendTableRequest{
				ProjectId: 2000,
				SearchInput: controllers.SearchInput{
					MaxResultCount: 1,
				},
			})

			assert.NoError(t, err)
			assert.Len(t, got.Items, 1)
			assert.Equal(t, tt.wantInvestmentBalance, got.Items[0].InvestmentBalance)
		})
	}
}

func Test_repaymentDividendUsecase_DeleteRepaymentDividendSchedule(t *testing.T) {
	type fields struct {
		repaymentDividendScheduleRepository func() *mocks.RepaymentDividendScheduleRepository
		dividendRepository                  func() *mocks.DividendRepository
	}
	type args struct {
		ctx     context.Context
		request adomain.DeleteRepaymentDividendRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    error
	}{
		{
			name: "スケジュール削除 - 成功",
			fields: fields{
				repaymentDividendScheduleRepository: func() *mocks.RepaymentDividendScheduleRepository {
					repo := new(mocks.RepaymentDividendScheduleRepository)
					repo.On("GetAll", context.Background(),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							projectIdStr := *opt.ProjectId()
							projectId, _ := strconv.Atoi(projectIdStr)
							return int64(projectId) == mockProject1.Id
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return opt.SortBy() == "execute_date,id"
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return opt.Order() == "desc"
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return opt.Limit() == 1
						}),
					).Return(int64(0), domain.RepaymentDividendSchedules{mockSchedule1}, nil)
					repo.On("Delete", mock.Anything, int64(1000)).Return(nil)
					return repo

				},

				dividendRepository: func() *mocks.DividendRepository {
					repo := new(mocks.DividendRepository)
					repo.On("DeleteByScheduleId", mock.Anything,
						mock.MatchedBy(func(field func(fields *cdomain.DividendQueryOptionFields)) bool {
							opt := cdomain.DividendQueryOptionFunc()
							field(&opt)
							return *opt.ScheduleId() == mockSchedule1.Id
						}),
					).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.DeleteRepaymentDividendRequest{
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
					ProjectId:  mockProject1.Id,
					ScheduleId: mockSchedule1.Id,
				},
			},
			wantErr: false,
			want:    nil,
		},
		{
			name: "スケジュール削除 - 失敗: スケジュールなし",
			fields: fields{
				repaymentDividendScheduleRepository: func() *mocks.RepaymentDividendScheduleRepository {
					repo := new(mocks.RepaymentDividendScheduleRepository)
					repo.On("GetAll", context.Background(),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							projectIdStr := *opt.ProjectId()
							projectId, _ := strconv.Atoi(projectIdStr)
							return int64(projectId) == mockProject1.Id
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return opt.SortBy() == "execute_date,id"
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return opt.Order() == "desc"
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return opt.Limit() == 1
						}),
					).Return(int64(0), domain.RepaymentDividendSchedules{}, nil)
					repo.On("Delete", mock.Anything, int64(1000)).Return(nil)
					return repo
				},

				dividendRepository: func() *mocks.DividendRepository {
					repo := new(mocks.DividendRepository)
					repo.On("DeleteByScheduleId", mock.Anything,
						mock.MatchedBy(func(field func(fields *cdomain.DividendQueryOptionFields)) bool {
							opt := cdomain.DividendQueryOptionFunc()
							field(&opt)
							return *opt.ScheduleId() == mockSchedule1.Id
						}),
					).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.DeleteRepaymentDividendRequest{
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
					ProjectId:  mockProject1.Id,
					ScheduleId: mockSchedule1.Id,
				},
			},
			wantErr: true,
			want:    cdomain.ErrItemNotFound,
		},
		{
			name: "スケジュール削除 - 失敗: 最新スケジュールの削除要求ではない",
			fields: fields{
				repaymentDividendScheduleRepository: func() *mocks.RepaymentDividendScheduleRepository {
					repo := new(mocks.RepaymentDividendScheduleRepository)
					repo.On("GetAll", context.Background(),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							projectIdStr := *opt.ProjectId()
							projectId, _ := strconv.Atoi(projectIdStr)
							return int64(projectId) == mockProject1.Id
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return opt.SortBy() == "execute_date,id"
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return opt.Order() == "desc"
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return opt.Limit() == 1
						}),
					).Return(int64(0), domain.RepaymentDividendSchedules{mockSchedule2}, nil)
					repo.On("Delete", mock.Anything, int64(1000)).Return(nil)
					return repo
				},

				dividendRepository: func() *mocks.DividendRepository {
					repo := new(mocks.DividendRepository)
					repo.On("DeleteByScheduleId", mock.Anything,
						mock.MatchedBy(func(field func(fields *cdomain.DividendQueryOptionFields)) bool {
							opt := cdomain.DividendQueryOptionFunc()
							field(&opt)
							return *opt.ScheduleId() == mockSchedule1.Id
						}),
					).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.DeleteRepaymentDividendRequest{
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
					ProjectId:  mockProject1.Id,
					ScheduleId: mockSchedule1.Id,
				},
			},
			wantErr: true,
			want:    cdomain.ErrBadRequest,
		},
		{
			name: "スケジュール削除 - 失敗: 最新スケジュールがすでに実施済み",
			fields: fields{
				repaymentDividendScheduleRepository: func() *mocks.RepaymentDividendScheduleRepository {
					repo := new(mocks.RepaymentDividendScheduleRepository)
					repo.On("GetAll", context.Background(),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							projectIdStr := *opt.ProjectId()
							projectId, _ := strconv.Atoi(projectIdStr)
							return int64(projectId) == mockProject1.Id
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return opt.SortBy() == "execute_date,id"
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return opt.Order() == "desc"
						}),
						mock.MatchedBy(func(field func(fields *cdomain.RepaymentDividendScheduleQueryOptionFields)) bool {
							opt := cdomain.RepaymentDividendQueryOptionFunc()
							field(&opt)
							return opt.Limit() == 1
						}),
					).Return(int64(0), domain.RepaymentDividendSchedules{mockSchedule3}, nil)
					repo.On("Delete", mock.Anything, int64(1000)).Return(nil)
					return repo
				},

				dividendRepository: func() *mocks.DividendRepository {
					repo := new(mocks.DividendRepository)
					repo.On("DeleteByScheduleId", mock.Anything,
						mock.MatchedBy(func(field func(fields *cdomain.DividendQueryOptionFields)) bool {
							opt := cdomain.DividendQueryOptionFunc()
							field(&opt)
							return *opt.ScheduleId() == mockSchedule3.Id
						}),
					).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.DeleteRepaymentDividendRequest{
					JWTUser: jwthelper.User{
						Id: mockUser1.Id,
					},
					ProjectId:  mockProject1.Id,
					ScheduleId: mockSchedule3.Id,
				},
			},
			wantErr: true,
			want:    cdomain.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := repaymentDividendUsecase{
				repaymentDividendScheduleRepository: tt.fields.repaymentDividendScheduleRepository(),
				dividendRepository:                  tt.fields.dividendRepository(),
			}

			err := r.DeleteRepaymentDividendSchedule(tt.args.ctx, tt.args.request)
			if tt.wantErr {
				assert.ErrorIs(t, err, tt.want)
				return
			}
			assert.NoError(t, err, "unexpected error returned from DeleteRepaymentDividendSchedule")
		})
	}
}

// Test_checkInvestmentExecutionForSchedule はMIR-1239: 出資実行が完了している
// かの判定(13755設計の状態表)をテーブル駆動で検証する。①(出資実行の記録が
// 無い)は呼び出し元でGetByProjectIdのErrItemNotFoundから判定するため、この
// 関数の対象外。
func Test_checkInvestmentExecutionForSchedule(t *testing.T) {
	completedExecute := domain.ProjectInvestmentExecute{
		ExecuteStatus: domain.COMPLETED,
		Amount:        2000,
	}
	scheduledExecute := domain.ProjectInvestmentExecute{
		ExecuteStatus: domain.SCHEDULED,
		Amount:        2000,
	}

	tests := []struct {
		name              string
		execute           domain.ProjectInvestmentExecute
		investedAmount    int64
		unprocessedCount  int64
		wantErr           bool
		wantErrIsSentinel bool
		wantErrMessage    string
	}{
		{
			name:             "完了済み・金額一致なら作れる",
			execute:          completedExecute,
			investedAmount:   2000,
			unprocessedCount: 0,
			wantErr:          false,
		},
		{
			name:             "完了後にSUCCESS_INVESTが紛れ込んでも、金額が一致していれば作れる(警告は呼び出し元の責務)",
			execute:          completedExecute,
			investedAmount:   2000,
			unprocessedCount: 3,
			wantErr:          false,
		},
		{
			name:              "分割実行の途中(ExecuteStatusがCOMPLETEDでない)は弾く(未処理件数を案内に含める)",
			execute:           scheduledExecute,
			investedAmount:    1000,
			unprocessedCount:  431,
			wantErr:           true,
			wantErrIsSentinel: true,
			wantErrMessage:    "出資実行が完了していません(未処理431件)。運用開始を最後まで進めてから作成してください",
		},
		{
			// MIR-1239: 未処理0件のときは「運用開始を最後まで進めてください」
			// という案内が実際には機能しない(取り消しでRaisedAmountが変わって
			// おり、押し直しても10106/10008で弾かれる)ため、「管理者に連絡して
			// ください」という別のメッセージになることを確認する。
			name: "途中で出資を取り消し、未処理0件だがSCHEDULEDのままなら弾く(案内は管理者連絡)",
			execute: domain.ProjectInvestmentExecute{
				ExecuteStatus: domain.SCHEDULED,
				Amount:        2000,
			},
			investedAmount:    1000,
			unprocessedCount:  0,
			wantErr:           true,
			wantErrIsSentinel: true,
			wantErrMessage:    "出資実行が完了していません。管理者に連絡してください",
		},
		{
			name:              "③の境界: 金額が1円少なければ弾く",
			execute:           completedExecute,
			investedAmount:    1999,
			unprocessedCount:  0,
			wantErr:           true,
			wantErrIsSentinel: true,
		},
		{
			name:              "③の境界: 金額が1円多ければ弾く",
			execute:           completedExecute,
			investedAmount:    2001,
			unprocessedCount:  0,
			wantErr:           true,
			wantErrIsSentinel: true,
		},
		// 「償還を終えたファンド」は入力(execute/investedAmount/unprocessedCount)
		// の組み合わせとしては「完了済み・金額一致なら作れる」と区別できない。
		// checkInvestmentExecutionForSchedule自体は③までしか判定せず、
		// 償還完了後の「no project share qty」は別の場所(集計後)でチェックされる
		// ため、この関数のテストとしては追加しない(13755設計の状態表)。
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkInvestmentExecutionForSchedule(tt.execute, tt.investedAmount, tt.unprocessedCount)
			if !tt.wantErr {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
			if tt.wantErrIsSentinel {
				assert.ErrorIs(t, err, domain.ErrInvestmentExecutionIncomplete)
			}
			if tt.wantErrMessage != "" {
				assert.Equal(t, tt.wantErrMessage, err.Error())
			}
		})
	}
}
