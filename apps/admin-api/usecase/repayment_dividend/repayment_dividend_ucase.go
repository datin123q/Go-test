package repayment_dividend

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/idgenerator"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type repaymentDividendUsecase struct {
	repaymentDividendScheduleRepository cdomain.RepaymentDividendScheduleRepository
	projectInvestmentExecuteRepository  cdomain.ProjectInvestmentExecuteRepository
	investmentRepository                cdomain.InvestmentRepository
	userAccountRepository               cdomain.UserAccountRepository
	projectRepository                   cdomain.ProjectRepository
	shareRepository                     cdomain.ShareRepository
	dividendRepository                  cdomain.DividendRepository
	operatorHistoryRepository           cdomain.OperatorHistoryRepository
	eventMessagesService                cdomain.EventMessageService
	translator                          i18nutils.Translator
	idGenerator                         idgenerator.IdGenerator
	calculateRepaymentDividendSchedule  func(ctx context.Context, request adomain.CreateRepaymentDividendRequest) (domain.RepaymentDividendSchedule, int64, error)
	createRepaymentDividendForInvestors func(ctx context.Context, sched domain.RepaymentDividendSchedule) (int64, error)
	calculateCustomerLedgerBalance      func(ctx context.Context, userId int64) (int64, error)
	withdrawalUserTransaction           func(ctx context.Context, userId, balance, amount, projectId int64, projectType domain.PROJECT_TYPE, accountNumber string, transactionDetailType domain.USER_TRANSACTION_DETAIL_TYPE, transactionDate jsonutils.JSONTime) error
	updateShareToOperation              func(ctx context.Context, share *domain.Share, operatorId int64) error
	timeNow                             func() time.Time
}

func New(
	repaymentDividendScheduleRepository cdomain.RepaymentDividendScheduleRepository,
	projectInvestmentExecuteRepository cdomain.ProjectInvestmentExecuteRepository,
	investmentRepository cdomain.InvestmentRepository,
	userAccountRepository cdomain.UserAccountRepository,
	projectRepository cdomain.ProjectRepository,
	shareRepository cdomain.ShareRepository,
	dividendRepository cdomain.DividendRepository,
	operatorHistoryRepository cdomain.OperatorHistoryRepository,
	eventMessagesService cdomain.EventMessageService,
	translator i18nutils.Translator,
	idGenerator idgenerator.IdGenerator,
) adomain.RepaymentDividendUsecase {
	usecase := &repaymentDividendUsecase{
		repaymentDividendScheduleRepository: repaymentDividendScheduleRepository,
		projectInvestmentExecuteRepository:  projectInvestmentExecuteRepository,
		investmentRepository:                investmentRepository,
		userAccountRepository:               userAccountRepository,
		projectRepository:                   projectRepository,
		shareRepository:                     shareRepository,
		dividendRepository:                  dividendRepository,
		operatorHistoryRepository:           operatorHistoryRepository,
		eventMessagesService:                eventMessagesService,
		translator:                          translator,
		idGenerator:                         idGenerator,
	}
	usecase.calculateRepaymentDividendSchedule = usecase._calculateRepaymentDividendSchedule
	usecase.createRepaymentDividendForInvestors = usecase._createRepaymentDividendForInvestors
	usecase.calculateCustomerLedgerBalance = domain.CalculateCustomerLedgerBalance
	usecase.withdrawalUserTransaction = (domain.UserTransaction{}).WithdrawalUserTransactionWithBalance
	usecase.updateShareToOperation = func(ctx context.Context, share *domain.Share, operatorId int64) error {
		return share.UpdateInvestStatus(ctx, domain.OPERATION_INVEST, operatorId, domain.NOT_IMMEDIATE)
	}
	usecase.timeNow = time.Now
	return usecase
}

// userFacingError は、message部分をそのまま利用者へ見せてよいエラー
// (MIR-1204)。errors.Wrapは"<message>: <cause>"の形になり原因の英語文が
// くっついてしまうため、Unwrap()でcauseへ辿れる薄いラッパーとして持つ。
type userFacingError struct {
	cause   error
	message string
}

func (e *userFacingError) Error() string { return e.message }
func (e *userFacingError) Unwrap() error { return e.cause }

// ExecuteInvestmentBatchSize は1回の実行で処理するshareの上限件数
// (MIR-1204)。本番でCloudFrontのオリジン応答タイムアウト(120秒、上限で
// 変更不可)に達し出資処理が完走できなかったため、1回のリクエストで処理
// する件数をここで区切り、上限を超えるプロジェクトは「運用開始」ボタンを
// 複数回押して続きから処理する運用にした。実測259ms/口を基準に、200件
// なら約52秒。実測より2倍遅くなっても120秒に収まるよう選んだ値で、暫定の
// 高速化対応(MIR-1202)には依存しない。
const ExecuteInvestmentBatchSize = 200

