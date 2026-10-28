package mysql

import (
	"context"
	"testing"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/stretchr/testify/assert"
)

func Test_lotteryRepository_Create(t *testing.T) {
	type args struct {
		ctx context.Context
		app domain.LotteryApplication
	}
	tests := []struct {
		name         string
		args         args
		wantInserted bool
		wantErr      bool
	}{
		{
			name: "新規作成成功",
			args: args{
				ctx: context.Background(),
				app: domain.LotteryApplication{
					ProjectId: 999, // 新規プロジェクト
					UserId:    100,
					Qty:       3,
					UnitPrice: 1000,
					Amount:    3000,
				},
			},
			wantInserted: true,
			wantErr:      false,
		},
		{
			name: "重複で挿入されない（同一user_id + project_id）",
			args: args{
				ctx: context.Background(),
				app: domain.LotteryApplication{
					ProjectId: 1,   // mockAppAppliedと重複
					UserId:    100, // mockAppAppliedと重複
					Qty:       5,
					UnitPrice: 1000,
					Amount:    5000,
				},
			},
			wantInserted: false,
			wantErr:      false,
		},
		{
			name: "同一ユーザー同一プロジェクトに、キャンセルの抽選がある状態では申し込める",
			args: args{
				ctx: context.Background(),
				app: domain.LotteryApplication{
					ProjectId: 3,   // mockAppCanceledForCreateTest と同じ
					UserId:    103, // mockAppCanceledForCreateTest と同じ
					Qty:       5,
					UnitPrice: 1000,
					Amount:    5000,
				},
			},
			wantInserted: true,
			wantErr:      false,
		},
		{
			name: "同一ユーザー同一プロジェクトに、申込中の抽選がある状態では申し込めない",
			args: args{
				ctx: context.Background(),
				app: domain.LotteryApplication{
					ProjectId: 3,   // mockAppAppliedForCreateTest と同じ
					UserId:    104, // mockAppAppliedForCreateTest と同じ
					Qty:       5,
					UnitPrice: 1000,
					Amount:    5000,
				},
			},
			wantInserted: false,
			wantErr:      false,
		},
		{
			name: "同一ユーザー同一プロジェクトに、キャンセルの抽選と申込中の抽選がある状態では申し込めない",
			args: args{
				ctx: context.Background(),
				app: domain.LotteryApplication{
					ProjectId: 3,   // mockAppCanceledForMixedTest / mockAppAppliedForMixedTest と同じ
					UserId:    105, // mockAppCanceledForMixedTest / mockAppAppliedForMixedTest と同じ
					Qty:       3,
					UnitPrice: 1000,
					Amount:    3000,
				},
			},
			wantInserted: false,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := lotteryApplicationRepository{
				dbFactory: &dbFactoryMock{},
			}

			id, inserted, err := repo.CreateIfNotExists(tt.args.ctx, tt.args.app)

			assert.Equalf(t, tt.wantErr, err != nil, "CreateIfNotExists() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.wantInserted, inserted, "CreateIfNotExists() inserted = %v, want %v", inserted, tt.wantInserted)

			if tt.wantInserted {
				assert.NotEmpty(t, id, "CreateIfNotExists() should return non-empty id")
			} else {
				assert.Empty(t, id, "CreateIfNotExists() should return empty id when not inserted")
			}
		})
	}
}

