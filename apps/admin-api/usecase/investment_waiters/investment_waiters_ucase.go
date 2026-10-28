package investment_waiters

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/idgenerator"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-share-core-api/service"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

const (
	investmentAmountLimit = int64(5_000_000)
)

type investmentWaitersUsecase struct {
	cfg                             *config.Config
	txManager                       factory.TxManager
	idGenerator                     idgenerator.IdGenerator
	investmentWaitersRepository     cdomain.InvestmentWaitersRepository
	userRepository                  cdomain.UserRepository
	projectRepository               cdomain.ProjectRepository
	investmentRequestRepository     cdomain.InvestmentRequestRepository
	investmentWaitHistoryRepository cdomain.InvestmentWaitHistoryRepository
	mailService                     cdomain.MailService
	investmentWaitersService        cdomain.InvestmentWaitersService
	investmentWaitHistoryService    cdomain.InvestmentWaitHistoryService
}

func New(
	cfg *config.Config,
	txManager factory.TxManager,
	idGenerator idgenerator.IdGenerator,
	investmentWaitersRepository cdomain.InvestmentWaitersRepository,
	userRepository cdomain.UserRepository,
	projectRepository cdomain.ProjectRepository,
	investmentRequestRepository cdomain.InvestmentRequestRepository,
	investmentWaitHistoryRepository cdomain.InvestmentWaitHistoryRepository,
	mailService cdomain.MailService,
	investmentWaitersService cdomain.InvestmentWaitersService,
	investmentWaitHistoryService cdomain.InvestmentWaitHistoryService,
) adomain.InvestmentWaitersUsecase {
	return &investmentWaitersUsecase{
		cfg:                             cfg,
		txManager:                       txManager,
		idGenerator:                     idGenerator,
		investmentWaitersRepository:     investmentWaitersRepository,
		userRepository:                  userRepository,
		projectRepository:               projectRepository,
		investmentRequestRepository:     investmentRequestRepository,
		investmentWaitHistoryRepository: investmentWaitHistoryRepository,
		mailService:                     mailService,
		investmentWaitersService:        investmentWaitersService,
		investmentWaitHistoryService:    investmentWaitHistoryService,
	}
}

func (i investmentWaitersUsecase) GetAll(ctx context.Context, request adomain.InvestmentWaitersGetAllRequest) (adomain.InvestmentWaitersGetAllResponse, error) {
	opts := cdomain.InvestmentWaitersQueryOptionFunc()

	totalCount, investmentWaiters, err := i.investmentWaitersRepository.GetAll(ctx,
		opts.WithProjectId(lo.FromPtr(request.ProjectId)),
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
	)
	if err != nil {
		return adomain.InvestmentWaitersGetAllResponse{}, errors.Wrap(err, "failed to get investment waiters")
	}

	items := make([]adomain.InvestmentWaitersResponse, 0, len(investmentWaiters))
	for _, investmentWaiter := range investmentWaiters {
		items = append(items, adomain.InvestmentWaitersResponse{}.FromInvestmentWaiters(investmentWaiter))
	}

	return adomain.InvestmentWaitersGetAllResponse{
		Items:      items,
		TotalCount: totalCount,
	}, nil
}

func (i investmentWaitersUsecase) GetOne(ctx context.Context, request adomain.InvestmentWaitersGetOneRequest) (adomain.InvestmentWaitersGetOneResponse, error) {
	investmentWaiters, err := i.investmentWaitersRepository.GetById(ctx, lo.FromPtr(request.InvestmentWaitersId))
	if err != nil {
		return adomain.InvestmentWaitersGetOneResponse{}, errors.Wrap(err, "failed to get investment waiter")
	}

	project, err := i.projectRepository.GetById(ctx, investmentWaiters.ProjectId)
	if err != nil {
		return adomain.InvestmentWaitersGetOneResponse{}, errors.Wrap(err, "failed to get project")
	}

	// 繰上げ履歴(MIR-1068)。当選が1件も無ければ空スライスのまま返す。
	winningItems, err := i.investmentWaitersRepository.GetWinningsByWaiterId(ctx, investmentWaiters.Id)
	if err != nil {
		return adomain.InvestmentWaitersGetOneResponse{}, errors.Wrap(err, "failed to get winnings")
	}
	winnings := lo.Map(winningItems, func(item domain.InvestmentWaitersWinningItem, _ int) adomain.InvestmentWaitersWinningItem {
		return adomain.InvestmentWaitersWinningItem{}.FromInvestmentWaitersWinningItem(item)
	})

	return adomain.InvestmentWaitersGetOneResponse{
		InvestmentWaitersResponse: adomain.InvestmentWaitersResponse{}.FromInvestmentWaiters(investmentWaiters),
		InvestmentProposalMaxQty:  min(investmentWaiters.Qty, project.RemainingQty),
		InvestmentProposalMinQty:  project.MinShareQty,
		IsInvestmentAmountLimit:   project.IsAmountLimit,
		InvestmentAmountLimit:     lo.If(project.IsAmountLimit, investmentAmountLimit).Else(-1),
		PricePerQty:               project.SharePrice,
		Winnings:                  winnings,
	}, nil
}