func (r repaymentDividendUsecase) ExecuteInvestment(
	ctx context.Context,
	request adomain.ExecuteInvestmentRequest,
) (adomain.ExecuteInvestmentResponse, error) {
	// プロジェクト単位で行ロックを取り、同じプロジェクトへの並行実行を直列化する
	// (MIR-1204)。分割実行では同じボタンを何度も押せるため、ロック無しだと
	// 2つのリクエストが同じSUCCESS_INVESTのshareを同時に読んでしまい、
	// 二重の出資取引・法定帳簿イベントを生む恐れがある。待機(FOR UPDATE)では
	// なくNOWAITにしているのは、実行中フラグのような状態を持たないため。
	// ECSのタスク入れ替え等でプロセスが落ちても、トランザクションが終われば
	// ロックも自動的に消えるので、手動での復旧が要らない(2026-09-08河内さん判断)。
	// ロックが取れない場合はErrItemConflictを返し、待たせずに「他の処理が
	// 実行中」と伝える。
	currentProject, err := r.projectRepository.TryLockById(ctx, request.ProjectId)
	if err != nil {
		return adomain.ExecuteInvestmentResponse{}, errors.Wrap(err, "failed to lock project by id")
	}
	if currentProject.ProjectStatus != domain.OPERATION_FUNDING {
		return adomain.ExecuteInvestmentResponse{}, errors.Wrapf(cdomain.ErrBadRequest, "project status should be OPERATION_FUNDING, but got %s", currentProject.ProjectStatus)
	}
	if currentProject.RaisedAmount < request.Amount {
		return adomain.ExecuteInvestmentResponse{}, errors.Wrap(cdomain.ErrBadRequest, "failed to execute investment: insufficient raised amount")
	}

	executeDate, err := jsonutils.ParseJSONTimeJST(time.DateOnly, request.ExecuteDate)
	if err != nil {
		return adomain.ExecuteInvestmentResponse{}, errors.Wrap(err, "failed to parse execute date")
	}

	// 分割実行(同じボタンの複数回押下)に対応するため、このプロジェクトの
	// 出資実行レコードが既にあれば使い回し、無ければ新規作成する(MIR-1204)。
	// 無条件にCreateすると、押すたびに重複したレコードが作られてしまう。
	// ExecuteDate・Amountは法定帳簿・投資額の判定に使われる値のため、
	// 既存レコードと異なる値が指定された場合は黙って上書き・無視せず
	// エラーにする(日や金額を変えて続きを実行したい場合は、このエラーを
	// 見て利用者が判断する)。
	execute, err := r.projectInvestmentExecuteRepository.GetByProjectId(ctx, request.ProjectId)
	if err != nil {
		if !errors.Is(err, cdomain.ErrItemNotFound) {
			return adomain.ExecuteInvestmentResponse{}, errors.Wrap(err, "failed to get project investment execute")
		}
		projectFile := domain.ProjectInvestmentExecute{
			ProjectId:   request.ProjectId,
			Amount:      request.Amount,
			ExecuteDate: executeDate,
			// 分割実行の途中はCOMPLETEDにしない(MIR-1204)。一覧画面が
			// ExecuteStatus==COMPLETEDのときだけ「実施済」と表示するため、
			// 1バッチ目が終わっただけの時点でCOMPLETEDにすると、まだ
			// 未処理のshareが残っているのに実施済に見えてしまう。全件
			// 処理し終えたときだけ、この後COMPLETEDへ更新する。
			ExecuteStatus: domain.SCHEDULED,
		}
		execute, err = r.projectInvestmentExecuteRepository.Create(ctx, projectFile)
		if err != nil {
			return adomain.ExecuteInvestmentResponse{}, err
		}
	} else if !execute.ExecuteDate.EqualDate(executeDate.Time()) {
		return adomain.ExecuteInvestmentResponse{}, &userFacingError{
			cause: cdomain.ErrExecuteInvestmentDateMismatch,
			message: fmt.Sprintf(
				"このファンドは%sで出資実行を開始しています。残りを処理するには、同じ日付を指定してください。",
				execute.ExecuteDate.ToFormat(time.DateOnly),
			),
		}
	} else if execute.Amount != request.Amount {
		return adomain.ExecuteInvestmentResponse{}, &userFacingError{
			cause: cdomain.ErrExecuteInvestmentAmountMismatch,
			message: fmt.Sprintf(
				"このファンドは金額%d円で出資実行を開始しています。残りを処理するには、同じ金額を指定してください。",
				execute.Amount,
			),
		}
	}

	// 法定帳簿向け Kafka イベント発行
	processedCount, err := r.publishLegalRecordEvent(ctx, &currentProject, execute, request.JWTUser.Id)
	if err != nil {
		return adomain.ExecuteInvestmentResponse{}, errors.Wrap(err, "failed to publish legal record event")
	}

	unprocessedShares, err := r.getUnprocessedShares(ctx, currentProject.Id)
	if err != nil {
		return adomain.ExecuteInvestmentResponse{}, errors.Wrap(err, "failed to get unprocessed shares")
	}
	remainingCount := int64(len(unprocessedShares))

	// 未処理のshareが0件になった時点で初めてCOMPLETEDへ更新する(MIR-1204)。
	if remainingCount == 0 && execute.ExecuteStatus != domain.COMPLETED {
		execute.ExecuteStatus = domain.COMPLETED
		if err := r.projectInvestmentExecuteRepository.Update(ctx, execute); err != nil {
			return adomain.ExecuteInvestmentResponse{}, errors.Wrap(err, "failed to update project investment execute status")
		}
	}

	return adomain.ExecuteInvestmentResponse{
		ProjectInvestmentExecute: execute,
		ProcessedCount:           processedCount,
		RemainingCount:           remainingCount,
	}, nil
}

// getUnprocessedShares はこのプロジェクトでまだ未処理(SUCCESS_INVEST)の
// まま残っているshareを返す(MIR-1204)。件数(旧countUnprocessedSharesが
// 返していたもの)は呼び出し側でlen()を取る。MIR-1209で元金残額(処理済み
// 金額を反映)の算出にも同じ未処理share一覧が必要になったため、件数専用
// だった形をやめてスライスを返す形にした。
func (r repaymentDividendUsecase) getUnprocessedShares(ctx context.Context, projectId int64) ([]domain.Share, error) {
	opts := cdomain.ShareQueryOptionFunc()
	remaining, err := r.shareRepository.GetAll(ctx,
		opts.WithProjectId(projectId),
		opts.WithInvestStatuses(domain.SUCCESS_INVEST),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get remaining project shares")
	}
	return remaining, nil
}

// investmentBalanceAfterUnprocessed はexecute.Amount(募集総額)から
// unprocessedAmount(未処理のまま残っているshareの金額合計)を差し引いた
// 元金残額を返す(MIR-1209)。execute.AmountはProject.RaisedAmountの生SQL
// 集計(SUCCESS_INVEST/OPERATION_INVEST/REFUND_INVESTの3ステータス合算・
// 論理削除を除外しない)に由来し、unprocessedAmountはXORM経由(論理削除を
// 除外する)で集計しているため、母集団が完全には一致しない。理論上マイナス
// になった場合でも、画面にマイナスの元金残額を出さないよう0で下げ止める。
func investmentBalanceAfterUnprocessed(ctx context.Context, projectId, totalAmount, unprocessedAmount int64) int64 {
	balance := totalAmount - unprocessedAmount
	if balance < 0 {
		factory.Logger(ctx).Errorf(
			"investment balance went negative: totalAmount=%d unprocessedAmount=%d projectId=%d",
			totalAmount, unprocessedAmount, projectId,
		)
		return 0
	}
	return balance
}

