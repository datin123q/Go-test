package service

import (
	"context"
	"strconv"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	investmentwaitersrepo "github.com/illuminarean/energy-shares-backend-repo/repository/investment_waiters/mysql"
)

type InvestmentService struct {
	txManager                   factory.TxManager
	investmentWaitersRepository cdomain.InvestmentWaitersRepository
}

func (InvestmentService) Invest(
	ctx context.Context, p *domain.Project, u *domain.User, qty int64, investMailType domain.INVEST_MAIL_TYPE,
	isWaitInvest bool, isProposalInvest bool, isWaitInvestFinalPortion bool,
) (*domain.Share, error) {

	if p == nil {
		return nil, factory.ErrorParameter.New(errors.New("project is nil"))
	}

	if u == nil {
		return nil, factory.ErrorParameter.New(errors.New("user is nil"))
	}

	amount := p.SharePrice * qty

	// 잔여 구매 수량 확인
	if !p.CheckShareQty(qty, amount) {
		return nil, factory.ErrorInsufficientQty.New(errors.New("insufficient qty"))
	}

	shares, err := p.InvestProject(u.Id, qty, p.SharePrice, amount, isWaitInvest, isProposalInvest, isWaitInvestFinalPortion)
	if err != nil {
		return nil, factory.ErrorFailInvest.New(errors.Wrap(err, "fail invest project"))
	}

	if _, err := shares.Create(ctx); err != nil {
		return nil, factory.ErrorDB.New(errors.Wrap(err, "fail create share"))
	}

	deposit, err := domain.DepositRequest{}.NewDepositRequest(ctx, shares, isWaitInvest)
	if err != nil {
		return nil, errors.Wrap(err, "fail create deposit request")
	}

	if err := p.CheckInvestmentWait(ctx, amount, qty); err != nil {
		return nil, errors.Wrap(err, "fail check investment wait")
	}

	if err := shares.UpdateInvestStatus(ctx, domain.REQUEST_INVEST, u.Id, investMailType); err != nil {
		return nil, factory.ErrorDB.New(errors.Wrap(err, "fail update share status"))
	}

	if p.ProjectType == domain.EQUITY {
		account, err := domain.UserAccount{}.GetByUserId(ctx, u.Id)
		if err != nil {
			return nil, factory.ErrorDB.New(errors.Wrap(err, "fail get user account"))
		}

		// 예치금에 잔액이 있는 경우 입금 처리
		if err = account.DepositCheckProcess(ctx, deposit, investMailType); err != nil {
			return nil, factory.ErrorDB.New(errors.Wrap(err, "fail deposit check process"))
		}
	}

	factory.Logger(ctx).WithFields(logrus.Fields{
		"share_id":               shares.Id,
		"user_id":                shares.OwnerId,
		"project_id":             shares.ProjectId,
		"project_type":           shares.ProjectType,
		"invest_status":          shares.InvestStatus,
		"invest_date":            shares.InvestDate,
		"invest_amount":          shares.Amount,
		"invest_qty":             shares.Qty,
		"deposit_request_id":     deposit.Id,
		"deposit_expire_date":    deposit.ExpireDepositDate,
		"deposit_account_number": deposit.DepositAccountNumber,
		"deposit_check_status":   deposit.DepositCheckStatus,
		"request_amount":         deposit.RequestAmount,
		"deposit_amount":         deposit.DepositAmount,
		"refund_amount":          deposit.RefundAmount,
		"deposit_date":           deposit.DepositDate,
	}).Info("Success Invest")

	return domain.Share{}.GetById(ctx, shares.Id)
}

