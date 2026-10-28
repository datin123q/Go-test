package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/actorcontext"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type historyAwareUserRepository struct {
	inner   cdomain.UserRepository
	history cdomain.UserInfoHistoryRepository
	logger  factory.LoggerFactory
}

// NewHistoryAware は UserRepository をラップし、Update() 呼び出しで audit_status・メールアドレス・パスワード・
// ニュースレター受信設定が変更された場合に
// 自動的に user_info_history へ変更を記録するデコレーターを返します。
// 変更を実行したアクターは actorcontext 経由でコンテキストから取得します。
// アクターが無い場合、またはアクターが会員本人の場合は、会員本人の操作として会員の行の氏名・メールアドレスを記録します。
func NewHistoryAware(
	inner cdomain.UserRepository,
	history cdomain.UserInfoHistoryRepository,
	logger factory.LoggerFactory,
) cdomain.UserRepository {
	return &historyAwareUserRepository{
		inner:   inner,
		history: history,
		logger:  logger,
	}
}

func (h historyAwareUserRepository) Create(ctx context.Context, user domain.User) (int64, error) {
	return h.inner.Create(ctx, user)
}

func (h historyAwareUserRepository) GetById(ctx context.Context, id int64) (domain.User, error) {
	return h.inner.GetById(ctx, id)
}

func (h historyAwareUserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	return h.inner.GetByEmail(ctx, email)
}

func (h historyAwareUserRepository) GetByEmailIncludeDeleted(ctx context.Context, email string) (domain.User, error) {
	return h.inner.GetByEmailIncludeDeleted(ctx, email)
}

func (h historyAwareUserRepository) GetAll(ctx context.Context, options ...cdomain.UserQueryOption) (int64, []domain.User, error) {
	return h.inner.GetAll(ctx, options...)
}

// UpdateLastLoginDate は last_login_date しか書かず audit_status に影響しないため、
// 履歴記録を挟まず内部リポジトリへそのまま委譲する。
func (h historyAwareUserRepository) UpdateLastLoginDate(ctx context.Context, id int64, lastLoginDate jsonutils.JSONTime) error {
	return h.inner.UpdateLastLoginDate(ctx, id, lastLoginDate)
}

// historyActor は履歴に記録する処理者情報です。
type historyActor struct {
	id            int64
	name          string
	email         string
	processMethod domain.USER_HISTORY_PROCESS_TYPE
}

// resolveActor はコンテキストから処理者を取得します。
// アクターが設定されていない場合（パスワード再設定など未ログインの操作を含む）は会員本人による操作として扱います。
// アクターが会員本人の場合も、氏名・メールアドレスは JWT ではなく会員の行から取ります。
// JWT はログイン時点の値のため、メールアドレス変更後に再ログインしていないと古いアドレスが記録されるためです。
func resolveActor(ctx context.Context, user domain.User) historyActor {
	actor, hasActor := actorcontext.ActorFrom(ctx)
	if !hasActor || (actor.Id == user.Id && actor.HasRole(domain.ROLE_USER)) {
		return historyActor{
			id:            user.Id,
			name:          user.Name,
			email:         user.Email,
			processMethod: domain.USER_HISTORY_PROCESS_MEMBER,
		}
	}

	processMethod := domain.USER_HISTORY_PROCESS_MEMBER
	if !actor.HasRole(domain.ROLE_USER) {
		processMethod = domain.USER_HISTORY_PROCESS_OPERATOR
	}
	return historyActor{
		id:            actor.Id,
		name:          actor.Name,
		email:         actor.Email,
		processMethod: processMethod,
	}
}

// checkTypeOf は bool 値を履歴表示用のコード（はい／いいえ）に変換します。
func checkTypeOf(v bool) string {
	if v {
		return string(domain.USER_HISTORY_CHECK_YES)
	}
	return string(domain.USER_HISTORY_CHECK_NO)
}

