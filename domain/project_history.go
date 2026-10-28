package domain

import (
	"context"
	"crypto/sha256"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type PROJECT_PUBLISH_TYPE string

const (
	PROJECT_PUBLISH_TYPE_CODE                      = "PROJECT_PUBLISH_TYPE"
	PROJECT_PUBLISH_POSTED    PROJECT_PUBLISH_TYPE = "PROJECT_PUBLISH_POSTED"   // 개시
	PROJECT_PUBLISH_UNPOSTED  PROJECT_PUBLISH_TYPE = "PROJECT_PUBLISH_UNPOSTED" // 미개시
)

func (ProjectInfoHistory) GetProjectPublishTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(PROJECT_PUBLISH_POSTED), Name: i18nutils.Translate("commoncode.PROJECT_PUBLISH_TYPE.PROJECT_PUBLISH_POSTED", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_PUBLISH_UNPOSTED), Name: i18nutils.Translate("commoncode.PROJECT_PUBLISH_TYPE.PROJECT_PUBLISH_UNPOSTED", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: PROJECT_PUBLISH_TYPE_CODE,
		Codes:     codes,
	}
}

type PROJECT_INVEST_TYPE string

const (
	PROJECT_INVEST_TYPE_CODE                     = "PROJECT_INVEST_TYPE"
	PROJECT_INVEST_BLOCK     PROJECT_INVEST_TYPE = "PROJECT_INVEST_BLOCK"    // 투자제한
	PROJECT_INVEST_NOTBLOCK  PROJECT_INVEST_TYPE = "PROJECT_INVEST_NOTBLOCK" // 투자제한 안함
)

func (ProjectInfoHistory) GetProjectInvestTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(PROJECT_INVEST_BLOCK), Name: i18nutils.Translate("commoncode.PROJECT_INVEST_TYPE.PROJECT_INVEST_BLOCK", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_INVEST_NOTBLOCK), Name: i18nutils.Translate("commoncode.PROJECT_INVEST_TYPE.PROJECT_INVEST_NOTBLOCK", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: PROJECT_INVEST_TYPE_CODE,
		Codes:     codes,
	}
}

type PROJECT_NEWS_TYPE string

const (
	PROJECT_NEWS_TYPE_CODE                   = "PROJECT_NEWS_TYPE"
	PROJECT_NEWS_ADD       PROJECT_NEWS_TYPE = "PROJECT_NEWS_ADD"    // 추가
	PROJECT_NEWS_DEL       PROJECT_NEWS_TYPE = "PROJECT_NEWS_DEL"    // 삭제
	PROJECT_NEWS_CHANGE    PROJECT_NEWS_TYPE = "PROJECT_NEWS_CHANGE" // 변경
)

func (ProjectInfoHistory) GetProjectNewsTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(PROJECT_NEWS_ADD), Name: i18nutils.Translate("commoncode.PROJECT_NEWS_TYPE.PROJECT_NEWS_ADD", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_NEWS_DEL), Name: i18nutils.Translate("commoncode.PROJECT_NEWS_TYPE.PROJECT_NEWS_DEL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_NEWS_CHANGE), Name: i18nutils.Translate("commoncode.PROJECT_NEWS_TYPE.PROJECT_NEWS_CHANGE", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: PROJECT_NEWS_TYPE_CODE,
		Codes:     codes,
	}
}

type PROJECT_SETTLEMENT_TYPE string

const (
	PROJECT_SETTLEMENT_TYPE_CODE                         = "PROJECT_SETTLEMENT_TYPE"
	PROJECT_SETTLEMENT_ADD       PROJECT_SETTLEMENT_TYPE = "PROJECT_SETTLEMENT_ADD" // 추가
	PROJECT_SETTLEMENT_DEL       PROJECT_SETTLEMENT_TYPE = "PROJECT_SETTLEMENT_DEL" // 삭제
	PROJECT_SETTLEMENT_COMPLETED PROJECT_SETTLEMENT_TYPE = "PROJECT_SETTLEMENT_COMPLETED"
	PROJECT_SETTLEMENT_SCHEDULED PROJECT_SETTLEMENT_TYPE = "PROJECT_SETTLEMENT_SCHEDULED"
	PROJECT_SETTLEMENT_CONFIRM   PROJECT_SETTLEMENT_TYPE = "PROJECT_SETTLEMENT_CONFIRM"
)

func (ProjectInfoHistory) GetProjectSettlementTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(PROJECT_SETTLEMENT_ADD), Name: i18nutils.Translate("commoncode.PROJECT_SETTLEMENT_TYPE.PROJECT_SETTLEMENT_ADD", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_SETTLEMENT_DEL), Name: i18nutils.Translate("commoncode.PROJECT_SETTLEMENT_TYPE.PROJECT_SETTLEMENT_DEL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_SETTLEMENT_COMPLETED), Name: i18nutils.Translate("commoncode.PROJECT_SETTLEMENT_TYPE.PROJECT_SETTLEMENT_COMPLETED", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_SETTLEMENT_SCHEDULED), Name: i18nutils.Translate("commoncode.PROJECT_SETTLEMENT_TYPE.PROJECT_SETTLEMENT_SCHEDULED", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_SETTLEMENT_CONFIRM), Name: i18nutils.Translate("commoncode.PROJECT_SETTLEMENT_TYPE.PROJECT_SETTLEMENT_CONFIRM", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: PROJECT_SETTLEMENT_TYPE_CODE,
		Codes:     codes,
	}
}

type PROJECT_MILESTONE_STATUS_TYPE string

