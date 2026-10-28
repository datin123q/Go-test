package domain

import (
	"context"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
)

// MIR-272: 運用終了・償還完了の前提チェック、金額の CASE 式、メールの宛先のテスト。

func newCloseTestEngine(t *testing.T) (*xorm.Engine, context.Context) {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync2(new(Project), new(Share), new(User), new(AdminUser), new(RepaymentDividendSchedule), new(ProjectInfoHistory)))
	return engine, context.WithValue(context.Background(), factory.ConextDBType, engine)
}

func insertCloseTestShare(t *testing.T, engine *xorm.Engine, id string, projectId, ownerId, amount int64, status INVEST_STATUS) {
	t.Helper()
	// 構造体で入れると、NOT NULL の列(日時など)にもゼロ値が入る。
	_, err := engine.Insert(&Share{Id: id, ProjectId: projectId, OwnerId: ownerId, Amount: amount, Qty: 1, InvestStatus: status, UpdateId: 1})
	require.NoError(t, err)
}

func insertCloseTestSchedule(t *testing.T, engine *xorm.Engine, projectId int64, executeDate string, returnType INVESTMENT_RETURN_TYPE, status INVESTMENT_RETURN_STATUS, balance int64, deleted bool) {
	t.Helper()
	deletedAt := "NULL"
	if deleted {
		deletedAt = "'2026-01-01 00:00:00'"
	}
	_, err := engine.Exec("INSERT INTO repayment_dividend_schedule (project_id, execute_date, return_type, status, investment_balance, deleted) VALUES (?, ?, ?, ?, ?, "+deletedAt+")",
		projectId, executeDate, string(returnType), string(status), balance)
	require.NoError(t, err)
}

func insertCloseTestUser(t *testing.T, engine *xorm.Engine, id int64, email string) {
	t.Helper()
	_, err := engine.Insert(&User{Id: id, Uid: email, CommonId: email, Email: email, Name: "user"})
	require.NoError(t, err)
}

func apiErrorCode(t *testing.T, err error) int {
	t.Helper()
	var apiErr *factory.Error
	require.ErrorAs(t, err, &apiErr)
	return int(apiErr.Code)
}

// 運用終了の前提:出資実行が要る出資(成立・申込中)が残っていなければ通す。
func TestCountInvestmentsBlockingOperationEnd(t *testing.T) {
	tests := []struct {
		name    string
		status  INVEST_STATUS
		blocked bool
	}{
		{"成立が残っていれば拒否", SUCCESS_INVEST, true},
		{"申込中が残っていれば拒否", REQUEST_INVEST, true},
		{"運用中だけなら通す", OPERATION_INVEST, false},
		{"返金済みだけなら通す", REFUND_INVEST, false},
		{"取消だけなら通す", CANCEL_INVEST, false},
		{"不成立だけなら通す", FAIL_INVEST, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, ctx := newCloseTestEngine(t)
			insertCloseTestShare(t, engine, "s1", 1, 1, 1000, tt.status)
			insertCloseTestShare(t, engine, "other", 2, 1, 1000, SUCCESS_INVEST) // 別のファンドは数えない

			count, err := countInvestmentsBlockingOperationEnd(ctx, 1)
			require.NoError(t, err)
			if tt.blocked {
				assert.Equal(t, int64(1), count)
				assert.Equal(t, int(factory.ErrorProjectOperationEndBlocked.Code), apiErrorCode(t, checkOperationEnd(count)))
			} else {
				assert.Zero(t, count)
				assert.NoError(t, checkOperationEnd(count))
			}
		})
	}

	t.Run("論理削除された出資は数えない", func(t *testing.T) {
		engine, ctx := newCloseTestEngine(t)
		insertCloseTestShare(t, engine, "s1", 1, 1, 1000, SUCCESS_INVEST)
		_, err := engine.Exec("UPDATE share SET deleted = '2026-01-01 00:00:00' WHERE id = 's1'")
		require.NoError(t, err)

		count, err := countInvestmentsBlockingOperationEnd(ctx, 1)
		require.NoError(t, err)
		assert.Zero(t, count)
	})
}

