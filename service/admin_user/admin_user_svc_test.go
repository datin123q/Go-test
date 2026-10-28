package admin_user

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"

	"github.com/stretchr/testify/mock"
)

var adminUserSvcTestJST = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	return loc
}()

func Test_adminUserService_ExportCSV(t *testing.T) {
	type args struct {
		ctx   context.Context
		items []domain.AdminUser
	}
	tests := []struct {
		name       string
		args       args
		translator i18nutils.Translator
		want       cdomain.CSVData
	}{
		{
			name: "(성공) CSV 데이터 생성",
			args: args{
				items: []domain.AdminUser{
					mockSuperAdmin,
				},
			},
			translator: i18nutils.New(pathutils.GetPathOfAssets()),
			want: [][]string{
				{"運用者名", "メールアドレス", "電話番号", "最終アクセス日時", "修正日時"},
				{"Admin", "admin@energyshares.jp", "012345678", "2024-02-01 09:00:00", ""},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := adminUserService{
				translator: tt.translator,
			}
			if got := a.ExportCSV(tt.args.ctx, tt.args.items); !cmp.Equal(got, tt.want) {
				t.Errorf("ExportCSV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_adminUserService_GetSuperAdmin(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name                string
		args                args
		adminUserRepository func() *mocks.AdminUserRepository
		want                domain.AdminUser
		wantErr             bool
	}{
		{
			name: "(성공) 슈퍼 어드민 조회",
			args: args{
				ctx: context.Background(),
			},
			adminUserRepository: func() *mocks.AdminUserRepository {
				repo := &mocks.AdminUserRepository{}
				repo.On("GetAll", mock.Anything,
					mock.MatchedBy(func(field func(*cdomain.AdminUserQueryOptionFields)) bool {
						opt := cdomain.AdminUserQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 1
					}), // Limit
					mock.MatchedBy(func(field func(*cdomain.AdminUserQueryOptionFields)) bool {
						opt := cdomain.AdminUserQueryOptionFunc()
						field(&opt)
						return opt.RoleCode() == string(domain.ROLE_SUPER_ADMIN)
					}), // RoleCode
				).Return(int64(1), []domain.AdminUser{mockSuperAdmin}, nil)
				return repo
			},
			want:    mockSuperAdmin,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := adminUserService{
				adminUserRepository: tt.adminUserRepository(),
			}
			got, err := a.GetSuperAdmin(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSuperAdmin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			ignoreType := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			if !cmp.Equal(got, tt.want, ignoreType) {
				t.Errorf("GetSuperAdmin() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_adminUserService_IsValidPeriod(t *testing.T) {
	type args struct {
		ctx       context.Context
		adminUser domain.AdminUser
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "(성공) 일반 어드민인 경우",
			args: args{
				ctx: context.Background(),
				adminUser: domain.AdminUser{
					IsSuperAdmin: false,
					ValidityStartDate: func() jsonutils.JSONTime {
						validityStartDate, _ := time.ParseInLocation(time.DateTime, "2024-01-01 00:00:00", adminUserSvcTestJST)
						return jsonutils.JSONTime(validityStartDate)
					}(),
					ValidityEndDate: func() jsonutils.JSONTime {
						validityStartDate, _ := time.ParseInLocation(time.DateTime, "2100-12-01 23:59:59", adminUserSvcTestJST)
						return jsonutils.JSONTime(validityStartDate)
					}(),
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "(성공) 슈퍼어드민인 경우",
			args: args{
				ctx: context.Background(),
				adminUser: domain.AdminUser{
					IsSuperAdmin: true,
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "(실패) 날짜가 지난 경우",
			args: args{
				ctx: context.Background(),
				adminUser: domain.AdminUser{
					IsSuperAdmin: false,
					ValidityStartDate: func() jsonutils.JSONTime {
						validityStartDate, _ := time.ParseInLocation(time.DateTime, "2023-01-01 00:00:00", adminUserSvcTestJST)
						return jsonutils.JSONTime(validityStartDate)
					}(),
					ValidityEndDate: func() jsonutils.JSONTime {
						validityStartDate, _ := time.ParseInLocation(time.DateTime, "2023-12-01 23:59:59", adminUserSvcTestJST)
						return jsonutils.JSONTime(validityStartDate)
					}(),
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := adminUserService{}
			got, err := a.IsValidPeriod(tt.args.ctx, tt.args.adminUser)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidPeriod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsValidPeriod() got = %v, want %v", got, tt.want)
			}
		})
	}
}
