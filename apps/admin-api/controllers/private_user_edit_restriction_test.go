package controllers

import (
	"testing"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/stretchr/testify/assert"
)

func TestRestrictPersonalInfoEditsAfterFinalAudit(t *testing.T) {
	beforeBirthday := jsonutils.JSONTime(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC))
	newBirthday := jsonutils.JSONTime(time.Date(1991, 3, 4, 0, 0, 0, 0, time.UTC))

	newBeforeUser := func(auditStatus domain.AUDIT_STATUS) *domain.PrivateUser {
		return &domain.PrivateUser{
			User: domain.User{
				AuditStatus: auditStatus,
				MobilePhone: "09000000000",
			},
			PrivateProfile: domain.PrivateProfile{
				Nationality: domain.JP,
				Gender:      domain.MALE,
				Birthday:    beforeBirthday,
			},
		}
	}
	newRequestedUser := func() *domain.PrivateUser {
		return &domain.PrivateUser{
			User: domain.User{
				MobilePhone: "08011112222",
			},
			PrivateProfile: domain.PrivateProfile{
				Nationality: domain.OTHERS,
				Gender:      domain.FEMALE,
				Birthday:    newBirthday,
			},
		}
	}

	tests := []struct {
		name        string
		auditStatus domain.AUDIT_STATUS
		wantChanged bool
	}{
		{name: "本登録承認完了は変更を戻す", auditStatus: domain.COMPLETED_AUDIT, wantChanged: false},
		{name: "無効は変更を戻す", auditStatus: domain.INVALIDITY_AUDIT, wantChanged: false},
		{name: "仮登録は変更を通す", auditStatus: domain.PROVISIONAL_AUDIT, wantChanged: true},
		{name: "本登録承認待機は変更を通す", auditStatus: domain.WAIT_AUDIT, wantChanged: true},
		{name: "本人確認待ちは変更を通す", auditStatus: domain.WAIT_IDENTITY_REGISTER_AUDIT, wantChanged: true},
		{name: "最終承認待機は変更を通す", auditStatus: domain.WAIT_FINAL_APPROVAL_AUDIT, wantChanged: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeUser := newBeforeUser(tt.auditStatus)
			user := newRequestedUser()

			restrictPersonalInfoEditsAfterFinalAudit(user, beforeUser)

			if tt.wantChanged {
				assert.Equal(t, domain.OTHERS, user.PrivateProfile.Nationality)
				assert.Equal(t, domain.FEMALE, user.PrivateProfile.Gender)
				assert.Equal(t, newBirthday, user.PrivateProfile.Birthday)
				assert.Equal(t, "08011112222", user.User.MobilePhone)
			} else {
				assert.Equal(t, domain.JP, user.PrivateProfile.Nationality)
				assert.Equal(t, domain.MALE, user.PrivateProfile.Gender)
				assert.Equal(t, beforeBirthday, user.PrivateProfile.Birthday)
				assert.Equal(t, "09000000000", user.User.MobilePhone)
			}
		})
	}
}
