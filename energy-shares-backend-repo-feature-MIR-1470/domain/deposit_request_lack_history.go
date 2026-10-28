package domain

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

// 부족 입금의뢰 내역
type DepositRequestLackHistory struct {
	Id                  int64              `json:"id" xorm:"pk autoincr"`
	UserId              int64              `json:"user_id" swagger:"desc(사용자 ID)"`
	DepositRequestId    int64              `json:"deposit_request_id" swagger:"desc(입금 요청 ID)"`
	RequestExecuteDate  jsonutils.JSONTime `json:"request_execute_date" swagger:"desc(입금 의뢰 실시일)"`
	DepositExpectedDate jsonutils.JSONTime `json:"deposit_expected_date" swagger:"desc(입금 완료 예정일)"`
	Amount              int64              `json:"amount" swagger:"desc(입금의뢰 금액)"`
	AccountNumber       string             `json:"account_number" swagger:"desc(입금용 가상계좌)"`
	ProjectType         PROJECT_TYPE       `json:"project_type" swagger:"desc(프로젝트 종류),enum(LOAN|EQUITY)"`
	ProjectId           int64              `json:"-" xorm:"project_id" swagger:"desc(프로젝트 ID - 상품 ID)"`
	ShareId             string             `json:"-" swagger:"desc(Share ID)"`
	Created             jsonutils.JSONTime `json:"-" xorm:"created"`
}

func (DepositRequestLackHistory) GetByUserId(ctx context.Context, user_id int64, sortby, order string, offset, limit int) (int64, []DepositRequestLackHistory, error) {

	query := factory.DB(ctx).Table("deposit_request_lack_history")
	query.Select("deposit_request_lack_history.*, deposit_request.deposit_check_status, project.end_at as end_at").
		Where("deposit_request_lack_history.user_id = ? and deposit_expected_date > ? ", user_id, time.Now()).
		Join("LEFT", "deposit_request", "deposit_request_lack_history.deposit_request_id = deposit_request.id").
		Join("LEFT", "project", "deposit_request_lack_history.project_id = project.id").
		Where("deposit_request.deposit_check_status = ? and project.end_at > ?", NO_DEPOSIT, time.Now())

	if err := modelhelper.SetSortOrder(query, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []DepositRequestLackHistory
	var totalCount int64
	totalCount, err := query.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (u *DepositRequestLackHistory) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(u)
}
func (DepositRequestLackHistory) CreateDepositRequestLackHistory(ctx context.Context, userId, amount, depositRequestId, projectId int64, projectType PROJECT_TYPE, accountNumber, shareId string, requestExecuteDate, depositExpectedDate jsonutils.JSONTime) error {

	drLackHistory := DepositRequestLackHistory{
		UserId:              userId,
		ProjectId:           projectId,
		ShareId:             shareId,
		DepositRequestId:    depositRequestId,
		RequestExecuteDate:  requestExecuteDate,
		DepositExpectedDate: depositExpectedDate,
		AccountNumber:       accountNumber,
		ProjectType:         projectType,
		Amount:              amount,
	}
	if _, err := drLackHistory.Create(ctx); err != nil {
		return err
	}
	return nil
}

func (DepositRequestLackHistory) RequestLackDepositMail(ctx context.Context, depositRequest *DepositRequest, project *Project) error {
	user, err := User{}.GetById(ctx, depositRequest.UserId)
	if err != nil {
		return err
	}

	share, err := Share{}.GetById(ctx, depositRequest.ShareId)
	if err != nil {
		return err
	}

	if share == nil {
		log.Println("share is nil")
		return nil
	}

	userVirtualAccount, err := UserVirtualAccount{}.GetVaByUserId(ctx, depositRequest.UserId, project.ProjectType)
	if err != nil {
		return err
	}

	var mail Mail
	lackAmount := strconv.FormatInt(depositRequest.RequestAmount-depositRequest.DepositAmount, 10)

	// InvestInsufficientBalanceMail [1.입금 요청]-[1B. 0 < 예치금 < 투자 신청 금액]: 예치금이 선 처리되고 잔여분에 대한 입금 요청
	err = mail.InvestInsufficientBalanceMail(ctx, serviceutils.GetServiceInfoMail(), user.Email, user.Name, lackAmount, project.Name,
		jsonutils.JSONTime.ToString(share.InvestDate),
		depositRequest.DepositBankName, depositRequest.DepositBranchName, depositRequest.DepositAccountNumber, userVirtualAccount.VaHolderNameKana,
		strconv.Itoa(int(share.Amount)), jsonutils.JSONTime.ToString(depositRequest.ExpireDepositDate))
	if err != nil {
		return err
	}

	return err
}
