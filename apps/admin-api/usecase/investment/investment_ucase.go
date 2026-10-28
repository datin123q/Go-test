package investment

import (
	"bytes"
	"context"
	"encoding/csv"
	"time"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"

	"github.com/pkg/errors"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
)

type investmentUsecase struct {
	txManager                          factory.TxManager
	shareRepository                    cdomain.ShareRepository
	shareService                       cdomain.ShareService
	investmentRepository               cdomain.InvestmentRepository
	investmentService                  cdomain.InvestmentService
	userAccountService                 cdomain.UserAccountService
	operatorHistoryRepository          cdomain.OperatorHistoryRepository
	depositRequestRepository           cdomain.DepositRequestRepository
	transactionInfoHistoryRepository   cdomain.TransactionInfoHistoryRepository
	projectRepository                  cdomain.ProjectRepository
	projectInvestmentExecuteRepository cdomain.ProjectInvestmentExecuteRepository
}

func New(
	txManager factory.TxManager,
	shareRepository cdomain.ShareRepository,
	shareService cdomain.ShareService,
	investmentRepository cdomain.InvestmentRepository,
	investmentService cdomain.InvestmentService,
	userAccountService cdomain.UserAccountService,
	operatorHistoryRepository cdomain.OperatorHistoryRepository,
	depositRequestRepository cdomain.DepositRequestRepository,
	transactionInfoHistoryRepository cdomain.TransactionInfoHistoryRepository,
	projectRepository cdomain.ProjectRepository,
	projectInvestmentExecuteRepository cdomain.ProjectInvestmentExecuteRepository,
) adomain.InvestmentUsecase {
	return &investmentUsecase{
		txManager:                          txManager,
		shareRepository:                    shareRepository,
		shareService:                       shareService,
		investmentRepository:               investmentRepository,
		investmentService:                  investmentService,
		userAccountService:                 userAccountService,
		operatorHistoryRepository:          operatorHistoryRepository,
		depositRequestRepository:           depositRequestRepository,
		transactionInfoHistoryRepository:   transactionInfoHistoryRepository,
		projectRepository:                  projectRepository,
		projectInvestmentExecuteRepository: projectInvestmentExecuteRepository,
	}
}

// userFacingError は、message部分をそのまま利用者へ見せてよいエラー
// (MIR-1204のrepayment_dividend_ucase.goと同じパターン)。errors.Wrapは
// "<message>: <cause>"の形になり原因の英語文がくっついてしまうため、
// Unwrap()でcauseへ辿れる薄いラッパーとして持つ。
type userFacingError struct {
	cause   error
	message string
}

func (e *userFacingError) Error() string { return e.message }
func (e *userFacingError) Unwrap() error { return e.cause }

func (i investmentUsecase) GetAll(ctx context.Context, request adomain.InvestmentGetAllRequest) (adomain.InvestmentGetAllResponse, error) {
	opts := cdomain.InvestmentQueryOptionFunc()

	totalCount, investments, err := i.investmentRepository.GetAll(ctx,
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
		opts.WithInvestStatus(domain.INVEST_STATUS(lo.FromPtrOr(request.InvestStatus, ""))),
		opts.WithInvestorName(request.InvestorName),
		opts.WithSpvName(request.SpvName),
		opts.WithProjectName(request.ProjectName),
		opts.WithProjectType(request.ProjectType),
		opts.WithStartAmount(request.StartAmount),
		opts.WithEndAmount(request.EndAmount),
	)
	if err != nil {
		return adomain.InvestmentGetAllResponse{}, errors.Wrap(err, "failed to get all investments")
	}

	items := lo.Map(investments, func(v domain.Investment, _ int) adomain.InvestmentItemResponse {
		return adomain.InvestmentItemResponse{
			Id:           v.Share.Id,
			ShareId:      v.Share.Id,
			SpvId:        v.SpvId,
			SpvName:      v.SpvName,
			UserName:     v.UserName,
			InvestStatus: v.InvestStatus,
			ProjectType:  v.Share.ProjectType,
			ProjectName:  v.Project.Name,
			Amount:       v.Amount,
			InvestDate:   v.InvestDate,
			OwnerId:      v.Share.OwnerId,
		}
	})

	if err = i.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_accounting.payment_list",
		Ip:     request.ClientIp,
		State:  domain.LOOKUP,
	}); err != nil {
		return adomain.InvestmentGetAllResponse{}, errors.Wrap(err, "failed to create operator history")
	}

	return adomain.InvestmentGetAllResponse{
		TotalCount: totalCount,
		Items:      items,
	}, nil
}