// 기존함수: func (l *ProjectInvestmentExecute) ExecuteInvestment(ctx context.Context, p *Project) error
// publishLegalRecordEvent 投資成立時に発火。処理したshareの件数を返す。
func (r repaymentDividendUsecase) publishLegalRecordEvent(
	ctx context.Context,
	project *domain.Project,
	execute domain.ProjectInvestmentExecute,
	operatorId int64,
) (int64, error) {
	// 未処理(SUCCESS_INVEST)のshareだけをExecuteInvestmentBatchSize件までに
	// 絞って取得する(MIR-1204)。前回までの実行で既にOPERATION_INVESTへ
	// 進んだshareはここに含まれないため、出金トランザクション・法定帳簿向け
	// イベントが再度発行されることはない。
	opts := cdomain.ShareQueryOptionFunc()
	projectShares, err := r.shareRepository.GetAll(ctx,
		opts.WithProjectId(project.Id),
		opts.WithInvestStatuses(domain.SUCCESS_INVEST),
		opts.WithSortBy("id"),
		opts.WithOrder("asc"),
		opts.WithLimit(ExecuteInvestmentBatchSize),
	)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get project shares")
	}

	// 送信失敗の件数はAfterCommitのクロージャの中でカウントする(MIR-1202の④)。
	// AfterCommitは登録順に逐次実行され並行実行はされないため、mutexなしで
	// 安全にインクリメントできる(pkg/factory/tx.goのafterCommitStore.run参照)。
	var publishFailedCount int64

	now := jsonutils.JSONTime(r.timeNow())
	for i := range projectShares {
		share := &projectShares[i]

		if err := r.updateShareToOperation(ctx, share, operatorId); err != nil {
			return 0, errors.Wrap(err, "failed to update share status to OPERATION_INVEST")
		}

		ua, err := r.userAccountRepository.GetByUserId(ctx, share.OwnerId)
		if err != nil {
			return 0, errors.Wrap(err, "failed to get user account")
		}

		balance, err := r.calculateCustomerLedgerBalance(ctx, share.OwnerId)
		if err != nil {
			return 0, errors.Wrap(err, "failed to calculate customer ledger balance")
		}

		// withdrawalUserTransactionは既に算出済みのbalanceをそのまま使う(MIR-1202)。
		// user_transactionへのINSERTはuser_asset_view(残高計算の元)に影響しないため、
		// ここでもう一度CalculateCustomerLedgerBalanceを呼んでも上のbalanceと同じ値に
		// なるだけの二重計算だった。本番の実測でuser_asset_view 1回あたり約58ms
		// かかっており、1口あたりの所要時間(実測259ms)の約22%を占めていた。
		if err = r.withdrawalUserTransaction(ctx, share.OwnerId, balance, share.Amount, project.Id, project.ProjectType, "", domain.WITHDRAWAL_INVESTMENT, now); err != nil {
			return 0, errors.Wrap(err, "failed to create user transaction for investment execute")
		}

		// 거래 이벤트 발행。event_messageへの記録(is_publish=false)はトランザクション内で
		// 行い、実際の送信はコミット後(AfterCommit)にまとめて行う(Outbox、MIR-1202)。
		// こうすることで、200件処理の途中でプロセスが落ちても、コミットが済んだ分は
		// event_messageの行が必ず存在し、RePublish(MIR-1114)で拾い直せる。
		transactionEvent, err := r.eventMessagesService.RecordEvent(ctx, domain.EVENT_TRANSACTION, domain.TransactionDto{
			EventDateTime:   lo.ToPtr(now.Time()),                  // システム上で出資処理が実行された日時
			CommitDateTime:  lo.ToPtr(share.CommitmentDate.Time()), // 약정일자
			ApplyDateTime:   lo.ToPtr(execute.ExecuteDate.Time()),  // ビジネス上の出資日
			TransactionType: domain.INVESTMENT,
			UserId:          share.OwnerId,
			UserName:        ua.UserName,
			InvestmentId:    share.Id, // shareの持つKSUID(PK)
			ProductName:     project.Name,
			SpvId:           project.SpvId,
			ProductId:       project.Id,
			UnitPrice:       project.SharePrice,
			Debt:            share.Amount,
			Balance:         balance,
			Amount:          share.Amount,
			Quantity:        share.Qty,
		})
		if err != nil {
			return 0, errors.Wrap(err, "failed to create customer ledger investment event: EVENT_TRANSACTION")
		}
		if err := factory.AfterCommit(ctx, func(afterCommitCtx context.Context) error {
			published, sendErr := r.eventMessagesService.SendRecordedEvent(afterCommitCtx, transactionEvent)
			// 送信自体は成功してもDB更新(is_publish=true)が失敗した場合、行は
			// is_publish=falseのまま残り再送の対象になる。published==falseの場合と
			// 同じく「まだ確実に送れていない」ことに変わりないため、この場合も
			// 失敗としてカウントする。
			if !published || sendErr != nil {
				publishFailedCount++
			}
			return sendErr
		}); err != nil {
			return 0, errors.Wrap(err, "failed to queue after-commit send: EVENT_TRANSACTION")
		}

		// EVENT_FUND_TRANSFERの発行はMIR-1202で暫定停止した(#543)。3法定帳簿アプリ
		// (statutory-books-jp-service/mirapro-statutory-books-service/
		// ikkaq/statutory-books-service)のいずれもEventCategoryにこの値を持たず、
		// consumerのswitchでdefault(invalid category)に落ちる(2026-09-08確認)。
		// 生イベントの保管テーブル(message)には他の種別と同様に行が作られるが、
		// 帳簿データ(fund_transfers)としては使われない。fund_transfersは
		// EVENT_TRANSACTION側の処理(本関数の直前のRecordEvent)から作られるため、
		// この発行を止めても帳簿データに影響は無い。1件あたりRecordEvent+
		// AfterCommitでの送信が1回減る(本番でadmin-api→user asset view
		// 参照のタイムアウトが起きたための応急処置、恒久対応はMIR-1202参照)。
	}

	// 全shareぶんのAfterCommit登録が終わった後、集計結果を実行レコードへ書き込む
	// 処理を最後に1回だけ登録する。AfterCommitは登録順に逐次実行される
	// (pkg/factory/tx.go)ため、この処理は必ず上のSendRecordedEvent呼び出し
	// すべての後に走り、publishFailedCountは確定した値になっている。
	executeId := execute.Id
	if err := factory.AfterCommit(ctx, func(afterCommitCtx context.Context) error {
		if publishFailedCount == 0 {
			return nil
		}
		if err := r.projectInvestmentExecuteRepository.IncrementPublishFailedCount(afterCommitCtx, executeId, publishFailedCount); err != nil {
			// publish_failed_countへの書き込み自体が失敗すると、送信失敗件数を
			// 追う手段がDBから消えるため、ログにも件数を残しておく。
			factory.Logger(afterCommitCtx).WithError(err).WithFields(logrus.Fields{
				"project_investment_execute_id": executeId,
				"publish_failed_count":          publishFailedCount,
			}).Error("failed to record publish failed count")
			return err
		}
		return nil
	}); err != nil {
		return 0, errors.Wrap(err, "failed to queue after-commit publish failed count update")
	}

	return int64(len(projectShares)), nil
}

