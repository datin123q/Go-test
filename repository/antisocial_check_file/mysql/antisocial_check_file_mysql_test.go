package mysql

import (
	"context"
	"testing"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_antisocialCheckFileMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.AntisocialCheckFileQueryOption
	}
	tests := []struct {
		name      string
		args      args
		want      domain.AntisocialCheckFiles
		wantTotal int64
		wantErr   bool
	}{
		{
			name: "AntisocialCheckFile 모델 조회 (User:2000) -> 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.AntisocialCheckFileQueryOption{
					cdomain.AntisocialCheckFileQueryOptionFunc().WithCheckId(mockAntisocialCheckFile1.CheckId),
				},
			},
			want:      domain.AntisocialCheckFiles{mockAntisocialCheckFile1},
			wantTotal: int64(1),
			wantErr:   false,
		},
		{
			name: "AntisocialCheckFile 모델 조회 (User:2001) -> 조회가 아무것도 되지 않아야 합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.AntisocialCheckFileQueryOption{
					cdomain.AntisocialCheckFileQueryOptionFunc().WithCheckId(2001),
				},
			},
			want:      domain.AntisocialCheckFiles{},
			wantTotal: int64(0),
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := antisocialCheckFileRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			total, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)
			assert.Equalf(t, tt.wantTotal, total, "antisocialCheckFileRepository.GetAll() total = %v, want %v", total, tt.wantTotal)
			assert.Equalf(t, tt.wantErr, err != nil, "antisocialCheckFileRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "antisocialCheckFileRepository.GetAll() got = %v, want %v", got, tt.want)
		})
	}
}
