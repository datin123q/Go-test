package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
)

func Test_projectMilestoneMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.ProjectQueryOption
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []domain.ProjectMilestone
	}{
		{
			name: "프로젝트 마일스톤 전체조회 => 3건이 조회되기를 기대합니다.",
			args: args{
				ctx:     context.Background(),
				options: []cdomain.ProjectQueryOption{},
			},
			want: []domain.ProjectMilestone{
				mockProjectMilestone1,
				mockProjectMilestone2,
				mockProjectMilestone3,
			},
		},
		{
			name: "프로젝트 마일스톤 특정 프로젝트 ID로 조회 : 1000 => 2건이 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.ProjectQueryOption{
					cdomain.ProjectQueryOptionFunc().WithProjectId(1000),
				},
			},
			want: []domain.ProjectMilestone{
				mockProjectMilestone1,
				mockProjectMilestone2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := projectMilestoneMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{}, time.Time{})

			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "GetAll() = %v, want %v", got, tt.want)
		})
	}
}
