package domain

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

const (
	EQUITY_EXPIRE_DAY             = 2
	WAIT_INVEST_EQUITY_EXPIRE_DAY = 1
)

// 입금 요청
// 사용자가 투자했을경우 입금 요청이 생성됨
type DepositRequest struct {
	Id                   int64                `json:"id" xorm:"pk autoincr"`
	UserId               int64                `json:"user_id" xorm:"user_id index" swagger:"desc(투자자 ID)"`
	ShareId              string               `json:"share_id" xorm:"UNIQUE NOT NULL index" swagger:"desc(구매 내역 ID)"`
	ProjectId            int64                `json:"project_id" xorm:"project_id NOT NULL" swagger:"desc(프로젝트 ID)"`
	ProjectType          PROJECT_TYPE         `json:"project_type" xorm:"project_type" swagger:"desc(프로젝트 타입)"`
	RequestAmount        int64                `json:"request_amount" swagger:"desc(입금 요청 금액)"`
	DepositAmount        int64                `json:"deposit_amount" swagger:"desc(실제 입금 금액: 일부 입금이거나 전체 입금일 수 있음)"`
	RefundAmount         int64                `json:"refund_amount" swagger:"desc(환불 금액)"`
	DepositDate          jsonutils.JSONTime   `json:"deposit_date" swagger:"desc(입금 일자)"`
	DepositCheckStatus   DEPOSIT_CHECK_STATUS `json:"deposit_check_status" swagger:"enum(OK_DEPOSIT|NO_DEPOSIT)"`
	UpdateDate           jsonutils.JSONTime   `json:"update_date" xorm:"update_date updated" swagger:"desc(은행 입금 내역 업데이트 일자)"`
	ExpireDepositDate    jsonutils.JSONTime   `json:"expire_deposit_date" swagger:"desc(입금 완료 예정일 휴일을 제외한 3일)"`
	DepositBankName      string               `json:"deposit_bank_name" swagger:"desc(입금 은행)"`
	DepositBranchName    string               `json:"deposit_branch" swagger:"desc(입금 지점)"`
	DepositBranchCode    string               `json:"deposit_branch_code" swagger:"desc(입금 지점)"`
	DepositAccountNumber string               `json:"deposit_account_number" swagger:"desc(입금 계좌 번호)"`
	Created              jsonutils.JSONTime   `json:"-" xorm:"created" swagger:"desc(생성일)"`
	Deleted              jsonutils.JSONTime   `json:"-" xorm:"deleted"`
}

func (DepositRequest) TableName() string {
	return "deposit_request"
}

func (d DepositRequest) SelectAll() string {
	return d.TableName() + ".*"
}

func (d DepositRequest) WithDepositCheckStatus(depositCheckStatus DEPOSIT_CHECK_STATUS) DepositRequest {
	d.DepositCheckStatus = depositCheckStatus
	return d
}

func (d DepositRequest) WithDepositAmount(depositAmount int64) DepositRequest {
	d.DepositAmount = depositAmount
	return d
}

func (d DepositRequest) WithDepositDate(depositDate jsonutils.JSONTime) DepositRequest {
	d.DepositDate = depositDate
	return d
}

// TODO: isWaitInvestを抽選の当選でも利用しているので整理する
// 先着の通常当選=>2日後の23:59
// キャンセル待ち繰上・抽選の当選=>1日後の13:00
func (DepositRequest) NewDepositRequest(ctx context.Context, share *Share, isWaitInvest bool) (*DepositRequest, error) {
	uv, err := UserVirtualAccount{}.GetVaByUserId(ctx, share.OwnerId, share.ProjectType)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}
	if uv == nil {
		return nil, factory.ErrorNotFoundVirtualAccount.New(err)
	}

	var expireDepositDate jsonutils.JSONTime
	if isWaitInvest {
		expireDepositDate = GetEquityExpireDepositDateForWaitInvest(share.InvestDate)
	} else {
		expireDepositDate = GetEquityExpireDepositDate(share.InvestDate)
	}

	deposit := &DepositRequest{
		UserId:               share.OwnerId,
		ShareId:              share.Id,
		ProjectId:            share.ProjectId,
		ProjectType:          share.ProjectType,
		RequestAmount:        share.Amount,
		DepositAmount:        0,
		DepositAccountNumber: uv.VaAccountNumber,
		DepositBankName:      uv.VaBankName,
		DepositBranchName:    uv.VaBranchName,
		DepositBranchCode:    uv.VaBranchCode,
		DepositCheckStatus:   NO_DEPOSIT,
		ExpireDepositDate:    expireDepositDate,
	}

	if _, err := deposit.Create(ctx); err != nil {
		return nil, factory.ErrorDB.New(err)
	}
	return deposit, nil
}

func (d *DepositRequest) OkDeposit(ctx context.Context) error {
	d.DepositCheckStatus = OK_DEPOSIT
	d.DepositAmount = d.RequestAmount
	d.DepositDate = jsonutils.JSONTime(time.Now())
	if err := d.Update(ctx); err != nil {
		return factory.ErrorDB.New(err)
	}
	return nil
}

