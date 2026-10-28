package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func TestUserMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx  context.Context
		user domain.User
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) 유저 생성",
			args: args{
				ctx: context.Background(),
				user: domain.User{
					Uid:          "UID100",
					IsNewsletter: false,
				},
			},
			wantErr: false,
		},
		{
			name: "(실패) 유저 생성 - 이미 존재하는 UID",
			args: args{
				ctx: context.Background(),
				user: domain.User{
					Uid:          "UID1",
					IsNewsletter: false,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			_, err := repo.Create(tt.args.ctx, tt.args.user)

			assert.Equalf(t, tt.wantErr, err != nil, "userMysqlRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx  context.Context
		user domain.User
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) 유저 정보 업데이트",
			args: args{
				ctx: context.Background(),
				user: domain.User{
					Id:           1000,
					Uid:          "UID100",
					IsNewsletter: false,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			err := repo.Update(tt.args.ctx, tt.args.user)

			assert.Equalf(t, tt.wantErr, err != nil, "userMysqlRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

// singleEngineDBFactory はdbFactoryMockと違い、呼び出すたびに新しいin-memory DBを
// 作らない。read-after-writeを1つのDB上で確かめるテスト専用。
type singleEngineDBFactory struct {
	engine *xorm.Engine
}

func (f singleEngineDBFactory) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	return f.engine, nil
}

func (f singleEngineDBFactory) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}

// TestUserMysqlRepository_UpdateLastLoginDate はMIR-1373の検証。
// Update(全列書き戻し)ではなく、last_login_date だけを書き換えることを確かめる。
// IsEkyc/AuditStatusに、ログイン開始時点の読み込みより後にeKYC Webhookが書いた
// 値を入れた状態でUpdateLastLoginDateを呼び、呼び出し後もその値が消えないことを見る。
func TestUserMysqlRepository_UpdateLastLoginDate(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { engine.Close() })

	assert.NoError(t, engine.Sync(new(domain.User)))
	_, err = engine.Insert(&domain.User{
		Id:    1000,
		Uid:   "UID1",
		Email: "gildong@email.com",
	})
	assert.NoError(t, err)

	repo := userMysqlRepository{
		dbFactory: singleEngineDBFactory{engine: engine},
	}
	ctx := context.Background()

	// eKYC Webhookが先に書いた状態を模す。ログインが読み込んだ時点ではこの値は
	// 反映されておらず、ログインの更新が走った後もこの値が残っていなければならない。
	_, err = engine.ID(int64(1000)).Cols("is_ekyc", "audit_status").Update(&domain.User{
		IsEkyc:      true,
		AuditStatus: domain.WAIT_FINAL_APPROVAL_AUDIT,
	})
	assert.NoError(t, err)

	newLastLoginDate := jsonutils.JSONTime(time.Now())
	err = repo.UpdateLastLoginDate(ctx, 1000, newLastLoginDate)
	assert.NoError(t, err)

	got, err := repo.GetById(ctx, 1000)
	assert.NoError(t, err)

	assert.True(t, got.IsEkyc, "UpdateLastLoginDateがis_ekycを書き戻してしまっている")
	assert.Equal(t, domain.WAIT_FINAL_APPROVAL_AUDIT, got.AuditStatus, "UpdateLastLoginDateがaudit_statusを書き戻してしまっている")
	assert.Equal(t, newLastLoginDate.Time().Unix(), got.LastLoginDate.Time().Unix(), "last_login_dateが更新されていない")
}

func Test_userMysqlRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.User
		wantErr bool
	}{
		{
			name: "(성공) 특정 유저 조회 {Id: 1000} => 결과 1건 있음, Roles는 2개",
			args: args{
				ctx: context.Background(),
				id:  1000,
			},
			want: domain.User{
				Id:           1000,
				Uid:          "UID1",
				CommonId:     "410d3741-7180-4024-a969-648723a57bbe",
				Email:        "gildong@email.com",
				IsNewsletter: true,
				IsSmsCert:    true,
				Roles: []string{
					string(domain.ROLE_USER),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetById(tt.args.ctx, tt.args.id)
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "userMysqlRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "userMysqlRepository.GetById() = %v, want %v", got, tt.want)
		})
	}
}

func Test_userMysqlRepository_GetByEmail(t *testing.T) {
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name    string
		args    args
		want    domain.User
		wantErr bool
	}{
		{
			name: "(성공)특정 유저를 조회 {Email: gildong@email.com} => 결과 1건 있음",
			args: args{
				ctx:   context.Background(),
				email: "gildong@email.com",
			},
			want: domain.User{
				Id:           1000,
				Uid:          "UID1",
				CommonId:     "410d3741-7180-4024-a969-648723a57bbe",
				Email:        "gildong@email.com",
				IsNewsletter: true,
				IsSmsCert:    true,
			},
			wantErr: false,
		},
		{
			name: "(성공) 특정 유저를 조회 {Email: dool2@email.com} => 결과 0건 있음",
			args: args{
				ctx:   context.Background(),
				email: "dool2@email.com",
			},
			want:    domain.User{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetByEmail(tt.args.ctx, tt.args.email)
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "userMysqlRepository.GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "userMysqlRepository.GetByEmail() = %v, want %v", got, tt.want)
		})
	}
}

func Test_userMysqlRepository_GetByEmailIncludeDeleted(t *testing.T) {
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name    string
		args    args
		want    domain.User
		wantErr bool
	}{
		{
			name: "(성공) 삭제된 특정 유저를 조회 {Email: minsu@email.com} => 결과 1건 있음",
			args: args{
				ctx:   context.Background(),
				email: "minsu@email.com",
			},
			want: domain.User{
				Id:           3000,
				Uid:          "UID3",
				CommonId:     "410d3741-4321-1234-a969-648723a57bba",
				Email:        "minsu@email.com",
				IsNewsletter: false,
				IsSmsCert:    false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetByEmailIncludeDeleted(tt.args.ctx, tt.args.email)
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "userMysqlRepository.GetByEmailIncludeDeleted() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "userMysqlRepository.GetByEmailIncludeDeleted() = %v, want %v", got, tt.want)
		})
	}
}