const (
	PROJECT_MILESTONE_TYPE_CODE                               = "PROJECT_MILESTONE_TYPE"
	PROJECT_MILESTONE_ADD       PROJECT_MILESTONE_STATUS_TYPE = "PROJECT_MILESTONE_ADD"
	PROJECT_MILESTONE_DEL       PROJECT_MILESTONE_STATUS_TYPE = "PROJECT_MILESTONE_DEL"
	PROJECT_MILESTONE_READY     PROJECT_MILESTONE_STATUS_TYPE = "PROJECT_MILESTONE_READY"
	PROJECT_MILESTONE_GOING     PROJECT_MILESTONE_STATUS_TYPE = "PROJECT_MILESTONE_GOING"
	PROJECT_MILESTONE_COMPLETED PROJECT_MILESTONE_STATUS_TYPE = "PROJECT_MILESTONE_COMPLETED"
)

func (ProjectInfoHistory) GetProjectMilestoneTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(PROJECT_MILESTONE_ADD), Name: i18nutils.Translate("commoncode.PROJECT_MILESTONE_STATUS_TYPE.PROJECT_MILESTONE_ADD", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_MILESTONE_DEL), Name: i18nutils.Translate("commoncode.PROJECT_MILESTONE_STATUS_TYPE.PROJECT_MILESTONE_DEL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_MILESTONE_READY), Name: i18nutils.Translate("commoncode.PROJECT_MILESTONE_STATUS_TYPE.PROJECT_MILESTONE_READY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_MILESTONE_GOING), Name: i18nutils.Translate("commoncode.PROJECT_MILESTONE_STATUS_TYPE.PROJECT_MILESTONE_GOING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_MILESTONE_COMPLETED), Name: i18nutils.Translate("commoncode.PROJECT_MILESTONE_STATUS_TYPE.PROJECT_MILESTONE_COMPLETED", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: PROJECT_MILESTONE_TYPE_CODE,
		Codes:     codes,
	}
}

type PROJECT_HISTORY_TYPE string

const (
	PROJECT_HISTORY_TYPE_CODE                           = "PROJECT_HISTORY_TYPE"
	PROJECT_HISTORY_CONTENTS       PROJECT_HISTORY_TYPE = "PROJECT_HISTORY_CONTENTS"       // 내용 변경
	PROJECT_HISTORY_PUBLISHED      PROJECT_HISTORY_TYPE = "PROJECT_HISTORY_PUBLISHED"      // 게시상태 변경
	PROJECT_HISTORY_LIMITED_INVEST PROJECT_HISTORY_TYPE = "PROJECT_HISTORY_LIMITED_INVEST" // 투자제한 변경
)

func (ProjectInfoHistory) GetProjectHistoryTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(PROJECT_HISTORY_CONTENTS), Name: i18nutils.Translate("commoncode.PROJECT_HISTORY_TYPE.PROJECT_HISTORY_CONTENTS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_HISTORY_PUBLISHED), Name: i18nutils.Translate("commoncode.PROJECT_HISTORY_TYPE.PROJECT_HISTORY_PUBLISHED", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_HISTORY_LIMITED_INVEST), Name: i18nutils.Translate("commoncode.PROJECT_HISTORY_TYPE.PROJECT_HISTORY_LIMITED_INVEST", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: PROJECT_HISTORY_TYPE_CODE,
		Codes:     codes,
	}
}

type PROJECT_CHANGED_CONTENT_TYPE string