// 償還完了の前提:未実施の予定があれば拒否、元本が残る・予定が無いときは確認が要る。
func TestCloseCheck_Check(t *testing.T) {
	tests := []struct {
		name      string
		check     CloseCheck
		confirmed bool
		wantCode  int // 0 なら通る
	}{
		{"最後の予定の残元本が0なら確認なしで通る", CloseCheck{HasSchedule: true, RemainingPrincipal: 0}, false, 0},
		{"残元本がマイナスでも(ジョブと同じ <= 0)確認なしで通る", CloseCheck{HasSchedule: true, RemainingPrincipal: -1}, false, 0},
		{"元本が残っていて同意が無ければ確認が要る", CloseCheck{HasSchedule: true, RemainingPrincipal: 100}, false, int(factory.ErrorProjectCloseNeedsConfirmation.Code)},
		{"元本が残っていても同意があれば通る", CloseCheck{HasSchedule: true, RemainingPrincipal: 100}, true, 0},
		{"予定が1件も無くて同意が無ければ確認が要る", CloseCheck{HasSchedule: false}, false, int(factory.ErrorProjectCloseNeedsConfirmation.Code)},
		{"予定が1件も無くても同意があれば通る", CloseCheck{HasSchedule: false}, true, 0},
		{"未実施の予定があれば同意があっても拒否", CloseCheck{UnexecutedScheduleCount: 1, HasSchedule: true, RemainingPrincipal: 0}, true, int(factory.ErrorProjectCloseBlockedBySchedule.Code)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.check.Check(tt.confirmed)
			if tt.wantCode == 0 {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Equal(t, tt.wantCode, apiErrorCode(t, err))
		})
	}
}

// 最後の予定は、予定の作成時と同じ並び(execute_date 降順・id 降順、論理削除を除く)で決める。
func TestGetCloseCheck(t *testing.T) {
	t.Run("予定が無ければ予定なし・残元本0", func(t *testing.T) {
		_, ctx := newCloseTestEngine(t)
		got, err := Project{}.GetCloseCheck(ctx, 1, false)
		require.NoError(t, err)
		assert.Equal(t, CloseCheck{ConfirmationRequired: true}, got)
		assert.True(t, got.NeedsConfirmation())
	})

	t.Run("execute_date が最も新しい予定の残元本を使い、同じ日は id が大きい方", func(t *testing.T) {
		engine, ctx := newCloseTestEngine(t)
		insertCloseTestSchedule(t, engine, 1, "2026-10-01 00:00:00", REPAYMENT_RETURN, RETURN_COMPLETED, 5000, false)
		insertCloseTestSchedule(t, engine, 1, "2026-10-13 00:00:00", DIVIDEND_RETURN, RETURN_COMPLETED, 700, false)
		insertCloseTestSchedule(t, engine, 1, "2026-10-13 00:00:00", REPAYMENT_RETURN, RETURN_COMPLETED, 0, false)
		// 別のファンド・論理削除された予定は見ない。
		insertCloseTestSchedule(t, engine, 2, "2026-12-01 00:00:00", REPAYMENT_RETURN, RETURN_SCHEDULED, 9999, false)
		insertCloseTestSchedule(t, engine, 1, "2026-12-01 00:00:00", REPAYMENT_RETURN, RETURN_SCHEDULED, 8888, true)

		got, err := Project{}.GetCloseCheck(ctx, 1, true)
		require.NoError(t, err)
		assert.Equal(t, CloseCheck{UnexecutedScheduleCount: 0, HasSchedule: true, RemainingPrincipal: 0, ConfirmationRequired: false}, got)
		assert.False(t, got.NeedsConfirmation())
	})

	t.Run("未実施の予定の件数を数える", func(t *testing.T) {
		engine, ctx := newCloseTestEngine(t)
		insertCloseTestSchedule(t, engine, 1, "2026-10-13 00:00:00", DIVIDEND_RETURN, RETURN_SCHEDULED, 1000, false)
		insertCloseTestSchedule(t, engine, 1, "2026-10-13 00:00:00", REPAYMENT_RETURN, RETURN_SCHEDULED, 0, false)
		insertCloseTestSchedule(t, engine, 1, "2026-10-14 00:00:00", REPAYMENT_RETURN, RETURN_SCHEDULED, 0, true)

		got, err := Project{}.GetCloseCheck(ctx, 1, false)
		require.NoError(t, err)
		assert.Equal(t, int64(2), got.UnexecutedScheduleCount)
	})
}

