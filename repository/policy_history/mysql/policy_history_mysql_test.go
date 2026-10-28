package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_policyHistoryRepository_Create(t *testing.T) {
	type args struct {
		ctx    context.Context
		policy domain.PolicyInfoHistory
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "약관 히스토리 생성",
			args: args{
				ctx: context.Background(),
				policy: domain.PolicyInfoHistory{
					PolicyId:        1,
					SubstanceType:   string(domain.POLICY_HISTORY_CONTENTS),
					Name:            "ADMIN",
					Email:           "admin@energyshares.jp",
					PrevContent:     "prev content",
					CurrContent:     "curr content",
					PrevContentHtml: "prev content html",
					CurrContentHtml: "curr content html",
					PrevUrl:         "prev url",
					CurrUrl:         "curr url",
					ContentType:     string(domain.POLICY_CHANGED_CONTENTS),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := policyHistoryRepository{
				dbFactory: dbFactoryMock{},
			}

			_, err := repo.Create(tt.args.ctx, tt.args.policy)

			assert.Equalf(t, tt.wantErr, err != nil, "PolicyHistoryRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_policyHistoryRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.PolicyHistoryQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    []domain.PolicyInfoHistory
		wantErr bool
	}{
		{
			name: "약관 히스토리 목록 조회 - 조건 없음",
			args: args{
				ctx: context.Background(),
				options: []cdomain.PolicyHistoryQueryOption{
					cdomain.PolicyHistoryQueryOptionFunc().WithSortBy("id"),
					cdomain.PolicyHistoryQueryOptionFunc().WithOrder("asc"),
					cdomain.PolicyHistoryQueryOptionFunc().WithLimit(10),
					cdomain.PolicyHistoryQueryOptionFunc().WithOffset(0),
				},
			},
			want: []domain.PolicyInfoHistory{
				mockPolicyHistory1, mockPolicyHistory2, mockPolicyHistory3,
			},
			wantErr: false,
		},
		{
			name: "약관 히스토리 목록 조회 - policyId로 조회",
			args: args{
				ctx: context.Background(),
				options: []cdomain.PolicyHistoryQueryOption{
					cdomain.PolicyHistoryQueryOptionFunc().WithPolicyId(1),
					cdomain.PolicyHistoryQueryOptionFunc().WithSortBy("id"),
					cdomain.PolicyHistoryQueryOptionFunc().WithOrder("asc"),
					cdomain.PolicyHistoryQueryOptionFunc().WithLimit(10),
					cdomain.PolicyHistoryQueryOptionFunc().WithOffset(0),
				},
			},
			want: []domain.PolicyInfoHistory{
				mockPolicyHistory1, mockPolicyHistory2,
			},
			wantErr: false,
		},
		{
			name: "약관 히스토리 목록 조회 - 다중 조건 - policyId, isPosted",
			args: args{
				ctx: context.Background(),
				options: []cdomain.PolicyHistoryQueryOption{
					cdomain.PolicyHistoryQueryOptionFunc().WithPolicyId(1),
					cdomain.PolicyHistoryQueryOptionFunc().WithIsPosted(lo.ToPtr(true)),
					cdomain.PolicyHistoryQueryOptionFunc().WithSortBy("id"),
					cdomain.PolicyHistoryQueryOptionFunc().WithOrder("desc"),
					cdomain.PolicyHistoryQueryOptionFunc().WithOffset(0),
					cdomain.PolicyHistoryQueryOptionFunc().WithLimit(10),
				},
			},
			want: []domain.PolicyInfoHistory{
				mockPolicyHistory1,
			},
			wantErr: false,
		},
		{
			name: "약관 히스토리 목록 조회 - 다중 조건 - policyId, label",
			args: args{
				ctx: context.Background(),
				options: []cdomain.PolicyHistoryQueryOption{
					cdomain.PolicyHistoryQueryOptionFunc().WithPolicyId(1),
					cdomain.PolicyHistoryQueryOptionFunc().WithLabel(lo.ToPtr("2024-02-09")),
					cdomain.PolicyHistoryQueryOptionFunc().WithSortBy("id"),
					cdomain.PolicyHistoryQueryOptionFunc().WithOrder("desc"),
					cdomain.PolicyHistoryQueryOptionFunc().WithOffset(0),
					cdomain.PolicyHistoryQueryOptionFunc().WithLimit(10),
				},
			},
			want: []domain.PolicyInfoHistory{
				mockPolicyHistory1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := policyHistoryRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			_, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			assert.Equalf(t, tt.wantErr, err != nil, "PolicyHistoryRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTypes), true, "PolicyHistoryRepository.GetAll() = %+v, want %+v", got, tt.want)
		})
	}
}

func Test_policyRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.PolicyInfoHistory
		wantErr bool
	}{
		{
			name: "약관 히스토리 조회",
			args: args{
				ctx: context.Background(),
				id:  mockPolicyHistory1.Id,
			},
			want:    mockPolicyHistory1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := policyHistoryRepository{
				dbFactory: dbFactoryMock{},
			}

			_, err := repo.GetById(tt.args.ctx, tt.args.id)

			assert.Equalf(t, tt.wantErr, err != nil, "PolicyHistoryRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_policyHistoryRepository_Update(t *testing.T) {
	type args struct {
		ctx     context.Context
		history domain.PolicyInfoHistory
	}
	tests := []struct {
		name    string
		args    args
		want    domain.PolicyInfoHistory
		wantErr bool
	}{
		{
			name: "약관 히스토리 수정",
			args: args{
				ctx: context.Background(),
				history: domain.PolicyInfoHistory{
					Id:       1,
					PolicyId: 1,
					IsPosted: false,
					Label:    jsonutils.JSONTime(time.Date(2024, 11, 11, 0, 0, 0, 0, time.UTC)),
				},
			},
			want: domain.PolicyInfoHistory{
				Id:              1,
				PolicyId:        1,
				SubstanceType:   string(domain.POLICY_HISTORY_CONTENTS),
				Name:            "ADMIN",
				Email:           "admin@energyshares.jp",
				IsPosted:        false,
				Label:           jsonutils.JSONTime(time.Date(2024, 11, 11, 0, 0, 0, 0, time.UTC)),
				PrevContent:     "prev content",
				CurrContent:     "curr content",
				PrevContentHtml: "prev content html",
				CurrContentHtml: "curr content html",
				PrevUrl:         "prev url",
				CurrUrl:         "curr url",
				ContentType:     string(domain.POLICY_CHANGED_CONTENTS),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, _ := dbFactoryMock{}.EnergySharesDB(tt.args.ctx)
			ctx := context.WithValue(tt.args.ctx, factory.ConextDBType, session)

			repo := policyHistoryRepository{
				dbFactory: factory.NewDBFactory(),
			}

			err := repo.Update(ctx, tt.args.history)

			assert.Equalf(t, tt.wantErr, err != nil, "PolicyHistoryRepository.Update() error = %v, wantErr %v", err, tt.wantErr)

			ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			result, err := repo.GetById(ctx, tt.args.history.Id)
			assert.Equalf(t, cmp.Equal(tt.want, result, ignoreTypes), true, "PolicyHistoryRepository.GetById() = %+v, want %+v", result, tt.want)
			assert.Equalf(t, tt.want.Label.ToFormat(time.DateOnly), result.Label.ToFormat(time.DateOnly), "PolicyHistoryRepository.GetById() = %+v, want %+v", result.Label.ToFormat(time.DateOnly), tt.want.Label.ToFormat(time.DateOnly))
		})
	}
}

// rawLabelMIR1384 は、labelカラムに実際に保存されている生の文字列を読む。
// CAST(... AS TEXT): sqlite3ドライバがDATETIMEカラムの値をtime.Timeとして
// 読み直し、テスト実行環境のtime.Localで再解釈してしまう(例: 末尾に
// 09:00:00が付く)ことがあるため、生の文字列を素通しさせる。
func rawLabelMIR1384(t *testing.T, ctx context.Context, id int64) string {
	t.Helper()
	db, err := factory.NewDBFactory().EnergySharesDB(ctx)
	require.NoError(t, err)

	rows, err := db.SQL("SELECT CAST(label AS TEXT) AS label FROM policy_info_history WHERE id = ?", id).Query()
	require.NoError(t, err)
	require.Len(t, rows, 1)

	return string(rows[0]["label"])
}

// MIR-1384: 保存される日付が1日前にずれていた理由はParseJSONDateJSTの
// godoc(pkg/jsonutils/json_time.go)を参照。実際にsqlite+xormへ書き込み、
// 生の値を直接読んでそのずれを再現し、専用のParseJSONDateJSTで直っている
// ことを確認する。
func Test_policyHistoryRepository_Update_MIR1384_DateColumn(t *testing.T) {
	tests := []struct {
		name  string
		parse func(layout, value string) (jsonutils.JSONTime, error)
		want  string
	}{
		{
			name:  "再現: ParseJSONTimeJST(旧)で解釈した2026-07-25は、前日の2026-07-24として保存される",
			parse: jsonutils.ParseJSONTimeJST,
			want:  "2026-07-24",
		},
		{
			name:  "修正: ParseJSONDateJST(新)で解釈した2026-07-25は、そのまま2026-07-25として保存される",
			parse: jsonutils.ParseJSONDateJST,
			want:  "2026-07-25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			session, err := dbFactoryMock{}.EnergySharesDB(ctx)
			require.NoError(t, err)
			ctx = context.WithValue(ctx, factory.ConextDBType, session)

			label, err := tt.parse(time.DateOnly, "2026-07-25")
			require.NoError(t, err)

			repo := policyHistoryRepository{dbFactory: factory.NewDBFactory()}
			err = repo.Update(ctx, domain.PolicyInfoHistory{Id: 1, PolicyId: 1, IsPosted: false, Label: label})
			require.NoError(t, err)

			assert.Equal(t, tt.want, rawLabelMIR1384(t, ctx, 1))
		})
	}

	t.Run("修正: 読み出し側(ToFormat)も正しい日付を返す", func(t *testing.T) {
		ctx := context.Background()
		session, err := dbFactoryMock{}.EnergySharesDB(ctx)
		require.NoError(t, err)
		ctx = context.WithValue(ctx, factory.ConextDBType, session)

		label, err := jsonutils.ParseJSONDateJST(time.DateOnly, "2026-07-25")
		require.NoError(t, err)

		repo := policyHistoryRepository{dbFactory: factory.NewDBFactory()}
		require.NoError(t, repo.Update(ctx, domain.PolicyInfoHistory{Id: 1, PolicyId: 1, IsPosted: false, Label: label}))

		got, err := repo.GetById(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, "2026-07-25", got.Label.ToFormat(time.DateOnly))
	})

	t.Run("掲載のオン・オフを繰り返しても日付が動かない", func(t *testing.T) {
		ctx := context.Background()
		session, err := dbFactoryMock{}.EnergySharesDB(ctx)
		require.NoError(t, err)
		ctx = context.WithValue(ctx, factory.ConextDBType, session)

		label, err := jsonutils.ParseJSONDateJST(time.DateOnly, "2026-10-08")
		require.NoError(t, err)

		repo := policyHistoryRepository{dbFactory: factory.NewDBFactory()}

		for i, isPosted := range []bool{true, false, true, false} {
			err = repo.Update(ctx, domain.PolicyInfoHistory{Id: 1, PolicyId: 1, IsPosted: isPosted, Label: label})
			require.NoErrorf(t, err, "更新 %d 回目", i+1)
			assert.Equalf(t, "2026-10-08", rawLabelMIR1384(t, ctx, 1), "更新 %d 回目", i+1)
		}
	})
}
