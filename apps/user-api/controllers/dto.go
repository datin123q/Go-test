package controllers

const (
	DefaultMaxResultCount = 30
)

// Quergy
type SearchInput struct {
	Sortby         string `query:"sortby"`
	Order          string `query:"order"`
	SkipCount      int    `query:"skip_count"`
	MaxResultCount int    `query:"max_result_count"`
}
