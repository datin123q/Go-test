package domain

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type ProjectNews struct {
	Id          int64              `json:"id"`
	ProjectId   int64              `json:"project_id"`
	Title       string             `json:"title"`
	OwnerId     int64              `json:"owner_id" xorm:"index"`  // 작성자
	UpdateId    int64              `json:"update_id" xorm:"index"` // 최근 수정자
	ContentHtml string             `json:"content_html" xorm:"LONGTEXT"`
	Posted      bool               `json:"posted"`    // 게시 여부
	PostDate    jsonutils.JSONTime `json:"post_date"` // 게시 일자 관리자 화면에서 수정가능
	Created     jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated     jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted     jsonutils.JSONTime `json:"deleted" xorm:"deleted"` // 삭제시 Deleted 에 시간이 입력되면 조회되지 않음 1년후에 삭제 할수 있도록 할수 잇음
}

func (ProjectNews) TableName() string {
	return "project_news"
}

func (p ProjectNews) SelectAll() string {
	return p.TableName() + ".*"
}

func (n *ProjectNews) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(n)
}

func (ProjectNews) GetAll(ctx context.Context, sortby, order string, offset, limit int) (int64, []ProjectNews, error) {
	var err error
	q := factory.DB(ctx)
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []ProjectNews
	var totalCount int64
	totalCount, err = q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}
	return totalCount, items, nil
}

func (ProjectNews) GetPostedList(ctx context.Context, id int64, offset, limit int) (int64, []*ProjectNews, error) {
	q := factory.DB(ctx)

	var items = make([]*ProjectNews, 0, 0)
	var totalCount int64
	totalCount, err := q.Where("project_id = ? AND posted = true AND DATE(CONVERT_TZ(post_date, 'UTC', 'Asia/Tokyo')) <= DATE(CONVERT_TZ(?, 'UTC', 'Asia/Tokyo'))", id, time.Now().Format(time.DateTime)).Desc("post_date").Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (ProjectNews) GetById(ctx context.Context, id int64) (*ProjectNews, error) {

	var pn ProjectNews
	if has, err := factory.DB(ctx).ID(id).Get(&pn); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &pn, nil
}

func (ProjectNews) GetProjectById(ctx context.Context, id int64, sortby, order string, offset, limit int) (int64, []*ProjectNews, error) {
	var err error
	q := factory.DB(ctx)
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []*ProjectNews
	var totalCount int64
	q = q.Where("project_id = ?", id)

	totalCount, err = q.Limit(limit, offset).FindAndCount(&items)

	if err != nil {
		return 0, nil, err
	}
	return totalCount, items, nil
}

func (ProjectNews) Delete(ctx context.Context, id int64) (err error) {
	_, err = factory.DB(ctx).ID(id).Delete(&ProjectNews{})
	return err
}