const (
	PROJECT_CHANGED_CONTENT_TYPE_CODE                                               = "PROJECT_CHANGED_CONTENT_TYPE"
	PROJECT_CHANGED_CONTENT_POSTED_STATE               PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_POSTED_STATE"               // 게시상태
	PROJECT_CHANGED_CONTENT_LIMITED_INVEST             PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_LIMITED_INVEST"             // 투자제한
	PROJECT_CHANGED_CONTENT_FUNDRAISING_STATUS         PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_PROJECT_STATUS"             // 모집상태, PROJECT_STATUS
	PROJECT_CHANGED_CONTENT_INVESTMENT_METHOD          PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_INVESTMENT_METHOD"          // 투자유형, PROJECT_TYPE
	PROJECT_CHANGED_CONTENT_SELECTION_METHOD           PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_SELECTION_METHOD"           // 当選方式, PROJECT_SELECTION_TYPE
	PROJECT_CHANGED_CONTENT_NAME                       PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_NAME"                       // 상품명
	PROJECT_CHANGED_CONTENT_DESCRIPTION                PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_DESCRIPTION"                // 상품설명, 추가 필요
	PROJECT_CHANGED_CONTENT_MAINIMAGE_URL              PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_MAINIMAGE_URL"              // 썸네일 URL
	PROJECT_CHANGED_CONTENT_SUBIMAGE_URL               PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_SUBIMAGE_URL"               // サブ画像 URL
	PROJECT_CHANGED_CONTENT_BANNERIMAGE_URL            PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_BANNERIMAGE_URL"            // 배너 URL
	PROJECT_CHANGED_CONTENT_SPV                        PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_SPV"                        // 개발회사
	PROJECT_CHANGED_CONTENT_POWER_PLANT                PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_POWER_PLANT"                // 발전 유형, PROJECT_CATEGORY
	PROJECT_CHANGED_CONTENT_SCALE                      PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_SCALE"                      // 규모(KW/MW), 추가 필요
	PROJECT_CHANGED_CONTENT_EXPECTED_RETURN            PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_EXPECTED_RETURN"            // 예상 수익률
	PROJECT_CHANGED_CONTENT_FUNDRAISING_START_DATE     PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_FUNDRAISING_START_DATE"     // 상품 개시일
	PROJECT_CHANGED_CONTENT_FUNDRAISING_END_DATE       PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_FUNDRAISING_END_DATE"       // 상품 종료일
	PROJECT_CHANGED_CONTENT_DUE_DATE                   PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_DUE_DATE"                   // 예상자금 회수일
	PROJECT_CHANGED_CONTENT_TOTAL_SHARES               PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_TOTAL_SHARES"               // 총 구수
	PROJECT_CHANGED_CONTENT_SHARES                     PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_SHARES"                     // 발행 구수, 추가 필요
	PROJECT_CHANGED_CONTENT_PRICE_PER_SHARE            PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_PRICE_PER_SHARE"            // 1구당 가격
	PROJECT_CHANGED_CONTENT_MINIMUM_FUNDRAISING_AMOUNT PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_MINIMUM_FUNDRAISING_AMOUNT" // 목표 모금액
	PROJECT_CHANGED_CONTENT_MINIMUM_INVESTMENT         PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_MINIMUM_INVESTMENT"         // 최소 투자수
	PROJECT_CHANGED_CONTENT_DIVIDEND_RATE              PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_DIVIDEND_RATE"              // 이자율
	PROJECT_CHANGED_CONTENT_OPERATION_TERM             PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_OPERATION_TERM"             // 운용 기간
	PROJECT_CHANGED_CONTENT_TARGET_FUNDRAISING_AMOUNT  PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_TARGET_FUNDRAISING_AMOUNT"  // 총 대출액, 추가 필요
	PROJECT_CHANGED_CONTENT_SCALE_FUND                 PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_SCALE_FUND"                 // 총 규모(MWP DC)
	PROJECT_CHANGED_CONTENT_IRR                        PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_IRR"                        // 전 분기 내부 수익률
	PROJECT_CHANGED_CONTENT_SUMMARY                    PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_SUMMARY"                    // 프로젝트 개요
	PROJECT_CHANGED_CONTENT_REFERENCE_FILE             PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_REFERENCE_FILE"             // 프로젝트 참고자료(다운로드)
	PROJECT_CHANGED_CONTENT_CAUTION                    PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_CAUTION"                    // 유의사항
	PROJECT_CHANGED_CONTENT_BANK_NAME                  PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_BANK_NAME"                  // 금융 기관명
	PROJECT_CHANGED_CONTENT_BRANCH_NAME                PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_BRANCH_NAME"                // 지점명
	PROJECT_CHANGED_CONTENT_BANK_ACCOUNT               PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_BANK_ACCOUNT"               // 계좌번호
	PROJECT_CHANGED_CONTENT_DEPOSITOR                  PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_DEPOSITOR"                  // 계좌주
	PROJECT_CHANGED_CONTENT_CONTRACT_BEFORE_PAGE       PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_CONTRACT_BEFORE_PAGE"       // 계약성립 전 서면
	PROJECT_CHANGED_CONTENT_CONTRACT_AFTER_PAGE        PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_CONTRACT_AFTER_PAGE"        // 계약성립 시 서면
	PROJECT_CHANGED_CONTENT_SILENT_PARTNERSHIPS        PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_SILENT_PARTNERSHIPS"        // 익명조합 계약서
	PROJECT_CHANGED_CONTENT_PROJECT_REPORT             PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_PROJECT_REPORT"             // 펀드 보고서
	PROJECT_CHANGED_CONTENT_NEWS                       PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_NEWS"                       // 뉴스
	PROJECT_CHANGED_CONTENT_QUESTIONS                  PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_QUESTIONS"                  // 문의사항
	PROJECT_CHANGED_CONTENT_REPAYMENT_SCHEDULE         PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_REPAYMENT_SCHEDULE"         // 상환 일정
	PROJECT_CHANGED_CONTENT_SETTLEMENT                 PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_SETTLEMENT"                 // 결산 실행
	PROJECT_CHANGED_CONTENT_LOAN_COMPLETED             PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_LOAN_COMPLETED"             // 대출 확정
	PROJECT_CHANGED_CONTENT_IS_INVESTOR_LIMIT          PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_IS_INVESTOR_LIMIT"          // 투자자수 제한 여부
	PROJECT_CHANGED_CONTENT_IS_AMOUNT_LIMIT            PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_IS_AMOUNT_LIMIT"            // 투자금액 제한 여부
	PROJECT_CHANGED_CONTENT_MILESTONE                  PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_MILESTONE"
	PROJECT_CHANGED_CONTENT_MILESTONE_NO               PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_MILESTONE_NO"
	PROJECT_CHANGED_CONTENT_MILESTONE_STATUS           PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_MILESTONE_STATUS"
	PROJECT_CHANGED_CONTENT_MILESTONE_ASSIGNMENT       PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_MILESTONE_ASSIGNMENT"
	PROJECT_CHANGED_CONTENT_MILESTONE_INFO             PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_MILESTONE_INFO"
	PROJECT_CHANGED_CONTENT_MILESTONE_NOTE             PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_MILESTONE_NOTE"
	PROJECT_CHANGED_CONTENT_IS_VISIBLE_INVESTMENT_INFO PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_IS_VISIBLE_INVESTMENT_INFO" // 투자 정보 공개 여부
	PROJECT_CHANGED_CONTENT_PROJECT_PROGRESS_PERCENT   PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_PROJECT_PROGRESS_PERCENT"   // 프로젝트 진행률
	PROJECT_CHANGED_CONTENT_FUND_NUMBER                PROJECT_CHANGED_CONTENT_TYPE = "PROJECT_CHANGED_CONTENT_FUND_NUMBER"                // ファンド番号
)

