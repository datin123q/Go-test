package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/pkg/errors"
)

type ProjectViewCount struct {
	Id        int64              `json:"id"`
	ProjectId int64              `json:"project_id" xorm:"index"`
	ViewCount int64              `json:"view_count" xorm:"view_count"`
	Updated   jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Created   jsonutils.JSONTime `json:"created" xorm:"created"`
}

func (ProjectViewCount) TableName() string {
	return "project_view_count"
}

func (p ProjectViewCount) SelectAll() string {
	return p.TableName() + ".*"
}

func (ProjectViewCount) GetByProjectId(ctx context.Context, pid int64) (*ProjectViewCount, error) {

	var pv ProjectViewCount
	if has, err := factory.DB(ctx).Where("project_id = ?", pid).Limit(1).Get(&pv); err != nil {
		return nil, err
	} else if !has {
		pv = ProjectViewCount{
			ProjectId: pid,
			ViewCount: 0,
		}
		err := pv.Create(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fail Project View Create")
		}
		return &pv, nil
	}

	return &pv, nil
}

func (ProjectViewCount) GetById(ctx context.Context, id int64) (*ProjectViewCount, error) {

	var pv ProjectViewCount
	if has, err := factory.DB(ctx).ID(id).Get(&pv); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &pv, nil
}

func (v *ProjectViewCount) UpdateCount(ctx context.Context) (err error) {
	v.ViewCount++
	_, err = factory.DB(ctx).ID(v.Id).Update(v)
	return err
}

func (r *ProjectViewCount) Create(ctx context.Context) (err error) {
	query := factory.DB(ctx)
	_, err = query.ID(r.Id).Insert(r)
	return err
}
