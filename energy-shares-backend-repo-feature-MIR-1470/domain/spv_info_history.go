package domain

import (
	"context"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type SPV_HISTORY_TYPE string

const (
	SPV_HISTORY_TYPE_CODE                  = "SPV_HISTORY_TYPE"
	SPV_HISTORY_CONTENTS  SPV_HISTORY_TYPE = "SPV_HISTORY_CONTENTS" // 내용 변경
)

func (SpvInfoHistory) GetSpvHistoryTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(SPV_HISTORY_CONTENTS), Name: i18nutils.Translate("commoncode.SPV_HISTORY_TYPE.SPV_HISTORY_CONTENTS", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: SPV_HISTORY_TYPE_CODE,
		Codes:     codes,
	}
}

type SPV_CHANGED_CONTENT_TYPE string

const (
	SPV_CHANGED_CONTENT_TYPE_CODE                          = "SPV_CHANGED_CONTENT_TYPE"
	SPV_CHANGED_CONTENT_NAME      SPV_CHANGED_CONTENT_TYPE = "SPV_CHANGED_CONTENT_CORP_NAME" // 법인명
	SPV_CHANGED_CONTENT_CORPNO    SPV_CHANGED_CONTENT_TYPE = "SPV_CHANGED_CONTENT_CORPNO"    // 법인번호
	SPV_CHANGED_CONTENT_ADDRESS   SPV_CHANGED_CONTENT_TYPE = "SPV_CHANGED_CONTENT_ADDRESS"   // 주소
	SPV_CHANGED_CONTENT_ADDRESS2  SPV_CHANGED_CONTENT_TYPE = "SPV_CHANGED_CONTENT_ADDRESS2"  // 상세주소
	SPV_CHANGED_CONTENT_PROVINCE  SPV_CHANGED_CONTENT_TYPE = "SPV_CHANGED_CONTENT_PROVINCE"  // 현
	SPV_CHANGED_CONTENT_CITY      SPV_CHANGED_CONTENT_TYPE = "SPV_CHANGED_CONTENT_CITY"      // 시 또는 구, 군
	SPV_CHANGED_CONTENT_EMAIL     SPV_CHANGED_CONTENT_TYPE = "SPV_CHANGED_CONTENT_EMAIL"     // 이메일
	SPV_CHANGED_CONTENT_PHONE     SPV_CHANGED_CONTENT_TYPE = "SPV_CHANGED_CONTENT_PHONE"     // 전화번호
)