func (i investmentWaitersUsecase) Download(ctx context.Context, request adomain.InvestmentWaitersDownloadRequest) (adomain.InvestmentWaitersDownloadResponse, error) {
	opts := cdomain.InvestmentWaitersQueryOptionFunc()

	_, investmentWaiters, err := i.investmentWaitersRepository.GetAll(ctx,
		opts.WithProjectId(lo.FromPtr(request.ProjectId)),
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
	)
	if err != nil {
		return adomain.InvestmentWaitersDownloadResponse{}, errors.Wrap(err, "failed to get investment waiters")
	}

	csvData := i.investmentWaitersService.ExportCSV(investmentWaiters)

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// CSV 데이터 쓰기
	if err = writer.WriteAll(csvData); err != nil {
		return adomain.InvestmentWaitersDownloadResponse{}, errors.Wrap(err, "failed to write csv data")
	}

	return adomain.InvestmentWaitersDownloadResponse{Data: buf.Bytes()}, nil
}

func (i investmentWaitersUsecase) StatusUpdate(ctx context.Context, request adomain.InvestmentWaitersStatusUpdateRequest) (adomain.InvestmentWaitersStatusUpdateResponse, error) {
	var response adomain.InvestmentWaitersStatusUpdateResponse

	investmentWaiters, err := i.investmentWaitersRepository.GetById(ctx, lo.FromPtr(request.InvestmentWaitersId))
	if err != nil {
		return response, errors.Wrap(err, "failed to get investment waiter")
	}

	switch request.Status {
	case domain.DEPOSIT_REQUEST:
		if err = i.investmentWaitersService.ChangeDepositRequestStatus(
			ctx,
			investmentWaiters,
			request.JWTUser.Id,
		); err != nil {
			return response, errors.Wrap(err, "failed to change deposit request status")
		}
	case domain.WAITING_END:
		if err = i.investmentWaitersService.ChangeWaitingEndStatus(
			ctx,
			investmentWaiters,
			request.JWTUser.Id,
		); err != nil {
			return response, errors.Wrap(err, "failed to change waiting end status")
		}
	case domain.PROPOSAL:
		if err = i.investmentWaitersService.ChangeProposalStatus(
			ctx,
			investmentWaiters,
			request.JWTUser.Id,
		); err != nil {
			return response, errors.Wrap(err, "failed to change proposal status")
		}
	case domain.CANCELLED:
		// この分岐はCancel usecase(/cancelエンドポイント)と違いLockByIdを使わず
		// トランザクションも張っていないため、同時に繰上げが実行された場合の完全な
		// 排他は無いが、明らかに繰上げ済みの待機者を誤ってキャンセルすることは防ぐ
		// (②の敵対的検証で、Cancel usecaseの修正後もこの分岐に同じ穴が残っている
		// ことが指摘された)。
		if err := i.ensureCancellable(ctx, investmentWaiters); err != nil {
			return response, err
		}
		if err = i.investmentWaitersService.ChangeCancelledStatus(
			ctx,
			investmentWaiters,
			request.JWTUser.Id,
		); err != nil {
			return response, errors.Wrap(err, "failed to change cancelled status")
		}
	default:
		return response, errors.New("not supported status")
	}

	return response.WithInvestmentWaiters(
		adomain.InvestmentWaitersResponse{}.FromInvestmentWaiters(investmentWaiters),
	), nil
}