func Test_userMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.UserQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    []domain.User
		wantErr bool
	}{
		{
			name: "(성공) 전체 유저 조회 => 결과 2건 있음",
			args: args{
				ctx: context.Background(),
				options: []cdomain.UserQueryOption{
					cdomain.UserQueryOptionFunc().WithIds([]int64{1000, 2000}),
				},
			},
			want: []domain.User{
				{
					Id:           1000,
					Uid:          "UID1",
					CommonId:     "410d3741-7180-4024-a969-648723a57bbe",
					Email:        "gildong@email.com",
					IsNewsletter: true,
					IsSmsCert:    true,
					Roles: []string{
						string(domain.ROLE_USER),
					},
				},
				{
					Id:           2000,
					Uid:          "UID2",
					CommonId:     "410d3741-4321-1234-a969-648723a57bbe",
					Email:        "chulsu@email.com",
					IsNewsletter: false,
					IsSmsCert:    false,
					Roles: []string{
						string(domain.ROLE_USER),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "(성공) 전체 유저 조건으로 조회 {Ids: [1000]} => 결과 1건 있음",
			args: args{
				ctx: context.Background(),
				options: []cdomain.UserQueryOption{
					cdomain.UserQueryOptionFunc().WithIds([]int64{1000}),
				},
			},
			want: []domain.User{
				{
					Id:           1000,
					Uid:          "UID1",
					CommonId:     "410d3741-7180-4024-a969-648723a57bbe",
					Email:        "gildong@email.com",
					IsNewsletter: true,
					IsSmsCert:    true,
					Roles: []string{
						string(domain.ROLE_USER),
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			count, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "userMysqlRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, int64(len(tt.want)), count, "userMysqlRepository.GetAll() count = %v, want %v", count, len(tt.want))
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "userMysqlRepository.GetAll() = %v, want %v", got, tt.want)
		})
	}
}
