package domain

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type TRANSACTION_HISTORY_TYPE string

const (
	TRANSACTION_HISTORY_TYPE_CODE                            = "TRANSACTION_HISTORY_TYPE"
	TRANSACTION_HISTORY_TRANSACTION TRANSACTION_HISTORY_TYPE = "TRANSACTION_HISTORY_TRANSACTION" // 거래상태 변경
)

func (TransactionInfoHistory) GetTransactionHistoryTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(TRANSACTION_HISTORY_TRANSACTION), Name: i18nutils.Translate("commoncode.TRANSACTION_HISTORY_TYPE.TRANSACTION_HISTORY_TRANSACTION", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: TRANSACTION_HISTORY_TYPE_CODE,
		Codes:     codes,
	}
}

type TRANSACTION_CHANGED_CONTENT_TYPE string

const (
	TRANSACTION_CHANGED_CONTENT_TYPE_CODE                                           = "TRANSACTION_CHANGED_CONTENT_TYPE"
	TRANSACTION_CHANGED_CONTENT_TRANSACTION_STATUS TRANSACTION_CHANGED_CONTENT_TYPE = "TRANSACTION_CHANGED_CONTENT_TRANSACTION_STATUS" // 거래상태
)

func (TransactionInfoHistory) GetTransactionChangedContentTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(TRANSACTION_CHANGED_CONTENT_TRANSACTION_STATUS), Name: i18nutils.Translate("commoncode.TRANSACTION_CHANGED_CONTENT_TYPE.TRANSACTION_CHANGED_CONTENT_TRANSACTION_STATUS", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: TRANSACTION_CHANGED_CONTENT_TYPE_CODE,
		Codes:     codes,
	}
}

type TransactionInfoHistory struct {
	Id            int64                     `json:"id"`                     // 고유아이디
	ShareId       string                    `json:"share_id"`               // 거래 상품 아이디
	SubstanceType string                    `json:"substance_type"`         // 내용 표기방식
	ProcessMethod USER_HISTORY_PROCESS_TYPE `json:"process_method"`         // 처리 방식
	Name          string                    `json:"name"`                   // 처리자
	Email         string                    `json:"email"`                  // 처리자 이메일
	ContentType   string                    `json:"content_type"`           // 상태변경 타입
	PrevContent   string                    `json:"prev_content"`           // 변경 전 내용
	CurrContent   string                    `json:"curr_content"`           // 변경 후 내용
	Created       jsonutils.JSONTime        `json:"created" xorm:"created"` // 처리일시
}

func (t TransactionInfoHistory) TableName() string {
	return "transaction_info_history"
}

func (t TransactionInfoHistory) SelectAll() string {
	return t.TableName() + ".*"
}

func (TransactionInfoHistory) Create(ctx context.Context, before, after, sharesId string, processId int64) (int64, error) {

	transactionInfoHistory := toTransactionInfoHistory(ctx, sharesId, processId)
	if transactionInfoHistory == nil {
		return 0, nil
	}

	// 내용 상태 비교 후 히스토리를 남김
	transactionInfoHistory.compareContents(ctx, before, after)

	return 0, nil
}

func (t *TransactionInfoHistory) compareContents(ctx context.Context, prevData, currData string) {

	t.Id = 0

	if prevData != currData {
		t.PrevContent = prevData
		t.CurrContent = currData
		t.createHistory(ctx)
	}

}

func (t *TransactionInfoHistory) createHistory(ctx context.Context) {

	if t.Id > 0 {
		return
	}

	t.SubstanceType = string(TRANSACTION_HISTORY_TRANSACTION)
	t.ContentType = string(TRANSACTION_CHANGED_CONTENT_TRANSACTION_STATUS)
	if _, err := factory.DB(ctx).Table("transaction_info_history").Insert(t); err != nil {
		logrus.Println("Error saving transaction's information history")
	}
}

func toTransactionInfoHistory(ctx context.Context, sharesId string, processId int64) *TransactionInfoHistory {

	var transactionInfoHistory TransactionInfoHistory
	transactionInfoHistory.ShareId = sharesId
	transactionInfoHistory.Name = ""
	transactionInfoHistory.Email = "-"

	if processId == 0 {
		var err error
		processId, err = AdminUser{}.GetSuperAdminId(ctx)
		if err != nil {
			return nil
		}
	}

	user, err := AdminUser{}.GetById(ctx, processId)
	if err != nil {
		factory.Logger(ctx).Println(err)
		return nil
	} else if user == nil {
		return nil
	}

	if modelhelper.Find(user.Roles, "ROLE_USER") {
		transactionInfoHistory.ProcessMethod = USER_HISTORY_PROCESS_MEMBER
	} else {
		transactionInfoHistory.ProcessMethod = USER_HISTORY_PROCESS_OPERATOR
		transactionInfoHistory.Name = user.Name
		transactionInfoHistory.Email = user.Email
	}

	return &transactionInfoHistory
}
