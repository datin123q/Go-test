package stringutils

import (
	"html"
	"strings"
)

// brTagReplacer はエスケープされた <br> タグを復元するための Replacer
var brTagReplacer = strings.NewReplacer(
	"&lt;br&gt;", "<br />",
	"&lt;br/&gt;", "<br />",
	"&lt;br /&gt;", "<br />",
)

// SanitizeHTML はHTMLタグをエスケープしてXSS攻撃を防ぐ
func SanitizeHTML(input string) string {
	return html.EscapeString(input)
}

// SanitizeHTMLKeepBR はHTMLタグをエスケープしつつ、改行タグ（<br>, <br/>, <br />）は保持する
func SanitizeHTMLKeepBR(input string) string {
	return brTagReplacer.Replace(html.EscapeString(input))
}
