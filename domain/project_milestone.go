package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type MILESTONE_STATUS string

const (
	MILESTONE_STATUS_CODE                  = "MILESTONE_STATUS_CODE"
	MILESTONE_READY       MILESTONE_STATUS = "MILESTONE_READY"
	MILESTONE_GOING       MILESTONE_STATUS = "MILESTONE_GOING"
	MILESTONE_COMPLETED   MILESTONE_STATUS = "MILESTONE_COMPLETED"
)

func (ProjectMilestone) GetMilestoneStatus(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(MILESTONE_READY), Name: i18nutils.Translate("commoncode.MILESTONE_STATUS_CODE.MILESTONE_READY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(MILESTONE_GOING), Name: i18nutils.Translate("commoncode.MILESTONE_STATUS_CODE.MILESTONE_DOING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(MILESTONE_COMPLETED), Name: i18nutils.Translate("commoncode.MILESTONE_STATUS_CODE.MILESTONE_DONE", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: MILESTONE_STATUS_CODE,
		Codes:     codes,
	}
}

type ProjectMilestone struct {
	Id                  int64              `json:"id" xorm:"pk autoincr" swagger:"desc(ID)"`
	ProjectId           int64              `json:"project_id" xorm:"NOT NULL INDEX" swagger:"desc(프로젝트 ID)"`
	MilestoneStatus     MILESTONE_STATUS   `json:"milestone_status" xorm:"NOT NULL INDEX" swagger:"desc(마일스톤 상태)"`
	MilestoneNo         int                `json:"milestone_no" xorm:"NOT NULL INDEX" swagger:"desc(마일스톤 순서 번호)"`
	MilestoneInfo       string             `json:"milestone_info" xorm:"NOT NULL" swagger:"desc(마일스톤 정보)"`
	MilestoneAssignment int                `json:"milestone_assignment" xorm:"NOT NULL" swagger:"desc(마일스톤 값)"`
	MilestoneNote       string             `json:"milestone_note" xorm:"NOT NULL" swagger:"desc(마일스톤 메모)"`
	Created             jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated             jsonutils.JSONTime `json:"updated" xorm:"updated"`
}

func (ProjectMilestone) TableName() string {
	return "project_milestone"
}

func (p ProjectMilestone) SelectAll() string {
	return p.TableName() + ".*"
}

func (p *ProjectMilestone) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(p)
}

func (ProjectMilestone) GetById(ctx context.Context, id int64) (*ProjectMilestone, error) {
	var item ProjectMilestone
	has, err := factory.DB(ctx).ID(id).Get(&item)
	if err != nil {
		return nil, err
	} else if has == false {
		return nil, nil
	}
	return &item, nil
}

func (ProjectMilestone) GetAllByProjectId(ctx context.Context, id int64) ([]ProjectMilestone, error) {
	var items []ProjectMilestone
	cnt, err := factory.DB(ctx).Where("project_id = ?", id).Asc("milestone_no").FindAndCount(&items)
	if err != nil {
		return nil, err
	}
	if cnt == 0 {
		return []ProjectMilestone{}, nil
	}
	return items, nil
}

func (p *ProjectMilestone) Update(ctx context.Context, id int64) (err error) {
	_, err = factory.DB(ctx).ID(id).UseBool("milestone_no").Update(p)
	return err
}

func (ProjectMilestone) Delete(ctx context.Context, id int64) (err error) {
	_, err = factory.DB(ctx).ID(id).Delete(&ProjectMilestone{})
	return err
}
