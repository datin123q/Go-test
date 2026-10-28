package jsonutils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var jst *time.Location

func init() {
	var err error
	jst, err = time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
}

type JSONTime time.Time

func (j JSONTime) MarshalJSON() ([]byte, error) {
	if time.Time(j).IsZero() {
		return []byte(`""`), nil
	}
	return []byte(`"` + time.Time(j).In(jst).Format(time.DateTime) + `"`), nil
}
func (j JSONTime) ToString() string {
	t := time.Time(j)
	if t.IsZero() {
		return ""
	}
	return t.In(jst).Format(time.DateTime)
}

func (j JSONTime) ToFormat(layout string) string {
	t := time.Time(j)
	if t.IsZero() {
		return ""
	}
	return t.In(jst).Format(layout)
}

func (j *JSONTime) UnmarshalJSON(value []byte) error {
	if string(value) == "null" {
		return nil
	}

	// 文字列を安全に抽出
	var timeString string
	if err := json.Unmarshal(value, &timeString); err != nil {
		return errors.Wrap(err, "failed to unmarshal JSONTime")
	}

	timeString = strings.TrimSpace(timeString)
	if timeString == "" {
		return nil
	}

	// タイムゾーン情報がない場合、Local(JST)として扱い設定する。
	// ESJPはタイムゾーン情報なしでデータベースに保存し、これをJSTとして扱う。
	t, err := time.ParseInLocation(time.DateTime, timeString, jst)
	if err != nil {
		// time.DateTime形式でない場合、RFC3339を試す。
		// 現在は統合メールサービスの返信の日時情報の場合のみである。
		t, err = time.Parse(time.RFC3339, timeString)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to parse JSONTime: %s", timeString))
		}
		t = t.In(jst)
	}

	*j = JSONTime(t)
	return nil
}

func (j *JSONTime) Unix() int64 {
	return (*time.Time)(j).Unix()
}

func (j *JSONTime) Time() time.Time {
	return time.Time(*j).In(jst)
}

// IsZero は time.Time のゼロ値判定を委譲する。
// XORM の utils.Zeroable インターフェースを満たすことで、
// Nullable カラムの更新時にゼロ値誤判定を防ぐ。
func (j JSONTime) IsZero() bool {
	return time.Time(j).IsZero()
}

func (j *JSONTime) IsEmpty() bool {
	if j == nil {
		return true
	}
	return time.Time(*j).IsZero()
}

func (j *JSONTime) EqualDate(t time.Time) bool {
	return t.Format(time.DateOnly) == j.ToFormat(time.DateOnly)
}

func (j *JSONTime) Equal(t time.Time) bool {
	return t.Format(time.DateTime) == j.ToFormat(time.DateTime)
}

func ParseDateTimeJST(layout string, value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errors.New("empty datetime")
	}
	t, err := time.ParseInLocation(layout, value, jst)
	if err != nil {
		return time.Time{}, errors.Wrap(err, fmt.Sprintf("failed to parse datetime: %s", value))
	}
	return t.UTC(), nil
}

func ParseJSONTimeJST(layout string, value string) (JSONTime, error) {
	t, err := ParseDateTimeJST(layout, value)
	if err != nil {
		return JSONTime{}, err
	}
	return JSONTime(t), nil
}

// ParseJSONDateJST は、xorm:"date" タグの付いたカラム(年月日だけを持つ)へ
// 保存する値を解釈する。MIR-1384: ParseJSONTimeJST は深夜0時JSTとして解釈して
// からUTCへ変換するが、xormはdate型のカラムへ書き込む際engine.DatabaseTZ
// (MySQLではtime.Local。本番はTZ未設定のためUTCになっている)で.In()した
// うえで年月日だけを取り出す(dialects/time.go FormatColumnTime)。深夜0時を
// 基準にすると、変換先がUTCの場合は前日の日付になってしまう。正午を基準に
// すれば、UTC・JSTのどちらに変換されても年月日は変わらない(他の呼び出し元は
// 日時として扱われるdatetimeカラム向けのため、ParseJSONTimeJST自体は直さず
// 専用の関数として分けている)。layoutには年月日のみの書式を渡すこと。
func ParseJSONDateJST(layout string, value string) (JSONTime, error) {
	t, err := ParseDateTimeJST(layout, value)
	if err != nil {
		return JSONTime{}, err
	}
	return JSONTime(t.Add(12 * time.Hour)), nil
}