func Test_lotteryRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.LotteryApplicationQueryOption
	}
	tests := []struct {
		name                  string
		args                  args
		wantCount             int64
		wantIds               []string // 期待されるIDリスト
		wantProjectStatusByID map[string]domain.PROJECT_STATUS
		wantErr               bool
	}{
		{
			name: "フィルターなし - 全件取得",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LotteryApplicationQueryOption{
					cdomain.LotteryApplicationQueryOptionFunc().WithOffset(0),
					cdomain.LotteryApplicationQueryOptionFunc().WithLimit(100),
				},
			},
			wantCount: 10,
			wantIds:   []string{"app-001", "app-002", "app-003", "app-004", "app-005", "app-006", "app-007", "app-008", "app-009", "app-010"},
			wantErr:   false,
		},
		{
			name: "projectIdフィルター",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LotteryApplicationQueryOption{
					cdomain.LotteryApplicationQueryOptionFunc().WithProjectId(1),
				},
			},
			wantCount: 4,
			wantIds:   []string{"app-001", "app-002", "app-003", "app-006"},
			wantProjectStatusByID: map[string]domain.PROJECT_STATUS{
				"app-001": domain.OPEN_FUNDING,
				"app-002": domain.OPEN_FUNDING,
				"app-003": domain.OPEN_FUNDING,
				"app-006": domain.OPEN_FUNDING,
			},
			wantErr: false,
		},
		{
			name: "userIdフィルター",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LotteryApplicationQueryOption{
					cdomain.LotteryApplicationQueryOptionFunc().WithUserId(100),
				},
			},
			wantCount: 3,
			wantIds:   []string{"app-001", "app-004", "app-006"},
			wantErr:   false,
		},
		{
			name: "statusフィルター: APPLIED",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LotteryApplicationQueryOption{
					cdomain.LotteryApplicationQueryOptionFunc().WithStatus(domain.LOTTERY_STATUS_APPLIED),
				},
			},
			wantCount: 4,
			wantIds:   []string{"app-001", "app-006", "app-008", "app-010"},
			wantErr:   false,
		},
		{
			name: "statusフィルター: WINNING（PARTIAL_WINNINGも含む）",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LotteryApplicationQueryOption{
					cdomain.LotteryApplicationQueryOptionFunc().WithStatus(domain.LOTTERY_STATUS_WINNING),
				},
			},
			wantCount: 2,
			wantIds:   []string{"app-002", "app-003"}, // DB上はWINNING/PARTIAL_WINNINGを区別しない
			wantErr:   false,
		},
		{
			name: "statusフィルター: LOST",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LotteryApplicationQueryOption{
					cdomain.LotteryApplicationQueryOptionFunc().WithStatus(domain.LOTTERY_STATUS_LOST),
				},
			},
			wantCount: 1,
			wantIds:   []string{"app-004"},
			wantErr:   false,
		},
		{
			name: "statusフィルター: CANCELED",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LotteryApplicationQueryOption{
					cdomain.LotteryApplicationQueryOptionFunc().WithStatus(domain.LOTTERY_STATUS_CANCELED),
				},
			},
			wantCount: 3,
			wantIds:   []string{"app-005", "app-007", "app-009"},
			wantErr:   false,
		},
		{
			name: "金額範囲フィルター: AmountMin",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LotteryApplicationQueryOption{
					cdomain.LotteryApplicationQueryOptionFunc().WithAmountMin(10000),
				},
			},
			wantCount: 8,
			wantIds:   []string{"app-001", "app-002", "app-003", "app-004", "app-005", "app-007", "app-008", "app-009"},
			wantErr:   false,
		},
		{
			name: "金額範囲フィルター: AmountMax",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LotteryApplicationQueryOption{
					cdomain.LotteryApplicationQueryOptionFunc().WithAmountMax(10000),
				},
			},
			wantCount: 9,
			wantIds:   []string{"app-001", "app-002", "app-003", "app-004", "app-006", "app-007", "app-008", "app-009", "app-010"},
			wantErr:   false,
		},
		{
			name: "金額範囲フィルター: AmountMin and AmountMax",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LotteryApplicationQueryOption{
					cdomain.LotteryApplicationQueryOptionFunc().WithAmountMin(8000),
					cdomain.LotteryApplicationQueryOptionFunc().WithAmountMax(11000),
				},
			},
			wantCount: 7,
			wantIds:   []string{"app-001", "app-002", "app-003", "app-004", "app-007", "app-008", "app-009"},
			wantErr:   false,
		},
		{
			name: "userNameフィルター（LIKE）",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LotteryApplicationQueryOption{
					cdomain.LotteryApplicationQueryOptionFunc().WithUserName("Tanaka"),
				},
			},
			wantCount: 3,
			wantIds:   []string{"app-001", "app-004", "app-006"},
			wantErr:   false,
		},
		{
			name: "projectNameフィルター（LIKE）",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LotteryApplicationQueryOption{
					cdomain.LotteryApplicationQueryOptionFunc().WithProjectName("Alpha"),
				},
			},
			wantCount: 4,
			wantIds:   []string{"app-001", "app-002", "app-003", "app-006"},
			wantErr:   false,
		},
		{
			name: "複合フィルター: projectId + userId + status",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LotteryApplicationQueryOption{
					cdomain.LotteryApplicationQueryOptionFunc().WithProjectId(1),
					cdomain.LotteryApplicationQueryOptionFunc().WithUserId(100),
					cdomain.LotteryApplicationQueryOptionFunc().WithStatus(domain.LOTTERY_STATUS_APPLIED),
				},
			},
			wantCount: 2,
			wantIds:   []string{"app-001", "app-006"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := lotteryApplicationRepository{
				dbFactory: &dbFactoryMock{},
			}

			count, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.wantCount, count, "GetAll() count = %v, want %v", count, tt.wantCount)
			assert.Equalf(t, int(tt.wantCount), len(got), "GetAll() len(items) = %v, want %v", len(got), tt.wantCount)

			// IDリストの確認
			gotIds := make([]string, len(got))
			for i, app := range got {
				gotIds[i] = app.Id
				if tt.wantProjectStatusByID != nil {
					assert.Equalf(t, tt.wantProjectStatusByID[app.Id], app.ProjectStatus, "GetAll() project_status for %s = %v, want %v", app.Id, app.ProjectStatus, tt.wantProjectStatusByID[app.Id])
				}
			}
			assert.ElementsMatchf(t, tt.wantIds, gotIds, "GetAll() ids = %v, want %v", gotIds, tt.wantIds)
		})
	}
}

