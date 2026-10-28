package controllers

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
)

func TestUserWaitlistResponseItem_StatusSwaggerEnum(t *testing.T) {
	field, ok := reflect.TypeOf(UserWaitlistResponseItem{}).FieldByName("Status")
	require.True(t, ok)

	assert.Contains(t, field.Tag.Get("swagger"), "enum(WAITING|END_FUNDING|CANCELLED|WINNING|WINNING_AFTER_SUCCESS|WINNING_AFTER_CANCELED|WINNING_AFTER_FAILED|PROPOSAL)")
}

func TestNewUserWaitlistResponseItem(t *testing.T) {
	waitingDate := jsonutils.JSONTime(time.Date(2026, time.July, 13, 12, 0, 0, 0, time.UTC))

	tests := []struct {
		name   string
		waiter domain.InvestmentWaiters
		want   UserWaitlistResponseItem
	}{
		{
			name: "Share作成済みの場合は全項目を設定する",
			waiter: domain.InvestmentWaiters{
				Id:                 100,
				ProjectId:          200,
				ProjectName:        "test project",
				Qty:                3,
				Amount:             30000,
				Status:             domain.WAITING_END,
				DepositCheckStatus: domain.OK_DEPOSIT,
				InvestStatus:       investStatusPtr(domain.SUCCESS_INVEST),
				ProjectStatus:      projectStatusPtr(domain.SUCCESS_FUNDING),
				WaitingDate:        waitingDate,
				ShareId:            "share-1",
			},
			want: UserWaitlistResponseItem{
				Id:            100,
				ProjectId:     200,
				ProjectName:   "test project",
				Qty:           3,
				Amount:        30000,
				Status:        domain.INVESTMENT_WAITING_STATUS_WINNING_AFTER_SUCCESS,
				InvestStatus:  investStatusPtr(domain.SUCCESS_INVEST),
				ProjectStatus: projectStatusPtr(domain.SUCCESS_FUNDING),
				WaitingDate:   waitingDate,
				ShareId:       "share-1",
			},
		},
		{
			name: "ステータス未設定の場合はnilのまま設定する",
			waiter: domain.InvestmentWaiters{
				Status:        domain.WAITING,
				InvestStatus:  nil,
				ProjectStatus: nil,
			},
			want: UserWaitlistResponseItem{
				Status:        domain.INVESTMENT_WAITING_STATUS_WAITING,
				InvestStatus:  nil,
				ProjectStatus: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newUserWaitlistResponseItem(tt.waiter)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserWaitlistResponseItem_MarshalJSON(t *testing.T) {
	tests := []struct {
		name              string
		item              UserWaitlistResponseItem
		wantInvestStatus  interface{}
		wantProjectStatus interface{}
	}{
		{
			name: "ステータスが設定済みの場合は文字列として出力する",
			item: UserWaitlistResponseItem{
				InvestStatus:  investStatusPtr(domain.SUCCESS_INVEST),
				ProjectStatus: projectStatusPtr(domain.SUCCESS_FUNDING),
			},
			wantInvestStatus:  string(domain.SUCCESS_INVEST),
			wantProjectStatus: string(domain.SUCCESS_FUNDING),
		},
		{
			name: "ステータスが未設定の場合はnullとして出力する",
			item: UserWaitlistResponseItem{
				InvestStatus:  nil,
				ProjectStatus: nil,
			},
			wantInvestStatus:  nil,
			wantProjectStatus: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.item)
			require.NoError(t, err)

			var got map[string]interface{}
			require.NoError(t, json.Unmarshal(data, &got))
			require.Contains(t, got, "invest_status")
			require.Contains(t, got, "project_status")
			assert.Equal(t, tt.wantInvestStatus, got["invest_status"])
			assert.Equal(t, tt.wantProjectStatus, got["project_status"])
		})
	}
}

func investStatusPtr(status domain.INVEST_STATUS) *domain.INVEST_STATUS {
	return &status
}

func projectStatusPtr(status domain.PROJECT_STATUS) *domain.PROJECT_STATUS {
	return &status
}

// 口座番号は金融機関を問わず7桁。ゆうちょも通帳の番号(8桁)ではなく振込用の
// 7桁で登録する。GMOあおぞらのAPIが7桁しか受け付けないため、画面を通さない
// 経路でも8桁が保存されないようにする。
func TestUserWithdrawalAccountInput_BankAccountValidation(t *testing.T) {
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
			name:        "(成功) 先頭ゼロの7桁も通る",
			bankAccount: "0000001",
			wantErr:     false,
		},
		{
			name:        "(失敗) ゆうちょの通帳番号8桁は弾く",
			bankAccount: "12345678",
			wantErr:     true,
		},
		{
			name:        "(失敗) 6桁は弾く",
			bankAccount: "123456",
			wantErr:     true,
		},
		{
			name:        "(失敗) 未入力は弾く",
			bankAccount: "",
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
		{
			name:        "(失敗) 小数点付きは弾く",
			bankAccount: "1.23456",
			wantErr:     true,
		},
	}

	cv := validator.NewCustomValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := UserWithdrawalAccountInput{
				BankName:    "ゆうちょ銀行",
				BranchName:  "〇一八支店",
				BankAccount: tt.bankAccount,
			}

			err := cv.Validate(&input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

// 管理画面は口座が未登録の会員も編集するため、空は通す。
// 入力があるときだけ7桁を求める。
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
			name:        "(成功) 未登録の会員を想定し空は通る",
			bankAccount: "",
			wantErr:     false,
		},
		{
			name:        "(失敗) 8桁は弾く",
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
			input := UserInput{
				FirstName:   "太郎",
				BankAccount: tt.bankAccount,
			}

			err := cv.Validate(&input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
