package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
)

// MIR-1420: SMS認証の確定で、書き換える相手と照合する会員をトークンの会員にする。本文の user_id に他の会員を入れても、
// その会員の電話番号は書き換わらない(他の会員の申請はトークンの会員とは照合しない)。

func newSmsVerificationTestEngine(t *testing.T) (*xorm.Engine, context.Context) {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync(
		new(domain.User), new(domain.Role), new(domain.CorporateProfile), new(domain.CorporateShareholder), new(domain.UserFile),
		new(domain.UserVirtualAccount), new(domain.UserAntisocialCheck), new(domain.UserAccount),
		new(domain.AntisocialCheckFile), new(domain.MobilePhoneVerificationRequest),
	))
	return engine, context.WithValue(context.Background(), factory.ConextDBType, engine)
}

func insertSmsTestUser(t *testing.T, engine *xorm.Engine, email, phone string) domain.User {
	t.Helper()
	u := domain.User{Uid: "uid-" + email, Email: email, MobilePhone: phone}
	_, err := engine.Insert(&u)
	require.NoError(t, err)
	return u
}

// insertOthersVerification は、会員 ownerId が申請したSMS認証を作る(コードも分かっている前提で、相手の決め方だけを見る)。
func insertOthersVerification(t *testing.T, engine *xorm.Engine, ownerId int64, phoneType domain.MOBILE_PHONE_TYPE) domain.MobilePhoneVerificationRequest {
	t.Helper()
	v := domain.MobilePhoneVerificationRequest{
		RequestId: "req-" + strconv.FormatInt(ownerId, 10), UserId: ownerId, Code: "246810", MobilePhone: "09099999999",
		PhoneType: phoneType, FirstReq: true, Expire: jsonutils.JSONTime(time.Now().Add(10 * time.Minute)),
	}
	_, err := engine.Insert(&v)
	require.NoError(t, err)
	return v
}

func newSmsConfirmContext(ctx context.Context, tokenUserId, bodyUserId int64, v domain.MobilePhoneVerificationRequest) echo.Context {
	e := echo.New()
	e.Validator = validator.NewCustomValidator()
	body := `{"user_id":` + strconv.FormatInt(bodyUserId, 10) + `,"request_id":"` + v.RequestId + `","code":"` + v.Code + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)).WithContext(ctx)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())
	c.Set("user", jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": strconv.FormatInt(tokenUserId, 10), "email": "attacker@example.com",
	}))
	return c
}

func assertFailSmsCert(t *testing.T, err error) {
	t.Helper()
	var fe *factory.Error
	require.True(t, errors.As(err, &fe), "err = %v", err)
	assert.Equal(t, factory.ErrorFailSmsCert.Code, fe.Code, fe.Details)
}

func TestConfirmSmsVerification_UsesTokenUserNotBody(t *testing.T) {
	engine, ctx := newSmsVerificationTestEngine(t)
	attacker := insertSmsTestUser(t, engine, "attacker@example.com", "09011111111")
	victim := insertSmsTestUser(t, engine, "victim@example.com", "09022222222")
	v := insertOthersVerification(t, engine, victim.Id, domain.MOBILE_PHONE_TYPE("DEFAULT"))

	c := newSmsConfirmContext(ctx, attacker.Id, victim.Id, v)
	err := NewUserApiController().ConfirmSmsVerification(c)
	assertFailSmsCert(t, err)

	var got domain.User
	_, err = engine.ID(victim.Id).Get(&got)
	require.NoError(t, err)
	assert.Equal(t, "09022222222", got.MobilePhone, "本文の user_id の会員の電話番号は変わらない")
	assert.False(t, got.IsSmsCert)
}

