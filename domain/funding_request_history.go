package domain

import (
	"context"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type FUNDING_HISTORY_TYPE string

const (
	FUNDING_HISTORY_TYPE_CODE                      = "FUNDING_HISTORY_TYPE"
	FUNDING_HISTORY_REVIEW    FUNDING_HISTORY_TYPE = "FUNDING_HISTORY_REVIEW"    // 검토상태 변경
	FUNDING_HISTORY_CONTENTS  FUNDING_HISTORY_TYPE = "FUNDING_HISTORY_CONTENTS"  // 내용 변경
	FUNDING_HISTORY_REFERENCE FUNDING_HISTORY_TYPE = "FUNDING_HISTORY_REFERENCE" // 참고자료 변경
)

func (FundingRequestHistory) GetFundingHistoryTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(FUNDING_HISTORY_REVIEW), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_TYPE.FUNDING_HISTORY_REVIEW", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_HISTORY_CONTENTS), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_TYPE.FUNDING_HISTORY_CONTENTS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_HISTORY_REFERENCE), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_TYPE.FUNDING_HISTORY_REFERENCE", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: FUNDING_HISTORY_TYPE_CODE,
		Codes:     codes,
	}
}

type FUNDING_HISTORY_STATE_TYPE string

const (
	FUNDING_HISTORY_STATE_TYPE_CODE                            = "FUNDING_HISTORY_STATE_TYPE"
	FUNDING_STATE_NEED              FUNDING_HISTORY_STATE_TYPE = "FUNDING_STATE_NEED"
	FUNDING_STATE_NOT_NEED          FUNDING_HISTORY_STATE_TYPE = "FUNDING_STATE_NOT_NEED"
)

func (FundingRequestHistory) GetFundingHistoryStateTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(FUNDING_STATE_NEED), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_STATE_TYPE.FUNDING_STATE_NEED", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_STATE_NOT_NEED), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_STATE_TYPE.FUNDING_STATE_NOT_NEED", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: FUNDING_HISTORY_STATE_TYPE_CODE,
		Codes:     codes,
	}
}

type FUNDING_HISTORY_CONTENT_TYPE string

const (
	FUNDING_HISTORY_CONTENT_TYPE_CODE                                 = "FUNDING_HISTORY_CONTENT_TYPE"
	FUNDING_REVIEW_STATE                 FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_REVIEW_STATE"                 // 검토상태
	FUNDING_CONTENT_REFERENCE            FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_CONTENT_REFERENCE"            // 참고자료
	FUNDING_CONTENT_NAME                 FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_CONTENT_NAME"                 // 법인명
	FUNDING_CONTENT_ZIP_CODE             FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_CONTENT_ZIP_CODE"             // 우편번호
	FUNDING_CONTENT_ADDRESS              FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_CONTENT_ADDRESS"              // 주소
	FUNDING_CONTENT_ADDRESS2             FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_CONTENT_ADDRESS2"             // 상세주소
	FUNDING_CONTENT_CORP_NO              FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_CONTENT_CORP_NO"              // 법인번호
	FUNDING_CONTENT_PROVINCE             FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_CONTENT_PROVINCE"             // 현
	FUNDING_CONTENT_CITY                 FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_CONTENT_CITY"                 // 시
	FUNDING_CONTENT_EMAIL                FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_CONTENT_EMAIL"                // 이메일
	FUNDING_CONTENT_PHONE                FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_CONTENT_PHONE"                // 전화번호
	FUNDING_CONTENT_EQUIPMENT_NAME       FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_CONTENT_EQUIPMENT_NAME"       // 프로젝트명
	FUNDING_CONTENT_EQUIPMENT_ADDRESS    FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_CONTENT_EQUIPMENT_ADDRESS"    // 소재지 주소
	FUNDING_CONTENT_EQUIPMENT_ADDRESS2   FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_CONTENT_EQUIPMENT_ADDRESS2"   // 소재지 상세주소
	FUNDING_CONTENT_EQUIPMENT_INFO       FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_CONTENT_EQUIPMENT_INFO"       // 프로젝트 개요
	FUNDING_CONTENT_EQUIPMENT_CONSULTING FUNDING_HISTORY_CONTENT_TYPE = "FUNDING_CONTENT_EQUIPMENT_CONSULTING" // 상담 내용
)

