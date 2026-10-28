package domain

type QueryResponse struct {
	Header       *QueryHeader `json:"header,omitempty"`
	Rows         []*QueryRow  `json:"row,omitempty"`
	FinalMessage *string      `json:"finalMessage,omitempty"`
}

type QueryResponseItem struct {
	Header       *QueryHeader `json:"header,omitempty"`
	Row          *QueryRow    `json:"row,omitempty"`
	FinalMessage *string      `json:"finalMessage,omitempty"`
}

type QueryHeader struct {
	QueryID string `json:"queryId"`
	Schema  string `json:"schema"`
}

type QueryRow struct {
	Columns []interface{} `json:"columns"`
}

type KsqlRequest struct {
	Ksql              string            `json:"ksql"`
	StreamsProperties map[string]string `json:"streamsProperties"`
}

type ChangelogRow struct {
	Id int64  `json:"id" description:"ID"`
	Op string `json:"op" description:"오퍼레이션 타입"`
	// TsMs          int64                  `json:"ts_ms" description:"이벤트 캡쳐 일시(ms)"`
	Updated string                 `json:"updated" description:"수정 일시"`
	Actor   ChangelogActor         `json:"actor" description:"행위자"`
	Columns map[string]interface{} `json:"columns" description:"변경된 컬럼 목록 및 before/after 값"`
}

type ChangelogActor struct {
	Id    int64
	Name  string
	Email string
}