func TestChangeStatus_OperationEnd(t *testing.T) {
	t.Run("運用中から運用終了へ。出資の状態は変えない", func(t *testing.T) {
		engine, ctx := newCloseTestEngine(t)
		insertTestProject(t, engine, 1, OPERATION_FUNDING)
		insertCloseTestUser(t, engine, 1, "a@example.com")
		insertCloseTestShare(t, engine, "s1", 1, 1, 1000, OPERATION_INVEST)
		insertCloseTestShare(t, engine, "s2", 1, 1, 1000, CANCEL_INVEST)
		_, err := engine.Exec("INSERT INTO admin_user (id, email, name) VALUES (1, 'admin@example.com', 'admin')")
		require.NoError(t, err)

		p := &Project{Id: 1, ProjectStatus: OPERATION_FUNDING}
		require.NoError(t, p.ChangeStatus(ctx, OPERATION_END_FUNDING, 1))

		assert.Equal(t, OPERATION_END_FUNDING, readProjectStatus(t, engine, 1))
		var status string
		_, err = engine.Table("share").Select("invest_status").Where("id = 's1'").Get(&status)
		require.NoError(t, err)
		assert.Equal(t, string(OPERATION_INVEST), status, "運用終了では出資を返金済みにしない(償還で変わる)")
	})

	t.Run("出資実行が済んでいない出資が残っていれば拒否し、状態も変えない", func(t *testing.T) {
		for _, blocking := range []INVEST_STATUS{SUCCESS_INVEST, REQUEST_INVEST} {
			engine, ctx := newCloseTestEngine(t)
			insertTestProject(t, engine, 1, OPERATION_FUNDING)
			insertCloseTestShare(t, engine, "s1", 1, 1, 1000, blocking)

			p := &Project{Id: 1, ProjectStatus: OPERATION_FUNDING}
			err := p.ChangeStatus(ctx, OPERATION_END_FUNDING, 1)
			require.Error(t, err, string(blocking))
			assert.Equal(t, int(factory.ErrorProjectOperationEndBlocked.Code), apiErrorCode(t, err))
			assert.Equal(t, OPERATION_FUNDING, readProjectStatus(t, engine, 1))
		}
	})

	t.Run("運用中以外からは進めない(成立から直接運用終了へは不可)", func(t *testing.T) {
		engine, ctx := newCloseTestEngine(t)
		insertTestProject(t, engine, 1, SUCCESS_FUNDING)

		p := &Project{Id: 1, ProjectStatus: SUCCESS_FUNDING}
		err := p.ChangeStatus(ctx, OPERATION_END_FUNDING, 1)
		require.Error(t, err)
		assert.Equal(t, int(factory.ErrorProjectStatusChanged.Code), apiErrorCode(t, err))
		assert.Equal(t, SUCCESS_FUNDING, readProjectStatus(t, engine, 1))
	})
}

