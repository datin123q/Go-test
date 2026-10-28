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
)

func Test_projectViewCountMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx              context.Context
		projectViewCount domain.ProjectViewCount
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "프로젝트 조회수 생성",
			args: args{
				ctx: context.Background(),
				projectViewCount: domain.ProjectViewCount{
					ProjectId: 1000,
					ViewCount: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "프로젝트 조회수 중복생성으로 에러",
			args: args{
				ctx: context.Background(),
				projectViewCount: domain.ProjectViewCount{
					Id:        3000,
					ProjectId: 1000,
					ViewCount: 0,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := projectViewCountMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Create(tt.args.ctx, tt.args.projectViewCount)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestProjectViewCountMysqlRepository_GetByProjectId(t *testing.T) {
	type args struct {
		ctx       context.Context
		projectId int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.ProjectViewCount
		wantErr bool
	}{
		{
			name: "프로젝트 ID로 조회 1000 => 1건이 조회되기를 기대합니다.",
			args: args{
				ctx:       context.Background(),
				projectId: 1000,
			},
			want: domain.ProjectViewCount{
				Id:        3000,
				ProjectId: 1000,
				ViewCount: 100,
			},
			wantErr: false,
		},
		{
			name: "프로젝트 ID로 조회 1001 => 조회되지 않기를 기대합니다.",
			args: args{
				ctx:       context.Background(),
				projectId: 1001,
			},
			want:    domain.ProjectViewCount{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := projectViewCountMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetByProjectId(tt.args.ctx, tt.args.projectId)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{}, time.Time{})

			assert.Equalf(t, tt.wantErr, err != nil, "GetByProjectId() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "GetByProjectId() = %v, want %v", got, tt.want)
		})
	}
}