func (i investmentUsecase) GetOne(ctx context.Context, request adomain.InvestmentGetOneRequest) (adomain.InvestmentGetOneResponse, error) {
	v, err := i.investmentService.GetDetail(ctx, request.InvestmentId)
	if err != nil {
		return adomain.InvestmentGetOneResponse{}, errors.Wrap(err, "failed to get investment by id")
	}

	if err = i.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_accounting.payment_list",
		Ip:     request.ClientIp,
		State:  domain.LOOKUP,
	}); err != nil {
		return adomain.InvestmentGetOneResponse{}, errors.Wrap(err, "failed to create operator history")
	}

	return adomain.InvestmentGetOneResponse{
		InvestmentItemDetailResponse: adomain.InvestmentItemDetailResponse{
			Id:                        v.Share.Id,
			OwnerId:                   v.Share.OwnerId,
			ShareId:                   v.Share.Id,
			ProjectId:                 v.ProjectId,
			CategoryType:              v.Share.CategoryType,
			ProjectType:               v.Share.ProjectType,
			InvestStatus:              v.InvestStatus,
			Amount:                    v.Amount,
			Qty:                       v.Qty,
			Price:                     v.Price,
			InvestDate:                v.InvestDate,
			SpvId:                     v.SpvId,
			SpvName:                   v.SpvName,
			ProjectName:               v.Project.Name,
			StartAt:                   v.StartAt,
			EndAt:                     v.EndAt,
			MainImageUrl:              v.MainImageUrl,
			ProjectStatus:             v.ProjectStatus,
			Capacity:                  v.Capacity,
			InvestorType:              v.InvestorType,
			Name:                      lo.If(v.IsUserTypePrivate(), v.FullName()).Else(v.UserName),
			NameKana:                  lo.If(v.IsUserTypePrivate(), v.FullNameKana()).Else(""),
			FirstName:                 v.FirstName,
			FamilyName:                v.FamilyName,
			FirstNameKana:             v.FirstNameKana,
			FamilyNameKana:            v.FamilyNameKana,
			Address:                   v.Address,
			Address2:                  v.Address2,
			City:                      v.City,
			ZipCode:                   v.ZipCode,
			Province:                  v.Province,
			Email:                     v.Email,
			Phone:                     v.MobilePhone,
			BankName:                  v.BankName,
			BankAccountNumber:         v.BankAccountNumber,
			Depositor:                 v.Depositor,
			InvestmentDate:            v.InvestDate,
			ContractAfterPageUrl:      v.ContractAfterPageUrl,
			ContractBeforePageUrl:     v.ContractBeforePageUrl,
			SilentPartnershipsPageUrl: v.SilentPartnershipsPageUrl,
			ProjectReportUrl:          v.ProjectReportUrl,
			ContractStatus: lo.If(v.ContractStatus == "", domain.CONTRACT_NONE).
				Else(domain.CONTRACT_STATUS(v.ContractStatus)),
			ContractBeforePageCheckDate: v.ContractBeforePageCheckDate,
			ContractAfterPageCheckDate:  v.ContractAfterPageCheckDate,
			SilentPartnershipsCheckDate: v.SilentPartnershipsCheckDate,
			ProjectReportCheckDate: lo.If(v.ProjectReportCheckDate.Time().After(v.ProjectReportDate.Time()), v.ProjectReportCheckDate).
				Else(jsonutils.JSONTime{}),
			PublicViewEndAt: v.EndAt,
		},
	}, nil
}

func (i investmentUsecase) DownloadHistories(ctx context.Context, request adomain.InvestmentDownloadHistoriesRequest) (adomain.InvestmentDownloadHistoriesResponse, error) {
	opts := cdomain.InvestmentQueryOptionFunc()

	_, items, err := i.investmentRepository.GetAll(ctx,
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
		opts.WithInvestStatus(domain.INVEST_STATUS(lo.FromPtrOr(request.InvestStatus, ""))),
		opts.WithInvestorName(request.InvestorName),
		opts.WithSpvName(request.SpvName),
		opts.WithProjectName(request.ProjectName),
		opts.WithStartAmount(request.StartAmount),
		opts.WithEndAmount(request.EndAmount),
	)
	if err != nil {
		return adomain.InvestmentDownloadHistoriesResponse{}, errors.Wrap(err, "failed to get all investments")
	}

	csvData := i.investmentService.ExportCSV(items)

	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)

	if err = w.WriteAll(csvData); err != nil {
		return adomain.InvestmentDownloadHistoriesResponse{}, errors.Wrap(err, "failed to write csv")
	}
	w.Flush()

	// TODO OperatorHistory 남기기

	return adomain.InvestmentDownloadHistoriesResponse{Data: buf.Bytes()}, nil
}