func TestChangeStatus_Close(t *testing.T) {
	setupAdmin := func(t *testing.T, engine *xorm.Engine) {
		_, err := engine.Exec("INSERT INTO admin_user (id, email, name) VALUES (1, 'admin@example.com', 'admin')")
		require.NoError(t, err)
	}

	t.Run("運用中から償還完了へは直接進めない", func(t *testing.T) {
		engine, ctx := newCloseTestEngine(t)
		insertTestProject(t, engine, 1, OPERATION_FUNDING)

		p := &Project{Id: 1, ProjectStatus: OPERATION_FUNDING}
		err := p.ChangeStatus(ctx, CLOSE_FUNDING, 1)
		require.Error(t, err)
		assert.Equal(t, int(factory.ErrorProjectStatusChanged.Code), apiErrorCode(t, err))
		assert.Equal(t, OPERATION_FUNDING, readProjectStatus(t, engine, 1))
	})

	t.Run("運用終了から運用中へは戻せない", func(t *testing.T) {
		engine, ctx := newCloseTestEngine(t)
		insertTestProject(t, engine, 1, OPERATION_END_FUNDING)

		p := &Project{Id: 1, ProjectStatus: OPERATION_END_FUNDING}
		err := p.ChangeStatus(ctx, OPERATION_FUNDING, 1)
		require.Error(t, err)
		assert.Equal(t, OPERATION_END_FUNDING, readProjectStatus(t, engine, 1))
	})

	t.Run("未実施の予定があれば拒否", func(t *testing.T) {
		engine, ctx := newCloseTestEngine(t)
		insertTestProject(t, engine, 1, OPERATION_END_FUNDING)
		insertCloseTestSchedule(t, engine, 1, "2026-10-13 00:00:00", REPAYMENT_RETURN, RETURN_SCHEDULED, 0, false)

		p := &Project{Id: 1, ProjectStatus: OPERATION_END_FUNDING}
		err := p.ChangeStatus(WithRemainingPrincipalConfirmed(ctx), CLOSE_FUNDING, 1)
		require.Error(t, err)
		assert.Equal(t, int(factory.ErrorProjectCloseBlockedBySchedule.Code), apiErrorCode(t, err))
		assert.Equal(t, OPERATION_END_FUNDING, readProjectStatus(t, engine, 1))
	})

	t.Run("元本が残っていれば確認が要る。同意すれば進める", func(t *testing.T) {
		engine, ctx := newCloseTestEngine(t)
		setupAdmin(t, engine)
		insertTestProject(t, engine, 1, OPERATION_END_FUNDING)
		insertCloseTestSchedule(t, engine, 1, "2026-10-13 00:00:00", REPAYMENT_RETURN, RETURN_COMPLETED, 300, false)

		p := &Project{Id: 1, ProjectStatus: OPERATION_END_FUNDING}
		err := p.ChangeStatus(ctx, CLOSE_FUNDING, 1)
		require.Error(t, err)
		assert.Equal(t, int(factory.ErrorProjectCloseNeedsConfirmation.Code), apiErrorCode(t, err))
		assert.Equal(t, OPERATION_END_FUNDING, readProjectStatus(t, engine, 1))

		p = &Project{Id: 1, ProjectStatus: OPERATION_END_FUNDING}
		require.NoError(t, p.ChangeStatus(WithRemainingPrincipalConfirmed(ctx), CLOSE_FUNDING, 1))
		assert.Equal(t, CLOSE_FUNDING, readProjectStatus(t, engine, 1))
	})

	// 設計どおりの流れ(運用終了→最後の償還をジョブが払う→償還完了)では、押す時点で出資は全部
	// 返金済みになっている。宛先は返金済みの出資者も含め、メールはコミットの後に送る。
	t.Run("ジョブで返金済みになった出資者にも、コミット後にメールが届く", func(t *testing.T) {
		engine, baseCtx := newCloseTestEngine(t)
		setupAdmin(t, engine)
		insertTestProject(t, engine, 1, OPERATION_END_FUNDING)
		insertCloseTestSchedule(t, engine, 1, "2026-10-13 00:00:00", REPAYMENT_RETURN, RETURN_COMPLETED, 0, false)
		insertCloseTestUser(t, engine, 1, "refund@example.com")
		insertCloseTestUser(t, engine, 2, "cancel@example.com")
		insertCloseTestShare(t, engine, "s1", 1, 1, 1000, REFUND_INVEST)
		insertCloseTestShare(t, engine, "s2", 1, 2, 1000, CANCEL_INVEST)

		var sent []string
		orig := sendOperationCloseMail
		sendOperationCloseMail = func(_ context.Context, email, _, _ string) error {
			sent = append(sent, email)
			return nil
		}
		t.Cleanup(func() { sendOperationCloseMail = orig })

		ctx, owns := factory.WithAfterCommit(baseCtx)
		require.True(t, owns)

		p := &Project{Id: 1, Name: "fund", ProjectStatus: OPERATION_END_FUNDING}
		require.NoError(t, p.ChangeStatus(ctx, CLOSE_FUNDING, 1))
		assert.Equal(t, CLOSE_FUNDING, readProjectStatus(t, engine, 1))
		assert.Empty(t, sent, "コミットの前には送らない")

		require.NoError(t, factory.RunAfterCommit(ctx, baseCtx))
		assert.Equal(t, []string{"refund@example.com"}, sent, "返金済みの出資者に届き、取消の人には届かない")
	})

	t.Run("メールの送信に失敗しても状態は変わったまま(コミット後のため)", func(t *testing.T) {
		engine, baseCtx := newCloseTestEngine(t)
		setupAdmin(t, engine)
		insertTestProject(t, engine, 1, OPERATION_END_FUNDING)
		insertCloseTestSchedule(t, engine, 1, "2026-10-13 00:00:00", REPAYMENT_RETURN, RETURN_COMPLETED, 0, false)
		insertCloseTestUser(t, engine, 1, "a@example.com")
		insertCloseTestUser(t, engine, 2, "b@example.com")
		insertCloseTestShare(t, engine, "s1", 1, 1, 1000, REFUND_INVEST)
		insertCloseTestShare(t, engine, "s2", 1, 2, 1000, REFUND_INVEST)

		var attempted int
		orig := sendOperationCloseMail
		sendOperationCloseMail = func(_ context.Context, _, _, _ string) error {
			attempted++
			return errors.New("send failed")
		}
		t.Cleanup(func() { sendOperationCloseMail = orig })

		ctx, _ := factory.WithAfterCommit(baseCtx)
		p := &Project{Id: 1, Name: "fund", ProjectStatus: OPERATION_END_FUNDING}
		require.NoError(t, p.ChangeStatus(ctx, CLOSE_FUNDING, 1))

		err := factory.RunAfterCommit(ctx, baseCtx)
		require.Error(t, err)
		assert.Equal(t, 2, attempted, "1人の失敗で残りの宛先を止めない")
		assert.Equal(t, CLOSE_FUNDING, readProjectStatus(t, engine, 1))
	})
}

