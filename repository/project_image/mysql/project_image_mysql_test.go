package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_projectImageMysqlRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.ProjectImage
		wantErr bool
	}{
		{
			name: "ProjectImage ID로 단건조회 : 3000 => 한건 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				id:  3000,
			},
			want: domain.ProjectImage{
				Id:            3000,
				ProjectId:     1000,
				FileName:      "test_report_13",
				OriginalName:  "test_report.pdf",
				FileType:      "PROJECT_REPORT",
				FileSize:      189117,
				FileExtension: "/project_file/project_report/",
				FilePath:      "https://test-cdn.dev.energyshares.jp/project_file/project_report/test_report.pdf",
				OwnerId:       2000,
				UpdateId:      2000,
				ImageType:     domain.MAIN_IMG,
			},
		},
		{
			name: "ProjectImage ID로 단건조회 : 3001 => 조회되지 않기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				id:  3001,
			},
			want:    domain.ProjectImage{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := projectImageMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetById(tt.args.ctx, tt.args.id)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{}, time.Time{})
			assert.Equalf(t, tt.wantErr, err != nil, "projectImageMysqlRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "projectImageMysqlRepository.GetById() = %v, want %v", got, tt.want)
		})
	}
}
