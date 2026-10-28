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

func Test_MyNumberRepository_Create(t *testing.T) {
	type args struct {
		ctx      context.Context
		myNumber domain.MyNumber
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) MyNumber 생성",
			args: args{
				ctx:      context.Background(),
				myNumber: mockMyNumberToCreate,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := myNumberRepository{
				dbFactory: dbFactoryMock{},
			}
			_, err := repo.Create(tt.args.ctx, tt.args.myNumber)

			assert.Equalf(t, tt.wantErr, err != nil, "myNumberRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_MyNumberRepository_GetByUserId(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.MyNumber
		wantErr bool
	}{
		{
			name: "MyNumber 조회 (User:2000) -> 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx:    context.Background(),
				userId: mockMyNumber1.UserId,
			},
			want:    mockMyNumber1,
			wantErr: false,
		},
		{
			name: "MyNumber 조회 (User:999) -> 존재하지 않는 User를 조회하므로 에러가 발생합니다.",
			args: args{
				ctx:    context.Background(),
				userId: 999,
			},
			want:    domain.MyNumber{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := myNumberRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := repo.GetByUserId(tt.args.ctx, tt.args.userId)
			assert.Equalf(t, tt.wantErr, err != nil, "myNumberRepository.GetByUserId() error = %v, wantErr %v",
				err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime),
				"myNumberRepository.GetByUserId() got = %v, want %v", got, tt.want)
		})
	}
}

func Test_MyNumberRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.MyNumber
		wantErr bool
	}{
		{
			name: "MyNumber 조회 (Id:1000) -> 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				id:  mockMyNumber1.Id,
			},
			want:    mockMyNumber1,
			wantErr: false,
		},
		{
			name: "MyNumber 조회 (Id:1) -> 존재하지 않으므로 에러가 발생합니다.",
			args: args{
				ctx: context.Background(),
				id:  999,
			},
			want:    domain.MyNumber{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := myNumberRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := repo.GetById(tt.args.ctx, tt.args.id)
			assert.Equalf(t, tt.wantErr, err != nil, "myNumberRepository.GetById() error = %v, wantErr %v",
				err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime),
				"myNumberRepository.GetById() diff = %v",
				cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}

func Test_MyNumberRepository_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "배너 삭제",
			args: args{
				ctx: context.Background(),
				id:  mockMyNumber2.Id,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := myNumberRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Delete(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("MyNumberRepository.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