func (SpvInfoHistory) GetSpvChangedContentTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(SPV_CHANGED_CONTENT_NAME), Name: i18nutils.Translate("commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_CORP_NAME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(SPV_CHANGED_CONTENT_CORPNO), Name: i18nutils.Translate("commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_CORPNO", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(SPV_CHANGED_CONTENT_ADDRESS), Name: i18nutils.Translate("commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_ADDRESS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(SPV_CHANGED_CONTENT_ADDRESS2), Name: i18nutils.Translate("commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_ADDRESS2", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(SPV_CHANGED_CONTENT_PROVINCE), Name: i18nutils.Translate("commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_PROVINCE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(SPV_CHANGED_CONTENT_CITY), Name: i18nutils.Translate("commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_CITY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(SPV_CHANGED_CONTENT_EMAIL), Name: i18nutils.Translate("commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_EMAIL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(SPV_CHANGED_CONTENT_PHONE), Name: i18nutils.Translate("commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_PHONE", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: SPV_CHANGED_CONTENT_TYPE_CODE,
		Codes:     codes,
	}
}

type SpvInfoHistory struct {
	Id            int64              `json:"id"`                     // 고유아이디
	SpvId         int64              `json:"spv_id"`                 // 상태변경된 개발회사 아이디
	SubstanceType string             `json:"substance_type"`         // 내용 표기방식
	Name          string             `json:"name"`                   // 처리자
	Email         string             `json:"email"`                  // 처리자 이메일
	Created       jsonutils.JSONTime `json:"created" xorm:"created"` // 처리일시
}

type SpvChangedContent struct {
	Id          int64  `json:"id"`
	InfoId      int64  `json:"info_id"`      // 개발회사 상태변경 히스토리 아이디
	ContentType string `json:"content_type"` // 변경된 내용
	PrevContent string `json:"prev_content"` // 변경 전 내용
	CurrContent string `json:"curr_content"` // 변경 후 내용
}

type SpvInfoHistoryExtends struct {
	SpvInfoHistory `xorm:"extends"`
	Content        SpvChangedContent `json:"content" xorm:"extends"`
}

func (SpvInfoHistory) GetAll(ctx context.Context, id int64, sortby, order string, offset, limit int) (int64, []SpvInfoHistoryExtends, error) {

	query := factory.DB(ctx).Table("spv_info_history").Select("spv_info_history.id, spv_info_history.spv_id, spv_info_history.substance_type, spv_info_history.name, spv_info_history.email, spv_info_history.created, spv_changed_content.*").
		Join("INNER", "spv_changed_content", "spv_changed_content.info_id = spv_info_history.id")
	if id > 0 {
		query.Where("spv_info_history.spv_id = ?", id)
	}

	if err := modelhelper.SetSortOrder(query, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var spvInfoHistoryList []SpvInfoHistoryExtends
	var totalCount int64
	totalCount, err := query.Limit(limit, offset).FindAndCount(&spvInfoHistoryList)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, spvInfoHistoryList, nil
}

func (SpvInfoHistory) Create(ctx context.Context, before, after interface{}, processId int64) (int64, error) {

	// i18n tag로 지정된 필드값을 저장하기 위한 맵
	spv := after.(*Spv)
	prevMapData := SpvInfoHistory{}.convertTagToMap(reflect.ValueOf(before.(*Spv)).Elem())
	currMapData := SpvInfoHistory{}.convertTagToMap(reflect.ValueOf(spv).Elem())

	// 처리자 정보를 가져옴
	user, _ := AdminUser{}.GetById(ctx, processId)
	var spvInfoHistory = &SpvInfoHistory{SpvId: spv.Id, SubstanceType: string(SPV_HISTORY_CONTENTS), Name: user.Name, Email: user.Email}

	// 내용 상태 비교 후 히스토리를 남김
	spvInfoHistory.compareContents(ctx, prevMapData, currMapData)

	return 0, nil
}

// i18n tag로 지정된 필드값을 저장
func (SpvInfoHistory) convertTagToMap(val reflect.Value) map[string]interface{} {

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
		default:
			if value, ok := val.Field(i).Interface().(jsonutils.JSONTime); ok {
				mapData[str] = value.Time().String()
			}
		}
	}

	return mapData
}

// 내용 상태 비교
func (sih *SpvInfoHistory) compareContents(ctx context.Context, prevData, currData map[string]interface{}) {

	// 정보를 확인하기 위한 key값을 슬라이스에 저장
	contentTypes := []string{string(SPV_CHANGED_CONTENT_NAME), string(SPV_CHANGED_CONTENT_CORPNO), string(SPV_CHANGED_CONTENT_ADDRESS), string(SPV_CHANGED_CONTENT_ADDRESS2),
		string(SPV_CHANGED_CONTENT_PROVINCE), string(SPV_CHANGED_CONTENT_CITY), string(SPV_CHANGED_CONTENT_EMAIL), string(SPV_CHANGED_CONTENT_PHONE)}

	sih.Id = 0
	for _, content := range contentTypes {

		if prevData[content] == nil {
			continue
		}

		// 정보 카테고리 및 정보 내용이 변경 되었다면 변경된 내용들을 테이블에 저장
		if prevData[content].(string) != currData[content].(string) {
			sih.createHistory(ctx, string(SPV_HISTORY_CONTENTS))
			SpvChangedContent{}.create(ctx, content, sih.Id, prevData[content].(string), currData[content].(string))
		}
	}
}

func (sih *SpvInfoHistory) createHistory(ctx context.Context, substanceType string) {

	if sih.Id > 0 {
		return
	}

	if _, err := factory.DB(ctx).Table("spv_info_history").Insert(sih); err != nil {
		logrus.Println("Error saving spv's information history")
	}
}

func (SpvChangedContent) create(ctx context.Context, contentType string, infoId int64, before, after string) {
	var spvChangedContent SpvChangedContent
	spvChangedContent.InfoId = infoId
	spvChangedContent.ContentType = contentType
	spvChangedContent.PrevContent = before
	spvChangedContent.CurrContent = after

	if _, err := factory.DB(ctx).Table("spv_changed_content").Insert(&spvChangedContent); err != nil {
		logrus.Println("Error saving spv's changed content")
	}
}
