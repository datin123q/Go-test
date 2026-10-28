package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_ekycRequestLogRepository_Create(t *testing.T) {
	type args struct {
		ctx            context.Context
		ekycRequestLog domain.EkycRequestLog
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공)EkycRequestLog 추가 -> 정상적으로 insert 되기를 기대합니다.",
			args: args{
				ctx:            context.Background(),
				ekycRequestLog: mockEkycRequestLogToInsert,
			},
			wantErr: false,
		},
		{
			name: "(실패)EkycRequestLog 추가 -> id 중복으로 실패하기를 기대합니다.",
			args: args{
				ctx:            context.Background(),
				ekycRequestLog: mockEkycRequestLog1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ekycRequestLogRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Create(tt.args.ctx, tt.args.ekycRequestLog)
			assert.Equalf(t, tt.wantErr, err != nil, "ekycRequestLogRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_ekycRequestLogRepository_Update(t *testing.T) {
	type args struct {
		ctx            context.Context
		ekycRequestLog domain.EkycRequestLog
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공)EkycRequestLog 업데이트 -> 정상적으로 update 되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				ekycRequestLog: domain.EkycRequestLog{
					Id:           1000,
					Email:        "gildong@gmail.com",
					UserId:       1,
					EkycId:       "EKYC-1110",
					ApplicantId:  1000,
					Status:       "SUCCESS",
					Result:       "MATCH",
					ErrorMessage: nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ekycRequestLogRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Update(tt.args.ctx, tt.args.ekycRequestLog)
			assert.Equalf(t, tt.wantErr, err != nil, "userFileRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_ekycRequestLogRepository_GetByUserId(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.EkycRequestLog
		wantErr bool
	}{
		{
			name: "EkycRequestLog 모델 조회 (UserId:1) -> 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx:    context.Background(),
				userId: mockEkycRequestLog1.UserId,
			},
			want:    mockEkycRequestLog2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ekycRequestLogRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			// ignoreFields := cmpopts.IgnoreFields(domain.UserFile{}, "Id")

			got, err := repo.GetByUserId(tt.args.ctx, tt.args.userId)
			assert.Equalf(t, tt.wantErr, err != nil, "GetByUserId() error = %v, wantErr %v",
				err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "GetByUserId() diff = %v",
				cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}

func Test_ekycRequestLogRepository_GetByEkycId(t *testing.T) {
	type args struct {
		ctx    context.Context
		ekycId string
	}
	tests := []struct {
		name    string
		args    args
		want    domain.EkycRequestLog
		wantErr bool
	}{
		{
			name: "EkycRequestLog 모델 조회 (RequestId:EKYC-1110) -> 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx:    context.Background(),
				ekycId: mockEkycRequestLog1.EkycId,
			},
			want:    mockEkycRequestLog1,
			wantErr: false,
		},
		{
			name: "EkycRequestLog 모델 조회 (RequestId:EKYC-1111) -> 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx:    context.Background(),
				ekycId: mockEkycRequestLog2.EkycId,
			},
			want:    mockEkycRequestLog2,
			wantErr: false,
		},
		{
			name: "EkycRequestLog 모델 조회 (RequestId:EKYC-1111) -> 조회 실패를 기대합니다.",
			args: args{
				ctx:    context.Background(),
				ekycId: mockEkycRequestLogToInsert.EkycId,
			},
			want:    domain.EkycRequestLog{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ekycRequestLogRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			// ignoreFields := cmpopts.IgnoreFields(domain.UserFile{}, "Id")

			got, err := repo.GetByEkycId(tt.args.ctx, tt.args.ekycId)
			assert.Equalf(t, tt.wantErr, err != nil, "mockEkycRequestLog2.GetByEkycId() error = %v, wantErr %v",
				err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "mockEkycRequestLog2.GetByEkycId() diff = %v",
				cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}
