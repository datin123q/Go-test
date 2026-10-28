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

func Test_projectFileMysqlRepository_GetByProjectIdAndFileType(t *testing.T) {
	type args struct {
		ctx       context.Context
		projectId int64
		fileType  domain.PROJECT_FILE_TYPE
	}
	tests := []struct {
		name    string
		args    args
		want    domain.ProjectFile
		wantErr bool
	}{
		{
			name: "프로젝트 ID와 파일 타입으로 조회 => 조회되기를 기대합니다.",
			args: args{
				ctx:       context.Background(),
				projectId: 1000,
				fileType:  domain.PROJECT_REPORT,
			},
			want: domain.ProjectFile{
				Id:            3000,
				ProjectId:     1000,
				FileName:      "test_report_13",
				OriginalName:  "test_report.pdf",
				FileType:      domain.PROJECT_REPORT,
				FileSize:      189117,
				FileExtension: "/project_file/project_report/",
				FilePath:      "https://test-cdn.dev.energyshares.jp/project_file/project_report/test_report.pdf",
				OwnerId:       2000,
				UpdateId:      2000,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := projectFileMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{}, time.Time{})
			got, err := repo.GetByProjectIdAndFileType(tt.args.ctx, tt.args.projectId, tt.args.fileType)

			assert.Equalf(t, tt.wantErr, err != nil, "GetByProjectIdAndFileType() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "GetByProjectIdAndFileType() = %v, want %v", got, tt.want)
		})
	}
}
