package controllers

import (
	"mime/multipart"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func TestShouldReopenFunding(t *testing.T) {
	now := time.Date(2026, 8, 12, 12, 0, 0, 0, time.Local)
	day := func(d int) time.Time {
		return time.Date(2026, 8, d, 23, 59, 59, 0, time.Local)
	}

	tests := []struct {
		name        string
		status      domain.PROJECT_STATUS
		beforeEndAt time.Time
		afterEndAt  time.Time
		want        bool
	}{
		{
			name:        "募集終了日を後ろへ変更し、現時点より後なら募集中へ戻す",
			status:      domain.END_FUNDING,
			beforeEndAt: day(20),
			afterEndAt:  day(31),
			want:        true,
		},
		{
			// 前倒しでも変更後が未来ならまだ募集期間中のため戻す。
			name:        "募集終了日を前倒しし、現時点より後なら募集中へ戻す",
			status:      domain.END_FUNDING,
			beforeEndAt: day(31),
			afterEndAt:  day(20),
			want:        true,
		},
		{
			// 早期募集終了の直後に保存しても再開しないことが今回の主目的。
			name:        "募集終了日を変更していなければ募集終了のまま",
			status:      domain.END_FUNDING,
			beforeEndAt: day(31),
			afterEndAt:  day(31),
			want:        false,
		},
		{
			name:        "前倒しでも変更後が現時点より前なら募集終了のまま",
			status:      domain.END_FUNDING,
			beforeEndAt: day(31),
			afterEndAt:  day(5),
			want:        false,
		},
		{
			name:        "変更しても現時点より前なら募集終了のまま",
			status:      domain.END_FUNDING,
			beforeEndAt: day(1),
			afterEndAt:  day(5),
			want:        false,
		},
		{
			name:        "募集中は対象外",
			status:      domain.OPEN_FUNDING,
			beforeEndAt: day(20),
			afterEndAt:  day(31),
			want:        false,
		},
		{
			name:        "成立は対象外",
			status:      domain.SUCCESS_FUNDING,
			beforeEndAt: day(20),
			afterEndAt:  day(31),
			want:        false,
		},
		{
			name:        "無効は対象外",
			status:      domain.INVALID_FUNDING,
			beforeEndAt: day(20),
			afterEndAt:  day(31),
			want:        false,
		},
		{
			// 現時点ちょうどは「後」ではないため戻さない。
			name:        "変更後が現時点ちょうどなら募集終了のまま",
			status:      domain.END_FUNDING,
			beforeEndAt: day(1),
			afterEndAt:  now,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldReopenFunding(tt.status, tt.beforeEndAt, tt.afterEndAt, now)

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestShouldAttemptReopenは、shouldReopenFundingと違い画面の状態を見ない
// ことを確認する(MIR-1314)。画面が募集中(古い)を表示していても、終了日時が
// 延長され現時点より後になっていれば再開を試みる対象になる。
func TestShouldAttemptReopen(t *testing.T) {
	now := time.Date(2026, 8, 12, 12, 0, 0, 0, time.Local)
	day := func(d int) time.Time {
		return time.Date(2026, 8, d, 23, 59, 59, 0, time.Local)
	}

	tests := []struct {
		name        string
		beforeEndAt time.Time
		afterEndAt  time.Time
		want        bool
	}{
		{
			name:        "終了日時を延長し現時点より後なら対象になる",
			beforeEndAt: day(20),
			afterEndAt:  day(31),
			want:        true,
		},
		{
			name:        "終了日時を変更していなければ対象外",
			beforeEndAt: day(31),
			afterEndAt:  day(31),
			want:        false,
		},
		{
			name:        "変更後が現時点より前なら対象外",
			beforeEndAt: day(1),
			afterEndAt:  day(5),
			want:        false,
		},
		{
			// Afterは同時刻を含まないため、境界のちょうど今は対象外。
			name:        "変更後が現時点ちょうどなら対象外",
			beforeEndAt: day(1),
			afterEndAt:  now,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldAttemptReopen(tt.beforeEndAt, tt.afterEndAt, now)

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestProjectReopenAlreadyConflictedは、画面が募集終了以外を表示していた
// 場合はDBの状態に関わらず競合として扱わないことを確認する(MIR-1314)。
func TestProjectReopenAlreadyConflicted(t *testing.T) {
	tests := []struct {
		name                 string
		explicitReopenIntent bool
		currentStatus        domain.PROJECT_STATUS
		want                 bool
	}{
		{
			name:                 "再開のつもりで、DBが既に募集終了でなくなっていれば競合",
			explicitReopenIntent: true,
			currentStatus:        domain.SUCCESS_FUNDING,
			want:                 true,
		},
		{
			name:                 "再開のつもりで、DBがまだ募集終了なら競合ではない",
			explicitReopenIntent: true,
			currentStatus:        domain.END_FUNDING,
			want:                 false,
		},
		{
			name:                 "再開のつもりでなければ、DBがどうであっても競合ではない",
			explicitReopenIntent: false,
			currentStatus:        domain.SUCCESS_FUNDING,
			want:                 false,
		},
		{
			// 画面は募集終了以外(古い)を表示、DBは募集終了、という延長経路。
			name:                 "再開のつもりでなく、DBが募集終了でも競合ではない",
			explicitReopenIntent: false,
			currentStatus:        domain.END_FUNDING,
			want:                 false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := projectReopenAlreadyConflicted(tt.explicitReopenIntent, tt.currentStatus)

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestProjectReopenResultは、再開のつもりが無かった場合は再開できなくても
// 保存を成功として扱う(conflict=false)ことを確認する(MIR-1314)。
func TestProjectReopenResult(t *testing.T) {
	tests := []struct {
		name                 string
		reopened             bool
		explicitReopenIntent bool
		wantConflict         bool
	}{
		{
			name:                 "再開できたなら競合ではない",
			reopened:             true,
			explicitReopenIntent: true,
			wantConflict:         false,
		},
		{
			name:                 "再開のつもりだったのに再開できなければ競合",
			reopened:             false,
			explicitReopenIntent: true,
			wantConflict:         true,
		},
		{
			name:                 "再開のつもりが無ければ、再開できなくても競合ではない",
			reopened:             false,
			explicitReopenIntent: false,
			wantConflict:         false,
		},
		{
			// 画面は募集終了以外(古い)を表示していたが、DBが募集終了だった
			// ため再開できた経路。再開のつもりで無くても競合にはしない。
			name:                 "再開のつもりが無くても、再開できたなら競合ではない",
			reopened:             true,
			explicitReopenIntent: false,
			wantConflict:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := projectReopenResult(tt.reopened, tt.explicitReopenIntent)

			assert.Equal(t, tt.wantConflict, got)
		})
	}
}

func TestCanDeleteProjectFile(t *testing.T) {
	newFile := func(key string) map[string][]*multipart.FileHeader {
		return map[string][]*multipart.FileHeader{key: {{Filename: "dummy.pdf"}}}
	}

	tests := []struct {
		name     string
		fileType domain.PROJECT_FILE_TYPE
		newFiles map[string][]*multipart.FileHeader
		want     bool
	}{
		{
			name:     "必須書面(契約締結時交付書面)を差し替えなしで削除しようとすると拒否する",
			fileType: domain.CONTRACT_AFTER_PAGE,
			newFiles: map[string][]*multipart.FileHeader{},
			want:     false,
		},
		{
			name:     "必須書面(契約成立前交付書面)を差し替えなしで削除しようとすると拒否する",
			fileType: domain.CONTRACT_BEFORE_PAGE,
			newFiles: map[string][]*multipart.FileHeader{},
			want:     false,
		},
		{
			name:     "必須書面(匿名組合契約書)を差し替えなしで削除しようとすると拒否する",
			fileType: domain.SILENT_PARTNERSHIPS,
			newFiles: map[string][]*multipart.FileHeader{},
			want:     false,
		},
		{
			name:     "同じ種別の新規ファイルが同時に送られていれば削除(差し替え)を許可する",
			fileType: domain.CONTRACT_AFTER_PAGE,
			newFiles: newFile("contract_after_page"),
			want:     true,
		},
		{
			name:     "別種別の新規ファイルが送られていても削除は拒否する",
			fileType: domain.CONTRACT_AFTER_PAGE,
			newFiles: newFile("silent_partnerships"),
			want:     false,
		},
		{
			name:     "必須でない書面(ファンドレポート)は差し替えなしでも削除できる",
			fileType: domain.PROJECT_REPORT,
			newFiles: map[string][]*multipart.FileHeader{},
			want:     true,
		},
		{
			name:     "newFilesがnilでも必須でない種別は削除できる",
			fileType: domain.PROJECT_REPORT,
			newFiles: nil,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := canDeleteProjectFile(tt.fileType, tt.newFiles)

			assert.Equal(t, tt.want, got)
		})
	}
}