func (s InvestmentService) WaitInvest(
	ctx context.Context, user *domain.User, waiter *domain.InvestmentWaiters, qty int64, isProposalInvest bool,
) (*domain.Share, error) {
	if user == nil {
		return nil, factory.ErrorParameter.New(errors.New("user is nil"))
	}
	if waiter == nil {
		return nil, factory.ErrorParameter.New(errors.New("investment waiter is nil"))
	}

	repository := s.investmentWaitersRepository
	if repository == nil {
		repository = investmentwaitersrepo.New(factory.NewDBFactory())
	}

	var share *domain.Share
	waitInvest := func(txCtx context.Context) error {
		lockedWaiter, err := repository.LockById(txCtx, waiter.Id)
		if err != nil {
			return errors.Wrap(err, "failed to lock investment waiter")
		}
		if lockedWaiter.UserId != user.Id {
			return errors.Wrapf(
				cdomain.ErrBadRequest,
				"investment waiter user does not match (waiter user: %d, investment user: %d)",
				lockedWaiter.UserId,
				user.Id,
			)
		}
		// 繰上げ済み累計(当選実績)は中間テーブル(投資済みかどうかの一次情報)を行ロックで
		// 確認する(MIR-1118)。旧列investment_waiters.share_idへのフォールバックは
		// PR3(MIR-1163)で削除した。
		advancedQty, err := repository.LockAndSumAdvancedQty(txCtx, lockedWaiter.Id)
		if err != nil {
			return errors.Wrap(err, "failed to resolve advanced qty")
		}
		remaining := lockedWaiter.Qty - advancedQty
		// MIR-1068: 繰上げ後もStatusがWAITINGのまま残るため、Status != WAITINGだけでは
		// 二重繰上げ・口数超過を防げない。繰上げ可能残(remaining)ベースの判定に置き換える。
		if !domain.InvestmentWaiterCanAdvance(lockedWaiter.Status, remaining, advancedQty > 0) {
			return errors.Wrapf(
				cdomain.ErrBadRequest,
				"investment waiter cannot be processed (status: %s, remaining: %d)",
				lockedWaiter.Status,
				remaining,
			)
		}
		// 申込口数(Qty)を超えて繰上げることはない(2026-08-25 河内さん・小倉さん確認済み)。
		if qty > remaining {
			return errors.Wrapf(
				cdomain.ErrBadRequest,
				"qty exceeds remaining advanceable qty (qty: %d, remaining: %d)",
				qty,
				remaining,
			)
		}

		share, err = s.waitInvest(txCtx, user, &lockedWaiter, qty, isProposalInvest, remaining)
		if err != nil {
			return err
		}

		*waiter = lockedWaiter
		return nil
	}

	run := func() error {
		if factory.InTransaction(ctx) {
			return waitInvest(ctx)
		}
		txManager := s.txManager
		if txManager == nil {
			engine, err := factory.Engine(ctx)
			if err != nil {
				return factory.ErrorDB.New(errors.Wrap(err, "failed to get db engine"))
			}
			txManager = factory.NewTxManager(engine)
		}
		return txManager.WithTx(ctx, waitInvest)
	}

	if err := run(); err != nil {
		// 別々の待機者への同時繰上げでデッドロックすることがある(敵対的検証で実測・
		// 再現)。データは壊れないため、専用のエラーで再試行を促す。判定はwaitInvest
		// 全体の戻り値に対して一括で行う(②レビューで発見: 個々の文単位で判定すると、
		// 実際にデッドロックが起きる中間テーブルへのINSERT(waitInvest内のCreate)を
		// 取りこぼす。errors.Wrapで包まれていてもerrors.Asでラップを辿って検出できる)。
		if factory.IsDeadlockError(err) {
			// ErrConcurrentUpdateへ置き換えると元のMySQLエラー(どの文でぶつかったか)が
			// ログから消えるため、警告として1行残す(②レビューで指摘)。
			factory.Logger(ctx).WithError(err).Warn("WaitInvest detected a deadlock, converting to ErrConcurrentUpdate")
			return nil, errors.Wrap(cdomain.ErrConcurrentUpdate, "concurrent advance conflict")
		}
		return nil, err
	}

	return share, nil
}