func TestCorporateMobilePhoneVerificationConfirm_UsesTokenUserNotBody(t *testing.T) {
	engine, ctx := newSmsVerificationTestEngine(t)
	attacker := insertSmsTestUser(t, engine, "attacker@example.com", "09011111111")
	victim := insertSmsTestUser(t, engine, "victim@example.com", "09022222222")
	for _, id := range []int64{attacker.Id, victim.Id} {
		empty := ""
		_, err := engine.Insert(&domain.CorporateProfile{
			UserId: id, CorpNo: "100000000000" + strconv.FormatInt(id, 10), RepresentativeMobilePhone: "0903333333" + strconv.FormatInt(id, 10),
			ManagerFamilyName: &empty, ManagerFirstName: &empty, ManagerFamilyNameKana: &empty, ManagerFirstNameKana: &empty,
			ManagerEmail: &empty, ManagerMobilePhone: &empty, Department: &empty, Position: &empty,
			ManagerBirthday: jsonutils.JSONTime(time.Now()),
		})
		require.NoError(t, err)
	}
	_, err := engine.Insert(&domain.UserAccount{UserId: attacker.Id}, &domain.UserAccount{UserId: victim.Id})
	require.NoError(t, err)
	v := insertOthersVerification(t, engine, victim.Id, domain.MOBILE_PHONE_TYPE("CEO"))

	c := newSmsConfirmContext(ctx, attacker.Id, victim.Id, v)
	c.SetParamNames("id", "corporation_type")
	c.SetParamValues(strconv.FormatInt(attacker.Id, 10), string(domain.REPRESENTATIVE))
	err = NewCorporateApiController().MobilePhoneVerificationConfirm(c)
	assertFailSmsCert(t, err)

	var got domain.CorporateProfile
	_, err = engine.Where("user_id = ?", victim.Id).Get(&got)
	require.NoError(t, err)
	assert.Equal(t, "0903333333"+strconv.FormatInt(victim.Id, 10), got.RepresentativeMobilePhone, "本文の user_id の法人の代表者の電話番号は変わらない")
}

// 正規の操作: トークンの会員が自分の申請を確定すると、本文の user_id に関係なく本人の番号が書き換わる。
func TestConfirmSmsVerification_OwnRequestUpdatesTokenUser(t *testing.T) {
	engine, ctx := newSmsVerificationTestEngine(t)
	require.NoError(t, engine.Sync(new(domain.MailRetry)))
	me := insertSmsTestUser(t, engine, "me@example.com", "09011111111")
	v := insertOthersVerification(t, engine, me.Id, domain.DEFAULT_PHONE)

	c := newSmsConfirmContext(ctx, me.Id, 0, v)
	require.NoError(t, NewUserApiController().ConfirmSmsVerification(c), "メールの送信の失敗はログだけで、応答は成功")

	var got domain.User
	_, err := engine.ID(me.Id).Get(&got)
	require.NoError(t, err)
	assert.Equal(t, v.MobilePhone, got.MobilePhone)
	assert.True(t, got.IsSmsCert)
}

func TestCorporateMobilePhoneVerificationConfirm_OwnRequestUpdatesTokenUser(t *testing.T) {
	for _, tt := range []struct {
		corpType  domain.CORP_TYPE
		phoneType domain.MOBILE_PHONE_TYPE
	}{{domain.REPRESENTATIVE, domain.CEO_PHONE}, {domain.MANAGER, domain.MANAGER_PHONE}} {
		t.Run(string(tt.corpType), func(t *testing.T) {
			engine, ctx := newSmsVerificationTestEngine(t)
			me := insertSmsTestUser(t, engine, "me@example.com", "09011111111")
			empty := ""
			_, err := engine.Insert(&domain.CorporateProfile{
				UserId: me.Id, CorpNo: "1000000000001", RepresentativeMobilePhone: "09033333333",
				ManagerFamilyName: &empty, ManagerFirstName: &empty, ManagerFamilyNameKana: &empty, ManagerFirstNameKana: &empty,
				ManagerEmail: &empty, ManagerMobilePhone: &empty, Department: &empty, Position: &empty,
				ManagerBirthday: jsonutils.JSONTime(time.Now()),
			}, &domain.UserAccount{UserId: me.Id})
			require.NoError(t, err)
			v := insertOthersVerification(t, engine, me.Id, tt.phoneType)

			c := newSmsConfirmContext(ctx, me.Id, 999999, v)
			c.SetParamNames("id", "corporation_type")
			c.SetParamValues(strconv.FormatInt(me.Id, 10), string(tt.corpType))
			require.NoError(t, NewCorporateApiController().MobilePhoneVerificationConfirm(c))

			var got domain.CorporateProfile
			_, err = engine.Where("user_id = ?", me.Id).Get(&got)
			require.NoError(t, err)
			if tt.corpType == domain.REPRESENTATIVE {
				assert.Equal(t, v.MobilePhone, got.RepresentativeMobilePhone)
			} else {
				require.NotNil(t, got.ManagerMobilePhone)
				assert.Equal(t, v.MobilePhone, *got.ManagerMobilePhone)
			}
		})
	}
}

