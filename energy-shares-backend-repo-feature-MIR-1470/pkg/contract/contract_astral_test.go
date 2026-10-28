package contract

import "testing"

func TestAstralAliasTableRemap(t *testing.T) {
	t.Run("補助面の文字が無ければそのまま返す", func(t *testing.T) {
		table := newAstralAliasTable(nil)
		got := table.remap("山田太郎")
		if got != "山田太郎" {
			t.Errorf("remap() = %q, want %q", got, "山田太郎")
		}
		if len(table.aliases) != 0 {
			t.Errorf("aliases should stay empty, got %v", table.aliases)
		}
	})

	t.Run("補助面の文字を私用領域の番号に置き換える", func(t *testing.T) {
		table := newAstralAliasTable(nil)
		got := table.remap("山田𠮷太郎")
		want := "山田" + string(rune(privateUseAreaStart)) + "太郎"
		if got != want {
			t.Errorf("remap() = %q, want %q", got, want)
		}
		if table.aliases['\U00020BB7'] != privateUseAreaStart {
			t.Errorf("aliases[𠮷] = %v, want %v", table.aliases['\U00020BB7'], privateUseAreaStart)
		}
	})

	t.Run("同じ文字には同じ番号を割り当てる", func(t *testing.T) {
		table := newAstralAliasTable(nil)
		first := table.remap("山田𠮷太郎")
		second := table.remap("佐藤𠮷次郎")
		if len(table.aliases) != 1 {
			t.Fatalf("aliases should have exactly 1 entry, got %d: %v", len(table.aliases), table.aliases)
		}
		wantAlias := string(rune(privateUseAreaStart))
		if first != "山田"+wantAlias+"太郎" {
			t.Errorf("first remap() = %q", first)
		}
		if second != "佐藤"+wantAlias+"次郎" {
			t.Errorf("second remap() = %q", second)
		}
	})

	t.Run("複数の異なる補助面文字には別々の番号を割り当てる", func(t *testing.T) {
		table := newAstralAliasTable(nil)
		got := table.remap("𥽜と𠩤と𠮷")
		runes := []rune(got)
		if len(runes) != 5 {
			t.Fatalf("unexpected rune count: %d (%q)", len(runes), got)
		}
		if runes[0] == runes[2] || runes[0] == runes[4] || runes[2] == runes[4] {
			t.Errorf("different astral characters must get different aliases, got %q", got)
		}
		for _, r := range runes {
			if r >= privateUseAreaStart && r <= privateUseAreaEnd {
				continue
			}
			if r == 'と' {
				continue
			}
			t.Errorf("unexpected rune %U in %q", r, got)
		}
	})

	t.Run("私用領域自体の文字はそのまま(補助面ではないので置き換えない)", func(t *testing.T) {
		table := newAstralAliasTable(nil)
		text := "山田" + string(rune(0xE000)) + "太郎"
		got := table.remap(text)
		if got != text {
			t.Errorf("remap() = %q, want unchanged %q", got, text)
		}
		if len(table.aliases) != 0 {
			t.Errorf("aliases should stay empty for a BMP private-use character, got %v", table.aliases)
		}
	})

	t.Run("reservedに登録された番号は補助面文字の割り当てに使わない", func(t *testing.T) {
		// MIR-1201: レビューで発見した不具合の再現。氏名等に本物の私用領域の
		// 外字(U+E000)が既にある場合、そこに別の補助面文字(𠮷)を割り当てて
		// しまうと、本物の外字が𠮷のグリフに置き換わって見えてしまう。
		table := newAstralAliasTable(map[rune]bool{privateUseAreaStart: true})
		got := table.remap("山田𠮷太郎")
		wantAlias := rune(privateUseAreaStart + 1)
		want := "山田" + string(wantAlias) + "太郎"
		if got != want {
			t.Errorf("remap() = %q, want %q (reservedのU+E000を飛ばしてU+E001を使うべき)", got, want)
		}
	})
}

func TestCollectPrivateUseRunes(t *testing.T) {
	values := map[string]string{
		"invest_name": "山田" + string(rune(0xE000)) + "太郎",
		"spv_name":    "test",
		"address":     "東京都" + string(rune(0xE010)),
	}
	got := collectPrivateUseRunes(values)
	want := map[rune]bool{0xE000: true, 0xE010: true}
	if len(got) != len(want) {
		t.Fatalf("collectPrivateUseRunes() = %v, want %v", got, want)
	}
	for r := range want {
		if !got[r] {
			t.Errorf("collectPrivateUseRunes() missing %U", r)
		}
	}
}
