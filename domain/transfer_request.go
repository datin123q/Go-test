package domain

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/pkg/errors"
)

type TRANSFER_RESULT_CODE string

const (
	TRANSFER_RESULT_SUCCESS TRANSFER_RESULT_CODE = "1" // 완료
	TRANSFER_RESULT_WAIT    TRANSFER_RESULT_CODE = "2" // 미완료
	TRANSFER_RESULT_FAIL    TRANSFER_RESULT_CODE = "8" // 마감기간 지남
)

type TRANSFER_REQUEST_STATUS string

const (
	TRANSFER_REQUEST_WAIT    TRANSFER_REQUEST_STATUS = "TRANSFER_REQUEST_WAIT"    // 미완료
	TRANSFER_REQUEST_SUCCESS TRANSFER_REQUEST_STATUS = "TRANSFER_REQUEST_SUCCESS" // 완료
	TRANSFER_REQUEST_FAIL    TRANSFER_REQUEST_STATUS = "TRANSFER_REQUEST_FAIL"    // 마감기간 지남

)

type COMPLETE_TRANSFER_STATUS string

const (
	COMPLETE_TRANSFER_NONE    COMPLETE_TRANSFER_STATUS = "COMPLETE_TRANSFER_NONE"    // 미실행
	COMPLETE_TRANSFER_WAIT    COMPLETE_TRANSFER_STATUS = "COMPLETE_TRANSFER_WAIT"    // 진행중
	COMPLETE_TRANSFER_SUCCESS COMPLETE_TRANSFER_STATUS = "COMPLETE_TRANSFER_SUCCESS" // 완료
)

type TransferRequest struct {
	Id                 int64                    `json:"id" xorm:"pk autoincr not null "`
	RequestDate        jsonutils.JSONTime       `json:"request_date" xorm:"transfer_date"`
	RequestTotalAmount int64                    `json:"request_total_amount" xorm:"request_total_amount not null"`
	RequestTotalCnt    int                      `json:"request_total_cnt" xorm:"request_total_cnt not null"`
	TransferTotalFee   int64                    `json:"transfer_fee" xorm:"transfer_fee" swagger:"desc(수수료)"`
	ErrorCode          string                   `json:"error_code" xorm:"error_code" swagger:"desc(에러 CODE)"`
	ErrorMsg           string                   `json:"error_msg" swagger:"desc(에러 메시지 )"`
	RequestStatus      TRANSFER_REQUEST_STATUS  `json:"request_status" xorm:"request_status"`
	ResultCode         TRANSFER_RESULT_CODE     `json:"result_code" xorm:"result_code"` //송금 상태
	ApplyNo            string                   `json:"apply_no" xorm:"apply_no"`
	ApplyEndDatetime   string                   `json:"apply_end_datetime" xorm:"apply_end_datetime"`
	CompletedTransfer  COMPLETE_TRANSFER_STATUS `json:"completed_transfer" xorm:"completed_transfer"`
	Transfers          []Transfer               `json:"transfers" xorm:"-"`
	Created            jsonutils.JSONTime       `json:"created" xorm:"created"`
	Updated            jsonutils.JSONTime       `json:"updated" xorm:"updated"`
}

// NewTransferRequest はtransfersをTRANSFER_LIMIT件ずつのTransferRequestへ分割して
// 作成する。以前はチャンクの境目でTransfer{}.InsertManyへ毎回transfers全体
// (チャンク分のスライスではなく)を渡しており、チャンクが2つ以上になる件数
// (100件以上)では、既に挿入済みの先頭チャンク分も含めて重複して挿入していた
// (MIR-1181)。今回は各チャンクの開始位置(start)を追いながら、その回で確定した
// 範囲transfers[start:i+1]だけを渡す形に直した。これによりRequestTotalCntも
// 「全体件数」ではなく「そのTransferRequestに実際に紐づく件数」になり、
// ItemIdも各チャンクの範囲内で採番済みの値だけが使われる。
// 境界条件はcnt >= TRANSFER_LIMIT(以前はcnt > TRANSFER_LIMIT)にしている。
// 以前の条件だとチャンクがTRANSFER_LIMIT(98)件ではなく99件になっていた。
// GMOの総合振込1リクエストあたりの明細数上限は9,999件で確認済みのため、
// 98でも99でも上限には抵触しない。TRANSFER_LIMIT=98がなぜこの値なのか
// (GMO側の制約ではなく、このシステム独自の分割単位)はコードからもチケットの
// 経緯からも分かっていない。境界条件をTRANSFER_LIMIT件ちょうどに揃えたのは、
// 定数の名前が示す値と実際のチャンクサイズを一致させるための整合であり、
// 何らかの上限を回避するためではない(MIR-1181)。
func (TransferRequest) NewTransferRequest(ctx context.Context, transfers []Transfer) (err error) {
	q := factory.DB(ctx).Table("transfer_request").Select("transfer_request.*")
	var cnt int
	var totalAmount int64
	totalCnt := len(transfers)
	start := 0
	for i, transfer := range transfers {
		cnt++
		totalAmount = totalAmount + transfer.TransferAmount
		transfers[i].ItemId = cnt

		if (cnt >= TRANSFER_LIMIT) || (totalCnt-1) == i {
			chunk := transfers[start : i+1]

			r := TransferRequest{
				RequestDate:        jsonutils.JSONTime(time.Now()),
				RequestTotalAmount: totalAmount,
				RequestTotalCnt:    len(chunk),
				RequestStatus:      TRANSFER_REQUEST_WAIT,
				CompletedTransfer:  COMPLETE_TRANSFER_NONE,
			}

			if _, err := q.Insert(&r); err != nil {
				return errors.Wrap(err, "Fail to Create Transfer Request Type ")
			}

			for j := range chunk {
				chunk[j].RequestId = r.Id
			}
			if _, err := (Transfer{}).InsertMany(ctx, chunk); err != nil {
				return err
			}
			r.Transfers = chunk
			factory.Logger(ctx).Info("Create Transfer Reset")

			start = i + 1
			cnt = 0
			totalAmount = 0
		}
	}

	return err
}