func Test_lotteryRepository_GetAll_CountMatchesFind(t *testing.T) {
	// Count と Find が同じフィルターを適用しているか確認
	repo := lotteryApplicationRepository{
		dbFactory: &dbFactoryMock{},
	}

	opts := []cdomain.LotteryApplicationQueryOption{
		cdomain.LotteryApplicationQueryOptionFunc().WithProjectId(1),
		cdomain.LotteryApplicationQueryOptionFunc().WithAmountMin(5000),
		cdomain.LotteryApplicationQueryOptionFunc().WithStatus(domain.LOTTERY_STATUS_APPLIED),
	}

	count, items, err := repo.GetAll(context.Background(), opts...)

	assert.NoError(t, err)
	assert.Equal(t, count, int64(len(items)), "Count should match len(items)")
}

func Test_lotteryRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name              string
		args              args
		wantStatus        domain.LOTTERY_STATUS
		wantProjectStatus domain.PROJECT_STATUS
		wantErr           bool
	}{
		{
			name: "APPLIED状態の取得",
			args: args{
				ctx: context.Background(),
				id:  "app-001",
			},
			wantStatus:        domain.LOTTERY_STATUS_APPLIED,
			wantProjectStatus: domain.OPEN_FUNDING,
			wantErr:           false,
		},
		{
			name: "WINNING状態の取得（全口当選）",
			args: args{
				ctx: context.Background(),
				id:  "app-002",
			},
			wantStatus:        domain.LOTTERY_STATUS_WINNING,
			wantProjectStatus: domain.OPEN_FUNDING,
			wantErr:           false,
		},
		{
			name: "PARTIAL_WINNING状態の取得（一部当選）",
			args: args{
				ctx: context.Background(),
				id:  "app-003",
			},
			wantStatus:        domain.LOTTERY_STATUS_PARTIAL_WINNING,
			wantProjectStatus: domain.OPEN_FUNDING,
			wantErr:           false,
		},
		{
			name: "LOST状態の取得",
			args: args{
				ctx: context.Background(),
				id:  "app-004",
			},
			wantStatus:        domain.LOTTERY_STATUS_LOST,
			wantProjectStatus: domain.END_FUNDING,
			wantErr:           false,
		},
		{
			name: "CANCELED状態の取得",
			args: args{
				ctx: context.Background(),
				id:  "app-005",
			},
			wantStatus:        domain.LOTTERY_STATUS_CANCELED,
			wantProjectStatus: domain.END_FUNDING,
			wantErr:           false,
		},
		{
			name: "存在しないID",
			args: args{
				ctx: context.Background(),
				id:  "app-999",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := lotteryApplicationRepository{
				dbFactory: &dbFactoryMock{},
			}

			got, err := repo.GetById(tt.args.ctx, tt.args.id)

			assert.Equalf(t, tt.wantErr, err != nil, "GetById() error = %v, wantErr %v", err, tt.wantErr)

			if !tt.wantErr {
				// Debug output for PARTIAL_WINNING test
				if tt.args.id == "app-003" {
					t.Logf("Debug app-003: Status=%s, Winnings=%+v, Qty=%d, UnitPrice=%d, Amount=%d",
						got.Status, got.Winnings, got.Qty, got.UnitPrice, got.Amount)
					if len(got.Winnings) > 0 {
						var totalQty int64
						for _, w := range got.Winnings {
							totalQty += w.Qty
						}
						t.Logf("Total winning qty: %d, winning amount: %d, applied amount: %d",
							totalQty, totalQty*got.UnitPrice, got.Amount)
					}
				}

				assert.Equalf(t, tt.wantStatus, got.Status, "GetById() status = %v, want %v", got.Status, tt.wantStatus)
				assert.Equalf(t, tt.wantProjectStatus, got.ProjectStatus, "GetById() project_status = %v, want %v", got.ProjectStatus, tt.wantProjectStatus)
				assert.NotEmpty(t, got.Id, "GetById() should have non-empty Id")

				// WINNING/PARTIAL_WINNINGの場合、Winningsもセットされているか確認
				if tt.wantStatus == domain.LOTTERY_STATUS_WINNING || tt.wantStatus == domain.LOTTERY_STATUS_PARTIAL_WINNING {
					assert.NotEmpty(t, got.Winnings, "GetById() should have Winnings for WINNING/PARTIAL_WINNING status")
				}
			}
		})
	}
}

