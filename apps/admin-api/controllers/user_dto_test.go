package controllers

import (
	"strings"
	"testing"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 口座番号は金融機関を問わず7桁。ゆうちょも通帳の番号(8桁)ではなく振込用の
// 7桁で登録する。管理画面は口座が未登録の会員も編集するため、空は通し、
// 入力があるときだけ桁数を求める。
//
// UserInput は他にも必須項目が多いため、構造体全体の可否ではなく
// 口座番号のエラーが出たかどうかで判定する。
func TestUserInput_BankAccountValidation(t *testing.T) {
	tests := []struct {
		name        string
		bankAccount string
		wantErr     bool
	}{
		{
			name:        "(成功) 7桁は通る",
			bankAccount: "1234567",
			wantErr:     false,
		},
		{
			name:        "(成功) 口座未登録を想定し空は通る",
			bankAccount: "",
			wantErr:     false,
		},
		{
			name:        "(失敗) ゆうちょの通帳番号8桁は弾く",
			bankAccount: "12345678",
			wantErr:     true,
		},
		{
			name:        "(失敗) 数字以外を含むものは弾く",
			bankAccount: "123456a",
			wantErr:     true,
		},
		{
			// numeric タグは符号や小数点を通すため number を使う。
			name:        "(失敗) 符号付きは弾く",
			bankAccount: "-123456",
			wantErr:     true,
		},
	}

	cv := validator.NewCustomValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := UserInput{BankAccount: tt.bankAccount}

			err := cv.Validate(&input)

			hasBankAccountError := err != nil &&
				strings.Contains(err.Error(), "UserInput.BankAccount")

			assert.Equal(t, tt.wantErr, hasBankAccountError)
		})
	}
}

// MIR-1173: 国籍・性別・生年月日・電話番号をToModel()経由でdomainへ反映する。
func TestUserInput_ToModel_PersonalInfoFields(t *testing.T) {
	input := UserInput{
		FirstName:   "太郎",
		FamilyName:  "山田",
		MobilePhone: "09011112222",
		Nationality: domain.JP,
		Gender:      domain.MALE,
		Birthday:    "1990-01-02 00:00:00",
	}

	user, err := input.ToModel()
	require.NoError(t, err)

	wantBirthday, err := jsonutils.ParseJSONTimeJST(time.DateTime, "1990-01-02 00:00:00")
	require.NoError(t, err)

	assert.Equal(t, "09011112222", user.User.MobilePhone)
	assert.Equal(t, domain.JP, user.PrivateProfile.Nationality)
	assert.Equal(t, domain.MALE, user.PrivateProfile.Gender)
	assert.Equal(t, wantBirthday, user.PrivateProfile.Birthday)
}

func TestUserInput_ToModel_InvalidBirthdayReturnsError(t *testing.T) {
	input := UserInput{
		FirstName:  "太郎",
		FamilyName: "山田",
		Birthday:   "not-a-date",
	}

	_, err := input.ToModel()

	assert.Error(t, err)
}

// MIR-1173: 生年月日未設定の会員(本登録承認完了・無効でdisabledのため
// 自分では直せない)を編集する際、空文字のままエラーにすると更新自体が
// 一切できなくなるため、空はゼロ値として通す。
func TestUserInput_ToModel_EmptyBirthdayIsZeroValueNotError(t *testing.T) {
	input := UserInput{
		FirstName:  "太郎",
		FamilyName: "山田",
		Birthday:   "",
	}

	user, err := input.ToModel()

	require.NoError(t, err)
	assert.Equal(t, jsonutils.JSONTime{}, user.PrivateProfile.Birthday)
}

// 法人会員も同様に、口座が未登録の状態で編集することがある。
func TestCorporateUserInput_BankAccountValidation(t *testing.T) {
	tests := []struct {
		name        string
		bankAccount string
		wantErr     bool
	}{
		{
			name:        "(成功) 7桁は通る",
			bankAccount: "1234567",
			wantErr:     false,
		},
		{
			name:        "(成功) 口座未登録を想定し空は通る",
			bankAccount: "",
			wantErr:     false,
		},
		{
			name:        "(失敗) ゆうちょの通帳番号8桁は弾く",
			bankAccount: "12345678",
			wantErr:     true,
		},
		{
			name:        "(失敗) 数字以外を含むものは弾く",
			bankAccount: "123456a",
			wantErr:     true,
		},
	}

	cv := validator.NewCustomValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := CorporateUserInput{BankAccount: tt.bankAccount}

			err := cv.Validate(&input)

			hasBankAccountError := err != nil &&
				strings.Contains(err.Error(), "CorporateUserInput.BankAccount")

			assert.Equal(t, tt.wantErr, hasBankAccountError)
		})
	}
}