func (r repaymentDividendUsecase) CreateRepaymentDividendSchedules(
	ctx context.Context,
	request adomain.CreateRepaymentDividendRequest,
) (adomain.CreateRepaymentDividendResponse, error) {
	// 배당상환 스케쥴 계산。ここではerrors.Wrapで包まない(MIR-1239)。
	// _calculateRepaymentDividendSchedule内の各エラーは既に固有のコンテキスト
	// 付きでラップ済みのため、ここでさらに包むと二重になる。userFacingError
	// (10107)をそのまま返すことで、http層がerrors.Causeのような階層を
	// 剥がす処理を持たずに済み、将来別のラップが挟まっても壊れにくくなる。
	calculatedSchedule, unmatchedInvestmentCount, err := r.calculateRepaymentDividendSchedule(ctx, request)
	if err != nil {
		return adomain.CreateRepaymentDividendResponse{}, err
	}

	// 배당상환 스케줄 DB 저장. 저장 후 회신한 sched에는 Id가 부여되어 있다.
	schedule, err := r.repaymentDividendScheduleRepository.Create(ctx, calculatedSchedule)
	if err != nil {
		return adomain.CreateRepaymentDividendResponse{},
			errors.Wrap(err, "failed to create repayment dividend schedule")
	}

	// 배당상환 스케쥴 기반 dividend 생성
	sumTax, err := r.createRepaymentDividendForInvestors(ctx, schedule)
	if err != nil {
		return adomain.CreateRepaymentDividendResponse{},
			errors.Wrap(err, "failed to create repayment dividend for investors")
	}

	// 投資家個別のdividendレコードの源泉税合計をファンド全体の源泉税として設定
	schedule.WithholdingTax = sumTax
	if err := r.repaymentDividendScheduleRepository.Update(ctx, schedule); err != nil {
		return adomain.CreateRepaymentDividendResponse{},
			errors.Wrap(err, "failed to update repayment dividend schedule withholding tax")
	}

	return adomain.CreateRepaymentDividendResponse{
		RepaymentDividendSchedule: schedule,
		UnmatchedInvestmentCount:  unmatchedInvestmentCount,
	}, nil
}

// checkInvestmentExecutionForSchedule は、分配・償還スケジュールを作成して
// よいだけ出資実行が完了しているかを判定する(MIR-1239)。②execute_statusが
// COMPLETEDでない場合と、③OPERATION_INVEST+REFUND_INVESTの金額合計が
// execute.Amountと一致しない場合にエラーを返す。①(出資実行の記録が無い)は
// 呼び出し元でGetByProjectIdのErrItemNotFoundから判定するため、ここでは
// 扱わない。userFacingErrorで包み、causeにdomain.ErrInvestmentExecutionIncomplete
// を持たせることでerrors.Isによる判別を可能にする。
func checkInvestmentExecutionForSchedule(execute domain.ProjectInvestmentExecute, investedAmount int64, unprocessedCount int64) error {
	if execute.ExecuteStatus != domain.COMPLETED {
		// MIR-1239: 未処理件数が0件のときは「途中で出資を取り消し、未処理は
		// 無くなったがExecuteStatusがCOMPLETEDに更新されないまま残っている」
		// ケースで、「運用開始を最後まで進めてください」と案内しても、
		// 取り消しでRaisedAmountが変わっているため再実行は10106(取り消し前の
		// 金額を指定)または10008(募集額不足、取り消し後の金額を指定)で弾かれ、
		// 案内どおりには進められない。この場合だけ管理者への連絡を案内する。
		message := fmt.Sprintf(
			"出資実行が完了していません(未処理%d件)。運用開始を最後まで進めてから作成してください",
			unprocessedCount,
		)
		if unprocessedCount == 0 {
			message = "出資実行が完了していません。管理者に連絡してください"
		}
		return &userFacingError{
			cause:   domain.ErrInvestmentExecutionIncomplete,
			message: message,
		}
	}
	if investedAmount != execute.Amount {
		return &userFacingError{
			cause: domain.ErrInvestmentExecutionIncomplete,
			message: fmt.Sprintf(
				"出資実行の金額(%d円)と、運用中の出資の合計(%d円)が一致しません。管理者に連絡してください",
				execute.Amount, investedAmount,
			),
		}
	}
	return nil
}