func Test_lotteryRepository_IsAlreadyApplied(t *testing.T) {
	type args struct {
		ctx       context.Context
		projectId int64
		userId    int64
	}
	tests := []struct {
		name        string
		args        args
		wantApplied bool
		wantErr     bool
	}{
		{
			name: "アクティブな申込が存在する（APPLIED）",
			args: args{
				ctx:       context.Background(),
				projectId: 1,
				userId:    100,
			},
			wantApplied: true,
			wantErr:     false,
		},
		{
			name: "申込が存在しない",
			args: args{
				ctx:       context.Background(),
				projectId: 999,
				userId:    999,
			},
			wantApplied: false,
			wantErr:     false,
		},
		{
			name: "キャンセル済みのみ存在する場合（申込可能）",
			args: args{
				ctx:       context.Background(),
				projectId: 3,
				userId:    103,
			},
			wantApplied: false,
			wantErr:     false,
		},
		{
			name: "キャンセル済みと申込中が混在する場合",
			args: args{
				ctx:       context.Background(),
				projectId: 3,
				userId:    105,
			},
			wantApplied: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := lotteryApplicationRepository{
				dbFactory: &dbFactoryMock{},
			}

			applied, err := repo.IsAlreadyApplied(tt.args.ctx, tt.args.projectId, tt.args.userId)

			assert.Equalf(t, tt.wantErr, err != nil, "IsAlreadyApplied() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.wantApplied, applied, "IsAlreadyApplied() = %v, want %v", applied, tt.wantApplied)
		})
	}
}

