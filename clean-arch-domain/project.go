package clean_arch_domain

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type ProjectRepository interface {
	GetById(ctx context.Context, id int64) (domain.Project, error)
	Update(ctx context.Context, project domain.Project) error
	LockById(ctx context.Context, id int64) (domain.Project, error)
	// TryLockById はLockByIdと違い、ロックを取得できなければ待たずに
	// ErrItemConflictを返す(FOR UPDATE NOWAIT)。同じプロジェクトへの
	// 分割実行(MIR-1204)を直列化する際、他の実行を待たせるのではなく
	// 「今は実行中」と即座に伝えるために使う。
	TryLockById(ctx context.Context, id int64) (domain.Project, error)
	ReduceFinalFundingAmount(ctx context.Context, projectID int64, amount int64) error
	AddFinalFundingAmount(ctx context.Context, projectId int64, amount int64) error
}

type ProjectImageRepository interface {
	GetById(ctx context.Context, id int64) (domain.ProjectImage, error)
}

type ProjectNewsRepository interface {
	GetAll(ctx context.Context, opts ...ProjectQueryOption) ([]domain.ProjectNews, error)
}

type ProjectFileRepository interface {
	GetByProjectIdAndFileType(ctx context.Context, projectId int64, fileType domain.PROJECT_FILE_TYPE) (domain.ProjectFile, error)
}

type ProjectViewCountRepository interface {
	Create(ctx context.Context, projectViewCount domain.ProjectViewCount) error
	GetByProjectId(ctx context.Context, projectId int64) (domain.ProjectViewCount, error)
}

type ProjectMilestoneRepository interface {
	GetAll(ctx context.Context, opts ...ProjectQueryOption) ([]domain.ProjectMilestone, error)
}

type ProjectService interface {
	GetDetail(ctx context.Context, id int64) (domain.Project, error)
}

type ProjectQueryOption = func(*ProjectQueryOptionFields)

func ProjectQueryOptionFunc() ProjectQueryOptionFields {
	return ProjectQueryOptionFields{}
}

type ProjectQueryOptionFields struct {
	projectId *int64
	QueryOptions
}

func (p ProjectQueryOptionFields) WithProjectId(projectId int64) func(*ProjectQueryOptionFields) {
	return func(pqo *ProjectQueryOptionFields) {
		pqo.projectId = &projectId
	}
}

func (p ProjectQueryOptionFields) WithSortBy(sortBy string) func(*ProjectQueryOptionFields) {
	return func(pqo *ProjectQueryOptionFields) {
		pqo.sortBy = &sortBy
	}
}

func (p ProjectQueryOptionFields) WithOrder(order string) func(*ProjectQueryOptionFields) {
	return func(pqo *ProjectQueryOptionFields) {
		pqo.order = &order
	}
}

func (p ProjectQueryOptionFields) WithOffset(offset int) func(*ProjectQueryOptionFields) {
	return func(pqo *ProjectQueryOptionFields) {
		pqo.offset = &offset
	}
}

func (p ProjectQueryOptionFields) WithLimit(limit int) func(*ProjectQueryOptionFields) {
	return func(pqo *ProjectQueryOptionFields) {
		pqo.limit = &limit
	}
}

func (p ProjectQueryOptionFields) ProjectId() *string {
	if p.projectId == nil {
		return nil
	}

	return lo.ToPtr(fmt.Sprintf("%d", *p.projectId))
}
