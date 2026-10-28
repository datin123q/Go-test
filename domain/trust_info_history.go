package domain

import (
	"context"
	"errors"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type PRINT_STATUS_TYPE string

const (
	PRINT_STATUS_TYPE_CODE                   = "PRINT_STATUS_TYPE"
	PRINT_STATUS_COMPLETE  PRINT_STATUS_TYPE = "PRINT_STATUS_COMPLETE" // 완료
	PRINT_STATUS_FAIL      PRINT_STATUS_TYPE = "PRINT_STATUS_FAIL"     // 실패
)

func (TrustInfoHistory) GetTrustInfoHistoryPrintStatusTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(PRINT_STATUS_COMPLETE), Name: i18nutils.Translate("commoncode.PRINT_STATUS_TYPE.PRINT_STATUS_COMPLETE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PRINT_STATUS_FAIL), Name: i18nutils.Translate("commoncode.PRINT_STATUS_TYPE.PRINT_STATUS_FAIL", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: PRINT_STATUS_TYPE_CODE,
		Codes:     codes,
	}
}

type TrustInfoHistory struct {
	Id          int64              `json:"id"`                     // 고유 아이디
	PrintStatus string             `json:"print_status"`           // 출력 상태
	Name        string             `json:"name"`                   // 처리자
	Email       string             `json:"email"`                  // 처리자 이메일
	Period      string             `json:"period"`                 // 대상 기간
	Created     jsonutils.JSONTime `json:"created" xorm:"created"` // 처리 일시
}

func (TrustInfoHistory) GetAll(ctx context.Context, sortby, order string, offset, limit int) (int64, []TrustInfoHistory, error) {

	query := factory.DB(ctx).Table("trust_info_history").Select("trust_info_history.*")

	if err := modelhelper.SetSortOrder(query, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var trustInfoHistoryList []TrustInfoHistory
	var totalCount int64
	totalCount, err := query.Limit(limit, offset).Desc("created").FindAndCount(&trustInfoHistoryList)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, trustInfoHistoryList, nil
}

func (TrustInfoHistory) GetLatest(ctx context.Context) (*TrustInfoHistory, error) {

	var trustInfoHistory TrustInfoHistory
	_, err := factory.DB(ctx).Table("trust_info_history").Desc("id").Get(&trustInfoHistory)
	if err != nil {
		return nil, err
	}

	return &trustInfoHistory, nil
}

func (TrustInfoHistory) Create(ctx context.Context, processId int64, printStatus, startDate, endDate string) (int64, error) {

	user, err := AdminUser{}.GetById(ctx, processId)
	if err != nil {
		return 0, err
	} else if user == nil {
		return 0, errors.New("Not Found User")
	}

	startTime, err := jsonutils.ParseDateTimeJST(time.DateTime, startDate)
	if err != nil {
		return 0, err
	}
	endTime, err := jsonutils.ParseDateTimeJST(time.DateTime, endDate)
	if err != nil {
		return 0, err
	}

	var trustInfoHistory = &TrustInfoHistory{
		Name:        user.Name,
		Email:       user.Email,
		PrintStatus: printStatus,
		Period:      startTime.Format(time.DateOnly) + " ~ " + endTime.Format(time.DateOnly),
	}

	return factory.DB(ctx).Table("trust_info_history").Insert(trustInfoHistory)
}