// GetEquityExpireDepositDate 入金期限日時を取得する
func GetEquityExpireDepositDate(investDate jsonutils.JSONTime) jsonutils.JSONTime {
	var expireDate time.Time
	if serviceutils.EnvIsProd() || serviceutils.EnvIsStg() {
		expireDate = investDate.Time().AddDate(0, 0, EQUITY_EXPIRE_DAY)
	} else {
		expireDate = investDate.Time()
	}

	return jsonutils.JSONTime(time.Date(expireDate.Year(), expireDate.Month(), expireDate.Day(), 23, 59, 59, 0, expireDate.Location()))
}

// 入金期限日時を取得する(キャンセル待ち繰上・抽選の当選用)
func GetEquityExpireDepositDateForWaitInvest(investDate jsonutils.JSONTime) jsonutils.JSONTime {
	expireDate := investDate.Time().AddDate(0, 0, WAIT_INVEST_EQUITY_EXPIRE_DAY)

	return jsonutils.JSONTime(time.Date(expireDate.Year(), expireDate.Month(), expireDate.Day(), 13, 0, 0, 0, expireDate.Location()))
}

func (DepositRequest) GetDepositRequestByShareId(ctx context.Context, shareId string) (*DepositRequest, error) {
	var item DepositRequest
	q := factory.DB(ctx).Table("deposit_request").Select("deposit_request.*")
	q = q.Where("deposit_request.share_id = ?", shareId)

	has, err := q.Get(&item)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, factory.ErrorNotFound.New(nil)
	}

	return &item, nil
}

func (DepositRequest) GetAll(ctx context.Context, sortby, order string, offset, limit int, startDate, endDate string) (int64, []DepositRequest, error) {
	q := factory.DB(ctx).Table("deposit_request").Select("deposit_request.*")
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(q, "deposit_date", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	items := make([]DepositRequest, 0, 0)
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (d *DepositRequest) NeedAmount() int64 {
	return d.RequestAmount - d.DepositAmount
}

func (DepositRequest) GetByID(ctx context.Context, id int64) (*DepositRequest, error) {
	var d DepositRequest
	has, err := factory.DB(ctx).ID(id).Get(&d)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &d, nil
}

// GetNoDepositRequestByUserId 만료되지 않은 입금된 입금요청 조회
func (DepositRequest) GetNoDepositRequestByUserId(ctx context.Context, userId int64) ([]DepositRequest, error) {
	items := make([]DepositRequest, 0)
	err := factory.DB(ctx).Where("user_id = ?", userId).And("deposit_check_status = ?", NO_DEPOSIT).
		Asc("created").Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (d *DepositRequest) CheckDepositSuccess() DEPOSIT_MATCH_STATUS {
	if d.DepositAmount == d.RequestAmount {
		return SAME_DEPOSIT
	}

	if d.RequestAmount > d.DepositAmount {
		return LACK_DEPOSIT
	}
	return OVER_DEPOSIT
}

func (d *DepositRequest) DepositCancel(ctx context.Context) error {
	switch d.ProjectType {
	case EQUITY:
		return d.EquityDepositCancel(ctx)
	case LOAN:
		return d.LoanDepositCancel(ctx)
	}
	return errors.New("Not Support Project Type")
}

func (d *DepositRequest) EquityDepositCancel(ctx context.Context) error {
	if d.DepositCheckStatus == CANCEL_DEPOSIT {
		return errors.New("already been canceled ")
	}

	amount := d.DepositAmount - d.RefundAmount

	if amount > 0 {
		account, err := UserAccount{}.GetByUserId(ctx, d.UserId)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
		// 투자금 입금
		_, err = account.DepositBalance(ctx, amount)
		if err != nil {
			return errors.Wrap(err, "failed to deposit to user account ")
		}

		d.RefundAmount = d.RefundAmount + amount
	}

	d.DepositCheckStatus = CANCEL_DEPOSIT

	return d.Update(ctx)
}

func (d *DepositRequest) LoanDepositCancel(ctx context.Context) error {
	if d.DepositCheckStatus == CANCEL_DEPOSIT {
		return errors.New("already been canceled ")
	}

	d.DepositCheckStatus = CANCEL_DEPOSIT

	return d.Update(ctx)
}

func (d *DepositRequest) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(d)
}

func (d *DepositRequest) Update(ctx context.Context) error {
	_, err := factory.DB(ctx).ID(d.Id).Update(d)
	return err
}

func (DepositRequest) Delete(ctx context.Context, id int64) (err error) {
	_, err = factory.DB(ctx).ID(id).Delete(&DepositRequest{})
	return
}

type DepositRequestDetail struct {
	InvestmentWaitersShareId string `xorm:"investment_waiters_share_id"`
	DepositRequest           `xorm:"extends"`
	Share                    `xorm:"extends"`
}