func (ProjectInfoHistory) GetProjectChangedContentTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_POSTED_STATE), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_POSTED_STATE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_LIMITED_INVEST), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_LIMITED_INVEST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_FUNDRAISING_STATUS), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_PROJECT_STATUS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_INVESTMENT_METHOD), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_INVESTMENT_METHOD", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_SELECTION_METHOD), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SELECTION_METHOD", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_NAME), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_NAME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_DESCRIPTION), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_DESCRIPTION", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_MAINIMAGE_URL), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_MAINIMAGE_URL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_SUBIMAGE_URL), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SUBIMAGE_URL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_BANNERIMAGE_URL), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_BANNERIMAGE_URL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_SPV), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SPV", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_POWER_PLANT), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_POWER_PLANT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_SCALE), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SCALE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_EXPECTED_RETURN), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_EXPECTED_RETURN", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_FUNDRAISING_START_DATE), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_FUNDRAISING_START_DATE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_FUNDRAISING_END_DATE), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_FUNDRAISING_END_DATE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_DUE_DATE), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_DUE_DATE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_TOTAL_SHARES), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_TOTAL_SHARES", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_SHARES), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SHARES", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_PRICE_PER_SHARE), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_PRICE_PER_SHARE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_MINIMUM_FUNDRAISING_AMOUNT), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_MINIMUM_FUNDRAISING_AMOUNT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_MINIMUM_INVESTMENT), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_MINIMUM_INVESTMENT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_DIVIDEND_RATE), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_DIVIDEND_RATE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_OPERATION_TERM), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_OPERATION_TERM", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_TARGET_FUNDRAISING_AMOUNT), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_TARGET_FUNDRAISING_AMOUNT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_SCALE_FUND), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SCALE_FUND", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_IRR), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_IRR", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_SUMMARY), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SUMMARY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_REFERENCE_FILE), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_REFERENCE_FILE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_NEWS), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_NEWS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_QUESTIONS), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_QUESTIONS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_CAUTION), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_CAUTION", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_BANK_NAME), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_BANK_NAME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_BRANCH_NAME), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_BRANCH_NAME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_BANK_ACCOUNT), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_BANK_ACCOUNT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_DEPOSITOR), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_DEPOSITOR", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_CONTRACT_BEFORE_PAGE), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_CONTRACT_BEFORE_PAGE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_CONTRACT_AFTER_PAGE), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_CONTRACT_AFTER_PAGE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_SILENT_PARTNERSHIPS), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SILENT_PARTNERSHIPS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_PROJECT_REPORT), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_PROJECT_REPORT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_REPAYMENT_SCHEDULE), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_REPAYMENT_SCHEDULE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_SETTLEMENT), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SETTLEMENT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_LOAN_COMPLETED), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_LOAN_COMPLETED", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_IS_INVESTOR_LIMIT), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_IS_INVESTOR_LIMIT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_IS_AMOUNT_LIMIT), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_IS_AMOUNT_LIMIT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_MILESTONE), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_MILESTONE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_MILESTONE_NO), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_MILESTONE_NO", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_MILESTONE_STATUS), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_MILESTONE_STATUS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_MILESTONE_ASSIGNMENT), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_MILESTONE_ASSIGNMENT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_MILESTONE_INFO), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_MILESTONE_INFO", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_MILESTONE_NOTE), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_MILESTONE_NOTE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_IS_VISIBLE_INVESTMENT_INFO), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_IS_VISIBLE_INVESTMENT_INFO", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_PROJECT_PROGRESS_PERCENT), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_PROJECT_PROGRESS_PERCENT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_CHANGED_CONTENT_FUND_NUMBER), Name: i18nutils.Translate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_FUND_NUMBER", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: PROJECT_CHANGED_CONTENT_TYPE_CODE,
		Codes:     codes,
	}
}

// 변경된 프로젝트 정보의 카테고리를 저장하기 위한 구조체
type ProjectInfoHistory struct {
	Id            int64              `json:"id"`                     // 고유아이디
	ProjectId     int64              `json:"project_id"`             // 상태변경된 투자상품 아이디
	SubstanceType string             `json:"substance_type"`         // 내용 표기방식
	Name          string             `json:"name"`                   // 처리자
	Email         string             `json:"email"`                  // 처리자 이메일
	Created       jsonutils.JSONTime `json:"created" xorm:"created"` // 처리일시
}

// 변경된 프로젝트 정보의 내용을 저장하기 위한 구조체
type ProjectChangedContent struct {
	Id              int64  `json:"id"`
	InfoId          int64  `json:"info_id"`                           // 투자상품 상태변경 히스토리 아이디
	ContentType     string `json:"content_type"`                      // 내용변경 타입
	ContentSuffix   string `json:"content_suffix"`                    // 내용 접미사
	PrevContent     string `json:"prev_content"`                      // 변경 전 내용
	CurrContent     string `json:"curr_content"`                      // 변경 후 내용
	PrevContentHtml string `json:"prev_long_content" xorm:"LONGTEXT"` // 변경 전 HTML
	CurrContentHtml string `json:"curr_content_html" xorm:"LONGTEXT"` // 변경 후 HTML
	PrevUrl         string `json:"prev_url"`                          // 변경 전 URL
	CurrUrl         string `json:"curr_url"`                          // 변경 후 URL
}

type ProjectInfoHistoryExtends struct {
	ProjectInfoHistory `xorm:"extends"`
	Content            ProjectChangedContent `json:"content" xorm:"extends"`
}