func (FundingRequestHistory) GetFundingHistoryContentTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(FUNDING_REVIEW_STATE), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_REVIEW_STATE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_CONTENT_REFERENCE), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_REFERENCE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_CONTENT_NAME), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_NAME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_CONTENT_ZIP_CODE), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_ZIP_CODE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_CONTENT_ADDRESS), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_ADDRESS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_CONTENT_ADDRESS2), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_ADDRESS2", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_CONTENT_CORP_NO), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_CORP_NO", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_CONTENT_PROVINCE), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_PROVINCE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_CONTENT_CITY), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_CITY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_CONTENT_EMAIL), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_EMAIL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_CONTENT_PHONE), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_PHONE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_CONTENT_EQUIPMENT_NAME), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_EQUIPMENT_NAME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_CONTENT_EQUIPMENT_ADDRESS), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_EQUIPMENT_ADDRESS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_CONTENT_EQUIPMENT_ADDRESS2), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_EQUIPMENT_ADDRESS2", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_CONTENT_EQUIPMENT_INFO), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_EQUIPMENT_INFO", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUNDING_CONTENT_EQUIPMENT_CONSULTING), Name: i18nutils.Translate("commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_EQUIPMENT_CONSULTING", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: FUNDING_HISTORY_CONTENT_TYPE_CODE,
		Codes:     codes,
	}
}

type FundingRequestHistory struct {
	Id               int64                     `json:"id"`                      // 고유아이디
	FundingRequestId int64                     `json:"funding_request_id"`      // 상태변경된 펀딩신청 아이디
	SubstanceType    string                    `json:"substance_type"`          // 내용 표기방식
	ProcessMethod    USER_HISTORY_PROCESS_TYPE `json:"process_method" xorm:"-"` // 처리 방식
	Name             string                    `json:"name"`                    // 처리자
	Email            string                    `json:"email"`                   // 처리자 이메일
	ContentType      string                    `json:"content_type"`            // 상태변경 타입
	PrevContent      string                    `json:"prev_content"`            // 변경 전 내용
	CurrContent      string                    `json:"curr_content"`            // 변경 후 내용
	PrevUrl          string                    `json:"prev_url"`                // 변경 전 첨부파일 url
	CurrUrl          string                    `json:"curr_url"`                // 변경 후 첨부파일 url
	Created          jsonutils.JSONTime        `json:"created" xorm:"created"`  // 처리일시
}