// Deprecated: Use Propose instead
func (i investmentWaitersUsecase) InvestmentPropose(ctx context.Context, request adomain.InvestmentWaitersInvestmentProposeRequest) (adomain.InvestmentWaitersInvestmentProposeResponse, error) {
	var response adomain.InvestmentWaitersInvestmentProposeResponse

	investmentWaiters, err := i.investmentWaitersRepository.GetById(ctx, lo.FromPtr(request.InvestmentWaitersId))
	if err != nil {
		return response, errors.Wrap(err, "failed to get investment waiter")
	}

	if err = i.investmentWaitersService.ChangeProposalStatus(
		ctx,
		investmentWaiters,
		request.JWTUser.Id,
	); err != nil {
		return response, errors.Wrap(err, "failed to change proposal status")
	}

	return response.WithInvestmentWaiters(
		adomain.InvestmentWaitersResponse{}.FromInvestmentWaiters(investmentWaiters),
	), nil
}

func (i investmentWaitersUsecase) Propose(ctx context.Context, request adomain.InvestmentWaitersProposeRequest) (adomain.InvestmentWaitersProposeResponse, error) {
	var response adomain.InvestmentWaitersProposeResponse

	investmentWaiters, err := i.investmentWaitersRepository.GetById(ctx, lo.FromPtr(request.InvestmentWaitersId))
	if err != nil {
		return adomain.InvestmentWaitersProposeResponse{}, errors.Wrap(err, "failed to get investment waiter")
	}

	project, err := i.projectRepository.GetById(ctx, investmentWaiters.ProjectId)
	if err != nil {
		return adomain.InvestmentWaitersProposeResponse{}, errors.Wrap(err, "failed to get project")
	}

	if investmentWaiters.Qty < request.Qty {
		// 申込口数＜提案口数であればエラー
		return adomain.InvestmentWaitersProposeResponse{}, errors.Wrapf(cdomain.ErrBadRequest, "err proposal qty is greater than applied qty (%v)", investmentWaiters.Qty)
	}

	if project.RemainingQty < request.Qty {
		// 残口数＜提案口数であればエラー
		return adomain.InvestmentWaitersProposeResponse{}, errors.Wrapf(cdomain.ErrBadRequest, "err proposal qty is greater than remaining qty (%v)", project.RemainingQty)
	}

	if project.MinShareQty < project.RemainingQty {
		if request.Qty < project.MinShareQty {
			// 提案口数＜最小投資口数の場合はエラー
			// NOTE: 残口数が最小投資口数未満の場合は、最小投資口数チェックをスキップ
			return adomain.InvestmentWaitersProposeResponse{}, errors.Wrapf(cdomain.ErrBadRequest, "err proposal qty is less than minimum share qty (%v)", project.MinShareQty)
		}
	}

	updatedInvestmentWaiters := investmentWaiters
	updatedInvestmentWaiters.Status = domain.PROPOSAL

	if _, err = i.investmentWaitersRepository.Update(ctx, updatedInvestmentWaiters); err != nil {
		return response, errors.Wrap(err, "failed to update investment waiters")
	}

	user, err := i.userRepository.GetById(ctx, investmentWaiters.UserId)
	if err != nil {
		return adomain.InvestmentWaitersProposeResponse{}, errors.Wrap(err, "failed to get user")
	}

	requestId := i.idGenerator.GenUuidV4().String()
	proposedUrl := fmt.Sprintf("%s/product/view/%v?request_id=%v", i.cfg.ServiceConfig.Url.UserServiceDomain, investmentWaiters.ProjectId, requestId)

	if _, err = i.investmentRequestRepository.Create(ctx, domain.InvestmentRequest{
		InvestmentWaitersId: investmentWaiters.Id,
		ProjectId:           investmentWaiters.ProjectId,
		RequestId:           requestId,
		Url:                 proposedUrl,
		IsExpire:            false,
		Email:               user.Email,
		Qty:                 request.Qty, // AS-IS 남은투자구수, TO-BE 투자제안구수
	}); err != nil {
		return adomain.InvestmentWaitersProposeResponse{}, errors.Wrap(err, "failed to create investment request")
	}

	// 메일발송
	// https://illuminarean.atlassian.net/wiki/spaces/ESJ/pages/891257292 -> 4D 템플릿 참조
	if err = i.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.INVEST_OFFER_TEMPLATE, // TODO : 4D 템플릿으로 변경이 필요합니다.
		Sender:       i.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{user.Email},
		MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
		Details: map[string]interface{}{
			"Receiver":              user.Name, // 투자자명
			"ServiceUrl":            i.cfg.ServiceConfig.Url.UserServiceDomain,
			"InvestedQty":           message.NewPrinter(language.Japanese).Sprintf("%d", investmentWaiters.Qty),          // 신청구수
			"InvestableQty":         message.NewPrinter(language.Japanese).Sprintf("%d", request.Qty),                    // 제안구수
			"TotalInvestmentAmount": message.NewPrinter(language.Japanese).Sprintf("%d", request.Qty*project.SharePrice), // 합계투자금액
			"ProductView":           proposedUrl,
		},
	}); err != nil {
		return adomain.InvestmentWaitersProposeResponse{}, errors.Wrap(err, "failed to send mail")
	}

	if err = i.investmentWaitHistoryService.CreateStatusHistory(
		ctx,
		domain.InvestmentWaitersComparer{Before: investmentWaiters, After: updatedInvestmentWaiters},
		domain.USER_HISTORY_PROCESS_OPERATOR,
		request.JWTUser.Id,
	); err != nil {
		return adomain.InvestmentWaitersProposeResponse{}, errors.Wrap(err, "failed to create status history")
	}

	if err = i.investmentWaitHistoryRepository.Create(ctx, domain.InvestmentWaitHistory{
		InvestmentWaitId: investmentWaiters.Id,
		SubstanceType:    string(domain.INVESTMENT_WAIT_HISTORY_MAIL),
		Name:             user.Name,
		Email:            user.Email,
		MailSendDate:     jsonutils.JSONTime(time.Now()),
		ProcessMethod:    domain.USER_HISTORY_PROCESS_OPERATOR,
	}); err != nil {
		return adomain.InvestmentWaitersProposeResponse{}, errors.Wrap(err, "failed to create mail send history")
	}

	return response.WithInvestmentWaiters(
		adomain.InvestmentWaitersResponse{}.FromInvestmentWaiters(investmentWaiters),
	), nil
}

