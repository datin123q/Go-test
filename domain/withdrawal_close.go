package domain

import (
	"context"
	"encoding/json"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/pkg/errors"
)

type WITHDRAWAL_REQUEST_STATUS_TYPE string

const (
	WITHDRAWAL_REQUEST_STATUS_TYPE_CODE                                = "WITHDRAWAL_REQUEST_STATUS_TYPE"
	WITHDRAWAL_REQUEST_STATUS_COMPLETE  WITHDRAWAL_REQUEST_STATUS_TYPE = "COMPLETE"
	WITHDRAWAL_REQUEST_STATUS_WAIT      WITHDRAWAL_REQUEST_STATUS_TYPE = "WAIT"
	WITHDRAWAL_REQUEST_STATUS_SENDING   WITHDRAWAL_REQUEST_STATUS_TYPE = "SENDING"
)

func (WithdrawalClose) GetWithdrawalRequestStatusTypeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(WITHDRAWAL_REQUEST_STATUS_COMPLETE), Name: i18nutils.Translate("commoncode.WITHDRAWAL_REQUEST_STATUS_TYPE.WITHDRAWAL_REQUEST_STATUS_COMPLETE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WITHDRAWAL_REQUEST_STATUS_WAIT), Name: i18nutils.Translate("commoncode.WITHDRAWAL_REQUEST_STATUS_TYPE.WITHDRAWAL_REQUEST_STATUS_WAIT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WITHDRAWAL_REQUEST_STATUS_SENDING), Name: i18nutils.Translate("commoncode.WITHDRAWAL_REQUEST_STATUS_TYPE.WITHDRAWAL_REQUEST_STATUS_SENDING", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: WITHDRAWAL_REQUEST_STATUS_TYPE_CODE,
		Codes:     codes,
	}
}

type WithdrawalClose struct {
	Id           int64                          `json:"id"`
	CloserId     int64                          `json:"closer_id" xorm:"not null"`
	CloseDate    jsonutils.JSONTime             `json:"close_date" xorm:"not null"`
	TransferDate jsonutils.JSONTime             `json:"transfer_date"`
	Total        int64                          `json:"total" xorm:"not null default 0"`
	Amount       int64                          `json:"amount" xorm:"not null default 0"`
	Fee          int64                          `json:"fee" xorm:"not null default 0"`
	Status       WITHDRAWAL_REQUEST_STATUS_TYPE `json:"status" xorm:"not null default 'WAIT'"`
}

func (wc WithdrawalClose) TableName() string {
	return "withdrawal_close"
}
func (wc WithdrawalClose) SelectAll() string {
	return wc.TableName() + ".*"
}

type WithdrawalCloseMapping struct {
	Id           int64 `json:"id"`
	CloseId      int64 `json:"close_id" xorm:"not null"`
	WithdrawalId int64 `json:"withdrawal_id" xorm:"not null"`
}

func (wc WithdrawalCloseMapping) TableName() string {
	return "withdrawal_close_mapping"
}

func (wc WithdrawalCloseMapping) SelectAll() string {
	return wc.TableName() + ".*"
}

func (wc *WithdrawalClose) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(wc)
}

// UpdateTransferDateAndStatus クリーンアーキテクチャではUpdateメソッドを使用するよう修正
// wc.Idの行を更新する(引数のidと分けると食い違いうるため、TryLockByIdが返した行の
// Idをそのまま使う)。更新対象をtransfer_date/statusの2列に.Colsで明示するのは、
// TryLockByIdが返す実データ入りのwcを渡すと、xormの非ゼロ値判定で他の列まで
// ロック取得時点の値で上書きされてしまうため(MIR-1175)。
func (wc *WithdrawalClose) UpdateTransferDateAndStatus(ctx context.Context, t jsonutils.JSONTime) error {
	wc.TransferDate = t
	wc.Status = WITHDRAWAL_REQUEST_STATUS_SENDING
	if _, err := factory.DB(ctx).ID(wc.Id).Cols("transfer_date", "status").Update(wc); err != nil {
		return err
	}
	return nil
}

// UpdateStatusByIds クリーンアーキテクチャではUpdateManyメソッドを使用するよう修正
func (wc *WithdrawalClose) UpdateStatusByIds(ctx context.Context, ids []int64, status WITHDRAWAL_REQUEST_STATUS_TYPE) error {
	if _, err := factory.DB(ctx).In("id", ids).Update(&WithdrawalClose{Status: status}); err != nil {
		return err
	}
	return nil
}

// ErrWithdrawalCloseLocked はTryLockByIdが、他のリクエストがロック中のため
// 待たずに諦めたことを表す。
var ErrWithdrawalCloseLocked = errors.New("withdrawal close is locked by another request")

// ErrWithdrawalCloseNotFound はTryLockByIdが対象行を見つけられなかったことを表す。
var ErrWithdrawalCloseNotFound = errors.New("withdrawal close not found")

