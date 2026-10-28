package mysql

import (
	"context"
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/stretchr/testify/assert"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_projectNewsMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.ProjectQueryOption
	}

	tests := []struct {
		name    string
		args    args
		want    []domain.ProjectNews
		wantErr bool
	}{
		{
			name: "ProjectNews 전체조회 => 전체 조회되기를 기대합니다.",
			args: args{
				ctx:     context.Background(),
				options: []cdomain.ProjectQueryOption{},
			},
			want: []domain.ProjectNews{
				mockProjectNews1, mockProjectNews2, mockProjectNews3,
			},
		},
		{
			name: "ProjectNews 프로젝트 ID로 필터링 조회 : projectId : 1000 => 2건이 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.ProjectQueryOption{
					cdomain.ProjectQueryOptionFunc().WithProjectId(1000),
				},
			},
			want: []domain.ProjectNews{
				mockProjectNews1, mockProjectNews2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := projectNewsMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "projectNewsMysqlRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "projectNewsMysqlRepository.GetAll() = %v, want %v", got, tt.want)
		})
	}
}
