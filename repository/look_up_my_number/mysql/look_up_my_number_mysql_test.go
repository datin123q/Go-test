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

func Test_lookUpMyNumberRepository_Create(t *testing.T) {
	type args struct {
		ctx            context.Context
		lookUpMyNumber domain.LookUpMyNumber
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) LookUpMyNumber 생성",
			args: args{
				ctx:            context.Background(),
				lookUpMyNumber: mockLookUpMyNumberCreate1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := lookUpMyNumberRepository{
				dbFactory: dbFactoryMock{},
			}
			_, err := repo.Create(tt.args.ctx, tt.args.lookUpMyNumber)

			assert.Equalf(t, tt.wantErr, err != nil, "lookUpMyNumberRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_lookUpMyNumberRepository_GetByMyNumberId(t *testing.T) {
	type args struct {
		ctx        context.Context
		myNumberId int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.LookUpMyNumber
		wantErr bool
	}{
		{
			name: "LookUpMyNumber(myNumberId:2000)를 가져옵니다. -> 성공해야 합니다.",
			args: args{
				ctx:        context.Background(),
				myNumberId: mockLookUpMyNumber1.MyNumberId,
			},
			want:    mockLookUpMyNumber1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := lookUpMyNumberRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			got, err := repo.GetByMyNumberId(tt.args.ctx, tt.args.myNumberId)
			assert.Equalf(t, tt.wantErr, err != nil, "lookUpMyNumberRepository.GetByMyNumberId() error = %v, wantErr %v", err, tt.wantErr)
			if diff := cmp.Diff(tt.want, got, ignoreTypes); diff != "" {
				t.Errorf("lookUpMyNumberRepository.GetByMyNumberId() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_lookUpMyNumberRepository_GetByMyNumberIdIncludeDeleted(t *testing.T) {
	type args struct {
		ctx        context.Context
		myNumberId int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.LookUpMyNumber
		wantErr bool
	}{
		{
			name: "LookUpMyNumber(myNumberId:2001)를 가져옵니다. 삭제한 것도 가져오도록 합니다. -> 성공해야 합니다.",
			args: args{
				ctx:        context.Background(),
				myNumberId: mockLookUpMyNumberDeleted1.MyNumberId,
			},
			want:    mockLookUpMyNumberDeleted1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := lookUpMyNumberRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			got, err := repo.GetByMyNumberIdIncludeDeleted(tt.args.ctx, tt.args.myNumberId)
			assert.Equalf(t, tt.wantErr, err != nil, "lookUpMyNumberRepository.GetByMyNumberIdIncludeDeleted() error = %v, wantErr %v", err, tt.wantErr)
			if diff := cmp.Diff(tt.want, got, ignoreTypes); diff != "" {
				t.Errorf("lookUpMyNumberRepository.GetByMyNumberIdIncludeDeleted() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_lookUpMyNumberRepository_GetByMyNumber(t *testing.T) {
	type args struct {
		ctx      context.Context
		myNumber string
	}
	tests := []struct {
		name    string
		args    args
		want    domain.LookUpMyNumber
		wantErr bool
	}{
		{
			name: "(성공) MyNumber 로 LookUpMyNumber 를 조회",
			args: args{
				ctx:      context.Background(),
				myNumber: mockLookUpMyNumber1.MyNumber,
			},
			want:    mockLookUpMyNumber1,
			wantErr: false,
		},
		{
			name: "(실패) MyNumber 로 삭제된 LookUpMyNumber 를 조회",
			args: args{
				ctx:      context.Background(),
				myNumber: mockLookUpMyNumberDeleted1.MyNumber,
			},
			want:    domain.LookUpMyNumber{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := lookUpMyNumberRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			got, err := repo.GetByMyNumber(tt.args.ctx, tt.args.myNumber)
			assert.Equalf(t, tt.wantErr, err != nil, "lookUpMyNumberRepository.GetByMyNumber() error = %v, wantErr %v", err, tt.wantErr)
			if diff := cmp.Diff(tt.want, got, ignoreTypes); diff != "" {
				t.Errorf("lookUpMyNumberRepository.GetByMyNumber() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_lookUpMyNumberRepository_GetByMyNumberIncludeDeleted(t *testing.T) {
	type args struct {
		ctx      context.Context
		myNumber string
	}
	tests := []struct {
		name    string
		args    args
		want    domain.LookUpMyNumber
		wantErr bool
	}{
		{
			name: "(성공) MyNumber 로 삭제된 것 포함하여 LookUpMyNumber 를 조회",
			args: args{
				ctx:      context.Background(),
				myNumber: mockLookUpMyNumberDeleted1.MyNumber,
			},
			want:    mockLookUpMyNumberDeleted1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := lookUpMyNumberRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			got, err := repo.GetByMyNumberIncludeDeleted(tt.args.ctx, tt.args.myNumber)
			assert.Equalf(t, tt.wantErr, err != nil, "lookUpMyNumberRepository.GetByMyNumberIncludeDeleted() error = %v, wantErr %v", err, tt.wantErr)
			if diff := cmp.Diff(tt.want, got, ignoreTypes); diff != "" {
				t.Errorf("lookUpMyNumberRepository.GetByMyNumberIncludeDeleted() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_lookUpMyNumberRepository_DeleteByMyNumberId(t *testing.T) {
	type args struct {
		ctx        context.Context
		myNumberId int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) MyNumberId 로 LookUpMyNumber 삭제",
			args: args{
				ctx:        context.Background(),
				myNumberId: mockLookUpMyNumber2.MyNumberId,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := lookUpMyNumberRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.DeleteByMyNumberId(tt.args.ctx, tt.args.myNumberId)
			assert.Equalf(t, tt.wantErr, err != nil, "lookUpMyNumberRepository.DeleteByMyNumberId() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