func (s InvestmentService) waitInvest(
	ctx context.Context, user *domain.User, waiter *domain.InvestmentWaiters, qty int64, isProposalInvest bool, remaining int64,
) (*domain.Share, error) {
	// 繰上げ可能かどうか(status・繰上げ可能残)は、唯一の呼び出し元であるWaitInvestが
	// LockAndSumAdvancedQtyとInvestmentWaiterCanAdvanceで既に検証済み(MIR-1068)。
	// waiter.ShareIdはLockByIdが返す実カラムの生値(旧繰上げ方式の名残)であり、解決済みの
	// 値ではないためここでは参照しない。

	project, err := domain.Project{}.GetById(ctx, waiter.ProjectId)
	if err != nil {
		return nil, factory.ErrorParameter.New(err)
	}

	amount := project.SharePrice * qty

	if !project.CheckShareQty(qty, amount) {
		return nil, factory.ErrorInsufficientQty.New(err)
	}

	// この繰上げで繰上げ可能残を使い切るか(=この待機者への最後の繰上げか)。WaitInvestで
	// qty <= remainingは検証済み。使い切る場合は最小投資口数チェックをスキップし
	// (2026-08-25 決定事項①)、この関数末尾でWAITING_ENDへ遷移させる(MIR-1068)。
	isWaitInvestFinalPortion := qty >= remaining
	share, err := s.Invest(ctx, project, user, qty, domain.WAITER_IMMEDIATE, true, isProposalInvest, isWaitInvestFinalPortion)
	if err != nil {
		return nil, factory.ErrorFailInvest.New(err)
	}
	// waiter.ShareId は中間テーブル経由のサブクエリでのみ埋まる読み取り専用フィールドなので、
	// ここでは更新されない(MIR-1118)。作成した share の Id が必要な呼び出し元は、戻り値の
	// share を参照すること。
	if _, err := (&domain.InvestmentWaitersWinning{
		InvestmentWaitersId: waiter.Id,
		ShareId:             share.Id,
	}).Create(ctx); err != nil {
		return nil, factory.ErrorDB.New(err)
	}

	if err := project.Update(ctx); err != nil {
		return nil, factory.ErrorDB.New(err)
	}

	// 繰上げ済み累計が申請口数に達したときだけ待機完了へ遷移する(MIR-1068)。①の敵対的
	// 検証で、この条件が抜けて毎回WAITING_ENDへ遷移していた実装漏れを発見・修正した。
	// メール送信より前に行っても最終結果は変わらない(メール送信が失敗すればこの関数は
	// エラーを返し、呼び出し元がトランザクション全体をロールバックするため)。
	// waiter.Statusが既にWAITING_END(手動での再繰上げ、MIR-1068)の場合はChangeStatusを
	// 呼ばない。同じ値への遷移でも履歴(InvestmentWaitHistory)に1行残ってしまうため
	// (②の敵対的検証で指摘)。
	if isWaitInvestFinalPortion && waiter.Status != domain.WAITING_END {
		operatorId, _ := domain.AdminUser{}.GetSuperAdminId(ctx)
		if err := waiter.ChangeStatus(ctx, domain.WAITING_END, operatorId); err != nil {
			return nil, factory.ErrorDB.New(err)
		}
	}

	if share.InvestStatus != domain.SUCCESS_INVEST {
		deposit, err := domain.DepositRequest{}.GetDepositRequestByShareId(ctx, share.Id)
		if err != nil {
			return nil, factory.ErrorDB.New(err)
		}

		va, err := domain.UserVirtualAccount{}.GetVaByUserId(ctx, user.Id, project.ProjectType)
		if err != nil {
			return nil, factory.ErrorDB.New(err)
		}

		if err := (&domain.Mail{}).InvestStandbyTerminationMail(ctx, serviceutils.GetServiceInfoMail(), user.Email, user.Name, project.Name, share.InvestDate.ToString(),
			deposit.ExpireDepositDate.ToString(), strconv.Itoa(int(amount)), deposit.DepositBankName, deposit.DepositBranchName, deposit.DepositAccountNumber,
			va.VaHolderNameKana, project.ProjectType); err != nil {
			return nil, factory.ErrorAWSSendMail.New(err)
		}
	}
	return share, nil
}
