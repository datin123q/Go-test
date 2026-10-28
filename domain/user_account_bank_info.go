package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/pkg/errors"
)

// 사용자 어카운트 정보에는 사용자의 계좌번호, 금융기관명, 지점명, 계좌종류.
type UserAccountBankInfo struct {
	UserId      int64        `json:"id" swagger:"desc(사용자 ID)"`
	BankAccount string       `json:"bank_account" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_BANK_ACCOUNT" swagger:"desc(은행계좌번호)"`
	BankName    string       `json:"bank_name" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_FINANCE_NAME" swagger:"desc(금융기관명)"`
	BranchName  string       `json:"branch_name" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_BANK_NAME" swagger:"desc(은행지점명)"`
	BranchCode  *string      `json:"branch_code" swagger:"desc(은행지점코드)"`
	DepositType DEPOSIT_TYPE `json:"deposit_type" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_DEPOSIT_TYPE" swagger:"desc(예금종류공통코드: [DEPOSIT_TYPE])"`
}

func (a UserAccountBankInfo) GetByUserId(ctx context.Context, id int64) (*UserAccountBankInfo, error) {
	var useraccountbankinfo UserAccountBankInfo

	q := factory.DB(ctx).Table("user").
		Select("user.*")
	if has, err := q.Where("user_id = ?", id).Get(&useraccountbankinfo); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("Not Found User Info ")
	}

	return &useraccountbankinfo, nil
}