func (r repaymentDividendUsecase) _calculateRepaymentDividendSchedule(
	ctx context.Context, request adomain.CreateRepaymentDividendRequest,
) (domain.RepaymentDividendSchedule, int64, error) {
	// 스케쥴 생성 - request 바인딩
	decideDate, err := jsonutils.ParseDateTimeJST(time.DateOnly, request.DecideDate)
	if err != nil {
		return domain.RepaymentDividendSchedule{}, 0, errors.Wrap(err, "failed to parse decide date")
	}
	executeDate, err := jsonutils.ParseDateTimeJST(time.DateOnly, request.ExecuteDate)
	if err != nil {
		return domain.RepaymentDividendSchedule{}, 0, errors.Wrap(err, "failed to parse execute date")
	}

	// 실시일(executeDate)은 결정일(decideDate) 이후이며, 생성일 이후여야 한다.
	today := time.Now().In(time.Local)
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	if !executeDate.After(decideDate) || !executeDate.After(today) {
		return domain.RepaymentDividendSchedule{}, 0, errors.Wrap(cdomain.ErrBadRequest, "execute date should be after decide date and today")
	}

	// MIR-1239: 出資実行が完了しているかを、集計より前に確認する。分割実行
	// (MIR-1204)の途中でスケジュールを作成すると、まだ処理していないshareが
	// SUCCESS_INVESTのまま残っているため、後段の集計(OPERATION_INVESTのみ
	// 対象)の分母が実際より小さくなり、一口あたりの還元額が過大に計算される。
	// ①出資実行の記録が無い、②ExecuteStatusがCOMPLETEDでない、③OPERATION_
	// INVESTとREFUND_INVESTの金額合計がexecute.Amountと一致しない、の
	// いずれかに該当すれば作成を止める。ここで取得したinvestmentは、後段の
	// 決定日と出資日の比較(旧: 502行目付近にあった二重取得)にも使い回す。
	investment, err := r.projectInvestmentExecuteRepository.GetByProjectId(ctx, request.ProjectId)
	if err != nil {
		if errors.Is(err, cdomain.ErrItemNotFound) {
			return domain.RepaymentDividendSchedule{}, 0, &userFacingError{
				cause:   domain.ErrInvestmentExecutionIncomplete,
				message: "出資実行が行われていません",
			}
		}
		return domain.RepaymentDividendSchedule{}, 0, errors.Wrap(err, "failed to get investment by project id")
	}

	unprocessedShares, err := r.getUnprocessedShares(ctx, request.ProjectId)
	if err != nil {
		return domain.RepaymentDividendSchedule{}, 0, errors.Wrap(err, "failed to get unprocessed shares")
	}

	// ③の合計はOPERATION_INVESTとREFUND_INVESTの両方を対象にする。1本目の
	// 償還ではexecute.Amountを分母に元金残額を計算するため(このあとの
	// InvestmentBalance計算)、償還後にOPERATION_INVESTからREFUND_INVESTへ
	// 移った分もexecute.Amountとの一致判定に含める必要がある。
	investedAmount, err := r.shareRepository.GetSum(ctx,
		cdomain.ShareQueryOptionFunc().WithProjectId(request.ProjectId),
		cdomain.ShareQueryOptionFunc().WithInvestStatuses(domain.OPERATION_INVEST, domain.REFUND_INVEST),
		cdomain.ShareQueryOptionFunc().WithSumColumn("amount"))
	if err != nil {
		return domain.RepaymentDividendSchedule{}, 0, errors.Wrap(err, "failed to get share sum for investment execution check")
	}

	if err := checkInvestmentExecutionForSchedule(investment, int64(investedAmount), int64(len(unprocessedShares))); err != nil {
		return domain.RepaymentDividendSchedule{}, 0, err
	}

	sched := domain.RepaymentDividendSchedule{
		ProjectId:       request.ProjectId,
		RepaymentAmount: request.RepaymentAmount,
		DividendAmount:  request.DividendAmount,
		BrokerFee:       request.BrokerFee,
		ReturnRate:      request.ReturnRate,
		DecideDate:      jsonutils.JSONTime(decideDate),
		ExecuteDate:     jsonutils.JSONTime(executeDate),
	}

	// 스케쥴 생성
	// 1. ProjectType
	p, err := r.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return domain.RepaymentDividendSchedule{}, 0, errors.Wrap(err, "failed to get project by id")
	}
	sched.ProjectType = p.ProjectType

	// 2. ProjectShareQty, ProjectShareAmount
	sumQty, err := r.shareRepository.GetSum(ctx,
		cdomain.ShareQueryOptionFunc().WithProjectId(request.ProjectId),
		cdomain.ShareQueryOptionFunc().WithInvestStatuses(domain.OPERATION_INVEST),
		cdomain.ShareQueryOptionFunc().WithSumColumn("qty"))
	if err != nil {
		return domain.RepaymentDividendSchedule{}, 0, errors.Wrap(err, "failed to get share sum")
	}
	sched.ProjectShareQty = int64(sumQty)

	if sched.ProjectShareQty == 0 {
		return domain.RepaymentDividendSchedule{}, 0, errors.New("no project share qty")
	}
	sumAmount, err := r.shareRepository.GetSum(ctx,
		cdomain.ShareQueryOptionFunc().WithProjectId(request.ProjectId),
		cdomain.ShareQueryOptionFunc().WithInvestStatuses(domain.OPERATION_INVEST),
		cdomain.ShareQueryOptionFunc().WithSumColumn("amount"))
	if err != nil {
		return domain.RepaymentDividendSchedule{}, 0, errors.Wrap(err, "failed to get share sum")
	}
	sched.ProjectShareAmount = int64(sumAmount)

	// 3. ReturnType, Status
	if sched.RepaymentAmount > 0 && sched.DividendAmount == 0 {
		sched.ReturnType = domain.REPAYMENT_RETURN
	} else if sched.RepaymentAmount == 0 && sched.DividendAmount > 0 {
		sched.ReturnType = domain.DIVIDEND_RETURN
	} else {
		return domain.RepaymentDividendSchedule{}, 0, errors.New("invalid repayment and/or dividend amount")
	}
	sched.Status = domain.RETURN_SCHEDULED

	// 검증: 추가하려는 스케줄보다 더 늦은 실시일의 스케줄이 있으면 안된다.
	_, latestSchedule, err := r.repaymentDividendScheduleRepository.GetAll(ctx,
		cdomain.RepaymentDividendQueryOptionFunc().WithProjectId(request.ProjectId),
		cdomain.RepaymentDividendQueryOptionFunc().WithSortBy("execute_date,id"),
		cdomain.RepaymentDividendQueryOptionFunc().WithOrder("desc"),
		cdomain.RepaymentDividendQueryOptionFunc().WithLimit(1),
	)
	if err != nil {
		return domain.RepaymentDividendSchedule{}, 0,
			errors.Wrap(err, "failed to get following schedules")
	}

	if len(latestSchedule) > 0 {
		requestExecuteDate, err := jsonutils.ParseDateTimeJST(time.DateOnly, request.ExecuteDate)
		if err != nil {
			return domain.RepaymentDividendSchedule{}, 0,
				errors.Wrap(err, "failed to parse execute date")
		}
		// 기존 스케줄의 실시일보다 더 빠른 실시일이면 안된다.
		latestExecuteDate := latestSchedule[0].ExecuteDate.Time()
		if requestExecuteDate.Before(latestExecuteDate) {
			return domain.RepaymentDividendSchedule{}, 0,
				errors.Wrap(cdomain.ErrBadRequest, "execute date should be after latest schedule")
		}
	}

	// 검증: 생성 스케쥴 일시(결정일)는 출자일시 이후여야 한다. investmentは
	// 冒頭のMIR-1239の判定で既に取得済みのものを使い回す。
	if sched.DecideDate.Time().Before(investment.ExecuteDate.Time()) {
		return domain.RepaymentDividendSchedule{}, 0, errors.Wrap(cdomain.ErrBadRequest, "decide date should be after investment execute date")
	}

	// 상환정보 - RepaymentAmount 는 이미 입력되어 있음
	if sched.ReturnType == domain.REPAYMENT_RETURN {
		// 상환액 검사: 이번 상환액 > 투자금액 - 지금까지의 상환액(실시전 포함) 이면 에러
		total, err := r.repaymentDividendScheduleRepository.GetSum(ctx,
			cdomain.RepaymentDividendQueryOptionFunc().WithProjectId(request.ProjectId),
			cdomain.RepaymentDividendQueryOptionFunc().WithSumColumn("repayment_amount"),
		)
		if err != nil {
			return domain.RepaymentDividendSchedule{}, 0, errors.Wrap(err, "failed to get sum repayment amount by project id")
		}
		totalRepaymentToDate := int64(total)

		if sched.RepaymentAmount > (investment.Amount - totalRepaymentToDate) { // totalRepaymentToDate 는 상환적립금을 포함한 금액이다.
			return domain.RepaymentDividendSchedule{}, 0, domain.ErrParameterRepaymentAmountExceeds
		}
		// 상환 적립금 계산
		if len(latestSchedule) > 0 {
			sched.PreviousRepaymentReserve = latestSchedule[0].RepaymentReserve
			sched.RepaymentPerShare = (sched.RepaymentAmount + sched.PreviousRepaymentReserve) / sched.ProjectShareQty                                        // 소수점 이하는 버려짐
			sched.RepaymentReserve = (sched.RepaymentAmount + sched.PreviousRepaymentReserve) - (sched.RepaymentPerShare * sched.ProjectShareQty)             // 이번회차 상환 적립금 = (상환금 + 이전 상환적립금) - (구당상환금 * 구수)
			sched.InvestmentBalance = latestSchedule[0].InvestmentBalance - (sched.RepaymentAmount + sched.PreviousRepaymentReserve - sched.RepaymentReserve) // 이번 상황액 + 이전 상환적립금을 상환하려 하였으나, 이번 상환적립금 만큼 남음
		} else { // 출자이후 첫 스케줄
			sched.RepaymentPerShare = sched.RepaymentAmount / sched.ProjectShareQty // 소수점 이하는 버려짐
			sched.RepaymentReserve = sched.RepaymentAmount - (sched.RepaymentPerShare * sched.ProjectShareQty)
			sched.InvestmentBalance = investment.Amount - (sched.RepaymentAmount - sched.RepaymentReserve) // 상환적립금을 뺀 상환액만 상환됨
		}
	}
	// 배당정보 - DividendAmount, BrokerFee, ReturnRate 는 이미 입력되어 있음
	investors, err := r.shareRepository.GetInvestor(ctx,
		cdomain.ShareQueryOptionFunc().WithProjectId(request.ProjectId),
		cdomain.ShareQueryOptionFunc().WithInvestStatuses(domain.OPERATION_INVEST),
		cdomain.ShareQueryOptionFunc().WithGroupBy("owner_id"),
	)
	if err != nil {
		return domain.RepaymentDividendSchedule{}, 0, errors.Wrap(err, "failed to get investors by project")
	}

	if sched.ReturnType == domain.DIVIDEND_RETURN {
		sched.NetDividendAfterBrokerFee = sched.DividendAmount - sched.BrokerFee
		// ESKK에서 배당적립금이 생기지 않도록 계산하여 입력 예정 - 따라서 배당적립금이 생겨서는 안됨.
		if sched.NetDividendAfterBrokerFee%sched.ProjectShareQty != 0 {
			return domain.RepaymentDividendSchedule{}, 0, domain.ErrParameterDividendAmountLogic
		}
		sched.DividendPerShare = sched.NetDividendAfterBrokerFee / sched.ProjectShareQty

		// WithholdingTax は投資家個別のdividendレコード作成後にsumして設定する

		if len(latestSchedule) > 0 {
			sched.InvestmentBalance = latestSchedule[0].InvestmentBalance
		} else {
			sched.InvestmentBalance = investment.Amount
		}
	}

	// 법정장부를 위한 부가정보
	sched.ProjectSpvId = p.SpvId
	sched.NumInvestors = int64(len(investors))

	return sched, int64(len(unprocessedShares)), nil
}

