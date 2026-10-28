package service

import (
	"context"
	"errors"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type IWithdrawalService interface {
	GetWithdrawalDetailListByUserId(context.Context, int64, string, string, int, int) (int64, []domain.WithdrawalDetail, error)
	GetWithdrawalDetailById(context.Context, int64) (*domain.WithdrawalDetail, error)
	ConvertWithdrawalStatus(context.Context, *domain.Withdrawal, domain.WITHDRAWAL_STATUS_TYPE) (*domain.Withdrawal, error)
	ReserveWithdrawalForTransfer(context.Context, int64) error
	GetWithdrawalCancelledReason(domain.WithdrawalDetail, string) *string
	GetTotalWithdrawalAmountByUserId(context.Context, int64) (int64, error)
}

var _ IWithdrawalService = (*WithdrawalService)(nil)

type WithdrawalService struct {
	// repository
}

func NewWithdrawalService() IWithdrawalService {
	return &WithdrawalService{}
}

func (s *WithdrawalService) GetWithdrawalDetailListByUserId(ctx context.Context, userId int64, sortBy, order string, offset, limit int) (int64, []domain.WithdrawalDetail, error) {
	total, withdrawalList, err := domain.Withdrawal{}.GetAllByUserId(ctx, userId, sortBy, order, offset, limit)
	if err != nil {
		return int64(0), nil, factory.ErrorDB.New(err)
	}

	return total, withdrawalList, nil
}

func (s *WithdrawalService) GetWithdrawalDetailById(ctx context.Context, id int64) (*domain.WithdrawalDetail, error) {
	withdrawal, err := domain.WithdrawalDetail{}.GetById(ctx, id)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}

	if withdrawal == nil {
		return nil, factory.ErrorNotFound.New(nil)
	}

	return withdrawal, nil
}

// ConvertWithdrawalStatus is used to change the status of the withdrawal.
//
// MIR-1250: checkStatusValidationはメモリ上のwithdrawal.Status(呼び出し元が
// 事前に取得した値)で可否を判定している。同じ出金へのステータス変更(例:
// ユーザーによるキャンセルの二度押し)が並行して届くと、両方ともこの判定を
// 通過してしまい、片方だけでなく両方とも後続の返金処理まで進んでしまう
// (口座に出金額の2倍が入る)。判定と更新の間の隙間そのものをなくすため、
// UPDATE自体を「判定に使った現在のステータス」を条件にしたDB側の条件付き
// 更新にする(Withdrawal.UpdateStatusIfCurrent)。並行した一方が先にコミット
// すると、もう一方のUPDATEはWHERE不一致で0行になり、
// ErrWithdrawalStatusChangedとして「既に処理済み」を呼び出し元へ伝える。
func (s *WithdrawalService) ConvertWithdrawalStatus(ctx context.Context, withdrawal *domain.Withdrawal, status domain.WITHDRAWAL_STATUS_TYPE) (*domain.Withdrawal, error) {
	// 既存の状態から希望の状態への変更が可能かどうかを確認
	if err := checkStatusValidation(withdrawal.Status, status); err != nil {
		return nil, factory.ErrorParameter.New(err)
	}
	currentStatus := withdrawal.Status

	// ステータス変更時に必要なデータ処理
	switch status {
	case domain.WITHDRAWAL_STATUS_COMPLETE:
		withdrawal.CompleteDate = jsonutils.JSONTime(time.Now())
	case domain.WITHDRAWAL_STATUS_CANCEL:
		withdrawal.CompleteDate = jsonutils.JSONTime(time.Now())
	}

	// ステータス変更
	withdrawal.Status = status
	if err := withdrawal.UpdateStatusIfCurrent(ctx, currentStatus); err != nil {
		if errors.Is(err, domain.ErrWithdrawalStatusChanged) {
			return nil, factory.ErrorWithdrawalStatusChanged.New(err)
		}
		return nil, factory.ErrorDB.New(err)
	}

	return withdrawal, nil
}