func (FundingRequestHistory) GetAll(ctx context.Context, id int64, sortby, order string, offset, limit int) (int64, []FundingRequestHistory, error) {
	q := factory.DB(ctx).Table("funding_request_history").Select("funding_request_history.*")
	if id > 0 {
		q.Where("funding_request_history.funding_request_id = ?", id)
	}

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []FundingRequestHistory
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (FundingRequestHistory) Create(ctx context.Context, before, after interface{}, processId int64) error {

	// i18n tag로 지정된 데이터 중
	// 변경된 부분이 어떤곳인지 확인하기 위하여
	// 변경 전 데이터와 변경 후 데이터를 가져온다.
	prevFR := before.(*FundingRequest)
	currFR := after.(*FundingRequest)
	prevData := FundingRequestHistory{}.convertTagToMap(reflect.ValueOf(prevFR).Elem())
	currData := FundingRequestHistory{}.convertTagToMap(reflect.ValueOf(currFR).Elem())

	user, _ := AdminUser{}.GetById(ctx, processId)

	err := FundingRequestHistory{}.compareReviewState(ctx, prevData, currData, user.Name, user.Email, currFR.Id)
	if err != nil {
		return errors.Wrap(err, "error saving funding request history of compareReviweState")
	}
	err = FundingRequestHistory{}.compareContents(ctx, prevData, currData, user.Name, user.Email, currFR.Id)
	if err != nil {
		return errors.Wrap(err, "error saving funding request history of compareContents")
	}
	err = FundingRequestHistory{}.compareReference(ctx, prevFR, currFR, user.Name, user.Email, currFR.Id)
	if err != nil {
		return errors.Wrap(err, "error saving funding request history of compareReference")
	}

	return nil
}

// i18n tag로 지정된 필드값을 저장
func (FundingRequestHistory) convertTagToMap(val reflect.Value) map[string]interface{} {

	var mapData = make(map[string]interface{}, 0)

	// 필드에 i18n tag로 지정된 필드를 찾아 mapData에 저장한다.
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		i18nStr := tag.Get("i18n")
		if len(i18nStr) <= 0 {
			continue
		}

		// i18n tag의 필드를 찾았다면 문자열 .을 기준으로 분리를 하여 마지막 값을 저장할 key로 사용한다.
		strs := strings.Split(i18nStr, ".")
		str := strs[len(strs)-1]
		if len(str) <= 0 {
			continue
		}

		switch t := val.Field(i).Interface().(type) {
		case string:
			mapData[str] = t
		case bool:
			// 필드가 bool 데이터인 경우 변경값을 true, false로 처리로그에 나타내지 않고
			// 번역된 값을 쓰기위해 코드값으로 저장해 준다.
			if t == true {
				mapData[str] = string(FUNDING_STATE_NEED)
			} else if t == false {
				mapData[str] = string(FUNDING_STATE_NOT_NEED)
			}
		default:
			// 문자열이 아닌 데이터들에 대해서는 문자열로 저장하기 위해 변환한다.
			if value, ok := val.Field(i).Interface().(REQUEST_STATUS); ok {
				mapData[str] = string(value)
			} else if value, ok := val.Field(i).Interface().(jsonutils.JSONTime); ok {
				mapData[str] = value.Time().String()
			}
		}
	}

	return mapData
}

func (FundingRequestHistory) compareReviewState(ctx context.Context, prevData, currData map[string]interface{}, name, email string, requestId int64) error {

	var prev, curr string
	if val, ok := prevData[string(FUNDING_REVIEW_STATE)]; ok {
		prev = val.(string)
	}

	if val, ok := currData[string(FUNDING_REVIEW_STATE)]; ok {
		curr = val.(string)
	}
	// 처리상태 비교 후 히스토리를 남김
	if prev != curr {

		// 어디가 변경 되었는지의 변경 상태(SubstanceType)와 변경 내용(ContentType)에 번역된 값을 쓰기 위해 번역 코드값으로 넣어준다
		// 처리자(Name)와 처리자 이메일(Email)을 저장한다.
		// 변경 전 내용(PrevContent)과 변경 후 내용(CurrContent)에 해당 내용을 저장한다.
		fundingRequestHistory := FundingRequestHistory{
			FundingRequestId: requestId,
			SubstanceType:    string(FUNDING_HISTORY_REVIEW),
			Name:             name,
			Email:            email,
			ContentType:      string(FUNDING_REVIEW_STATE),
			PrevContent:      prev,
			CurrContent:      curr,
		}

		if _, err := factory.DB(ctx).Table("funding_request_history").Insert(&fundingRequestHistory); err != nil {
			return err
		}
	}

	return nil
}

