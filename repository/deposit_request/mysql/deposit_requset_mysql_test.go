package mysql

import (
	"context"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/stretchr/testify/assert"
)

func Test_depositRequestMysqlRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.DepositRequest
		wantErr bool
	}{
		{
			name: "DepositRequest를 ID로 단건 조회합니다 : 1000 => 조회되는 객체가 있어야 합니다.",
			args: args{
				ctx: context.Background(),
				id:  1000,
			},
			want:    mockDepositRequest1,
			wantErr: false,
		},
		{
			name: "DepositRequest를 ID로 단건 조회합니다 : 1001 => 객체가 없어서 에러를 반환합니다.",
			args: args{
				ctx: context.Background(),
				id:  1001,
			},
			want:    domain.DepositRequest{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := depositRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := d.GetById(tt.args.ctx, tt.args.id)
			assert.Equalf(t, tt.wantErr, err != nil, "depositRequestMysqlRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "depositRequestMysqlRepository.GetById() = %v, want %v", got, tt.want)
		})
	}
}

func Test_depositRequestMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx            context.Context
		depositRequest domain.DepositRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "DepositRequest 모델에서 DepositCheckStatus를 변경합니다 : domain.OK_DEPOSIT => 정상적으로 변경되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				depositRequest: domain.DepositRequest{
					Id:                 1000,
					DepositCheckStatus: domain.OK_DEPOSIT,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := depositRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := d.Update(tt.args.ctx, tt.args.depositRequest)
			assert.Equalf(t, tt.wantErr, err != nil, "depositRequestMysqlRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_depositRequestMysqlRepository_GetDepositRequestByShareId(t *testing.T) {
	type args struct {
		ctx     context.Context
		shareId string
	}
	tests := []struct {
		name    string
		args    args
		want    domain.DepositRequest
		wantErr bool
	}{
		{
			name: "SHARE_ID1 값으로 조회했을 때 DepositRequest 객체가 리턴이됩니다.",
			args: args{
				ctx:     context.Background(),
				shareId: "SHARE_ID1",
			},
			want:    mockDepositRequest1,
			wantErr: false,
		},
		{
			name: "SHARE_ID2 값으로 조회했을 때 조회되는 객체가 없어서 에러를 반환합니다.",
			args: args{
				ctx:     context.Background(),
				shareId: "SHARE_ID2",
			},
			want:    domain.DepositRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := depositRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := d.GetDepositRequestByShareId(tt.args.ctx, tt.args.shareId)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "depositRequestMysqlRepository.GetDepositRequestByShareId() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "depositRequestMysqlRepository.GetDepositRequestByShareId() = %v, want %v", got, tt.want)
		})
	}
}

func Test_depositRequestMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.DepositRequestQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    []domain.DepositRequest
		wantErr bool
	}{
		{
			name: "DepositRequest 모델 조회 (User:2000) => 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.DepositRequestQueryOption{
					cdomain.DepositRequestQueryOptionFunc().WithUserId(2000),
				},
			},
			want:    []domain.DepositRequest{mockDepositRequest1},
			wantErr: false,
		},
		{
			name: "DepositRequest 모델 조회 (User:2000, 입금날짜가 만료되었는지 확인_현재시간기준) => 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.DepositRequestQueryOption{
					cdomain.DepositRequestQueryOptionFunc().WithUserId(2000),
					cdomain.DepositRequestQueryOptionFunc().WithExpireDepositDate(jsonutils.JSONTime(time.Now())),
				},
			},
			want:    []domain.DepositRequest{mockDepositRequest1},
			wantErr: false,
		},
		{
			name: "DepositRequest 모델 조회 (User:2000, 입금날짜가 만료되었는지 확인_입금만료일보다 하루 더 지남) => 조회되는 것이 없습니다",
			args: args{
				ctx: context.Background(),
				options: []cdomain.DepositRequestQueryOption{
					cdomain.DepositRequestQueryOptionFunc().WithUserId(2000),
					cdomain.DepositRequestQueryOptionFunc().WithExpireDepositDate(jsonutils.JSONTime(time.Now().AddDate(0, 0, 2))),
				},
			},
			want:    []domain.DepositRequest{},
			wantErr: false,
		},
		{
			name: "DepositRequest 모델 조회 (User: 2001) => 조회가 아무것도 되지 않아야 합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.DepositRequestQueryOption{
					cdomain.DepositRequestQueryOptionFunc().WithUserId(2001),
				},
			},
			want:    []domain.DepositRequest{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := depositRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := d.GetAll(tt.args.ctx, tt.args.options...)
			assert.Equalf(t, tt.wantErr, err != nil, "depositRequestMysqlRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "depositRequestMysqlRepository.GetAll() got = %v, want %v", got, tt.want)
		})
	}
}
