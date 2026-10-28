package domain

import (
	"context"
	"crypto/sha256"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type INVESTMENT_WAIT_HISTORY_TYPE string

const (
	INVESTMENT_WAIT_HISTORY_TYPE_CODE                              = "INVESTMENT_WAIT_HISTORY_TYPE"
	INVESTMENT_WAIT_HISTORY_STATUS    INVESTMENT_WAIT_HISTORY_TYPE = "INVESTMENT_WAIT_HISTORY_STATUS" // 待機状態変更
	INVESTMENT_WAIT_HISTORY_MAIL      INVESTMENT_WAIT_HISTORY_TYPE = "INVESTMENT_WAIT_HISTORY_MAIL"   // 投資提案メール送信
)

func (InvestmentWaitHistory) GetTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(INVESTMENT_WAIT_HISTORY_STATUS), Name: i18nutils.Translate("commoncode.INVESTMENT_WAIT_HISTORY_TYPE.INVESTMENT_WAIT_HISTORY_STATUS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(INVESTMENT_WAIT_HISTORY_MAIL), Name: i18nutils.Translate("commoncode.INVESTMENT_WAIT_HISTORY_TYPE.INVESTMENT_WAIT_HISTORY_MAIL", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: INVESTMENT_WAIT_HISTORY_TYPE_CODE,
		Codes:     codes,
	}
}

type InvestmentWaitHistory struct {
	Id               int64                     `json:"id" xorm:"pk autoincr"`  // 고유아이디
	InvestmentWaitId int64                     `json:"investment_wait_id"`     // 상태변경된 투자대기자 아이디
	SubstanceType    string                    `json:"substance_type"`         // 내용 표기방식
	Name             string                    `json:"name"`                   // 처리자
	Email            string                    `json:"email"`                  // 처리자 이메일
	PrevStatus       INVESTMENT_WAITING_TYPE   `json:"prev_status"`            // 변경 전 상태
	CurrStatus       INVESTMENT_WAITING_TYPE   `json:"curr_status"`            // 변경 후 상태
	MailSendDate     jsonutils.JSONTime        `json:"mail_send_date"`         // 처리일시
	ProcessMethod    USER_HISTORY_PROCESS_TYPE `json:"process_method"`         // 처리자
	Created          jsonutils.JSONTime        `json:"created" xorm:"created"` // 처리일시
}

func (InvestmentWaitHistory) TableName() string {
	return "investment_wait_history"
}

func (i InvestmentWaitHistory) SelectAll() string {
	return i.TableName() + ".*"
}

func (InvestmentWaitHistory) CreateStatusHistory(ctx context.Context, before, after *InvestmentWaiters, processMethod USER_HISTORY_PROCESS_TYPE, uid int64) (int64, error) {
	var email, name string
	// 관리자와 사용자 분리 하면서 처리자 타입에 따라서 처리자 정보 가져오기
	switch processMethod {
	case USER_HISTORY_PROCESS_OPERATOR:
		admin, err := AdminUser{}.GetById(ctx, uid)
		if err != nil {
			return 0, err
		}
		email = admin.Email
		name = admin.Name
	case USER_HISTORY_PROCESS_MEMBER:
		user, err := User{}.GetById(ctx, uid)
		if err != nil {
			return 0, err
		}
		email = user.Email
		name = user.Name

	}

	// 내용을 hash 변경 후 비교, 내용 다른것만 처리로그 남김
	s1 := sha256.Sum256([]byte(before.Status))
	s2 := sha256.Sum256([]byte(after.Status))
	if s1 == s2 {
		return 0, nil
	}

	var InvestmentWaitHistory = &InvestmentWaitHistory{
		InvestmentWaitId: after.Id,
		SubstanceType:    string(INVESTMENT_WAIT_HISTORY_STATUS),
		Name:             name,
		Email:            email,
		PrevStatus:       before.Status,
		CurrStatus:       after.Status,
		ProcessMethod:    processMethod,
	}

	if _, err := factory.DB(ctx).Table("investment_wait_history").Insert(InvestmentWaitHistory); err != nil {
		logrus.Println("Error saving investment Wait status change history")
		return 0, err
	}

	return 0, nil
}

func (InvestmentWaitHistory) CreateMailSendHistory(ctx context.Context, sendDate jsonutils.JSONTime, investmentWaiters *InvestmentWaiters, uid int64) (int64, error) {

	user, err := AdminUser{}.GetById(ctx, uid)
	if err != nil {
		return 0, err
	}

	var InvestmentWaitHistory = &InvestmentWaitHistory{
		InvestmentWaitId: investmentWaiters.Id,
		SubstanceType:    string(INVESTMENT_WAIT_HISTORY_MAIL),
		Name:             user.Name,
		Email:            user.Email,
		MailSendDate:     sendDate,
		ProcessMethod:    USER_HISTORY_PROCESS_OPERATOR,
	}

	if _, err := factory.DB(ctx).Table("investment_wait_history").Insert(InvestmentWaitHistory); err != nil {
		logrus.Println("Error saving investment Wait Send Mail history")
		return 0, err
	}

	return 0, nil
}