func (ProjectInfoHistory) GetAll(ctx context.Context, id int64, sortby, order string, offset, limit int) (int64, []ProjectInfoHistoryExtends, error) {

	query := factory.DB(ctx).Table("project_info_history").Select("project_info_history.id, project_info_history.project_id, project_info_history.substance_type, project_info_history.name, project_info_history.email, project_info_history.created, project_changed_content.*").
		Join("INNER", "project_changed_content", "project_changed_content.info_id = project_info_history.id")
	if id > 0 {
		query.Where("project_info_history.project_id = ?", id)
	}

	if err := modelhelper.SetSortOrder(query, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var projectInfoHistoryList []ProjectInfoHistoryExtends
	var totalCount int64
	totalCount, err := query.Limit(limit, offset).FindAndCount(&projectInfoHistoryList)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, projectInfoHistoryList, nil
}

func (ProjectInfoHistory) Create(ctx context.Context, before, after interface{}, processId int64) (int64, error) {

	// i18n tag로 지정된 필드값을 저장하기 위한 맵
	// key = tag 명, value = 필드의 값
	var prevMapData map[string]interface{}
	var currMapData map[string]interface{}

	if processId == 0 {
		var err error
		processId, err = AdminUser{}.GetSuperAdminId(ctx)
		if err != nil {
			return 0, err
		}
	}

	// 처리자 정보를 가져옴
	user, _ := AdminUser{}.GetById(ctx, processId)
	var projectInfoHistory = &ProjectInfoHistory{
		ProjectId: after.(*Project).Id,
		Name:      user.Name,
		Email:     user.Email,
	}

	// i18n tag로 지정된 필드들을 저장한다.
	prevMapData = ProjectInfoHistory{}.ConvertTagToMap(reflect.ValueOf(before.(*Project)).Elem())
	currMapData = ProjectInfoHistory{}.ConvertTagToMap(reflect.ValueOf(after.(*Project)).Elem())

	// 게시상태 비교 후 히스토리를 남김
	projectInfoHistory.ComparePostedState(ctx, prevMapData, currMapData)

	// 투자제한 상태 비교 후 히스토리를 남김
	projectInfoHistory.CompareBlockInvest(ctx, prevMapData, currMapData)

	// 내용 상태 비교 후 히스토리를 남김
	projectInfoHistory.CompareContents(ctx, before.(*Project), after.(*Project), prevMapData, currMapData)

	// 프로젝트 파일 비교
	projectInfoHistory.CompareAttachment(ctx, before.(*Project), after.(*Project))

	// 서면 파일 비교
	projectInfoHistory.CompareDocument(ctx, before.(*Project), after.(*Project))

	// 뉴스 비교
	projectInfoHistory.CompareNews(ctx, before.(*Project), after.(*Project))

	// 대출 상태 비교
	projectInfoHistory.CompareLoan(ctx, before.(*Project), after.(*Project))

	// 마일스톤 비교
	projectInfoHistory.CompareMilestone(ctx, before.(*Project), after.(*Project))

	return 0, nil
}

// i18n tag로 지정된 필드값을 저장
func (ProjectInfoHistory) ConvertTagToMap(val reflect.Value) map[string]interface{} {

	var mapData = make(map[string]interface{}, 0)

	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		i18nStr := tag.Get("i18n")
		if len(i18nStr) <= 0 {
			continue
		}

		// .으로 구분하여 마지막 string을 key로 설정
		strs := strings.Split(i18nStr, ".")
		str := strs[len(strs)-1]
		if len(str) <= 0 {
			continue
		}

		switch t := val.Field(i).Interface().(type) {
		case string:
			mapData[str] = t
		case bool:
			// true, false로 value에 저장하는 대신 번역 코드값으로 저장
			if t == true {
				mapData[str] = string(USER_HISTORY_CHECK_YES)
			} else {
				mapData[str] = string(USER_HISTORY_CHECK_NO)
			}
		case int64:
			mapData[str] = strconv.FormatInt(t, 10)
		case *int64:
			// 未設定は空文字として扱い、設定/解除も差分として残す。
			if t == nil {
				mapData[str] = ""
			} else {
				mapData[str] = strconv.FormatInt(*t, 10)
			}
		case int:
			mapData[str] = strconv.Itoa(t)
		case float64:
			mapData[str] = strconv.FormatFloat(t, 'f', 1, 64)
		case float32:
			mapData[str] = fmt.Sprintf("%f", t)
		default:
			// 사용자에 의해 지정된 데이터형을 string으로 변환하여 value에 저장
			if value, ok := val.Field(i).Interface().(PROJECT_CATEGORY); ok {
				mapData[str] = string(value)
			} else if value, ok := val.Field(i).Interface().(PROJECT_TYPE); ok {
				mapData[str] = string(value)
			} else if value, ok := val.Field(i).Interface().(PROJECT_STATUS); ok {
				mapData[str] = string(value)
			} else if value, ok := val.Field(i).Interface().(jsonutils.JSONTime); ok {
				mapData[str] = value.Time().String()
			}
		}
	}

	return mapData
}

// 게시상태 비교
func (pih *ProjectInfoHistory) ComparePostedState(ctx context.Context, prevData, currData map[string]interface{}) {

	var prev, curr string
	if val, ok := prevData[string(PROJECT_CHANGED_CONTENT_POSTED_STATE)]; ok {
		prev = val.(string)
	}
	if val, ok := currData[string(PROJECT_CHANGED_CONTENT_POSTED_STATE)]; ok {
		curr = val.(string)
	}

	if prev == curr {
		return
	}

	// 변경된 정보 및 정보의 내용 저장
	pih.Id = 0
	pih.createHistory(ctx, string(PROJECT_HISTORY_PUBLISHED))
	if prev == "true" && curr == "false" {
		ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_POSTED_STATE), pih.Id, string(PROJECT_PUBLISH_POSTED), string(PROJECT_PUBLISH_UNPOSTED), "", "", "", "", "")
	} else if prev == "false" && curr == "true" {
		ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_POSTED_STATE), pih.Id, string(PROJECT_PUBLISH_UNPOSTED), string(PROJECT_PUBLISH_POSTED), "", "", "", "", "")
	}
}

