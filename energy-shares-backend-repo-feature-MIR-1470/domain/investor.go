package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type Investor struct {
	Id           int64              `json:"id" swagger:"desc(ID)"`
	Uid          string             `json:"uid" xorm:"notnull unique" swagger:"desc(사용자 유니크 ID)"`
	CommonId     string             `json:"common_id" xorm:"notnull UUID" swagger:"desc(공용 ID)"`
	InvestorType INVESTOR_TYPE      `json:"investor_type" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INVEST_GRADE" swagger:"desc(投資者タイプ)"` // 投資者タイプ
	InvestDate   jsonutils.JSONTime `json:"invest_date" xorm:"invest_date"`
	Name         string             `json:"name" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CORP_NAME" swagger:"desc(이름)" `
	Email        string             `json:"email" xorm:"notnull index unique" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_EMAIL" swagger:"desc(Email)"`
	Created      jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated      jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted      jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
	ProjectId    int64              `json:"project_id"`
}

type IInvestorRepository interface {
	FindByAcceptedNewsletterAndCompletedInvestors(ctx context.Context, projectId int64) ([]Investor, error)
	FindOperationInvestors(ctx context.Context, projectId int64) ([]Investor, error)
	FindAcceptedNewsletterInvestors(ctx context.Context, projectId int64) ([]Investor, error)
	FindInvestorByProjectIdAndStatus(ctx context.Context, projectId int64, projectStatus PROJECT_STATUS) ([]Investor, error)
	FindInvestorByProjectId(ctx context.Context, projectId int64) ([]Investor, error)
}