// InvestmentRequest は管理者が投資待機者に対して投資申請を実行するユースケース。
// 待機者のステータス検証（WAITINGであること）を行い、WaitInvestで投資処理・メール送信・ステータス更新を一括実行する。
// WaitInvestはShare作成、DepositRequest作成、入金チェック、InvestStandbyTerminationMail送信、ステータスWAITING_END更新を行う。
func (i investmentWaitersUsecase) InvestmentRequest(ctx context.Context, request adomain.InvestmentWaitersInvestmentRequestRequest) (adomain.InvestmentWaitersInvestmentRequestResponse, error) {
	var response adomain.InvestmentWaitersInvestmentRequestResponse

	// 待機者を取得
	investmentWaiters, err := i.investmentWaitersRepository.GetById(ctx, lo.FromPtr(request.InvestmentWaitersId))
	if err != nil {
		return response, errors.Wrap(err, "failed to get investment waiter")
	}

	// ステータスがWAITING・WAITING_END以外(PROPOSAL・CANCELLED等)なら投資申請不可。
	// WAITING_ENDは「繰上げ完了」と「打ち切り」の両方を意味するため、ここでは弾かず、
	// 繰上げ実績の有無による判定はWaitInvest(LockById配下のInvestmentWaiterCanAdvance)に
	// 委ねる。設計書の「既存データの再繰上げは手動操作に限る」を実現する経路がここ(MIR-1068)。
	if investmentWaiters.Status != domain.WAITING && investmentWaiters.Status != domain.WAITING_END {
		return response, errors.Wrapf(cdomain.ErrBadRequest, "investment waiter status is not WAITING or WAITING_END (current: %s)", investmentWaiters.Status)
	}

	// 申込口数を超える口数では投資申請不可
	if investmentWaiters.Qty < request.Qty {
		return response, errors.Wrapf(cdomain.ErrBadRequest, "investment request qty is greater than applied qty (%v)", investmentWaiters.Qty)
	}

	// ユーザーを取得
	user, err := i.userRepository.GetById(ctx, investmentWaiters.UserId)
	if err != nil {
		return response, errors.Wrap(err, "failed to get user")
	}

	// WaitInvestで投資処理・メール送信・ステータス更新を一括実行
	investmentService := service.InvestmentService{}
	if _, err = investmentService.WaitInvest(ctx, &user, &investmentWaiters, request.Qty, false); err != nil {
		return response, errors.Wrap(err, "failed to execute wait invest")
	}

	return adomain.InvestmentWaitersInvestmentRequestResponse{
		ProjectId: investmentWaiters.ProjectId,
		UserId:    investmentWaiters.UserId,
		Qty:       request.Qty,
	}, nil
}