func TestGetProjectInvestorsByProjectStatus_Close(t *testing.T) {
	engine, ctx := newCloseTestEngine(t)
	insertTestProject(t, engine, 1, CLOSE_FUNDING)
	insertCloseTestUser(t, engine, 1, "operation@example.com")
	insertCloseTestUser(t, engine, 2, "refund@example.com")
	insertCloseTestUser(t, engine, 3, "cancel@example.com")
	insertCloseTestShare(t, engine, "s1", 1, 1, 1000, OPERATION_INVEST)
	insertCloseTestShare(t, engine, "s2", 1, 2, 1000, REFUND_INVEST)
	insertCloseTestShare(t, engine, "s3", 1, 3, 1000, CANCEL_INVEST)

	t.Run("償還完了は運用中と返金済みの出資者へ", func(t *testing.T) {
		p := &Project{Id: 1, ProjectStatus: CLOSE_FUNDING}
		got, err := p.GetProjectInvestorsByProjectStatus(ctx)
		require.NoError(t, err)
		emails := make([]string, 0, len(got))
		for _, inv := range got {
			emails = append(emails, inv.Email)
		}
		assert.ElementsMatch(t, []string{"operation@example.com", "refund@example.com"}, emails)
	})

	t.Run("運用終了はメールを送らない(宛先なし)", func(t *testing.T) {
		p := &Project{Id: 1, ProjectStatus: OPERATION_END_FUNDING}
		got, err := p.GetProjectInvestorsByProjectStatus(ctx)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

// 公開用の金額・進捗率は、運用中・運用終了とも、最後の償還を払って出資が返金済みになっても
// 変わらない(0円・0% にならない)。償還完了は返金済みを数える。
func TestPublicViewAmountAndPercent_OperationEnd(t *testing.T) {
	query := "SELECT " + PUBLIC_VIEW_SHARE_AMOUNT + " AS amount, " + PUBLIC_PROGRESS_INVEST_PERCENT + " AS percent FROM project WHERE project.id = 1"

	read := func(t *testing.T, engine *xorm.Engine) (string, string) {
		rows, err := engine.QueryString(query)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		return rows[0]["amount"], rows[0]["percent"]
	}

	for _, status := range []PROJECT_STATUS{OPERATION_FUNDING, OPERATION_END_FUNDING} {
		t.Run(string(status)+": 償還の前後で金額・進捗率が変わらない", func(t *testing.T) {
			engine, _ := newCloseTestEngine(t)
			insertTestProject(t, engine, 1, status)
			_, err := engine.Exec("UPDATE project SET target_share_amount = 4000 WHERE id = 1")
			require.NoError(t, err)
			insertCloseTestShare(t, engine, "s1", 1, 1, 2000, OPERATION_INVEST)
			insertCloseTestShare(t, engine, "s2", 1, 2, 2000, OPERATION_INVEST)
			insertCloseTestShare(t, engine, "s3", 1, 3, 1000, CANCEL_INVEST) // 取消は数えない

			// SQLite は整数どうしの割り算が切り捨てになるため、目標額と同額にして 100% で見る。
			amountBefore, percentBefore := read(t, engine)
			assert.Equal(t, "4000", amountBefore)
			assert.Equal(t, "100", percentBefore)

			_, err = engine.Exec("UPDATE share SET invest_status = 'REFUND_INVEST' WHERE invest_status = 'OPERATION_INVEST'")
			require.NoError(t, err)

			amountAfter, percentAfter := read(t, engine)
			assert.Equal(t, amountBefore, amountAfter)
			assert.Equal(t, percentBefore, percentAfter)
		})
	}

	t.Run("償還完了は返金済みを数える", func(t *testing.T) {
		engine, _ := newCloseTestEngine(t)
		insertTestProject(t, engine, 1, CLOSE_FUNDING)
		_, err := engine.Exec("UPDATE project SET target_share_amount = 4000 WHERE id = 1")
		require.NoError(t, err)
		insertCloseTestShare(t, engine, "s1", 1, 1, 4000, REFUND_INVEST)

		amount, percent := read(t, engine)
		assert.Equal(t, "4000", amount)
		assert.Equal(t, "100", percent)
	})
}

// 最後の予定の元本残高が NULL(本番のスキーマは null を許す)のとき、0 とみなして確認なしで進めない。
func TestGetCloseCheck_NullBalance(t *testing.T) {
	engine, ctx := newCloseTestEngine(t)
	_, err := engine.Exec("INSERT INTO repayment_dividend_schedule (project_id, execute_date, return_type, status, investment_balance) VALUES (1, '2026-10-13 00:00:00', 'REPAYMENT_RETURN', 'RETURN_COMPLETED', NULL)")
	require.NoError(t, err)

	got, err := Project{}.GetCloseCheck(ctx, 1, false)
	require.NoError(t, err)
	assert.True(t, got.HasSchedule)
	assert.True(t, got.BalanceUnknown)
	assert.True(t, got.NeedsConfirmation())
	assert.True(t, got.ConfirmationRequired, "画面は確認が要るかをこの値だけで決める")
	require.Error(t, got.Check(false))
	assert.NoError(t, got.Check(true))
}

// ユーザー画面の一覧タブは、運用終了(OPERATION_END_FUNDING_LIST_TYPE)と償還完了(CLOSE_FUNDING_LIST_TYPE)を
// 別々に絞り込む。運用中のタブは運用終了を含まない。
func TestGetByProjectType_OperationEndAndCloseTabsAreSeparate(t *testing.T) {
	engine, ctx := newCloseTestEngine(t)
	require.NoError(t, engine.Sync2(new(ProjectViewCount)))
	statuses := map[int64]PROJECT_STATUS{
		1: OPERATION_FUNDING, 2: OPERATION_END_FUNDING, 3: CLOSE_FUNDING, 4: INVALID_FUNDING, 5: SUCCESS_FUNDING,
	}
	for id, status := range statuses {
		insertTestProject(t, engine, id, status)
		_, err := engine.Exec("UPDATE project SET posted = 1, is_closed_fund = 0 WHERE id = ?", id)
		require.NoError(t, err)
	}

	idsOf := func(listType PROJECT_STATUS_LIST_TYPE) []int64 {
		_, items, err := Project{}.GetByProjectType(ctx, listType, 0, 50, "", "", nil, "", "")
		require.NoError(t, err)
		got := make([]int64, 0, len(items))
		for _, item := range items {
			got = append(got, item.Id)
		}
		return got
	}
	assert.ElementsMatch(t, []int64{2}, idsOf(OPERATION_END_FUNDING_LIST_TYPE))
	assert.ElementsMatch(t, []int64{3}, idsOf(CLOSE_FUNDING_LIST_TYPE))
	assert.ElementsMatch(t, []int64{1}, idsOf(OPERATION_FUNDING_LIST_TYPE))
}

// 管理画面のファンド一覧の「募集状態」は、状態がぴったり同じファンドだけを返す。運用終了と償還完了は別々に選べる
// (選択肢は、コード表(GetStatusCodeGroup)から作られる)。
func TestGetAll_ProjectStatusFilterSeparatesOperationEndAndClose(t *testing.T) {
	engine, _ := newCloseTestEngine(t)
	require.NoError(t, engine.Sync2(new(ProjectViewCount)))
	// 検索条件の組み立て(modelhelper)は *xorm.Session を要る。
	session := engine.NewSession()
	t.Cleanup(func() { _ = session.Close() })
	ctx := context.WithValue(context.Background(), factory.ConextDBType, session)
	for id, status := range map[int64]PROJECT_STATUS{1: OPERATION_FUNDING, 2: OPERATION_END_FUNDING, 3: CLOSE_FUNDING} {
		insertTestProject(t, engine, id, status)
	}

	idsOf := func(status PROJECT_STATUS) []int64 {
		st := string(status)
		_, items, err := Project{}.GetAll(ctx, "", "", 0, 50, nil, &st, nil, nil, nil, nil, nil, "", "", false, nil)
		require.NoError(t, err)
		got := make([]int64, 0, len(items))
		for _, item := range items {
			got = append(got, item.Id)
		}
		return got
	}
	assert.Equal(t, []int64{2}, idsOf(OPERATION_END_FUNDING))
	assert.Equal(t, []int64{3}, idsOf(CLOSE_FUNDING))
}

// 表示名: 新しい状態は「運用終了」、CLOSE は「償還完了」。出資の状態 REFUND_INVEST は、ファンドの「償還完了」と
// 区別するため「償還済み」。
func TestProjectStatusNames(t *testing.T) {
	require.NoError(t, i18nutils.InitI18nFiles("../assets"))

	names := func(group commoncode.CodeGroup) map[string]string {
		m := map[string]string{}
		for _, c := range group.Codes {
			m[c.Code] = c.Name
		}
		return m
	}

	for _, loc := range []string{"ja", "en", "ko"} {
		status := names(Project{}.GetStatusCodeGroup(loc))
		require.Contains(t, status, string(OPERATION_END_FUNDING), loc)
		assert.NotEmpty(t, status[string(OPERATION_END_FUNDING)], loc)
		assert.NotEqual(t, status[string(OPERATION_END_FUNDING)], status[string(CLOSE_FUNDING)], "%s: 運用終了と償還完了は別の名前", loc)
		assert.NotEqual(t, status[string(OPERATION_FUNDING)], status[string(OPERATION_END_FUNDING)], loc)
	}

	ja := names(Project{}.GetStatusCodeGroup("ja"))
	assert.Equal(t, "運用終了", ja[string(OPERATION_END_FUNDING)])
	assert.Equal(t, "償還完了", ja[string(CLOSE_FUNDING)])
	// 一覧タブも、運用終了と償還完了で別の名前(状態の表示名と同じ訳語)。
	tabs := names(Project{}.GetProjrctListStatusCodeGroup("ja"))
	assert.Equal(t, "運用終了", tabs[string(OPERATION_END_FUNDING_LIST_TYPE)])
	assert.Equal(t, "償還完了", tabs[string(CLOSE_FUNDING_LIST_TYPE)])
	assert.Equal(t, "償還済み", names(Share{}.GetCodeGroup("ja"))[string(REFUND_INVEST)])
	assert.Equal(t, "償還済み", REFUND_INVEST.Label(), "エラー文言の表示名も同じ")
	for _, loc := range []string{"en", "ko"} {
		assert.NotEqual(t, names(Project{}.GetStatusCodeGroup(loc))[string(CLOSE_FUNDING)], names(Share{}.GetCodeGroup(loc))[string(REFUND_INVEST)], "%s: ファンドの償還完了と出資の償還済みは別の名前", loc)
	}
}

// 元本が残ったまま(元本割れなどでジョブが出資を返金済みにしなかった)、出資が運用中のファンドを
// 同意のうえで償還完了にすると、その出資は返金済みになり、宛先は付け替えの前に取るので、その出資者にも
// メールが届く。
func TestChangeStatus_Close_RefundsRemainingOperationInvest(t *testing.T) {
	engine, baseCtx := newCloseTestEngine(t)
	require.NoError(t, engine.Sync2(new(DepositRequest), new(Role), new(Spv), new(ProjectNews), new(Attachment), new(ProjectFile), new(ProjectImage), new(ProjectMilestone), new(ProjectInvestmentExecute), new(ProjectViewCount)))
	_, err := engine.Exec("INSERT INTO admin_user (id, email, name) VALUES (1, 'admin@example.com', 'admin')")
	require.NoError(t, err)
	insertTestProject(t, engine, 1, OPERATION_END_FUNDING)
	insertCloseTestSchedule(t, engine, 1, "2026-10-13 00:00:00", REPAYMENT_RETURN, RETURN_COMPLETED, 300, false)
	insertCloseTestUser(t, engine, 1, "operation@example.com")
	insertCloseTestShare(t, engine, "s1", 1, 1, 1000, OPERATION_INVEST)
	_, err = engine.Insert(&DepositRequest{ShareId: "s1", UserId: 1, ProjectId: 1})
	require.NoError(t, err)

	var sent []string
	orig := sendOperationCloseMail
	sendOperationCloseMail = func(_ context.Context, email, _, _ string) error {
		sent = append(sent, email)
		return nil
	}
	t.Cleanup(func() { sendOperationCloseMail = orig })

	ctx, _ := factory.WithAfterCommit(WithRemainingPrincipalConfirmed(baseCtx))
	p := &Project{Id: 1, Name: "fund", ProjectStatus: OPERATION_END_FUNDING}
	require.NoError(t, p.ChangeStatus(ctx, CLOSE_FUNDING, 1))

	var status string
	_, err = engine.Table("share").Select("invest_status").Where("id = 's1'").Get(&status)
	require.NoError(t, err)
	assert.Equal(t, string(REFUND_INVEST), status, "残っていた運用中の出資は返金済みになる")

	require.NoError(t, factory.RunAfterCommit(ctx, baseCtx))
	assert.Equal(t, []string{"operation@example.com"}, sent)
}

// 論理削除されたファンドは、前提チェックの前のロックで見つからないものとして拒否する(生 SQL は
// XORM の論理削除フィルタを通らないため、deleted IS NULL を自分で付けている)。
func TestChangeStatus_RejectsDeletedProject(t *testing.T) {
	engine, ctx := newCloseTestEngine(t)
	insertTestProject(t, engine, 1, OPERATION_FUNDING)
	_, err := engine.Exec("UPDATE project SET deleted = '2026-10-15 00:00:00' WHERE id = 1")
	require.NoError(t, err)

	p := &Project{Id: 1, ProjectStatus: OPERATION_FUNDING}
	err = p.ChangeStatus(ctx, OPERATION_END_FUNDING, 1)
	require.Error(t, err)
	assert.Equal(t, int(factory.ErrorNotFound.Code), apiErrorCode(t, err))
	// 論理削除された行は engine.Get で見えないため、生 SQL で状態が変わっていないことを見る。
	var status string
	_, err = engine.SQL("SELECT project_status FROM project WHERE id = 1").Get(&status)
	require.NoError(t, err)
	assert.Equal(t, string(OPERATION_FUNDING), status)
}

// 償還済みへの付け替えは、ファンドの状態とは別の処理。ファンドの運用中の出資だけを償還済みにし、取消などは変えない。
// 分配償還ジョブ(updateId=0)が呼ぶ形でも動く(スーパー管理者の取得が失敗しない)。
func TestRefundOperationInvestsByProject(t *testing.T) {
	for _, updateId := range []int64{1, 0} {
		engine, baseCtx := newCloseTestEngine(t)
		require.NoError(t, engine.Sync2(new(DepositRequest), new(Role), new(Spv), new(ProjectNews), new(Attachment), new(ProjectFile), new(ProjectImage), new(ProjectMilestone), new(ProjectInvestmentExecute), new(ProjectViewCount)))
		_, err := engine.Exec("INSERT INTO admin_user (id, email, name) VALUES (1, 'admin@example.com', 'admin')")
		require.NoError(t, err)
		_, err = engine.Insert(&Role{UserId: 1, RoleCode: string(ROLE_SUPER_ADMIN)})
		require.NoError(t, err)
		insertTestProject(t, engine, 1, OPERATION_END_FUNDING)
		insertTestProject(t, engine, 2, OPERATION_END_FUNDING)
		insertCloseTestUser(t, engine, 1, "a@example.com")
		insertCloseTestShare(t, engine, "s1", 1, 1, 1000, OPERATION_INVEST)
		insertCloseTestShare(t, engine, "s2", 1, 1, 1000, CANCEL_INVEST)
		insertCloseTestShare(t, engine, "other", 2, 1, 1000, OPERATION_INVEST) // 別のファンドは変えない
		_, err = engine.Insert(&DepositRequest{ShareId: "s1", UserId: 1, ProjectId: 1})
		require.NoError(t, err)
		_, err = engine.Insert(&DepositRequest{ShareId: "other", UserId: 1, ProjectId: 2})
		require.NoError(t, err)

		require.NoError(t, Share{}.RefundOperationInvestsByProject(baseCtx, 1, updateId), "updateId=%d", updateId)

		statusOf := func(id string) string {
			var st string
			_, err := engine.Table("share").Select("invest_status").Where("id = ?", id).Get(&st)
			require.NoError(t, err)
			return st
		}
		assert.Equal(t, string(REFUND_INVEST), statusOf("s1"))
		assert.Equal(t, string(CANCEL_INVEST), statusOf("s2"))
		assert.Equal(t, string(OPERATION_INVEST), statusOf("other"))
		assert.Equal(t, OPERATION_END_FUNDING, readProjectStatus(t, engine, 1), "ファンドの状態は変えない")
	}
}

// 償還完了の状態を借りた古い呼び方(ChangeOfInvestStatusByProject(CLOSE_FUNDING))は、黙って何もしないで通さずエラーにする。
func TestChangeOfInvestStatusByProject_CloseIsNotSupported(t *testing.T) {
	_, ctx := newCloseTestEngine(t)
	err := Share{}.ChangeOfInvestStatusByProject(ctx, CLOSE_FUNDING, 1, 1)
	require.Error(t, err)
}