// 변경된 내용을 찾기 위한 메서드
func (FundingRequestHistory) compareContents(ctx context.Context, prevData, currData map[string]interface{}, name, email string, requestId int64) error {

	// 변경 된 내용을 비교하기 위해 i18n tag로 지정한 값을 비교하기 위한 슬라이스에 넣어준다.
	contentTypes := []string{string(FUNDING_CONTENT_NAME), string(FUNDING_CONTENT_ZIP_CODE), string(FUNDING_CONTENT_ADDRESS), string(FUNDING_CONTENT_ADDRESS2),
		string(FUNDING_CONTENT_CORP_NO), string(FUNDING_CONTENT_PROVINCE), string(FUNDING_CONTENT_CITY), string(FUNDING_CONTENT_EMAIL),
		string(FUNDING_CONTENT_PHONE), string(FUNDING_CONTENT_EQUIPMENT_NAME), string(FUNDING_CONTENT_EQUIPMENT_ADDRESS),
		string(FUNDING_CONTENT_EQUIPMENT_ADDRESS2), string(FUNDING_CONTENT_EQUIPMENT_INFO),
		string(FUNDING_CONTENT_EQUIPMENT_CONSULTING)}

	// 이전 데이터와 현재 데이터를 비교 후 다른 것들만 funding_request_history 테이블에 저장한다.
	for _, content := range contentTypes {

		if prevData[content] == nil {
			continue
		}

		prev := prevData[content].(string)
		curr := currData[content].(string)

		if prev != curr {
			fundingRequestHistory := FundingRequestHistory{
				FundingRequestId: requestId,
				SubstanceType:    string(FUNDING_HISTORY_CONTENTS),
				Name:             name,
				Email:            email,
				ContentType:      content,
				PrevContent:      prev,
				CurrContent:      curr,
			}

			if _, err := factory.DB(ctx).Table("funding_request_history").Insert(&fundingRequestHistory); err != nil {
				return err
			}
		}
	}
	return nil
}

// 변경된 첨부파일 비교
func (FundingRequestHistory) compareReference(ctx context.Context, before, after interface{}, name, email string, requestId int64) error {

	// 변경 전 첨부파일과 변경 후 첨부파일을 슬라이스에 저장
	var prevFiles []*Attachment
	var currFiles []*Attachment
	if v, ok := after.(*FundingRequest); ok {
		prevFiles = before.(*FundingRequest).Attachments
		currFiles = v.Attachments
	}

	if len(prevFiles) <= 0 && len(currFiles) <= 0 {
		return nil
	}

	// 삭제 처리, ( ID가 달라도 같은 이름이 존재하면 무시)
	for _, pu := range prevFiles {

		// 변경 전과 변경 후 첨부파일 제목을 비교하여 같은게 있는지 검사
		var find bool
		for _, cu := range currFiles {
			if (pu.OriginalName == cu.OriginalName) || (pu.Id == cu.Id) {
				find = true
				break
			}
		}

		// 변경 전 첨부파일과 같은 변경 후 첨부파일이 없다면 처리로그에 삭제 되었다고 적용 후 테이블에 저장
		if !find {
			fundingRequestHistory := FundingRequestHistory{
				FundingRequestId: requestId,
				SubstanceType:    string(FUNDING_HISTORY_REFERENCE),
				Name:             name,
				Email:            email,
				ContentType:      string(FUNDING_CONTENT_REFERENCE),
				PrevContent:      pu.OriginalName,
				CurrContent:      string(USER_HISTORY_DOCUMENT_DEL),
				PrevUrl:          pu.FileUrl,
			}

			if _, err := factory.DB(ctx).Table("funding_request_history").Insert(&fundingRequestHistory); err != nil {
				return err
			}
		}
	}

	// 추가 처리
	for _, cu := range currFiles {
		var find bool
		for _, pu := range prevFiles {
			if cu.Id == pu.Id {
				find = true
				break
			}
		}

		// 변경 후 첨부파일을 변경 전 첨부파일에서 찾을 수 없다면 추가 후 테이블에 저장
		if !find {
			fundingRequestHistory := FundingRequestHistory{
				FundingRequestId: requestId,
				SubstanceType:    string(FUNDING_HISTORY_REFERENCE),
				Name:             name,
				Email:            email,
				ContentType:      string(FUNDING_CONTENT_REFERENCE),
				PrevContent:      string(USER_HISTORY_DOCUMENT_ADD),
				CurrContent:      cu.OriginalName,
				CurrUrl:          cu.FileUrl,
			}

			if _, err := factory.DB(ctx).Table("funding_request_history").Insert(&fundingRequestHistory); err != nil {
				return err
			}
		}
	}
	return nil
}
