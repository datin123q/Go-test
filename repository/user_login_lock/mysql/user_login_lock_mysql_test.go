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

func TestUserLoginLockMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.UserLoginLockQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		want1   []domain.UserLoginLock
		wantErr bool
	}{
		{
			name: "로그인 실패 기록 조회 - 조건 없음",
			args: args{
				ctx:     context.Background(),
				options: []cdomain.UserLoginLockQueryOption{},
			},
			want: 3,
			want1: []domain.UserLoginLock{
				{
					Id:           1,
					UserId:       1001,
					LoginFailCnt: 1,
				},
				{
					Id:           2,
					UserId:       1002,
					LoginFailCnt: 4,
				},
				{
					Id:           3,
					UserId:       1003,
					LoginFailCnt: 7,
				},
			},
			wantErr: false,
		},
		{
			name: "로그인 실패 기록 조회- userId, created",
			args: args{
				ctx: context.Background(),
				options: []cdomain.UserLoginLockQueryOption{
					cdomain.UserLoginLockQueryOptionFunc().WithUserId(1003),
					cdomain.UserLoginLockQueryOptionFunc().WithSortBy("id"),
					cdomain.UserLoginLockQueryOptionFunc().WithOrder("desc"),
					cdomain.UserLoginLockQueryOptionFunc().WithLimit(1),
					cdomain.UserLoginLockQueryOptionFunc().WithOffset(0),
				},
			},
			want:    1,
			want1:   []domain.UserLoginLock{mockUserLoginLock3},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userLoginLockMysqlRepository{
				dbFactory: &dbFactoryMock{},
			}
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, got1, err := repo.GetAll(tt.args.ctx, tt.args.options...)
			assert.Equalf(t, tt.wantErr, err != nil, "userLoginLockMysqlRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, int(tt.want), int(got), "userLoginLockMysqlRepository.GetAll() = %v, want %v", got, tt.want)
			assert.Equalf(t, true, cmp.Equal(tt.want1, got1, ignoreTime), "userLoginLockMysqlRepository.GetAll() = %v, want %v", got1, tt.want1)
		})
	}
}

// mir1233DbFactoryMockはMIR-1233の検証専用。dbFactoryMock(共有の固定3件)を
// 使うと、既存テストケース("조건 없음"、want=3件)の件数が変わって壊れるため、
// 独立したuser_idのデータを別のsqliteインスタンスに用意する。
type mir1233DbFactoryMock struct{}

func (d mir1233DbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}
	xormEngine.Sync(new(domain.UserLoginLock))
	xormEngine.ShowSQL(true)

	// user_id=3001: 過去にロックされた行(id=100)がログイン成功で論理削除され、
	// 同じ日にまた1回失敗して新しい行(id=101、未削除)ができた状態を再現する。
	// xormのInsertはDeletedフィールドを明示的にセットしても無視してNULLで
	// 挿入するため、挿入後にDeleteを呼んで論理削除する。
	xormEngine.Insert(&domain.UserLoginLock{
		Id:           100,
		UserId:       3001,
		LoginFailCnt: 7,
		LockAt:       jsonutils.JSONTime(baseTime),
	})
	xormEngine.ID(int64(100)).Delete(&domain.UserLoginLock{})
	xormEngine.Insert(&domain.UserLoginLock{
		Id:           101,
		UserId:       3001,
		LoginFailCnt: 2,
	})

	// user_id=3002: 未削除の行が2つ同時に存在する状態(将来の同時アクセスを
	// 想定)。idの大きい方(200→201)が新しい行であることを固定する。
	xormEngine.Insert(&domain.UserLoginLock{
		Id:           200,
		UserId:       3002,
		LoginFailCnt: 3,
	})
	xormEngine.Insert(&domain.UserLoginLock{
		Id:           201,
		UserId:       3002,
		LoginFailCnt: 5,
	})

	return xormEngine, nil
}

func (d mir1233DbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}