func (r repaymentDividendUsecase) _createRepaymentDividendForInvestors(ctx context.Context, sched domain.RepaymentDividendSchedule) (int64, error) {
	// 투자자 정보 확인
	investors, err := r.shareRepository.GetInvestor(ctx,
		cdomain.ShareQueryOptionFunc().WithProjectId(sched.ProjectId),
		cdomain.ShareQueryOptionFunc().WithGroupBy("owner_id"),
	)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get investors by project")
	}

	// 배당, 상황에 따라 투자자별 배당, 상환 레코드 생성
	var sumTax int64
	dividends := make([]domain.Dividend, 0, len(investors))
	for _, investor := range investors {
		shares, err := r.shareRepository.GetAll(ctx,
			cdomain.ShareQueryOptionFunc().WithProjectId(sched.ProjectId),
			cdomain.ShareQueryOptionFunc().WithUserId(investor.Id),
			cdomain.ShareQueryOptionFunc().WithInvestStatuses(domain.OPERATION_INVEST),
		)
		if err != nil {
			return 0, err
		}
		if len(shares) == 0 {
			continue
		}

		// 한 투자자의 하나 이상의 투자에 대해 각각 처리
		for _, share := range shares {
			var receivingAmount, investorTax int64
			var transactionType domain.USER_TRANSACTION_DETAIL_TYPE
			if sched.ReturnType == domain.REPAYMENT_RETURN {
				receivingAmount = sched.RepaymentPerShare * share.Qty
				transactionType = domain.DEPOSIT_REPAYMENT

			} else if sched.ReturnType == domain.DIVIDEND_RETURN {
				receivingAmount = sched.DividendPerShare * share.Qty
				investorTax = int64(float64(receivingAmount) * domain.WITHHOLDING_TAX)
				transactionType = domain.DEPOSIT_DIVIDEND
			} else {
				continue
			}

			sumTax += investorTax

			// 레코드 생성 및 추가
			dividend := domain.Dividend{
				UserId:            investor.Id,
				SpvId:             sched.ProjectSpvId,
				ProjectId:         sched.ProjectId,
				ProjectType:       sched.ProjectType,
				SettlementId:      sched.Id,
				InvestmentAmount:  share.Amount,
				InvestmentQty:     share.Qty,
				Amount:            receivingAmount,
				Qty:               share.Qty,
				DepositAmount:     receivingAmount - investorTax,
				Tax:               investorTax,
				DividendDate:      sched.ExecuteDate,
				ConfirmedDate:     sched.DecideDate,
				TransactionStatus: domain.TRANSACTION_NO,
				TransactionType:   transactionType,
				TransactionId:     r.idGenerator.GenKsuid().String(),
			}

			dividends = append(dividends, dividend)
		}
	}

	if err := r.dividendRepository.CreateMany(ctx, dividends); err != nil {
		return 0, errors.Wrap(err, "failed to create dividends")
	}

	return sumTax, nil
}

func (r repaymentDividendUsecase) GetRepaymentDividendSchedulesTable(
	ctx context.Context, request adomain.GetRepaymentDividendTableRequest,
) (adomain.GetRepaymentDividendTableResponse, error) {
	total, schedules, err := r.fetchAndFormatRepaymentDividendData(ctx, request)
	if err != nil {
		return adomain.GetRepaymentDividendTableResponse{}, errors.Wrap(err, "failed to get repayment dividend schedules")
	}

	return adomain.GetRepaymentDividendTableResponse{
		Items:      schedules,
		TotalCount: total,
	}, nil
}

