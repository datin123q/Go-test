package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var signUpRequestTestJST = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	return loc
}()

func Test_signUpRequestMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx           context.Context
		signUpRequest domain.SignUpRequest
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "회원가입요청 정상생성",
			args: args{
				ctx: context.Background(),
				signUpRequest: domain.SignUpRequest{
					RequestId: "bcdb9d00-fb92-11ee-a951-0242ac120001",
					Email:     "gildong@email.com",
					Url:       "test-url",
					SendCount: 1,
				},
			},
			want:    3,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &signUpRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := s.Create(tt.args.ctx, tt.args.signUpRequest)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "Create() = %v, want %v", got, tt.want)
		})
	}
}

func Test_signUpRequestMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx           context.Context
		signUpRequest domain.SignUpRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "회원가입요청 정상수정",
			args: args{
				ctx: context.Background(),
				signUpRequest: domain.SignUpRequest{
					Id:          1,
					EmailExpire: jsonutils.JSONTime(time.Date(2024, 4, 15, 0, 0, 0, 0, signUpRequestTestJST)),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &signUpRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := s.Update(tt.args.ctx, tt.args.signUpRequest)
			assert.Equalf(t, tt.wantErr, err != nil, "Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_signUpRequestMysqlRepository_GetResourceId(t *testing.T) {
	type args struct {
		ctx       context.Context
		requestId string
	}
	tests := []struct {
		name    string
		args    args
		want    domain.SignUpRequest
		wantErr bool
	}{
		{
			name: "회원가입요청 조회",
			args: args{
				ctx:       context.Background(),
				requestId: "bcdb9d00-fb92-11ee-a951-0242ac120002",
			},
			want: mockSignUpRequest1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &signUpRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := s.GetByRequestId(tt.args.ctx, tt.args.requestId)
			assert.Equalf(t, tt.wantErr, err != nil, "GetByRequestId() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, cmpopts.IgnoreTypes(jsonutils.JSONTime{})), "GetByRequestId() = %v, want %v", got, tt.want)
		})
	}
}

func TestSignUpRequestMysqlRepository_Delete(t *testing.T) {
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
			name: "회원가입요청 삭제",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &signUpRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := s.Delete(tt.args.ctx, tt.args.id)
			assert.Equalf(t, tt.wantErr, err != nil, "Delete() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestSignUpRequestMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.SignUpRequestQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    cdomain.SignUpRequests
		wantErr bool
	}{
		{
			name: "특정 이메일로 필터링",
			args: args{
				ctx: context.Background(),
				options: []cdomain.SignUpRequestQueryOption{
					cdomain.SignUpRequestQueryOptionFunc().WithEmail("gildong@email.com"),
				},
			},
			wantErr: false,
			want: cdomain.SignUpRequests{
				mockSignUpRequest1,
			},
		},
		{
			name: "(url_expire) 특정날짜 (2024/04/13) 보다 이후인지 필터링",
			args: args{
				ctx: context.Background(),
				options: []cdomain.SignUpRequestQueryOption{
					cdomain.SignUpRequestQueryOptionFunc().WithUrlExpire(
						jsonutils.JSONTime(time.Date(2024, 4, 13, 0, 0, 0, 0, signUpRequestTestJST)),
					),
				},
			},
			wantErr: false,
			want: cdomain.SignUpRequests{
				mockSignUpRequest1,
			},
		},
		{
			name: "(email_expire) 특정날짜 (2024/04/13) 보다 이후인지 필터링",
			args: args{
				ctx: context.Background(),
				options: []cdomain.SignUpRequestQueryOption{
					cdomain.SignUpRequestQueryOptionFunc().WithEmailExpire(
						jsonutils.JSONTime(time.Date(2024, 4, 13, 0, 0, 0, 0, signUpRequestTestJST)),
					),
				},
			},
			wantErr: false,
			want: cdomain.SignUpRequests{
				mockSignUpRequest2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &signUpRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := s.GetAll(tt.args.ctx, tt.args.options...)
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)

			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTypes), "GetAll() = %v, want %v", got, tt.want)
		})
	}
}