// 투자제한 상태 비교
func (pih *ProjectInfoHistory) CompareBlockInvest(ctx context.Context, prevData, currData map[string]interface{}) {

	var prev, curr string
	if val, ok := prevData[string(PROJECT_CHANGED_CONTENT_LIMITED_INVEST)]; ok {
		prev = val.(string)
	}
	if val, ok := currData[string(PROJECT_CHANGED_CONTENT_LIMITED_INVEST)]; ok {
		curr = val.(string)
	}

	if prev == curr {
		return
	}

	// 변경된 정보 및 정보의 내용 저장
	pih.Id = 0
	pih.createHistory(ctx, string(PROJECT_HISTORY_LIMITED_INVEST))
	if prev == "true" && curr == "false" {
		ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_LIMITED_INVEST), pih.Id, string(PROJECT_INVEST_BLOCK), string(PROJECT_INVEST_NOTBLOCK), "", "", "", "", "")
	} else if prev == "false" && curr == "true" {
		ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_LIMITED_INVEST), pih.Id, string(PROJECT_INVEST_NOTBLOCK), string(PROJECT_INVEST_BLOCK), "", "", "", "", "")
	}
}

// ProjectChangedContentTypesForCompare は変更履歴として差分を残す項目。
// ここに載っていない項目は、値を拾えていても履歴に出ない。
var ProjectChangedContentTypesForCompare = []string{string(PROJECT_CHANGED_CONTENT_FUNDRAISING_STATUS), string(PROJECT_CHANGED_CONTENT_INVESTMENT_METHOD), string(PROJECT_CHANGED_CONTENT_NAME), string(PROJECT_CHANGED_CONTENT_SELECTION_METHOD),
	string(PROJECT_CHANGED_CONTENT_DESCRIPTION), string(PROJECT_CHANGED_CONTENT_MAINIMAGE_URL), string(PROJECT_CHANGED_CONTENT_SUBIMAGE_URL), string(PROJECT_CHANGED_CONTENT_BANNERIMAGE_URL),
	string(PROJECT_CHANGED_CONTENT_SPV), string(PROJECT_CHANGED_CONTENT_POWER_PLANT), string(PROJECT_CHANGED_CONTENT_SCALE), string(PROJECT_CHANGED_CONTENT_EXPECTED_RETURN),
	string(PROJECT_CHANGED_CONTENT_FUNDRAISING_START_DATE), string(PROJECT_CHANGED_CONTENT_FUNDRAISING_END_DATE), string(PROJECT_CHANGED_CONTENT_DUE_DATE),
	string(PROJECT_CHANGED_CONTENT_TOTAL_SHARES), string(PROJECT_CHANGED_CONTENT_SHARES), string(PROJECT_CHANGED_CONTENT_PRICE_PER_SHARE),
	string(PROJECT_CHANGED_CONTENT_MINIMUM_FUNDRAISING_AMOUNT),
	string(PROJECT_CHANGED_CONTENT_MINIMUM_INVESTMENT), string(PROJECT_CHANGED_CONTENT_DIVIDEND_RATE),
	string(PROJECT_CHANGED_CONTENT_OPERATION_TERM), string(PROJECT_CHANGED_CONTENT_TARGET_FUNDRAISING_AMOUNT), string(PROJECT_CHANGED_CONTENT_SUMMARY),
	string(PROJECT_CHANGED_CONTENT_IS_INVESTOR_LIMIT), string(PROJECT_CHANGED_CONTENT_IS_AMOUNT_LIMIT),
	string(PROJECT_CHANGED_CONTENT_REFERENCE_FILE), string(PROJECT_CHANGED_CONTENT_QUESTIONS),
	string(PROJECT_CHANGED_CONTENT_SCALE_FUND), string(PROJECT_CHANGED_CONTENT_IRR), string(PROJECT_CHANGED_CONTENT_BANK_NAME),
	string(PROJECT_CHANGED_CONTENT_BRANCH_NAME), string(PROJECT_CHANGED_CONTENT_BANK_ACCOUNT),
	string(PROJECT_CHANGED_CONTENT_DEPOSITOR), string(PROJECT_CHANGED_CONTENT_CAUTION),
	string(PROJECT_CHANGED_CONTENT_IS_VISIBLE_INVESTMENT_INFO), string(PROJECT_CHANGED_CONTENT_PROJECT_PROGRESS_PERCENT),
	string(PROJECT_CHANGED_CONTENT_FUND_NUMBER)}

// 내용 상태 비교
func (pih *ProjectInfoHistory) CompareContents(ctx context.Context, prevProject, currProject *Project, prevData, currData map[string]interface{}) {

	// 変更を比較する項目は ProjectChangedContentTypesForCompare を参照する。
	contentTypes := ProjectChangedContentTypesForCompare

	pih.Id = 0
	for _, content := range contentTypes {

		if prevData[content] == nil {
			continue
		}

		if prevData[content].(string) != currData[content].(string) {

			// 변경 된 정보명 저장
			pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))

			before := prevData[content].(string)
			after := currData[content].(string)
			beforeHtml := ""
			aterHtml := ""
			beforeUrl := ""
			afterUrl := ""

			switch content {
			case string(PROJECT_CHANGED_CONTENT_MAINIMAGE_URL):
				before = prevProject.MainImage.OriginalName
				after = currProject.MainImage.OriginalName
				beforeUrl = prevProject.MainImageUrl
				afterUrl = currProject.MainImageUrl
			case string(PROJECT_CHANGED_CONTENT_SUBIMAGE_URL):
				for i := 0; i < len(prevProject.SubImageInfoList); i++ {
					before = prevProject.SubImageInfoList[i].SubImage.OriginalName
					after = currProject.SubImageInfoList[i].SubImage.OriginalName
					if before == after {
						break
					}
					beforeUrl = prevProject.SubImageInfoList[i].SubImageUrl
					afterUrl = currProject.SubImageInfoList[i].SubImageUrl
				}
			case string(PROJECT_CHANGED_CONTENT_SPV):
				before = prevProject.Spv.Name
				after = currProject.Spv.Name
			}

			// 변경 된 정보의 내용 저장
			ProjectChangedContent{}.create(ctx, content, pih.Id, before, after, "", beforeHtml, aterHtml, beforeUrl, afterUrl)
		}
	}
}

