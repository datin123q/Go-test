package domain

import (
	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type FAQ_TYPE string

const (
	FAQ_GORUP_CODE          = "FAQ_TYPE"
	INVESTOR       FAQ_TYPE = "INVESTOR"  // 일반 투자자
	DEVELOPER      FAQ_TYPE = "DEVELOPER" // 특수 목적으로 설립된 회사
	OTHER          FAQ_TYPE = "OTHER"     // 기타
)

func (Faq) GetTypeCodeGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)

	codes = append(codes, commoncode.Code{Code: string(INVESTOR), Name: i18nutils.Translate("commoncode.FAQ_TYPE.INVESTOR", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(DEVELOPER), Name: i18nutils.Translate("commoncode.FAQ_TYPE.DEVELOPER", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(OTHER), Name: i18nutils.Translate("commoncode.FAQ_TYPE.OTHER", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: FAQ_GORUP_CODE,
		Codes:     codes,
	}
}

type FAQ_CATEGORY_TYPE string

const (
	FAQ_CATEGORY_GORUP_CODE                             = "FAQ_CATEGORY_TYPE"
	FAQ_CATEGORY_ACCOUNT_REGISTRATION FAQ_CATEGORY_TYPE = "FAQ_CATEGORY_ACCOUNT_REGISTRATION" // 회원등록
	FAQ_CATEGORY_PUT_TAKE_MONEY       FAQ_CATEGORY_TYPE = "FAQ_CATEGORY_PUT_TAKE_MONEY"       // 입금·출금
	FAQ_CATEGORY_INVESTING            FAQ_CATEGORY_TYPE = "FAQ_CATEGORY_INVESTING"            // 투자
	FAQ_CATEGORY_FUNDS                FAQ_CATEGORY_TYPE = "FAQ_CATEGORY_FUNDS"                // 펀드
	FAQ_CATEGORY_REPAY_DISTRIBUTION   FAQ_CATEGORY_TYPE = "FAQ_CATEGORY_REPAY_DISTRIBUTION"   // 상환·분배
)

func (Faq) GetCategoryTypeCodeGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)

	codes = append(codes, commoncode.Code{Code: string(FAQ_CATEGORY_ACCOUNT_REGISTRATION), Name: i18nutils.Translate("commoncode.FAQ_CATEGORY_TYPE.FAQ_CATEGORY_ACCOUNT_REGISTRATION", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FAQ_CATEGORY_PUT_TAKE_MONEY), Name: i18nutils.Translate("commoncode.FAQ_CATEGORY_TYPE.FAQ_CATEGORY_PUT_TAKE_MONEY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FAQ_CATEGORY_INVESTING), Name: i18nutils.Translate("commoncode.FAQ_CATEGORY_TYPE.FAQ_CATEGORY_INVESTING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FAQ_CATEGORY_FUNDS), Name: i18nutils.Translate("commoncode.FAQ_CATEGORY_TYPE.FAQ_CATEGORY_FUNDS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FAQ_CATEGORY_REPAY_DISTRIBUTION), Name: i18nutils.Translate("commoncode.FAQ_CATEGORY_TYPE.FAQ_CATEGORY_REPAY_DISTRIBUTION", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: FAQ_CATEGORY_GORUP_CODE,
		Codes:     codes,
	}
}

type Faq struct {
	Id            int64              `json:"id" description:"FAQ ID" example:"61"`
	CommonId      string             `json:"common_id" description:"공용 ID" xorm:"notnull UUID" swagger:"desc(공용 ID)" example:"aeee0ebc-cb8a-43bd-9eb6-94c34b8686ae"`
	Pin           bool               `json:"pin" description:"고정핀" xorm:"not null" swagger:"desc(고정핀)" example:"false"`
	Title         string             `json:"title" description:"제목" xorm:"varchar(255) not null" swagger:"desc(FAQ 제목)" example:"FAQ 테스트 입니다"`
	OwnerId       int64              `json:"owner_id" description:"작성자" xorm:"index" swagger:"desc(FAQ 작성자)" example:"110511"`
	UpdateId      int64              `json:"update_id" description:"최근 수정자" xorm:"index" swagger:"desc(FAQ 최근 수정자)" example:"110511"`
	ContentHtml   string             `json:"content_html" description:"컨텐츠 HTML" xorm:"LONGTEXT not null" swagger:"desc(FAQ 컨텐츠)" example:"<p>안녕하세요</p>\r\n<p><img title=\"Untitled.png\" src=\"https://cdn.dev.energyshares.jp/temp_file/2024-07-26/cqhkgsanulri1lp6i7eg.png\" alt=\"\" width=\"878\" height=\"228\" /></p>\r\n<p>FAQ 테스트입니다.</p>\r\n<p>감사합니다.</p>"`
	Posted        bool               `json:"posted" description:"게시여부" swagger:"desc(게시여부)" example:"true"`
	CategoryType  FAQ_CATEGORY_TYPE  `json:"category_type" description:"카테고리타입" enum:"FAQ_CATEGORY_ACCOUNT_REGISTRATION,FAQ_CATEGORY_PUT_TAKE_MONEY,FAQ_CATEGORY_INVESTING,FAQ_CATEGORY_FUNDS,FAQ_CATEGORY_REPAY_DISTRIBUTION" example:"FAQ_CATEGORY_ACCOUNT_REGISTRATION"`
	ClosingDate   jsonutils.JSONTime `json:"closing_date" description:"마감 날짜" xorm:"not null" example:"0001-01-01 09:18:59"`
	Created       jsonutils.JSONTime `json:"created" description:"생성 일자" xorm:"created" swagger:"desc(생성 일자)" example:"2024-07-26 15:56:32"`
	Updated       jsonutils.JSONTime `json:"updated" description:"수정 일자" xorm:"updated" swagger:"desc(수정 일자)" example:"2024-07-26 15:56:32"`
	Deleted       jsonutils.JSONTime `json:"deleted" description:"삭제 일자" xorm:"deleted" swagger:"desc(삭제 일자)" example:""`
	Attachments   []*Attachment      `json:"attachments" description:"첨부파일 목록" xorm:"-" swagger:"desc(첨부 파일 목록)"`
	AttachmentCnt int                `json:"attachment_cnt" description:"첨부파일 개수" example:"1"`
}

func (Faq) TableName() string {
	return "faq"
}

func (f Faq) SelectAll() string {
	return f.TableName() + ".*"
}

func (f Faq) WithAttachments(attachments []Attachment) Faq {
	for _, attachment := range attachments {
		f.Attachments = append(f.Attachments, &attachment)
	}

	return f
}
