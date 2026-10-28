package domain

import (
	"reflect"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestProjectInfoHistory_ConvertTagToMap_FundNumber(t *testing.T) {
	key := string(PROJECT_CHANGED_CONTENT_FUND_NUMBER)

	tests := []struct {
		name       string
		fundNumber *int64
		want       string
	}{
		{
			name:       "未設定は空文字として記録する",
			fundNumber: nil,
			want:       "",
		},
		{
			name:       "設定済みは数値の文字列として記録する",
			fundNumber: lo.ToPtr(int64(202600123)),
			want:       "202600123",
		},
		{
			name:       "0 も記録する",
			fundNumber: lo.ToPtr(int64(0)),
			want:       "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := Project{FundNumber: tt.fundNumber}

			got := ProjectInfoHistory{}.ConvertTagToMap(reflect.ValueOf(&project).Elem())

			value, ok := got[key]
			assert.True(t, ok, "ファンド番号が変更履歴の対象に含まれていない")
			assert.Equal(t, tt.want, value)
		})
	}
}

// 未設定から設定、設定から未設定のどちらも差分として拾えることを確かめる。
func TestProjectInfoHistory_ConvertTagToMap_FundNumberDiff(t *testing.T) {
	key := string(PROJECT_CHANGED_CONTENT_FUND_NUMBER)

	before := Project{FundNumber: nil}
	after := Project{FundNumber: lo.ToPtr(int64(1))}

	beforeMap := ProjectInfoHistory{}.ConvertTagToMap(reflect.ValueOf(&before).Elem())
	afterMap := ProjectInfoHistory{}.ConvertTagToMap(reflect.ValueOf(&after).Elem())

	assert.NotEqual(t, beforeMap[key], afterMap[key])
	assert.Equal(t, "", beforeMap[key])
	assert.Equal(t, "1", afterMap[key])
}

// 値を拾えていても比較対象に入れ忘れると履歴に出ない。
// ConvertTagToMap と ProjectChangedContentTypesForCompare の両方に
// 載っていることを確かめる。
func TestProjectChangedContentTypesForCompare_ContainsFundNumber(t *testing.T) {
	key := string(PROJECT_CHANGED_CONTENT_FUND_NUMBER)

	project := Project{FundNumber: lo.ToPtr(int64(1))}
	converted := ProjectInfoHistory{}.ConvertTagToMap(reflect.ValueOf(&project).Elem())

	_, ok := converted[key]
	assert.True(t, ok, "ConvertTagToMap がファンド番号を拾えていない")

	assert.Contains(
		t,
		ProjectChangedContentTypesForCompare,
		key,
		"比較対象に入っていないため履歴に出ない",
	)
}

// 比較時に空文字を nil と取り違えないことを確かめる。
// 未設定は空文字で記録するため、nil 判定でスキップされると
// 「未設定 → 設定」が履歴に残らなくなる。
func TestProjectInfoHistory_ConvertTagToMap_FundNumberNotNilInterface(t *testing.T) {
	key := string(PROJECT_CHANGED_CONTENT_FUND_NUMBER)

	project := Project{FundNumber: nil}
	converted := ProjectInfoHistory{}.ConvertTagToMap(reflect.ValueOf(&project).Elem())

	assert.NotNil(t, converted[key])
	assert.Equal(t, "", converted[key])
}