// 프로젝트 첨부파일 비교
func (pih *ProjectInfoHistory) CompareAttachment(ctx context.Context, prevProject, currProject *Project) {

	prevAttachments := prevProject.Attachments
	currAttachments := currProject.Attachments

	if len(prevAttachments) <= 0 && len(currAttachments) <= 0 {
		return
	}

	pih.Id = 0

	// 삭제 처리 ( ID가 달라도 같은 이름이면 무시)
	for _, pa := range prevAttachments {

		var find bool = false
		for _, ca := range currAttachments {
			if (pa.OriginalName == ca.OriginalName) || (pa.Id == ca.Id) {
				find = true
				break
			}
		}

		if !find {
			pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
			ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_REFERENCE_FILE), pih.Id, pa.OriginalName, string(PROJECT_NEWS_DEL), "", "", "", pa.FileUrl, "")
		}
	}

	// 추가 처리
	for _, ca := range currAttachments {
		var find bool = false
		for _, pa := range prevAttachments {
			if ca.Id == pa.Id {
				find = true
				break
			}
		}

		if !find {
			pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
			ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_REFERENCE_FILE), pih.Id, string(PROJECT_NEWS_ADD), ca.OriginalName, "", "", "", "", ca.FileUrl)
		}
	}
}

// 프로젝트 뉴스 비교
func (pih *ProjectInfoHistory) CompareNews(ctx context.Context, prevProject, currProject *Project) {

	prevNews := prevProject.News
	currNews := currProject.News

	pih.Id = 0
	// 삭제, 변경 처리
	for _, pn := range prevNews {

		var find bool = false
		for _, cn := range currNews {

			if pn.Title == cn.Title {

				find = true
				// hash 변경 후 비교
				s1 := sha256.Sum256([]byte(pn.ContentHtml))
				s2 := sha256.Sum256([]byte(cn.ContentHtml))
				if s1 != s2 {

					pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
					ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_NEWS), pih.Id, pn.Title, string(PROJECT_NEWS_CHANGE), "", pn.ContentHtml, cn.ContentHtml, "", "")
				}
				break
			}
		}

		if !find {
			pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
			ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_NEWS), pih.Id, pn.Title, string(PROJECT_NEWS_DEL), "", "", "", "", "")
		}
	}

	// 추가 처리
	for _, cn := range currNews {
		var find bool = false
		for _, pn := range prevNews {
			if cn.Title == pn.Title {
				find = true
				break
			}
		}

		if !find {
			pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
			ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_NEWS), pih.Id, cn.Title, string(PROJECT_NEWS_ADD), "", "", "", "", "")
		}
	}
}

// 서면 파일 비교
func (pih *ProjectInfoHistory) CompareDocument(ctx context.Context, prevProject, currProject *Project) {

	// 계약성립 전 서면 비교
	prevDoc := prevProject.ContractBeforePage
	currDoc := currProject.ContractBeforePage
	if prevDoc == nil && currDoc != nil {
		pih.Id = 0
		pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
		ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_CONTRACT_BEFORE_PAGE), pih.Id, string(PROJECT_NEWS_ADD), currDoc.OriginalName, "", "", "", "", currDoc.FileUrl)
	} else if prevDoc != nil && currDoc != nil {
		pih.Id = 0
		if prevDoc.OriginalName != currDoc.OriginalName {
			pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
			ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_CONTRACT_BEFORE_PAGE), pih.Id, prevDoc.OriginalName, currDoc.OriginalName, "", "", "", prevDoc.FileUrl, currDoc.FileUrl)
		}
	} else if prevDoc != nil && currDoc == nil {
		pih.Id = 0
		pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
		ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_CONTRACT_BEFORE_PAGE), pih.Id, prevDoc.OriginalName, string(PROJECT_NEWS_DEL), "", "", "", prevDoc.FileUrl, "")
	}

	// 계약성립 시 서면 비교
	prevDoc = prevProject.ContractAfterPage
	currDoc = currProject.ContractAfterPage
	if prevDoc == nil && currDoc != nil {
		pih.Id = 0
		pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
		ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_CONTRACT_AFTER_PAGE), pih.Id, string(PROJECT_NEWS_ADD), currDoc.OriginalName, "", "", "", "", currDoc.FileUrl)
	} else if prevDoc != nil && currDoc != nil {
		pih.Id = 0
		if prevDoc.OriginalName != currDoc.OriginalName {
			pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
			ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_CONTRACT_AFTER_PAGE), pih.Id, prevDoc.OriginalName, currDoc.OriginalName, "", "", "", prevDoc.FileUrl, currDoc.FileUrl)
		}
	} else if prevDoc != nil && currDoc == nil {
		pih.Id = 0
		pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
		ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_CONTRACT_AFTER_PAGE), pih.Id, prevDoc.OriginalName, string(PROJECT_NEWS_DEL), "", "", "", prevDoc.FileUrl, "")
	}

	// 익명조합 계약서 비교
	prevDoc = prevProject.SilentPartnerships
	currDoc = currProject.SilentPartnerships
	if prevDoc == nil && currDoc != nil {
		pih.Id = 0
		pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
		ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_SILENT_PARTNERSHIPS), pih.Id, string(PROJECT_NEWS_ADD), currDoc.OriginalName, "", "", "", "", currDoc.FileUrl)
	} else if prevDoc != nil && currDoc != nil {
		pih.Id = 0

		if prevDoc.OriginalName != currDoc.OriginalName {
			pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
			ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_SILENT_PARTNERSHIPS), pih.Id, prevDoc.OriginalName, currDoc.OriginalName, "", "", "", prevDoc.FileUrl, currDoc.FileUrl)
		}
	} else if prevDoc != nil && currDoc == nil {
		pih.Id = 0
		pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
		ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_SILENT_PARTNERSHIPS), pih.Id, prevDoc.OriginalName, string(PROJECT_NEWS_DEL), "", "", "", prevDoc.FileUrl, "")
	}

	// 펀드 보고서 비교
	prevDoc = prevProject.ProjectReport
	currDoc = currProject.ProjectReport
	if prevDoc == nil && currDoc != nil {
		pih.Id = 0
		pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
		ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_PROJECT_REPORT), pih.Id, string(PROJECT_NEWS_ADD), currDoc.OriginalName, "", "", "", "", currDoc.FileUrl)
	} else if prevDoc != nil && currDoc != nil {
		pih.Id = 0

		if prevDoc.OriginalName != currDoc.OriginalName {
			pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
			ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_PROJECT_REPORT), pih.Id, prevDoc.OriginalName, currDoc.OriginalName, "", "", "", prevDoc.FileUrl, currDoc.FileUrl)
		}
	} else if prevDoc != nil && currDoc == nil {
		pih.Id = 0
		pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
		ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_PROJECT_REPORT), pih.Id, prevDoc.OriginalName, string(PROJECT_NEWS_DEL), "", "", "", prevDoc.FileUrl, "")
	}
}