// collectContentChanges は変更前後のユーザーを比較し、内容変更として記録すべき項目を返します。
// XORM の Update は空文字のカラムを更新しないため、空文字の Email/Password は変更として扱いません。
// bool カラムは UseBool() により常に更新されるため、値が異なれば変更として扱います。
func collectContentChanges(prev, curr domain.User) []cdomain.ContentChange {
	changes := make([]cdomain.ContentChange, 0, 3)

	if curr.Email != "" && curr.Email != prev.Email {
		changes = append(changes, cdomain.ContentChange{
			ContentType: domain.USER_CHANGED_CONTENT_EMAIL,
			Prev:        prev.Email,
			Curr:        curr.Email,
		})
	}

	// パスワードはハッシュ値も含めて履歴に残さず、変更があった事実のみをマスク値で記録する
	if curr.Password != "" && curr.Password != prev.Password {
		changes = append(changes, cdomain.ContentChange{
			ContentType: domain.USER_CHANGED_CONTENT_PASSWORD,
			Prev:        domain.USER_HISTORY_MASKED_VALUE,
			Curr:        domain.USER_HISTORY_MASKED_VALUE,
		})
	}

	if curr.IsNewsletter != prev.IsNewsletter {
		changes = append(changes, cdomain.ContentChange{
			ContentType: domain.USER_CHANGED_CONTENT_NEWSLETTER,
			Prev:        checkTypeOf(prev.IsNewsletter),
			Curr:        checkTypeOf(curr.IsNewsletter),
		})
	}

	return changes
}

// Update は内部リポジトリに処理を委譲し、audit_status・メールアドレス・パスワード・ニュースレター受信設定が
// 変更された場合に user_info_history へ変更を記録します。
// 履歴記録が失敗しても Update 自体は失敗しません — ログに記録して握り潰します。
func (h historyAwareUserRepository) Update(ctx context.Context, user domain.User) error {
	// 変更前後を比較するために現在の状態を取得する。
	prev, err := h.inner.GetById(ctx, user.Id)
	if err != nil {
		// 変更前の状態を取得できない場合でも更新処理は続行する。
		h.logger.Entry(ctx).WithError(err).Warn("historyAwareUserRepository: could not fetch previous user state for history recording")
		return h.inner.Update(ctx, user)
	}

	// 実際の更新処理を実行する。
	if err := h.inner.Update(ctx, user); err != nil {
		return err
	}

	actor := resolveActor(ctx, user)

	// audit_status が実際に変更された場合のみ記録する。
	if user.AuditStatus != "" && prev.AuditStatus != user.AuditStatus {
		entry := cdomain.AuditStatusChangeEntry{
			UserId:        user.Id,
			ActorId:       actor.id,
			ActorName:     actor.name,
			ActorEmail:    actor.email,
			ProcessMethod: actor.processMethod,
			PrevStatus:    prev.AuditStatus,
			CurrStatus:    user.AuditStatus,
		}
		if err := h.history.RecordAuditStatusChange(ctx, entry); err != nil {
			// 履歴記録の失敗によってメインの更新処理をロールバックしてはいけない。
			h.logger.Entry(ctx).WithError(err).Error("historyAwareUserRepository: failed to record audit_status change history")
		}
	}

	// メールアドレス・パスワード・ニュースレター受信設定の変更を1件の内容変更履歴として記録する。
	if changes := collectContentChanges(prev, user); len(changes) > 0 {
		entry := cdomain.ContentChangeEntry{
			UserId:        user.Id,
			ActorId:       actor.id,
			ActorName:     actor.name,
			ActorEmail:    actor.email,
			ProcessMethod: actor.processMethod,
			Changes:       changes,
		}
		if err := h.history.RecordContentChanges(ctx, entry); err != nil {
			// 履歴記録の失敗によってメインの更新処理をロールバックしてはいけない。
			h.logger.Entry(ctx).WithError(err).Error("historyAwareUserRepository: failed to record content change history")
		}
	}

	return nil
}
