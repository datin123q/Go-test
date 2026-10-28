package commoncode

type Code struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}
type CodeGroup struct {
	GroupCode string `json:"group_code"`
	Codes     []Code `json:"codes"`
}
