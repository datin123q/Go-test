package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// AuditStatusChangeEntry は audit_status の変更を履歴に記録するために必要なデータを保持します。
type AuditStatusChangeEntry struct {
	UserId        int64
	ActorId       int64
	ActorName     string
	ActorEmail    string
	ProcessMethod domain.USER_HISTORY_PROCESS_TYPE
	PrevStatus    domain.AUDIT_STATUS
	CurrStatus    domain.AUDIT_STATUS
}

// ContentChange は user_changed_content に記録する1項目分の変更内容です。
type ContentChange struct {
	ContentType domain.USER_CHANGED_CONTENT_TYPE
	Prev        string
	Curr        string
}

// ContentChangeEntry は内容変更（USER_HISTORY_CONTENTS）を履歴に記録するために必要なデータを保持します。
type ContentChangeEntry struct {
	UserId        int64
	ActorId       int64
	ActorName     string
	ActorEmail    string
	ProcessMethod domain.USER_HISTORY_PROCESS_TYPE
	Changes       []ContentChange
}

// UserInfoHistoryRepository はユーザー情報の変更履歴を永続化するリポジトリのインターフェースです。
type UserInfoHistoryRepository interface {
	RecordAuditStatusChange(ctx context.Context, entry AuditStatusChangeEntry) error
	RecordContentChanges(ctx context.Context, entry ContentChangeEntry) error
}