func (TransferRequest) GetAll(ctx context.Context, sortby, order string, offset, limit int, startDate, endDate string) (int64, []TransferRequest, error) {
	q := factory.DB(ctx).Table("transfer_request").Select("transfer_request.*")
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(q, "request_date", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	items := make([]TransferRequest, 0, 0)
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (TransferRequest) GetTransferRequestByRequestStatus(ctx context.Context, status TRANSFER_REQUEST_STATUS) (int64, []TransferRequest, error) {
	q := factory.DB(ctx).Table("transfer_request").Select("transfer_request.*")
	items := make([]TransferRequest, 0, 0)
	q.Where("request_status = ?", status)
	var totalCount int64
	totalCount, err := q.FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	for i, _ := range items {
		_, items[i].Transfers, err = Transfer{}.GetByRequestId(ctx, items[i].Id)
		if err != nil {
			return 0, nil, err
		}
	}

	return totalCount, items, nil
}

func (TransferRequest) GetTransferRequestByResultCode(ctx context.Context, code TRANSFER_RESULT_CODE) (int64, []TransferRequest, error) {
	q := factory.DB(ctx).Table("transfer_request").Select("transfer_request.*")
	items := make([]TransferRequest, 0)
	q.Where("result_code = ?", code)
	var totalCount int64
	totalCount, err := q.FindAndCount(&items)
	if err != nil {
		return 0, nil, errors.Wrap(err, "Fail to Get Transfer Request By Result Code")
	}

	for i, _ := range items {
		_, items[i].Transfers, err = Transfer{}.GetByRequestId(ctx, items[i].Id)
		if err != nil {
			return 0, nil, errors.Wrap(err, "Fail to Get Transfer By  Transfer Request Id")
		}
	}

	return totalCount, items, nil
}

func (TransferRequest) GetTransferRequestByCompletedTransfer(ctx context.Context, code COMPLETE_TRANSFER_STATUS) (int64, []TransferRequest, error) {
	q := factory.DB(ctx).Table("transfer_request").Select("transfer_request.*")
	items := make([]TransferRequest, 0)
	q.Where("completed_transfer = ?", code)
	var totalCount int64
	totalCount, err := q.FindAndCount(&items)
	if err != nil {
		return 0, nil, errors.Wrap(err, "Fail to Get Transfer Request By Completed Transfer")
	}

	for i, _ := range items {
		_, items[i].Transfers, err = Transfer{}.GetByRequestId(ctx, items[i].Id)
		if err != nil {
			return 0, nil, errors.Wrap(err, "Fail to Get Transfer By  Transfer Request Id")
		}
	}

	return totalCount, items, nil
}

func (TransferRequest) GetByID(ctx context.Context, id int64) (*TransferRequest, error) {
	var t TransferRequest
	has, err := factory.DB(ctx).ID(id).Get(&t)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	_, t.Transfers, err = Transfer{}.GetByRequestId(ctx, t.Id)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (t *TransferRequest) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(t)
}

func (t *TransferRequest) Update(ctx context.Context) error {
	_, err := factory.DB(ctx).ID(t.Id).Update(t)
	return err
}

// UpdateTransferTotalFeeは、TransferTotalFee(手数料合計)だけを強制的にUPDATE文へ含める。
// 通常のUpdate()はXORMの既定どおりゼロ値のフィールドをUPDATE対象から除くため、受取口座が
// 全員GMOあおぞら宛て(同行宛ては0円)で手数料合計が本当に0円になったケース(MIR-1123)でも
// 書き込まれず、古い合計が残ってしまう。ほかのフィールドは通常どおり自動検出させたいため、
// Cols(全体を絞り込む)ではなくMustCols(この1カラムだけ強制する)を使う。
func (t *TransferRequest) UpdateTransferTotalFee(ctx context.Context) error {
	_, err := factory.DB(ctx).ID(t.Id).MustCols("transfer_fee").Update(t)
	return err
}