// 대출 상태 비교
func (pih *ProjectInfoHistory) CompareLoan(ctx context.Context, prevProject, currProject *Project) {

	pih.Id = 0
	if prevProject.ProjectInvestmentExecute == nil && currProject.ProjectInvestmentExecute != nil {
		pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
		ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_LOAN_COMPLETED), pih.Id, "", "", "", "", "", "", "")
	}
}

// 마일스톤 비교
func (pih *ProjectInfoHistory) CompareMilestone(ctx context.Context, prevProject, currProject *Project) {

	prevMilestones := prevProject.ProjectMilestones
	currMilestones := currProject.ProjectMilestones

	for _, prev := range prevMilestones {

		var find bool
		for _, curr := range currMilestones {
			if prev.MilestoneInfo == curr.MilestoneInfo {
				find = true

				// 마일스톤 번호 비교
				if prev.MilestoneNo != curr.MilestoneNo {
					pih.Id = 0
					pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
					ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_MILESTONE_NO), pih.Id, strconv.Itoa(prev.MilestoneNo), strconv.Itoa(curr.MilestoneNo), " ("+prev.MilestoneInfo+")", "", "", "", "")
				}

				// 마일스톤 상태 비교
				if prev.MilestoneStatus != curr.MilestoneStatus {
					pih.Id = 0
					pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
					ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_MILESTONE_STATUS), pih.Id, "PROJECT_"+string(prev.MilestoneStatus), "PROJECT_"+string(curr.MilestoneStatus), " ("+prev.MilestoneInfo+")", "", "", "", "")
				}

				// 마일스톤 값 비교
				if prev.MilestoneAssignment != curr.MilestoneAssignment {
					pih.Id = 0
					pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
					ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_MILESTONE_ASSIGNMENT), pih.Id, strconv.Itoa(prev.MilestoneAssignment), strconv.Itoa(curr.MilestoneAssignment), " ("+prev.MilestoneInfo+")", "", "", "", "")
				}

				// 마일스톤 비고 비교
				if prev.MilestoneNote != curr.MilestoneNote {
					pih.Id = 0
					pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
					ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_MILESTONE_NOTE), pih.Id, prev.MilestoneNote, curr.MilestoneNote, " ("+prev.MilestoneInfo+")", "", "", "", "")
				}

				break
			}
		}
		// 찾지 못햇다면 삭제
		if !find {
			pih.Id = 0
			pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
			ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_MILESTONE), pih.Id, prev.MilestoneInfo, string(PROJECT_MILESTONE_DEL), "", "", "", "", "")
		}
	}

	// 새롭게 추가 된 마일스톤인지를 확인 후 테이블에 저장
	for _, curr := range currMilestones {

		var find bool
		for _, prev := range prevMilestones {
			if curr.MilestoneInfo == prev.MilestoneInfo {
				find = true
				break
			}
		}
		if !find {
			pih.Id = 0
			pih.createHistory(ctx, string(PROJECT_HISTORY_CONTENTS))
			ProjectChangedContent{}.create(ctx, string(PROJECT_CHANGED_CONTENT_MILESTONE), pih.Id, curr.MilestoneInfo, string(PROJECT_MILESTONE_ADD), "", "", "", "", "")
		}
	}
}

func (pih *ProjectInfoHistory) createHistory(ctx context.Context, substanceType string) {

	if pih.Id > 0 {
		return
	}

	pih.SubstanceType = substanceType
	if _, err := factory.DB(ctx).Table("project_info_history").Insert(pih); err != nil {
		logrus.Println("Error saving projects's information history")
	}
}

func (ProjectChangedContent) create(ctx context.Context, contentType string, infoId int64, before, after, suffix, beforeHtml, afterHtml, beforeUrl, afterUrl string) {
	var projectChangedContent ProjectChangedContent
	projectChangedContent.InfoId = infoId
	projectChangedContent.ContentType = contentType
	projectChangedContent.ContentSuffix = suffix
	projectChangedContent.PrevContent = before
	projectChangedContent.CurrContent = after
	projectChangedContent.PrevContentHtml = beforeHtml
	projectChangedContent.CurrContentHtml = afterHtml
	projectChangedContent.PrevUrl = beforeUrl
	projectChangedContent.CurrUrl = afterUrl

	if _, err := factory.DB(ctx).Table("project_changed_content").Insert(&projectChangedContent); err != nil {
		logrus.Println("Error saving project's changed content")
	}
}