func (i investmentUsecase) UpdateStatus(ctx context.Context, request adomain.InvestmentUpdateStatusRequest) (adomain.InvestmentUpdateStatusResponse, error) {
	if request.InvestStatus != domain.CANCEL_INVEST {
		return i.updateStatus(ctx, request, 0)
	}

	// MIR-1295: project_idの解決はWithTxで新しいトランザクションを始める前に
	// 行う。本番の分離レベルはREPEATABLE-READで、ロック無しの読み取り
	// (consistent read)はそのトランザクション内で最初に行われた時点の
	// スナップショットに固定される。これをWithTx内で行うと、後段の
	// projectInvestmentExecuteRepository.GetAllByProjectId(こちらもロック
	// 無し)まで同じ古いスナップショットを見てしまい、ロックの直前に
	// コミットされた出資実行の完了を見落とす恐れがある。project_idはshareの
	// 不変なキーなので、この時点(ロックの前)で読んでも安全。
	unlocked, err := i.shareRepository.GetById(ctx, request.InvestmentId)
	if err != nil {
		return adomain.InvestmentUpdateStatusResponse{}, errors.Wrap(err, "failed to get share by id")
	}

	var response adomain.InvestmentUpdateStatusResponse
	if err := i.txManager.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		response, err = i.updateStatus(txCtx, request, unlocked.ProjectId)
		return err
	}); err != nil {
		return adomain.InvestmentUpdateStatusResponse{}, err
	}

	return response, nil
}

func (i investmentUsecase) updateStatus(ctx context.Context, request adomain.InvestmentUpdateStatusRequest, projectId int64) (adomain.InvestmentUpdateStatusResponse, error) {
	var (
		share domain.Share
		err   error
	)

	if request.InvestStatus == domain.CANCEL_INVEST {
		// MIR-1295: 出資実行(運用開始、ExecuteInvestment)と成立→取消が
		// 同時に走ると、会員の残高とファンドの出資額の両方に同じ金額が
		// 残る競合がある(取消がshareだけロックし、出資実行の記録がまだ
		// コミット前で「記録なし」に見えて通ってしまうため)。project の
		// 行 → share の順でロックしてから判定・読み込みを行う。このトラン
		// ザクション内で最初に行う読み取りが両方ともロック付き(FOR UPDATE)に
		// なるよう、project_idの解決(ロック無し読み取り)はこの関数の外
		// (呼び出し元のUpdateStatus、WithTxの外側)で済ませてある。
		//
		// project側はExecuteInvestmentがTryLockById(FOR UPDATE NOWAIT)を
		// 使うため、ここでは待機するLockByIdを使う。取消の処理中に出資実行を
		// 押すとNOWAITで即座に「他の処理が実行中」になって止まり、出資実行の
		// 処理中に取消を押すとそのコミットを待ってから判定できる(逆にすると
		// デッドロックの恐れがあるため、project→shareの順を守る)。
		if _, err = i.projectRepository.LockById(ctx, projectId); err != nil {
			return adomain.InvestmentUpdateStatusResponse{}, errors.Wrap(err, "failed to lock project by id")
		}

		share, err = i.shareRepository.LockById(ctx, request.InvestmentId)
	} else {
		share, err = i.shareRepository.GetById(ctx, request.InvestmentId)
	}
	if err != nil {
		return adomain.InvestmentUpdateStatusResponse{}, errors.Wrap(err, "failed to get share by id")
	}

	if request.InvestStatus == share.InvestStatus {
		return adomain.InvestmentUpdateStatusResponse{}, errors.Wrap(cdomain.ErrItemConflict, "The same state cannot be changed.")
	}

	// MIR-1295: 手動での状態変更は、申込中→取消、成立→取消(出資実行が
	// 途中でない場合)の2つだけを許す。判定・読み込みは(CANCEL_INVESTの
	// 経路では)ロックのあとに行う。本番の分離レベルはREPEATABLE-READのため、
	// ロックより先に読むとスナップショットに固定され、直前にコミットされた
	// 状態を見落とす恐れがある。
	executionState := domain.InvestmentExecutionNone
	if request.InvestStatus == domain.CANCEL_INVEST {
		executes, executesErr := i.projectInvestmentExecuteRepository.GetAllByProjectId(ctx, share.ProjectId)
		if executesErr != nil {
			return adomain.InvestmentUpdateStatusResponse{}, errors.Wrap(executesErr, "failed to get project investment executes")
		}
		executionState = domain.DetermineInvestmentExecutionState(executes)
	}

	// isSelfFund: ミラリタには自己募集(ZOLOME固有)が無いため常にfalse。
	if err := domain.CanManuallyChangeInvestStatus(share.InvestStatus, request.InvestStatus, executionState, false); err != nil {
		return adomain.InvestmentUpdateStatusResponse{}, &userFacingError{
			cause:   cdomain.ErrCannotManuallyChangeInvestStatus,
			message: err.Error(),
		}
	}

	switch request.InvestStatus {
	case domain.CANCEL_INVEST:
		if err = i.shareService.CancelInvest(ctx, share.Id, request.JWTUser.Id); err != nil {
			return adomain.InvestmentUpdateStatusResponse{}, errors.Wrap(err, "failed to cancel invest")
		}

		// 입금해소로직 수행
		if err = i.resolveDepositAmount(ctx, share.Id, share.OwnerId); err != nil {
			return adomain.InvestmentUpdateStatusResponse{}, errors.Wrap(err, "failed to resolve deposit amount")
		}
	default:
		// CanManuallyChangeInvestStatusでCANCEL_INVEST以外への遷移はすべて
		// 拒否されるため、ここには到達しないはずだが、安全側の網として残す。
		return adomain.InvestmentUpdateStatusResponse{}, errors.Wrap(cdomain.ErrBadRequest, "The state is not supported.")
	}

	// 히스토리 생성 (추후에 좀 더 개선 필요)
	if err = i.transactionInfoHistoryRepository.Create(ctx, domain.TransactionInfoHistory{
		ShareId:       share.Id,
		SubstanceType: string(domain.TRANSACTION_HISTORY_TRANSACTION),
		ProcessMethod: domain.USER_HISTORY_PROCESS_OPERATOR,
		ContentType:   string(domain.TRANSACTION_CHANGED_CONTENT_TRANSACTION_STATUS),
		PrevContent:   string(share.InvestStatus),
		CurrContent:   string(request.InvestStatus),
		Email:         request.JWTUser.Email,
		Name:          request.JWTUser.Name,
		Created:       jsonutils.JSONTime(time.Now()),
	}); err != nil {
		return adomain.InvestmentUpdateStatusResponse{}, errors.Wrap(err, "failed to create transaction info history")
	}

	investment, err := i.investmentRepository.GetById(ctx, request.InvestmentId)
	if err != nil {
		return adomain.InvestmentUpdateStatusResponse{}, errors.Wrap(err, "failed to get investment by id")
	}

	return adomain.InvestmentUpdateStatusResponse{Investment: investment}, nil
}