// ensureCancellable は、待機者がキャンセル可能かどうかを判定する(MIR-1068)。
// Cancel usecase(/cancelエンドポイント)とStatusUpdate usecase(CANCELLED分岐)の
// 両方から呼ばれる。判定式を2箇所に複製すると片方だけ直して食い違う事故が起きる
// ため、共通化した(②の敵対的検証で指摘)。
// 当選済みかどうかは中間テーブル(投資済みかどうかの一次情報)を確認する(MIR-1118)。
// MIR-1068で繰上げ後もStatusがWAITINGのまま残るようになったため、
// Status != WAITINGだけでは一部繰上げ済みの待機者のキャンセルを防げない
// (①の敵対的検証で発見。user-apiのCancelは既に対応済みだったが、admin-apiは
// 対応漏れだった)。1件でも当選実績があればキャンセル不可という判定自体は
// user-apiと同じ(一部繰上げ済みの残りだけをキャンセルする機能は無い)。
// LockAndSumAdvancedQtyはFOR UPDATEを使うため、行ロックとして機能するのは
// LockByIdで待機者行をロックした後、同一トランザクション内(txCtx)でこの関数を
// 呼んだ場合だけ(Cancel usecaseの経路)。StatusUpdateのCANCELLED分岐のように
// トランザクション外(ctx)で呼ぶと、ロックはこのSELECT文の間しか持たず、
// 同時に繰上げが実行された場合の完全な排他にはならない。
func (i investmentWaitersUsecase) ensureCancellable(ctx context.Context, waiter domain.InvestmentWaiters) error {
	if waiter.Status != domain.WAITING {
		return errors.Wrapf(
			cdomain.ErrBadRequest,
			"investment waiter is not cancellable (waiter status: %s)",
			waiter.Status,
		)
	}

	advancedQty, err := i.investmentWaitersRepository.LockAndSumAdvancedQty(ctx, waiter.Id)
	if err != nil {
		return errors.Wrap(err, "failed to resolve advanced qty")
	}
	if !domain.InvestmentWaiterIsStillPending(waiter.Status, advancedQty) {
		return errors.Wrapf(
			cdomain.ErrBadRequest,
			"investment waiter is not cancellable (waiter status: %s, advanced qty: %d)",
			waiter.Status,
			advancedQty,
		)
	}

	return nil
}

// Cancel は当選前のキャンセル待ち申込をキャンセルするユースケース。
func (i investmentWaitersUsecase) Cancel(ctx context.Context, request adomain.InvestmentWaitersCancelRequest) (adomain.InvestmentWaitersCancelResponse, error) {
	var response adomain.InvestmentWaitersCancelResponse

	if err := i.txManager.WithTx(ctx, func(txCtx context.Context) error {
		lockedInvestmentWaiters, err := i.investmentWaitersRepository.LockById(
			txCtx,
			lo.FromPtr(request.InvestmentWaitersId),
		)
		if err != nil {
			return errors.Wrap(err, "failed to lock investment waiter")
		}

		if err := i.ensureCancellable(txCtx, lockedInvestmentWaiters); err != nil {
			return err
		}

		return i.investmentWaitersService.ChangeCancelledStatus(
			txCtx,
			lockedInvestmentWaiters,
			request.JWTUser.Id,
		)
	}); err != nil {
		return response, errors.Wrap(err, "failed to change cancelled status")
	}

	return adomain.InvestmentWaitersCancelResponse{}, nil
}
