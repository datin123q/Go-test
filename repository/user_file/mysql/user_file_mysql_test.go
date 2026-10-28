package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_userFileRepositoryRepository_Create(t *testing.T) {
	type args struct {
		ctx      context.Context
		userFile domain.UserFile
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공)UserFile 추가 -> 정상적으로 insert 되기를 기대합니다.",
			args: args{
				ctx:      context.Background(),
				userFile: mockUserFileCreate,
			},
			wantErr: false,
		},
		{
			name: "(실패)UserFile 추가 -> FineName 중복으로 실패를 기대합니다.",
			args: args{
				ctx:      context.Background(),
				userFile: mockUserFile1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userFileRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Create(tt.args.ctx, tt.args.userFile)
			assert.Equalf(t, tt.wantErr, err != nil, "userFileRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userFileRepositoryRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.UserFileQueryOption
	}
	tests := []struct {
		name      string
		args      args
		wantTotal int64
		want      domain.UserFiles
		wantErr   bool
	}{
		{
			name: "UserFile 모델 조회 (UserId:2000) -> 정상적으로 2개가 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.UserFileQueryOption{
					cdomain.UserFileQueryOptionFunc().WithUserId(2000),
				},
			},
			wantTotal: int64(2),
			want:      mockUserFiles1,
			wantErr:   false,
		},
		{
			name: "UserFile 모델 조회 (Id:1001,UserId:2000) -> 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.UserFileQueryOption{
					cdomain.UserFileQueryOptionFunc().WithId(1001),
					cdomain.UserFileQueryOptionFunc().WithUserId(2000),
				},
			},
			wantTotal: int64(1),
			want:      mockUserFiles2,
			wantErr:   false,
		},
		{
			name: "UserFile 모델 조회 (Id:1001,UserId:2002) -> 실패를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.UserFileQueryOption{
					cdomain.UserFileQueryOptionFunc().WithId(1001),
					cdomain.UserFileQueryOptionFunc().WithUserId(2002),
					cdomain.UserFileQueryOptionFunc().WithFileType(domain.IDENTIFICATION1_BACK),
				},
			},
			wantTotal: int64(0),
			want:      nil,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userFileRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			// ignoreFields := cmpopts.IgnoreFields(domain.UserFile{}, "Id")

			total, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)
			assert.Equalf(t, tt.wantTotal, total, "userFileRepository.GetAll() total = %v, wantTotal %v", total, tt.wantTotal)
			assert.Equalf(t, tt.wantErr, err != nil, "userFileRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "userFileRepository.GetAll() diff = %v", cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}

func Test_userFileRepository_ExistsByFileName(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     bool
	}{
		{
			name:     "(成功)登録済みの file_name は存在すると判定する",
			fileName: mockUserFile1.FileName,
			want:     true,
		},
		{
			name: "(成功)論理削除済みの file_name も存在すると判定する",
			// unique 制約は deleted を見ないため、削除済みでも同名は作れない。
			// Unscoped を外すとこのケースが false になる。
			fileName: mockUserFileDeleted.FileName,
			want:     true,
		},
		{
			name:     "(成功)未登録の file_name は存在しないと判定する",
			fileName: "not_registered_file_name",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userFileRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.ExistsByFileName(context.Background(), tt.fileName)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_userFileRepository_GetById(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		want    *domain.UserFile
		wantErr error
	}{
		{
			name: "(成功)登録済みのidは取得できる",
			id:   mockUserFile1.Id,
			want: &mockUserFile1,
		},
		{
			// unique 制約は deleted を見ないため ExistsByFileName は拾うが、
			// GetById は通常の取得なので論理削除済みは対象外になる。
			name:    "(失敗)論理削除済みのidは見つからない",
			id:      mockUserFileDeleted.Id,
			wantErr: cdomain.ErrItemNotFound,
		},
		{
			name:    "(失敗)存在しないidは見つからない",
			id:      99999,
			wantErr: cdomain.ErrItemNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userFileRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.GetById(context.Background(), tt.id)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "userFileRepository.GetById() diff = %v", cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}

func Test_userFileRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "(成功)登録済みのidは削除できる",
			id:      mockUserFile2.Id,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userFileRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Delete(context.Background(), tt.id)
			assert.Equalf(t, tt.wantErr, err != nil, "userFileRepository.Delete() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