// MIR-1233: WithSortBy("id")・WithOrder("desc")で「今日の最新の失敗記録」を
// 1件取得できることを、実際に起きうる2つの状況で確認する。
func TestUserLoginLockMysqlRepository_GetAll_LatestRecord(t *testing.T) {
	tests := []struct {
		name       string
		userId     int64
		wantTotal  int64
		wantFailId int64
		wantCnt    int64
	}{
		{
			// 論理削除済みの行(fail_cnt=7)が混ざっていても、xormの標準フィルタで
			// 除外され、未削除の最新行(fail_cnt=2)が返ることを確認する。totalも
			// 未削除の1件だけがカウントされる。
			name:       "(成功) 論理削除済みの古い行を無視し、未削除の最新行を返す",
			userId:     3001,
			wantTotal:  1,
			wantFailId: 101,
			wantCnt:    2,
		},
		{
			// 未削除の行が2つ同時に存在する場合(totalは2件)、idの大きい方
			// (新しい方)をLIMIT 1で返す。WithSortBy/WithOrderが入れ替わっていた
			// 修正前は、ORDER BYが一切積まれずLIMIT 1で不定の行が返っていた
			// (MIR-1233)。
			name:       "(成功) 未削除行が2件あるとき、idが大きい新しい方を返す",
			userId:     3002,
			wantTotal:  2,
			wantFailId: 201,
			wantCnt:    5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userLoginLockMysqlRepository{
				dbFactory: &mir1233DbFactoryMock{},
			}

			total, got, err := repo.GetAll(context.Background(),
				cdomain.UserLoginLockQueryOptionFunc().WithUserId(tt.userId),
				cdomain.UserLoginLockQueryOptionFunc().WithSortBy("id"),
				cdomain.UserLoginLockQueryOptionFunc().WithOrder("desc"),
				cdomain.UserLoginLockQueryOptionFunc().WithLimit(1),
				cdomain.UserLoginLockQueryOptionFunc().WithOffset(0),
			)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantTotal, total)
			assert.Len(t, got, 1)
			assert.Equal(t, tt.wantFailId, got[0].Id)
			assert.Equal(t, tt.wantCnt, got[0].LoginFailCnt)
		})
	}
}

func TestUserLoginLockMysqlRepository_GetByUserId(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		want1   domain.UserLoginLock
		wantErr bool
	}{
		{
			name: "get by user id",
			args: args{
				ctx:    context.Background(),
				userId: 1001,
			},
			want: 1,
			want1: domain.UserLoginLock{

				Id:           1,
				UserId:       1001,
				LoginFailCnt: 1,
				Created:      jsonutils.JSONTime(baseTime),
				Updated:      jsonutils.JSONTime(baseTime.Add(5 * time.Minute)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userLoginLockMysqlRepository{
				dbFactory: &dbFactoryMock{},
			}
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			ignoreFields := cmpopts.IgnoreFields(domain.UserLoginLock{}, "Created", "Updated")

			got, err := repo.GetByUserId(tt.args.ctx, tt.args.userId)
			assert.Equalf(t, tt.wantErr, err != nil, "userLoginLockMysqlRepository.GetByUserId() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want1, got, ignoreTime, ignoreFields), "userLoginLockMysqlRepository.GetByUserId() = %v, want %v", got, tt.want1)
		})
	}
}

func TestUserLoginLockMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx       context.Context
		loginLock domain.UserLoginLock
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "create new login lock",
			args: args{
				ctx: context.Background(),
				loginLock: domain.UserLoginLock{
					UserId:       1004,
					LoginFailCnt: 0,
				},
			},
			want:    4,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userLoginLockMysqlRepository{
				dbFactory: &dbFactoryMock{},
			}

			got, err := repo.Create(tt.args.ctx, tt.args.loginLock)
			assert.Equalf(t, tt.wantErr, err != nil, "userLoginLockMysqlRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "userLoginLockMysqlRepository.Create() = %v, want %v", got, tt.want)
		})
	}
}

func TestUserLoginLockMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx       context.Context
		loginLock domain.UserLoginLock
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "update login lock",
			args: args{
				ctx: context.Background(),
				loginLock: domain.UserLoginLock{
					Id:           2,
					UserId:       1002,
					LoginFailCnt: 7,
					LockAt:       jsonutils.JSONTime(time.Now()),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userLoginLockMysqlRepository{
				dbFactory: &dbFactoryMock{},
			}

			err := repo.Update(tt.args.ctx, tt.args.loginLock)
			assert.Equalf(t, tt.wantErr, err != nil, "userLoginLockMysqlRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestUserLoginLockMysqlRepository_Delete(t *testing.T) {
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
			name: "delete login lock",
			args: args{
				ctx: context.Background(),
				id:  3,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userLoginLockMysqlRepository{
				dbFactory: &dbFactoryMock{},
			}

			err := repo.Delete(tt.args.ctx, tt.args.id)
			assert.Equalf(t, tt.wantErr, err != nil, "userLoginLockMysqlRepository.Delete() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