// ReserveWithdrawalForTransfer is used by the admin's individual re-transfer
// action (FAIL -> WAIT) to atomically claim the withdrawal before creating
// transfer records.
//
// MIR-1275: 呼び出し元(RequestTransferByWithdrawalId)は事前のGetByIdで読んだ
// メモリ上のstatus(FAIL)だけで再送金の可否を判定していたため、個別再送金ボタンの
// 二度押しで同じ出金への再送金リクエストが並行して届くと、両方ともこの判定を
// 通過してしまい、両方ともTransfer/WithdrawalTransferMappingを作成してしまう
// (口座へ二重に送金される)。ConvertWithdrawalStatus(MIR-1250)と同じ
// Withdrawal.UpdateStatusIfCurrentを使い、ステータスの条件付き更新自体を
// 「送金レコードを作ってよいかの予約」として使う。並行した一方が先にコミット
// すると、もう一方の更新はWHERE不一致で0行に終わり、ErrWithdrawalStatusChanged
// を返す。呼び出し元はこのメソッドが成功した場合だけTransfer/
// WithdrawalTransferMappingの作成へ進む。
func (s *WithdrawalService) ReserveWithdrawalForTransfer(ctx context.Context, withdrawalId int64) error {
	withdrawal := domain.Withdrawal{Id: withdrawalId, Status: domain.WITHDRAWAL_STATUS_WAIT}
	if err := withdrawal.UpdateStatusIfCurrent(ctx, domain.WITHDRAWAL_STATUS_FAIL); err != nil {
		if errors.Is(err, domain.ErrWithdrawalStatusChanged) {
			return factory.ErrorWithdrawalStatusChanged.New(err)
		}
		return factory.ErrorDB.New(err)
	}
	return nil
}

func (s *WithdrawalService) GetWithdrawalCancelledReason(withdrawal domain.WithdrawalDetail, by string) *string {
	reason := ""
	switch by {
	case "ADMIN":
		// 出金口座情報に誤りがあり、管理者によってキャンセルされました。出金口座情報を変更した後に依頼してください。
		reason = "出金口座情報に誤りがあり、管理者によってキャンセルされました。出金口座情報を変更してから再度依頼してください。"
	case "USER":
		// YYYY.mm.dd HH:mm にキャンセルを要請しました。
		cancelTime := withdrawal.CompleteDate.Time().Format(time.DateTime)
		reason = cancelTime + "にキャンセルを要請しました。"
	}

	return &reason
}

func (s *WithdrawalService) GetTotalWithdrawalAmountByUserId(ctx context.Context, userId int64) (int64, error) {
	// 出金申請金額 (= 未完了 + クローズ + 送金失敗)

	withdrawalIncompleteSum, err := domain.Withdrawal{}.SumIncompleteAmountByInvestorId(ctx, userId)
	if err != nil {
		return 0, factory.ErrorDB.New(err)
	}

	withdrawalWaitSum, err := domain.Withdrawal{}.SumWaitAmountByInvestorId(ctx, userId)
	if err != nil {
		return 0, factory.ErrorDB.New(err)
	}

	withdrawalFailSum, err := domain.Withdrawal{}.SumFailAmountByInvestorId(ctx, userId)
	if err != nil {
		return 0, factory.ErrorDB.New(err)
	}

	total := withdrawalIncompleteSum + withdrawalWaitSum + withdrawalFailSum

	return total, nil
}

// checkStatusValidation is used to check whether the status can be changed.
func checkStatusValidation(current, target domain.WITHDRAWAL_STATUS_TYPE) error {
	switch target {
	case domain.WITHDRAWAL_STATUS_COMPLETE:
		// FAIL -> COMPLETE
		if current != domain.WITHDRAWAL_STATUS_FAIL {
			return errors.New("only FAIL withdrawal can be COMPLETE")
		}
	case domain.WITHDRAWAL_STATUS_INCOMPLETE:
		// FAIL -> INCOMPLETE
		if current != domain.WITHDRAWAL_STATUS_FAIL {
			return errors.New("only FAIL withdrawal can be INCOMPLETE")
		}
	case domain.WITHDRAWAL_STATUS_CANCEL:
		// FAIL -> CANCEL || INCOMPLETE -> CANCEL
		if current != domain.WITHDRAWAL_STATUS_FAIL && current != domain.WITHDRAWAL_STATUS_INCOMPLETE {
			return errors.New("only FAIL or INCOMPLETE withdrawal can be CANCEL")
		}
	default:
		return errors.New("invalid status")
	}

	return nil
}