// TryLockById は締め行をロックして、RequestTransferByClosedIdの多重実行を防ぐ。
// 他のリクエストがロック中なら待たずにErrWithdrawalCloseLockedを返す
// (FOR UPDATE NOWAIT)。MIR-1204のTryLockByIdと同じ作法(NOWAIT+MySQLエラー3572判定)
// を踏襲する(MIR-1175)。sqliteはNOWAIT構文を持たないため、単体テストでは
// FOR UPDATE NOWAIT自体を付けない(User.LockByIdと同じ考え方)。
func (WithdrawalClose) TryLockById(ctx context.Context, id int64) (WithdrawalClose, error) {
	query := "SELECT * FROM withdrawal_close WHERE id = ?"
	engine, err := factory.Engine(ctx)
	if err != nil {
		return WithdrawalClose{}, err
	}
	if engine.DriverName() != "sqlite3" {
		query += " FOR UPDATE NOWAIT"
	}

	var wc WithdrawalClose
	has, err := factory.DB(ctx).SQL(query, id).Get(&wc)
	if err != nil {
		if factory.IsLockNotAvailableError(err) {
			return WithdrawalClose{}, errors.Wrap(ErrWithdrawalCloseLocked, "failed to lock withdrawal close: locked by another request")
		}
		return WithdrawalClose{}, errors.Wrap(err, "failed to try-lock withdrawal close by id")
	}
	if !has {
		return WithdrawalClose{}, errors.Wrap(ErrWithdrawalCloseNotFound, "failed to try-lock withdrawal close by id")
	}

	return wc, nil
}

// CanRequestTransfer は、締めがまだ一括送金を実行できる状態(WAIT)かを返す。
// TryLockByIdでロックした後にこれを確認することで、既に実行済み・進行中の締めへの
// 多重実行を防ぐ(MIR-1175)。
func (wc WithdrawalClose) CanRequestTransfer() bool {
	return wc.Status == WITHDRAWAL_REQUEST_STATUS_WAIT
}

// UpdateFeeById クリーンアーキテクチャでは以下のメソッドを使用するよう修正
// WithdrawalCloseMapping.GetAll -> Withdrawal.GetSum -> WithdrawalClose.Update
func (wc *WithdrawalClose) UpdateFeeById(ctx context.Context, id int64) error {
	query := "UPDATE withdrawal_close " +
		"SET fee = (SELECT SUM(fee) FROM withdrawal WHERE id IN " +
		"(SELECT withdrawal_id FROM withdrawal_close_mapping WHERE close_id = ?)) " +
		"WHERE id = ? "

	_, err := factory.DB(ctx).Exec(query, id, id)
	if err != nil {
		return err
	}

	return nil
}

type WithdrawalCloseDetail struct {
	WithdrawalClose `xorm:"extends"`
	CloserName      string `json:"closer_name"`
}

func (wcd *WithdrawalCloseDetail) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"total":         wcd.WithdrawalClose.Total,
		"amount":        wcd.WithdrawalClose.Amount,
		"fee":           wcd.WithdrawalClose.Fee,
		"closer_id":     wcd.WithdrawalClose.CloserId,
		"closer_name":   wcd.CloserName,
		"close_date":    wcd.WithdrawalClose.CloseDate,
		"transfer_date": wcd.WithdrawalClose.TransferDate,
		"status":        wcd.WithdrawalClose.Status,
	})
}

func (wcd *WithdrawalCloseDetail) GetAll(ctx context.Context, sortBy, order string, offset, limit int) (int64, []WithdrawalCloseDetail, error) {
	var (
		total int64
		list  []WithdrawalCloseDetail
	)

	q := factory.DB(ctx).Table("withdrawal_close")
	q = q.Join("LEFT", "admin_user", "withdrawal_close.closer_id = admin_user.id")
	q = q.Select("withdrawal_close.*, admin_user.name as closer_name")

	if err := modelhelper.SetSortOrder(q, sortBy, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	total, err := q.Limit(limit, offset).FindAndCount(&list)
	if err != nil {
		return 0, nil, err
	}

	return total, list, nil
}

func (wcm *WithdrawalCloseMapping) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(wcm)
}

func (wcm *WithdrawalCloseMapping) CreateMulti(ctx context.Context, list []WithdrawalCloseMapping) (int64, error) {
	return factory.DB(ctx).Insert(&list)
}

// GetCloseIdsByWithdrawalIds クリーンアーキテクチャではGetAllを使用するよう修正
func (wcm *WithdrawalCloseMapping) GetCloseIdsByWithdrawalIds(ctx context.Context, withdrawalIds []int64) ([]int64, error) {

	var closeIds []int64
	if err := factory.DB(ctx).Table("withdrawal_close_mapping").Distinct("close_id").In("withdrawal_id", withdrawalIds).Find(&closeIds); err != nil {
		return nil, err
	}

	return closeIds, nil
}