func (r repaymentDividendUsecase) fetchAndFormatRepaymentDividendData(ctx context.Context, request adomain.GetRepaymentDividendTableRequest) (int64, []domain.InvestmentAndReturns, error) {
	var result []domain.InvestmentAndReturns
	execute, err := r.projectInvestmentExecuteRepository.GetByProjectId(ctx, request.ProjectId)
	if err != nil {
		if errors.Is(err, cdomain.ErrItemNotFound) {
			return 0, result, nil
		}
		return 0, nil, err
	}

	// SkipCount가 0이면 출자정보 레코드를 추가하므로 분배상환에서 가져올 레코드 갯수를 하나 줄인다.
	// SkipCount가 0보다 크면 오프셋을 하나 줄여서(SkipCount = SkipCount -1) 분배상환에서 가져온다. MaxResultCount는 그대로 유지한다.
	if request.SkipCount == 0 {
		// MIR-1209: 分割実行の途中では、募集総額(execute.Amount)からまだ
		// 未処理(SUCCESS_INVEST)のまま残っているshareの金額合計を差し引いた
		// ものを元金残額として見せる。全件処理済みならunprocessedAmountは0に
		// なり、従来どおりexecute.Amountと一致する。一覧・件数ではなく合計額
		// だけが要るため、getUnprocessedShares(全行取得)ではなくGetSumで
		// DB側に集計させる(:376-395と同じ形)。
		unprocessedAmount, err := r.shareRepository.GetSum(ctx,
			cdomain.ShareQueryOptionFunc().WithProjectId(execute.ProjectId),
			cdomain.ShareQueryOptionFunc().WithInvestStatuses(domain.SUCCESS_INVEST),
			cdomain.ShareQueryOptionFunc().WithSumColumn("amount"))
		if err != nil {
			return 0, nil, errors.Wrap(err, "failed to get unprocessed share amount")
		}

		investmentBalance := investmentBalanceAfterUnprocessed(ctx, execute.ProjectId, execute.Amount, int64(unprocessedAmount))

		record := domain.InvestmentAndReturns{
			Id:                0,
			ProjectId:         execute.ProjectId,
			Type:              domain.INVESTMENT_EXECUTE,
			DecideDate:        execute.ExecuteDate.ToFormat(time.DateOnly),
			ExecuteDate:       execute.ExecuteDate.ToFormat(time.DateOnly),
			Status:            domain.RETURN_SCHEDULED,
			InvestmentBalance: investmentBalance,
		}
		if execute.ExecuteStatus == domain.COMPLETED {
			record.Status = domain.RETURN_COMPLETED
		}
		result = append(result, record)
		request.MaxResultCount = request.MaxResultCount - 1
	} else {
		request.SkipCount = request.SkipCount - 1
	}

	// MaxResultCount가 0이면 1을 반환하고 레코드를 추가하지 않는다.
	if request.MaxResultCount == 0 {
		return 1, result, nil
	}

	total, schedules, err := r.repaymentDividendScheduleRepository.GetAll(ctx,
		cdomain.RepaymentDividendQueryOptionFunc().WithProjectId(request.ProjectId),
		cdomain.RepaymentDividendQueryOptionFunc().WithStatus(request.Status),
		cdomain.RepaymentDividendQueryOptionFunc().WithSortBy("execute_date,id"),
		cdomain.RepaymentDividendQueryOptionFunc().WithOrder("asc"),
		cdomain.RepaymentDividendQueryOptionFunc().WithOffset(request.SkipCount),
		cdomain.RepaymentDividendQueryOptionFunc().WithLimit(request.MaxResultCount),
	)
	if err != nil {
		return 0, nil, err
	}

	for _, sched := range schedules {
		record := domain.InvestmentAndReturns{
			Id:                sched.Id,
			ProjectId:         sched.ProjectId,
			Type:              sched.ReturnType,
			DecideDate:        sched.DecideDate.ToFormat(time.DateOnly),
			ExecuteDate:       sched.ExecuteDate.ToFormat(time.DateOnly),
			Status:            sched.Status,
			Repayment:         sched.RepaymentAmount,
			Dividend:          sched.DividendAmount,
			BrokerFee:         sched.BrokerFee,
			WithholdingTax:    sched.WithholdingTax,
			RepaymentReserve:  sched.RepaymentReserve,
			InvestmentBalance: sched.InvestmentBalance,
		}
		result = append(result, record)
	}

	return total + 1, result, nil
}

func (r repaymentDividendUsecase) GetRepaymentDividendSchedulesTableCsv(
	ctx context.Context, request adomain.GetRepaymentDividendTableRequest,
) (adomain.GetRepaymentDividendTableCsvResponse, error) {
	_, schedules, err := r.fetchAndFormatRepaymentDividendData(ctx, request)
	if err != nil {
		return adomain.GetRepaymentDividendTableCsvResponse{},
			errors.Wrap(err, "fail Settlement detail search by project id")
	}

	csvData := r.exportCsv(schedules)

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// CSV 데이터 쓰기
	if err = writer.WriteAll(csvData); err != nil {
		return adomain.GetRepaymentDividendTableCsvResponse{}, errors.Wrap(err, "failed to write csv data")
	}

	return adomain.GetRepaymentDividendTableCsvResponse{Data: buf.Bytes()}, nil
}

var investmentAndReturnsCSVColumns = []string{
	"model.repayment_dividend_schedule.type",               // 내용 内容
	"model.repayment_dividend_schedule.decide_date",        // 일시 日時
	"model.repayment_dividend_schedule.execute_date",       // 결산일 決算日
	"model.repayment_dividend_schedule.status",             // 상태 ステータス
	"model.repayment_dividend_schedule.repayment",          // 출자상환액 出資償還額
	"model.repayment_dividend_schedule.dividend",           // 배당금 利益の分当
	"model.repayment_dividend_schedule.broker_fee",         // 공제(원칙０) 控除(原則０)
	"model.repayment_dividend_schedule.withholding_tax",    // 원천징수세 源泉徴収税
	"model.repayment_dividend_schedule.repayment_reserve",  // 상환적립금 償還プール金
	"model.repayment_dividend_schedule.investment_balance", // 원금잔액 元金残額
}

func (r repaymentDividendUsecase) exportCsv(items []domain.InvestmentAndReturns) cdomain.CSVData {
	csvHeaders := r.translator.Translates(investmentAndReturnsCSVColumns)

	records := [][]string{csvHeaders}
	for _, item := range items {
		record := []string{
			i18nutils.DefaultTranslate("commoncode.INVESTMENT_RETURN_TYPE_CODE." + string(item.Type)),
			item.DecideDate,
			item.ExecuteDate,
			i18nutils.DefaultTranslate("commoncode.INVESTMENT_RETURN_STATUS_CODE." + string(item.Status)),
			fmt.Sprint(item.Repayment),
			fmt.Sprint(item.Dividend),
			fmt.Sprint(item.BrokerFee),
			fmt.Sprint(item.WithholdingTax),
			fmt.Sprint(item.RepaymentReserve),
			fmt.Sprint(item.InvestmentBalance),
		}
		records = append(records, record)
	}

	return records
}

