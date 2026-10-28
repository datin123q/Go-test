package dto

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type MyPage struct {
	domain.User              `xorm:"extends"`
	domain.CorporateProfile  `xorm:"extends"`
	Birthday                 jsonutils.JSONTime          `json:"birthday" swagger:"desc(生年月日)"`
	Balance                  int64                       `json:"balance" swagger:"desc(預け金残高)" xorm:"-"`
	Withdrawal               int64                       `json:"withdrawal" swagger:"desc(出金申請金額)" xorm:"-"`
	AuditStatus              domain.AUDIT_STATUS         `json:"audit_status" swagger:"desc(審査状態)" xorm:"-"`
	UserTotalAsset           int64                       `json:"user_total_asset" swagger:"desc(顧客資産総額)" xorm:"-"`
	UserTotalUnpaid          int64                       `json:"user_total_unpaid" swagger:"desc(顧客未払額総額)" xorm:"-"`
	UserShareAmountInfo      domain.UserShareAmountInfo  `json:"user_share_amount_info" xorm:"-"`
	AntisocialCheck          *domain.UserAntisocialCheck `json:"antisocial_check" xorm:"-" swagger:"desc(反社会勢力チェック)"`
	UserAccount              *domain.UserAccount         `json:"user_account" xorm:"-" swagger:"desc(預け金口座)"`
	EquityVaAccount          *domain.UserVirtualAccount  `json:"equity_va_account" xorm:"-" swagger:"desc(事業型仮想口座)"`
	Identification1FrontFile *domain.UserFile            `json:"identification1_front_file" xorm:"-" swagger:"desc(本人確認書類正面1)"`
	Identification1BackFile  *domain.UserFile            `json:"identification1_back_file" xorm:"-" swagger:"desc(本人確認書類背面1)"`
	Identification2FrontFile *domain.UserFile            `json:"identification2_front_file" xorm:"-" swagger:"desc(本人確認書類正面2)"`
	Identification2BackFile  *domain.UserFile            `json:"identification2_back_file" xorm:"-" swagger:"desc(本人確認書類背面2)"`
}