func (i investmentUsecase) GetAllTransactionInfoHistories(ctx context.Context, request adomain.InvestmentGetAllTransactionInfoHistoriesRequest) (adomain.InvestmentGetAllTransactionInfoHistoriesResponse, error) {
	opts := cdomain.TransactionInfoHistoryQueryOptionFunc()

	totalCount, items, err := i.transactionInfoHistoryRepository.GetAll(ctx,
		opts.WithShareId(request.InvestmentId),
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
	)
	if err != nil {
		return adomain.InvestmentGetAllTransactionInfoHistoriesResponse{}, errors.Wrap(err, "failed to get all transaction info history")
	}

	return adomain.InvestmentGetAllTransactionInfoHistoriesResponse{
		TotalCount: totalCount,
		Items:      items,
	}, nil
}

func (i investmentUsecase) resolveDepositAmount(ctx context.Context, investmentId string, userId int64) error {
	depositRequests, err := i.depositRequestRepository.GetAll(ctx,
		cdomain.DepositRequestQueryOptionFunc().WithUserId(userId),
		cdomain.DepositRequestQueryOptionFunc().WithExpireDepositDate(jsonutils.JSONTime(time.Now())),
	)
	if err != nil {
		return errors.Wrap(err, "fail to get deposit requests")
	}

	for _, deposit := range depositRequests {

		// 현재 투자는 입금해소 로직에서 제외
		// 입금이 이미 되었거나 취소가 된 경우는 제외 (일반적인 경우에는 ExpireDepositDate로 제외되지만, 예외적인 경우가 있을 수 있음)
		if deposit.ShareId == investmentId || deposit.DepositCheckStatus != domain.NO_DEPOSIT {
			continue
		}

		if err = i.userAccountService.DepositCheckProcess(ctx, deposit); err != nil {
			return errors.Wrap(err, "fail to deposit check process")
		}
	}

	return nil
}
