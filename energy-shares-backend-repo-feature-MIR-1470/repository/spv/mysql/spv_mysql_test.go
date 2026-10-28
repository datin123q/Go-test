package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/stretchr/testify/assert"
)

func Test_spvMysqlRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.Spv
		wantErr bool
	}{
		{
			name: "SpvのIDで単件照会 : 1000 => 1件が照会されることが期待されます。",
			args: args{
				ctx: context.Background(),
				id:  1000,
			},
			want: domain.Spv{
				Id:      1000,
				Name:    "水資源公社",
				Email:   "gildong@email.com",
				Phone:   "010-1234-5678",
				OwnerId: 2000,
			},
			wantErr: false,
		},
		{
			name: "SpvのIDで単件照会 : 1001 => 照会されないことが期待されます。",
			args: args{
				ctx: context.Background(),
				id:  1001,
			},
			want:    domain.Spv{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := spvMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetById(tt.args.ctx, tt.args.id)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "spvMysqlRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "spvMysqlRepository.GetById() = %v, want %v", got, tt.want)

		})
	}
}
