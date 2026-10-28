package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/actorcontext"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

// historyPrevUser は変更前の会員です。
func historyPrevUser() domain.User {
	return domain.User{
		Id:           1001,
		Name:         "山田太郎",
		Email:        "old@example.com",
		Password:     "oldhash",
		Salt:         "oldsalt",
		IsNewsletter: true,
		AuditStatus:  domain.PROVISIONAL_AUDIT,
	}
}

// historyCurrUser は変更前の会員に modify を適用した変更後の会員を返します。
func historyCurrUser(modify func(u *domain.User)) domain.User {
	u := historyPrevUser()
	modify(&u)
	return u
}

// 会員本人のログイン情報（個人の PatchOne などで WithActor に設定される）。
// DB の値と区別できるよう、氏名とメールアドレスを変えている。
var historyMemberActor = jwthelper.User{
	Id: 1001, Name: "山田太郎(JWT)", Email: "jwt@example.com",
	Roles: map[string]int{string(domain.ROLE_USER): 1},
}

// 管理画面の運用者。
var historyOperatorActor = jwthelper.User{
	Id: 9, Name: "管理者", Email: "admin@example.com",
	Roles: map[string]int{string(domain.ROLE_ADMIN): 1},
}

func Test_historyAwareUserRepository_Update(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		curr        domain.User
		userRepo    func(t *testing.T, curr domain.User) *mocks.UserRepository
		historyRepo func(t *testing.T) *mocks.UserInfoHistoryRepository
		wantErr     bool
	}{
		{
			name: "アクター無しのメール変更は会員本人・メンバー直接処理として記録し、処理者メールは保存後の値になる",
			ctx:  context.Background(),
			curr: historyCurrUser(func(u *domain.User) { u.Email = "new@example.com" }),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(historyPrevUser(), nil)
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				m := mocks.NewUserInfoHistoryRepository(t)
				m.On("RecordContentChanges", mock.Anything, cdomain.ContentChangeEntry{
					UserId:        1001,
					ActorId:       1001,
					ActorName:     "山田太郎",
					ActorEmail:    "new@example.com",
					ProcessMethod: domain.USER_HISTORY_PROCESS_MEMBER,
					Changes: []cdomain.ContentChange{
						{ContentType: domain.USER_CHANGED_CONTENT_EMAIL, Prev: "old@example.com", Curr: "new@example.com"},
					},
				}).Return(nil)
				return m
			},
		},
		{
			name: "パスワード変更はハッシュ値を残さずマスク値で記録する",
			ctx:  context.Background(),
			curr: historyCurrUser(func(u *domain.User) { u.Password = "newhash"; u.Salt = "newsalt" }),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(historyPrevUser(), nil)
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				m := mocks.NewUserInfoHistoryRepository(t)
				m.On("RecordContentChanges", mock.Anything, cdomain.ContentChangeEntry{
					UserId:        1001,
					ActorId:       1001,
					ActorName:     "山田太郎",
					ActorEmail:    "old@example.com",
					ProcessMethod: domain.USER_HISTORY_PROCESS_MEMBER,
					Changes: []cdomain.ContentChange{
						{ContentType: domain.USER_CHANGED_CONTENT_PASSWORD, Prev: "********", Curr: "********"},
					},
				}).Return(nil)
				return m
			},
		},
		{
			name: "ニュースレターの停止は はい → いいえ で記録する",
			ctx:  context.Background(),
			curr: historyCurrUser(func(u *domain.User) { u.IsNewsletter = false }),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(historyPrevUser(), nil)
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				m := mocks.NewUserInfoHistoryRepository(t)
				m.On("RecordContentChanges", mock.Anything, cdomain.ContentChangeEntry{
					UserId:        1001,
					ActorId:       1001,
					ActorName:     "山田太郎",
					ActorEmail:    "old@example.com",
					ProcessMethod: domain.USER_HISTORY_PROCESS_MEMBER,
					Changes: []cdomain.ContentChange{
						{
							ContentType: domain.USER_CHANGED_CONTENT_NEWSLETTER,
							Prev:        string(domain.USER_HISTORY_CHECK_YES),
							Curr:        string(domain.USER_HISTORY_CHECK_NO),
						},
					},
				}).Return(nil)
				return m
			},
		},
		{
			name: "ニュースレターの再登録は いいえ → はい で記録する",
			ctx:  context.Background(),
			curr: historyCurrUser(func(u *domain.User) { u.IsNewsletter = true }),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				prev := historyPrevUser()
				prev.IsNewsletter = false
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(prev, nil)
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				m := mocks.NewUserInfoHistoryRepository(t)
				m.On("RecordContentChanges", mock.Anything, cdomain.ContentChangeEntry{
					UserId:        1001,
					ActorId:       1001,
					ActorName:     "山田太郎",
					ActorEmail:    "old@example.com",
					ProcessMethod: domain.USER_HISTORY_PROCESS_MEMBER,
					Changes: []cdomain.ContentChange{
						{
							ContentType: domain.USER_CHANGED_CONTENT_NEWSLETTER,
							Prev:        string(domain.USER_HISTORY_CHECK_NO),
							Curr:        string(domain.USER_HISTORY_CHECK_YES),
						},
					},
				}).Return(nil)
				return m
			},
		},
		{
			name: "複数項目の変更はメール・パスワード・ニュースレターの順に1件の履歴にまとめる",
			ctx:  context.Background(),
			curr: historyCurrUser(func(u *domain.User) {
				u.Email = "new@example.com"
				u.Password = "newhash"
				u.IsNewsletter = false
			}),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(historyPrevUser(), nil)
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				m := mocks.NewUserInfoHistoryRepository(t)
				m.On("RecordContentChanges", mock.Anything, cdomain.ContentChangeEntry{
					UserId:        1001,
					ActorId:       1001,
					ActorName:     "山田太郎",
					ActorEmail:    "new@example.com",
					ProcessMethod: domain.USER_HISTORY_PROCESS_MEMBER,
					Changes: []cdomain.ContentChange{
						{ContentType: domain.USER_CHANGED_CONTENT_EMAIL, Prev: "old@example.com", Curr: "new@example.com"},
						{ContentType: domain.USER_CHANGED_CONTENT_PASSWORD, Prev: "********", Curr: "********"},
						{
							ContentType: domain.USER_CHANGED_CONTENT_NEWSLETTER,
							Prev:        string(domain.USER_HISTORY_CHECK_YES),
							Curr:        string(domain.USER_HISTORY_CHECK_NO),
						},
					},
				}).Return(nil).Once()
				return m
			},
		},
		{
			name: "会員本人のアクターがいる場合は JWT ではなく会員の行の氏名・メールをメンバー直接処理として記録する",
			ctx:  actorcontext.WithActor(context.Background(), historyMemberActor),
			curr: historyCurrUser(func(u *domain.User) { u.IsNewsletter = false }),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(historyPrevUser(), nil)
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				m := mocks.NewUserInfoHistoryRepository(t)
				m.On("RecordContentChanges", mock.Anything, cdomain.ContentChangeEntry{
					UserId:        1001,
					ActorId:       1001,
					ActorName:     "山田太郎",
					ActorEmail:    "old@example.com",
					ProcessMethod: domain.USER_HISTORY_PROCESS_MEMBER,
					Changes: []cdomain.ContentChange{
						{
							ContentType: domain.USER_CHANGED_CONTENT_NEWSLETTER,
							Prev:        string(domain.USER_HISTORY_CHECK_YES),
							Curr:        string(domain.USER_HISTORY_CHECK_NO),
						},
					},
				}).Return(nil)
				return m
			},
		},
		{
			name: "会員本人のアクターがメールを変えた場合は処理者メールを JWT の旧アドレスではなく変更後のアドレスで記録する",
			ctx:  actorcontext.WithActor(context.Background(), historyMemberActor),
			curr: historyCurrUser(func(u *domain.User) { u.Email = "new@example.com" }),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(historyPrevUser(), nil)
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				m := mocks.NewUserInfoHistoryRepository(t)
				m.On("RecordContentChanges", mock.Anything, cdomain.ContentChangeEntry{
					UserId:        1001,
					ActorId:       1001,
					ActorName:     "山田太郎",
					ActorEmail:    "new@example.com",
					ProcessMethod: domain.USER_HISTORY_PROCESS_MEMBER,
					Changes: []cdomain.ContentChange{
						{ContentType: domain.USER_CHANGED_CONTENT_EMAIL, Prev: "old@example.com", Curr: "new@example.com"},
					},
				}).Return(nil)
				return m
			},
		},
		{
			name: "会員ロールでも本人以外のアクターの場合は JWT の情報をメンバー直接処理として記録する",
			ctx: actorcontext.WithActor(context.Background(), jwthelper.User{
				Id: 2002, Name: "別会員", Email: "other@example.com",
				Roles: map[string]int{string(domain.ROLE_USER): 1},
			}),
			curr: historyCurrUser(func(u *domain.User) { u.IsNewsletter = false }),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(historyPrevUser(), nil)
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				m := mocks.NewUserInfoHistoryRepository(t)
				m.On("RecordContentChanges", mock.Anything, cdomain.ContentChangeEntry{
					UserId:        1001,
					ActorId:       2002,
					ActorName:     "別会員",
					ActorEmail:    "other@example.com",
					ProcessMethod: domain.USER_HISTORY_PROCESS_MEMBER,
					Changes: []cdomain.ContentChange{
						{
							ContentType: domain.USER_CHANGED_CONTENT_NEWSLETTER,
							Prev:        string(domain.USER_HISTORY_CHECK_YES),
							Curr:        string(domain.USER_HISTORY_CHECK_NO),
						},
					},
				}).Return(nil)
				return m
			},
		},
		{
			name: "運用者による審査ステータス変更は運用者処理として記録する",
			ctx:  actorcontext.WithActor(context.Background(), historyOperatorActor),
			curr: historyCurrUser(func(u *domain.User) { u.AuditStatus = domain.INVALIDITY_AUDIT }),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(historyPrevUser(), nil)
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				m := mocks.NewUserInfoHistoryRepository(t)
				m.On("RecordAuditStatusChange", mock.Anything, cdomain.AuditStatusChangeEntry{
					UserId:        1001,
					ActorId:       9,
					ActorName:     "管理者",
					ActorEmail:    "admin@example.com",
					ProcessMethod: domain.USER_HISTORY_PROCESS_OPERATOR,
					PrevStatus:    domain.PROVISIONAL_AUDIT,
					CurrStatus:    domain.INVALIDITY_AUDIT,
				}).Return(nil)
				return m
			},
		},
		{
			name: "アクター無しの審査ステータス変更は処理者を会員本人の氏名・メールで記録する",
			ctx:  context.Background(),
			curr: historyCurrUser(func(u *domain.User) { u.AuditStatus = domain.INVALIDITY_AUDIT }),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(historyPrevUser(), nil)
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				m := mocks.NewUserInfoHistoryRepository(t)
				m.On("RecordAuditStatusChange", mock.Anything, cdomain.AuditStatusChangeEntry{
					UserId:        1001,
					ActorId:       1001,
					ActorName:     "山田太郎",
					ActorEmail:    "old@example.com",
					ProcessMethod: domain.USER_HISTORY_PROCESS_MEMBER,
					PrevStatus:    domain.PROVISIONAL_AUDIT,
					CurrStatus:    domain.INVALIDITY_AUDIT,
				}).Return(nil)
				return m
			},
		},
		{
			name: "審査ステータスと内容が同じ Update で変わった場合は履歴を2件記録する",
			ctx:  actorcontext.WithActor(context.Background(), historyOperatorActor),
			curr: historyCurrUser(func(u *domain.User) {
				u.AuditStatus = domain.INVALIDITY_AUDIT
				u.Email = "new@example.com"
			}),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(historyPrevUser(), nil)
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				m := mocks.NewUserInfoHistoryRepository(t)
				m.On("RecordAuditStatusChange", mock.Anything, cdomain.AuditStatusChangeEntry{
					UserId:        1001,
					ActorId:       9,
					ActorName:     "管理者",
					ActorEmail:    "admin@example.com",
					ProcessMethod: domain.USER_HISTORY_PROCESS_OPERATOR,
					PrevStatus:    domain.PROVISIONAL_AUDIT,
					CurrStatus:    domain.INVALIDITY_AUDIT,
				}).Return(nil).Once()
				m.On("RecordContentChanges", mock.Anything, cdomain.ContentChangeEntry{
					UserId:        1001,
					ActorId:       9,
					ActorName:     "管理者",
					ActorEmail:    "admin@example.com",
					ProcessMethod: domain.USER_HISTORY_PROCESS_OPERATOR,
					Changes: []cdomain.ContentChange{
						{ContentType: domain.USER_CHANGED_CONTENT_EMAIL, Prev: "old@example.com", Curr: "new@example.com"},
					},
				}).Return(nil).Once()
				return m
			},
		},
		{
			name: "追跡対象外の項目だけの変更（ログイン等）は履歴を記録しない",
			ctx:  context.Background(),
			curr: historyCurrUser(func(u *domain.User) { u.Name = "山田花子" }),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(historyPrevUser(), nil)
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				return mocks.NewUserInfoHistoryRepository(t)
			},
		},
		{
			name: "空の Email・Password・AuditStatus は XORM で更新されないため変更として扱わない",
			ctx:  context.Background(),
			curr: historyCurrUser(func(u *domain.User) {
				u.Email = ""
				u.Password = ""
				u.AuditStatus = ""
			}),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(historyPrevUser(), nil)
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				return mocks.NewUserInfoHistoryRepository(t)
			},
		},
		{
			name: "審査ステータスの履歴記録に失敗しても内容変更の履歴は記録し、Update は成功する",
			ctx:  actorcontext.WithActor(context.Background(), historyOperatorActor),
			curr: historyCurrUser(func(u *domain.User) {
				u.AuditStatus = domain.INVALIDITY_AUDIT
				u.Email = "new@example.com"
			}),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(historyPrevUser(), nil)
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				m := mocks.NewUserInfoHistoryRepository(t)
				m.On("RecordAuditStatusChange", mock.Anything, cdomain.AuditStatusChangeEntry{
					UserId:        1001,
					ActorId:       9,
					ActorName:     "管理者",
					ActorEmail:    "admin@example.com",
					ProcessMethod: domain.USER_HISTORY_PROCESS_OPERATOR,
					PrevStatus:    domain.PROVISIONAL_AUDIT,
					CurrStatus:    domain.INVALIDITY_AUDIT,
				}).Return(errors.New("db down"))
				m.On("RecordContentChanges", mock.Anything, cdomain.ContentChangeEntry{
					UserId:        1001,
					ActorId:       9,
					ActorName:     "管理者",
					ActorEmail:    "admin@example.com",
					ProcessMethod: domain.USER_HISTORY_PROCESS_OPERATOR,
					Changes: []cdomain.ContentChange{
						{ContentType: domain.USER_CHANGED_CONTENT_EMAIL, Prev: "old@example.com", Curr: "new@example.com"},
					},
				}).Return(nil).Once()
				return m
			},
		},
		{
			name: "内容変更の履歴記録に失敗しても Update は成功する",
			ctx:  context.Background(),
			curr: historyCurrUser(func(u *domain.User) { u.Email = "new@example.com" }),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(historyPrevUser(), nil)
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				m := mocks.NewUserInfoHistoryRepository(t)
				m.On("RecordContentChanges", mock.Anything, cdomain.ContentChangeEntry{
					UserId:        1001,
					ActorId:       1001,
					ActorName:     "山田太郎",
					ActorEmail:    "new@example.com",
					ProcessMethod: domain.USER_HISTORY_PROCESS_MEMBER,
					Changes: []cdomain.ContentChange{
						{ContentType: domain.USER_CHANGED_CONTENT_EMAIL, Prev: "old@example.com", Curr: "new@example.com"},
					},
				}).Return(errors.New("db down"))
				return m
			},
		},
		{
			name: "本体の Update が失敗した場合はエラーを返し、履歴は記録しない",
			ctx:  context.Background(),
			curr: historyCurrUser(func(u *domain.User) { u.Email = "new@example.com" }),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(historyPrevUser(), nil)
				m.On("Update", mock.Anything, curr).Return(errors.New("update failed"))
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				return mocks.NewUserInfoHistoryRepository(t)
			},
			wantErr: true,
		},
		{
			name: "変更前の取得に失敗しても Update は実行し、履歴は記録しない",
			ctx:  context.Background(),
			curr: historyCurrUser(func(u *domain.User) { u.Email = "new@example.com" }),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(domain.User{}, errors.New("select failed"))
				m.On("Update", mock.Anything, curr).Return(nil)
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				return mocks.NewUserInfoHistoryRepository(t)
			},
		},
		{
			name: "変更前の取得と本体の Update が両方失敗した場合はエラーを返し、履歴は記録しない",
			ctx:  context.Background(),
			curr: historyCurrUser(func(u *domain.User) { u.Email = "new@example.com" }),
			userRepo: func(t *testing.T, curr domain.User) *mocks.UserRepository {
				m := mocks.NewUserRepository(t)
				m.On("GetById", mock.Anything, int64(1001)).Return(domain.User{}, errors.New("select failed"))
				m.On("Update", mock.Anything, curr).Return(errors.New("update failed"))
				return m
			},
			historyRepo: func(t *testing.T) *mocks.UserInfoHistoryRepository {
				return mocks.NewUserInfoHistoryRepository(t)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mocks.NewXxx(t) は終了時に期待した呼び出しを検証し、期待していない呼び出しはテストを失敗させる
			repo := NewHistoryAware(tt.userRepo(t, tt.curr), tt.historyRepo(t), factory.NewLoggerFactory())

			err := repo.Update(tt.ctx, tt.curr)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

// MIR-1373: UpdateLastLoginDateはaudit_statusを触らないため、Updateと違い
// 履歴記録(UserInfoHistoryRepository)を一切呼ばずに内部リポジトリへ委譲する。
func TestHistoryAwareUserRepository_UpdateLastLoginDate(t *testing.T) {
	inner := &mocks.UserRepository{}
	history := &mocks.UserInfoHistoryRepository{}
	logger := &mocks.LoggerFactory{}

	lastLoginDate := jsonutils.JSONTime(time.Now())
	inner.On("UpdateLastLoginDate", mock.Anything, int64(1000), lastLoginDate).Return(nil)

	repo := NewHistoryAware(inner, history, logger)
	err := repo.UpdateLastLoginDate(context.Background(), 1000, lastLoginDate)

	assert.NoError(t, err)
	inner.AssertExpectations(t)
	history.AssertNotCalled(t, "RecordAuditStatusChange", mock.Anything, mock.Anything)
}