func (r repaymentDividendUsecase) GetInvestmentAndReturnsInfo(ctx context.Context, request adomain.GetInvestmentAndReturnsInfoRequest) (
	adomain.GetInvestmentAndReturnsInfoResponse, error,
) {
	// project 정보 가져오기
	projectId := request.ProjectId
	p, err := r.projectRepository.GetById(ctx, projectId)
	if err != nil {
		return adomain.GetInvestmentAndReturnsInfoResponse{}, errors.Wrap(err, "failed to get project by id")
	}

	// 출자 정보 가져오기
	execute, err := r.projectInvestmentExecuteRepository.GetByProjectId(ctx, projectId)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return adomain.GetInvestmentAndReturnsInfoResponse{}, errors.Wrap(err, "failed to get investment execute by project id")
	}

	// 마지막 배당/상환 스케쥴 가져오기 - 실시한 상환이 있을 경우 원금 잔액과 상환적립금을 적용하기 위함이다. 없으면 초기값 그대로
	_, schedules, err := r.repaymentDividendScheduleRepository.GetAll(ctx,
		cdomain.RepaymentDividendQueryOptionFunc().WithProjectId(projectId),
		cdomain.RepaymentDividendQueryOptionFunc().WithStatus(domain.RETURN_COMPLETED),
		cdomain.RepaymentDividendQueryOptionFunc().WithSortBy("execute_date,id"),
		cdomain.RepaymentDividendQueryOptionFunc().WithOrder("desc"),
		cdomain.RepaymentDividendQueryOptionFunc().WithLimit(1),
	)
	if err != nil {
		return adomain.GetInvestmentAndReturnsInfoResponse{}, errors.Wrap(err, "failed to get last repayment dividend schedule")
	}

	// 배당/상환 스케쥴 합계 가져오기, 코드 최적화보다는 가독성을 위해 호출조건 체크하지 않았다.
	columnNames := []string{"repayment_amount", "net_dividend_after_broker_fee", "broker_fee", "withholding_tax"}
	sums, err := r.repaymentDividendScheduleRepository.GetSums(ctx,
		cdomain.RepaymentDividendQueryOptionFunc().WithProjectId(projectId),
		cdomain.RepaymentDividendQueryOptionFunc().WithStatus(domain.RETURN_COMPLETED),
		cdomain.RepaymentDividendQueryOptionFunc().WithSumColumns(columnNames),
	)
	if err != nil {
		return adomain.GetInvestmentAndReturnsInfoResponse{}, errors.Wrap(err, "failed to get sums of repayment dividend schedule")
	}

	if len(sums) != 4 {
		return adomain.GetInvestmentAndReturnsInfoResponse{}, errors.New("no sums of repayment dividend schedule")
	}

	result := adomain.GetInvestmentAndReturnsInfoResponse{
		ProjectAccount:                 p.BankAccount,
		ProjectDepositor:               p.Depositor,
		ServiceAccount:                 serviceutils.GetServiceAccountNo(),
		TargetShareAmount:              p.TargetShareAmount,
		RaisedAmount:                   p.RaisedAmount,
		InvestedAmount:                 execute.Amount,
		InvestmentBalance:              execute.Amount, // 초기값이다. 상환 내역이 있으면 업데이트된다.
		TotalRepayment:                 int64(sums[0]),
		TotalNetDividendAfterBrokerFee: int64(sums[1]),
		TotalBrokerFee:                 int64(sums[2]),
		TotalWithholdingTax:            int64(sums[3]),
	}

	// 배당 상환이 한 번이라도 이루어진 경우 현재의 잔고와 상환후 자투리 남은 금액을 업데이트 한다.
	if len(schedules) > 0 {
		result.InvestmentBalance = schedules[0].InvestmentBalance
		result.RepaymentReserve = schedules[0].RepaymentReserve
	}

	unprocessedShares, err := r.getUnprocessedShares(ctx, projectId)
	if err != nil {
		return adomain.GetInvestmentAndReturnsInfoResponse{}, errors.Wrap(err, "failed to get unprocessed shares")
	}
	result.UnprocessedShareCount = int64(len(unprocessedShares))

	if err := r.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_product.tab1",
		Title:  fmt.Sprintf("ID :%v", projectId),
		Ip:     request.ClientIp,
		State:  domain.LOOKUP,
	}); err != nil {
		return adomain.GetInvestmentAndReturnsInfoResponse{}, errors.Wrap(err, "failed to create operator history")
	}
	return result, nil
}

func (r repaymentDividendUsecase) DeleteRepaymentDividendSchedule(ctx context.Context, request adomain.DeleteRepaymentDividendRequest) error {
	// 검증: 가장 실시일이 늦은 스케줄이면서 실시하지 않은 스케줄만 삭제할 수 있다.
	_, latestSchedule, err := r.repaymentDividendScheduleRepository.GetAll(ctx,
		cdomain.RepaymentDividendQueryOptionFunc().WithProjectId(request.ProjectId),
		cdomain.RepaymentDividendQueryOptionFunc().WithSortBy("execute_date,id"),
		cdomain.RepaymentDividendQueryOptionFunc().WithOrder("desc"),
		cdomain.RepaymentDividendQueryOptionFunc().WithLimit(1),
	)
	if err != nil {
		return errors.Wrap(err, "failed to get the latest schedule")
	}

	if len(latestSchedule) == 0 {
		return errors.Wrap(cdomain.ErrItemNotFound, "no schedule found")
	}

	if latestSchedule[0].Id != request.ScheduleId {
		return errors.Wrap(cdomain.ErrBadRequest, "only the latest schedule can be deleted")
	}

	if latestSchedule[0].Status == domain.RETURN_COMPLETED {
		return errors.Wrap(cdomain.ErrBadRequest, "cannot delete completed schedule")
	}

	// 배당상환 스케쥴 삭제
	if err := r.repaymentDividendScheduleRepository.Delete(ctx, request.ScheduleId); err != nil {
		return errors.Wrap(err, "failed to delete repayment dividend schedule")
	}

	// 삭제한 스케쥴에 예정되어 있던 배당 삭제
	if err := r.dividendRepository.DeleteByScheduleId(ctx,
		cdomain.DividendQueryOptionFunc().WithScheduleId(request.ScheduleId),
	); err != nil {
		return errors.Wrap(err, "failed to delete dividends by schedule id")
	}

	return nil
}