func Test_lotteryRepository_MarkAsWinning(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name        string
		args        args
		wantSuccess bool
		wantErr     bool
	}{
		{
			name: "winning_atを更新",
			args: args{
				ctx: context.Background(),
				id:  "app-001",
			},
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name: "キャンセル済みの場合は更新されない",
			args: args{
				ctx: context.Background(),
				id:  "app-005",
			},
			wantSuccess: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbFactory := &dbFactoryMock{}
			repo := lotteryApplicationRepository{
				dbFactory: dbFactory,
			}

			success, err := repo.MarkAsWinning(tt.args.ctx, tt.args.id)

			assert.Equalf(t, tt.wantErr, err != nil, "MarkAsWinning() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.wantSuccess, success, "MarkAsWinning() success = %v, wantSuccess %v", success, tt.wantSuccess)

			// 更新後の値を確認
			if !tt.wantErr && tt.wantSuccess {
				got, err := repo.GetById(context.Background(), tt.args.id)
				assert.NoError(t, err)
				assert.False(t, got.WinningAt.IsZero(), "MarkAsWinning() should set WinningAt")
			}
		})
	}
}

func Test_lotteryRepository_MarkAsLost(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name        string
		args        args
		wantSuccess bool
		wantErr     bool
	}{
		{
			name: "lost_atを更新",
			args: args{
				ctx: context.Background(),
				id:  "app-001",
			},
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name: "キャンセル済みの場合は更新されない",
			args: args{
				ctx: context.Background(),
				id:  "app-005",
			},
			wantSuccess: false,
			wantErr:     false,
		},
		{
			name: "当選済みでも落選更新できる",
			args: args{
				ctx: context.Background(),
				id:  "app-002",
			},
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name: "部分当選でも落選更新できる",
			args: args{
				ctx: context.Background(),
				id:  "app-003",
			},
			wantSuccess: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbFactory := &dbFactoryMock{}
			repo := lotteryApplicationRepository{
				dbFactory: dbFactory,
			}

			success, err := repo.MarkAsLost(tt.args.ctx, tt.args.id)

			assert.Equalf(t, tt.wantErr, err != nil, "MarkAsLost() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.wantSuccess, success, "MarkAsLost() success = %v, wantSuccess %v", success, tt.wantSuccess)

			// 更新後の値を確認
			if !tt.wantErr && tt.wantSuccess {
				got, err := repo.GetById(context.Background(), tt.args.id)
				assert.NoError(t, err)
				assert.False(t, got.LostAt.IsZero(), "MarkAsLost() should set LostAt")
			}
		})
	}
}

func Test_lotteryRepository_MarkAsCanceled(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name        string
		args        args
		wantSuccess bool
		wantErr     bool
	}{
		{
			name: "canceled_atを更新",
			args: args{
				ctx: context.Background(),
				id:  "app-001",
			},
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name: "当選済みの場合は更新されない",
			args: args{
				ctx: context.Background(),
				id:  "app-002",
			},
			wantSuccess: false,
			wantErr:     false,
		},
		{
			name: "落選済みの場合は更新されない",
			args: args{
				ctx: context.Background(),
				id:  "app-004",
			},
			wantSuccess: false,
			wantErr:     false,
		},
		{
			name: "キャンセル済みの場合は更新されない",
			args: args{
				ctx: context.Background(),
				id:  "app-005",
			},
			wantSuccess: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbFactory := &dbFactoryMock{}
			repo := lotteryApplicationRepository{
				dbFactory: dbFactory,
			}

			success, err := repo.MarkAsCanceled(tt.args.ctx, tt.args.id)

			assert.Equalf(t, tt.wantErr, err != nil, "MarkAsCanceled() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.wantSuccess, success, "MarkAsCanceled() success = %v, wantSuccess %v", success, tt.wantSuccess)

			// 更新後の値を確認
			if !tt.wantErr && tt.wantSuccess {
				got, err := repo.GetById(context.Background(), tt.args.id)
				assert.NoError(t, err)
				assert.False(t, got.CanceledAt.IsZero(), "MarkAsCanceled() should set CanceledAt")
			}
		})
	}
}