// staging では固定のコードが効くので、攻撃者は自分の申請と固定のコードで通せる。本文の user_id に他の会員を入れても、
// 書き換わるのは攻撃者自身の番号だけで、他の会員の番号は変わらない(staging でこれを止めているのはトークンの会員を使う直し)。
func TestSmsVerificationConfirm_FixedCodeInStagingDoesNotReachOthers(t *testing.T) {
	prev := serviceutils.GetEnv()
	serviceutils.SetEnv("stg")
	t.Cleanup(func() { serviceutils.SetEnv(prev) })

	t.Run("個人", func(t *testing.T) {
		engine, ctx := newSmsVerificationTestEngine(t)
		require.NoError(t, engine.Sync(new(domain.MailRetry)))
		attacker := insertSmsTestUser(t, engine, "attacker@example.com", "09011111111")
		victim := insertSmsTestUser(t, engine, "victim@example.com", "09022222222")
		v := insertOthersVerification(t, engine, attacker.Id, domain.DEFAULT_PHONE)
		v.Code = domain.FixedSmsCode

		require.NoError(t, NewUserApiController().ConfirmSmsVerification(newSmsConfirmContext(ctx, attacker.Id, victim.Id, v)))
		var got domain.User
		_, err := engine.ID(victim.Id).Get(&got)
		require.NoError(t, err)
		assert.Equal(t, "09022222222", got.MobilePhone)
		assert.False(t, got.IsSmsCert)
	})

	t.Run("法人", func(t *testing.T) {
		engine, ctx := newSmsVerificationTestEngine(t)
		attacker := insertSmsTestUser(t, engine, "attacker@example.com", "09011111111")
		victim := insertSmsTestUser(t, engine, "victim@example.com", "09022222222")
		empty := ""
		for i, id := range []int64{attacker.Id, victim.Id} {
			_, err := engine.Insert(&domain.CorporateProfile{
				UserId: id, CorpNo: "100000000000" + strconv.Itoa(i), RepresentativeMobilePhone: "0903333333" + strconv.Itoa(i),
				ManagerFamilyName: &empty, ManagerFirstName: &empty, ManagerFamilyNameKana: &empty, ManagerFirstNameKana: &empty,
				ManagerEmail: &empty, ManagerMobilePhone: &empty, Department: &empty, Position: &empty,
				ManagerBirthday: jsonutils.JSONTime(time.Now()),
			}, &domain.UserAccount{UserId: id})
			require.NoError(t, err)
		}
		v := insertOthersVerification(t, engine, attacker.Id, domain.CEO_PHONE)
		v.Code = domain.FixedSmsCode

		c := newSmsConfirmContext(ctx, attacker.Id, victim.Id, v)
		c.SetParamNames("id", "corporation_type")
		c.SetParamValues(strconv.FormatInt(victim.Id, 10), string(domain.REPRESENTATIVE))
		require.NoError(t, NewCorporateApiController().MobilePhoneVerificationConfirm(c))
		var got domain.CorporateProfile
		_, err := engine.Where("user_id = ?", victim.Id).Get(&got)
		require.NoError(t, err)
		assert.Equal(t, "09033333331", got.RepresentativeMobilePhone)
	})
}
