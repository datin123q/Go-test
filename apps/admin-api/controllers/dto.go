package controllers

import "github.com/illuminarean/energy-shares-backend-repo/domain"

const (
	DefaultMaxResultCount = 30
	UserMaxResultCount    = 100000
)

// ProfileSearchInput は会員管理(個人・法人)の一覧・件数・CSV で受けるプロフィール側の検索条件(MIR-534)。
type ProfileSearchInput struct {
	Kana         *string `query:"kana" swagger:"desc(フリガナ(部分一致)。個人: 姓・名 / 法人: 代表者・取引担当者の姓・名)"`
	BirthdayFrom string  `query:"birthday_from" swagger:"desc(生年月日-from YYYY-MM-DD。法人は取引担当者)"`
	BirthdayTo   string  `query:"birthday_to" swagger:"desc(生年月日-to YYYY-MM-DD。fromと同じ日なら完全一致)"`
}

func (p ProfileSearchInput) ToCondition() domain.ProfileSearchCondition {
	return domain.ProfileSearchCondition{
		Kana:         p.Kana,
		BirthdayFrom: p.BirthdayFrom,
		BirthdayTo:   p.BirthdayTo,
	}
}

type SearchInput struct {
	Sortby         string `query:"sortby"`
	Order          string `query:"order"`
	SkipCount      int    `query:"skip_count"`
	MaxResultCount int    `query:"max_result_count"`
}

type QuerySearchInput struct {
	SearchKeys   string `query:"searchKeys" swagger:"desc(key1,key2)"`
	SearchValues string `query:"searchValues" swagger:"desc(value1,value2)"`
}
