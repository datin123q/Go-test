package ksql

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// Test_New_UsesIndependentClientWithExpectedTimeout はMIR-1327の回帰テスト。
// New()を呼ぶたびに独立したhttp.Clientを使い、修正前に共有クライアント経由で
// 実質的に効いていた値(10秒)を明示的に保つことを確認する。
func Test_New_UsesIndependentClientWithExpectedTimeout(t *testing.T) {
	cfg := &config.Config{}
	logger := factory.NewLoggerFactory()

	a := New(cfg, logger).(*changelogRepository)
	b := New(cfg, logger).(*changelogRepository)

	assert.NotSame(t, a.restyClient.GetClient(), b.restyClient.GetClient(), "New()を呼ぶたびに独立したhttp.Clientを使うこと")
	assert.Equal(t, 10*time.Second, a.restyClient.GetClient().Timeout)
}

// Test_RunKsqlQuery_DebugStaysOffRegardlessOfConfigはMIR-1365の回帰テスト。以前は
// runKsqlQueryがSetDebug(r.cfg.Debug)で設定値を見ており、cfg.Debugがtrueに
// なるとksqlの問い合わせ内容(個人情報を含みうる)がログに出ていた。New()は
// クライアントのDebugを設定しないため、実際に送るリクエストのDebugを左右するのは
// 呼び出し後のSetDebugの有無だけであり、それをOnBeforeRequestフックで確認する。
func Test_RunKsqlQuery_DebugStaysOffRegardlessOfConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"header":{"queryId":"q","schema":"s"}},{"finalMessage":"Limit Reached"}]`))
	}))
	defer server.Close()

	cfg := &config.Config{}
	cfg.Debug = true
	logger := factory.NewLoggerFactory()

	r := New(cfg, logger).(*changelogRepository)

	var gotDebug bool
	r.restyClient.OnBeforeRequest(func(_ *resty.Client, req *resty.Request) error {
		gotDebug = req.Debug
		return nil
	})

	_, err := r.runKsqlQuery(context.Background(), server.URL, "SELECT * FROM x;")

	assert.NoError(t, err)
	assert.False(t, gotDebug, "cfg.Debug=trueでも実際に送るリクエストのDebugはfalseのままであるべき")
}

func Test_changelogRepository_GetAll(t *testing.T) {
	type fields struct {
		cfg         *config.Config
		restyClient *resty.Client
		logger      factory.LoggerFactory
	}
	type args struct {
		ctx     context.Context
		kStream string
		options []cdomain.ChangelogQueryOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []domain.ChangelogRow
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := changelogRepository{
				cfg:         tt.fields.cfg,
				restyClient: tt.fields.restyClient,
				logger:      tt.fields.logger,
			}
			got, err := r.GetAll(tt.args.ctx, tt.args.options...)

			assert.NoErrorf(t, err, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetAll() got = %v, want %v", got, tt.want)
		})
	}
}

func Test_changelogRepository_runKsqlQuery(t *testing.T) {
	options := []cdomain.ChangelogQueryOption{
		cdomain.ChangelogQueryOptionFields{}.WithStartDate(lo.ToPtr("2024-10-01 00:00:00")),
		cdomain.ChangelogQueryOptionFields{}.WithEndDate(lo.ToPtr("2024-11-02 16:47:00")),
		cdomain.ChangelogQueryOptionFields{}.WithLimit(10),
		cdomain.ChangelogQueryOptionFields{}.WithId(lo.ToPtr("133")),
	}
	queryOpts := cdomain.ChangelogQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	type fields struct {
		cfg         *config.Config
		restyClient *resty.Client
		logger      factory.LoggerFactory
	}
	type args struct {
		ctx  context.Context
		url  string
		stmt string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *resty.Response
		wantErr bool
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := changelogRepository{
				cfg:         tt.fields.cfg,
				restyClient: tt.fields.restyClient,
				logger:      tt.fields.logger,
			}

			got, err := r.runKsqlQuery(tt.args.ctx, tt.args.url, tt.args.stmt)

			assert.NoErrorf(t, err, "runKsqlQuery() error = %v, wantErr %v", err, tt.wantErr)
			assert.NotNilf(t, got, "runKsqlQuery() got = %v, want %v", got, tt.want)
			assert.Equalf(t, got, tt.want, "runKsqlQuery() got = %v, want %v", got, tt.want)
		})
	}
}

func Test_filterNilValueElement(t *testing.T) {
	type args struct {
		columnMap map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "맵에서 nil값 요소를 제거하고, 필터링된 맵을 반환",
			args: args{
				columnMap: map[string]interface{}{
					"key1": "value1",
					"key2": nil,
					"key3": "value3",
				},
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key3": "value3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterNilValueElement(tt.args.columnMap)

			assert.Equalf(t, tt.want, got, "filterNilValueElement() = %v, want %v", got, tt.want)
		})
	}
}

func Test_makeKsqlSelectStatement(t *testing.T) {
	options := []cdomain.ChangelogQueryOption{
		cdomain.ChangelogQueryOptionFields{}.WithStartDate(lo.ToPtr("2024-10-01 00:00:00")),
		cdomain.ChangelogQueryOptionFields{}.WithEndDate(lo.ToPtr("2024-11-02 16:47:00")),
		cdomain.ChangelogQueryOptionFields{}.WithLimit(10),
		cdomain.ChangelogQueryOptionFields{}.WithId(lo.ToPtr("133")),
	}
	queryOpts := cdomain.ChangelogQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	type args struct {
		kStream   string
		queryOpts cdomain.ChangelogQueryOptionFields
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "스트림 이름, 쿼리 옵션으로 부터 KSQL 쿼리문을 생성",
			args: args{
				kStream:   "dev_answer_diff",
				queryOpts: queryOpts,
			},
			want: "SELECT * FROM dev_answer_diff WHERE op != 'r'" +
				" AND updated >= PARSE_TIMESTAMP('2024-10-01 00:00:00', 'yyyy-MM-dd HH:mm:ss', 'Asia/Tokyo')" +
				" AND updated <= PARSE_TIMESTAMP('2024-11-02 16:47:00', 'yyyy-MM-dd HH:mm:ss', 'Asia/Tokyo')" +
				" AND id = 133" +
				" LIMIT 10;",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeKsqlSelectStatement(tt.args.kStream, tt.args.queryOpts)

			assert.NoErrorf(t, err, "makeKsqlSelectStatement() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "makeKsqlSelectStatement() got = %v, want %v", got, tt.want)
		})
	}
}

func Test_unmarshallResponse(t *testing.T) {
	type args struct {
		respBody []byte
	}
	tests := []struct {
		name    string
		args    args
		want    domain.QueryResponse
		wantErr bool
	}{
		{
			name: "응답 body를 QueryResponse로 변환",
			args: args{

				respBody: []byte(`
[
  {
    "header": {
      "queryId": "transient_DEV_ANSWER_DIFF_7521268954572835650",
      "schema": "` + ("`ID` BIGINT, `OP` STRING, `TS_MS` BIGINT, `UPDATED` TIMESTAMP, `SOURCE` STRING, `ACTOR` STRING, `DIFF` STRING") + `"
    }
  },
  {
    "row": {
      "columns": [
        133,
        "u",
        1730532614408,
        "2024-10-08T08:03:10.000",
        "{\"id\":1,\"name\":\"admin\",\"email\":\"admin@energyshares.jp\"}",
        {
          "ANSWER_COUNT": null,
          "COMMON_ID": null,
          "CONTENT": {
            "AFTER": "test11",
            "BEFORE": "test10"
          },
          "CREATED": null,
          "EMAIL": null,
          "ID": null,
          "MOBILE_PHONE": null,
          "NAME": null,
          "READ": null,
          "TITLE": null,
          "UPDATED": null,
          "USER_ID": null
        }
      ]
    }
  },
  {
    "finalMessage": "Limit Reached"
  }
]`),
			},
			want: domain.QueryResponse{
				Header: &domain.QueryHeader{
					QueryID: "transient_DEV_ANSWER_DIFF_7521268954572835650",
					Schema:  "`ID` BIGINT, `OP` STRING, `TS_MS` BIGINT, `UPDATED` TIMESTAMP, `SOURCE` STRING, `ACTOR` STRING, `DIFF` STRING",
				},
				Rows: []*domain.QueryRow{
					{
						Columns: []interface{}{
							float64(133),
							"u",
							float64(1730532614408),
							"2024-10-08T08:03:10.000",
							"{\"id\":1,\"name\":\"admin\",\"email\":\"admin@energyshares.jp\"}",
							map[string]interface{}{
								"ANSWER_COUNT": nil,
								"COMMON_ID":    nil,
								"CONTENT": map[string]interface{}{
									"AFTER":  "test11",
									"BEFORE": "test10",
								},
								"CREATED":      nil,
								"EMAIL":        nil,
								"ID":           nil,
								"MOBILE_PHONE": nil,
								"NAME":         nil,
								"READ":         nil,
								"TITLE":        nil,
								"UPDATED":      nil,
								"USER_ID":      nil,
							},
						},
					},
				},
				FinalMessage: func() *string {
					s := "Limit Reached"
					return &s
				}(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshallBody(tt.args.respBody)

			assert.NoErrorf(t, err, "unmarshallBody() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, *tt.want.Header, *got.Header, "unmarshallBody() got.Header = %v, want.Header %v", got.Header, tt.want.Header)
			assert.Equalf(t, *tt.want.FinalMessage, *got.FinalMessage, "unmarshallBody() got.FinalMessage = %v, want.FinalMessage %v", got.FinalMessage, tt.want.FinalMessage)
			for i, row := range tt.want.Rows {
				assert.Equalf(t, row.Columns, got.Rows[i].Columns, "unmarshallBody() got.Rows[%d].Columns = %v, want.Rows[%d].Columns %v", i, got.Rows[i].Columns, i, row.Columns)
			}
		})
	}
}

func Test_unmarshallRows(t *testing.T) {
	type args struct {
		queryRows []*domain.QueryRow
	}
	tests := []struct {
		name    string
		args    args
		want    []domain.ChangelogRow
		wantErr bool
	}{
		{
			name: "응답의 rows(QueryRow)를 변경로그 rows(ChangelogRow)로 변환",
			args: args{
				queryRows: []*domain.QueryRow{
					{
						Columns: []interface{}{
							float64(133),
							"u",
							float64(1730532614408),
							"2024-10-08T08:03:10.000",
							"{\"id\":1,\"name\":\"admin\",\"email\":\"admin@energyshares.jp\"}",
							map[string]interface{}{
								"ANSWER_COUNT": nil,
								"COMMON_ID":    nil,
								"CONTENT": map[string]interface{}{
									"AFTER":  "test11",
									"BEFORE": "test10",
								},
								"CREATED":      nil,
								"EMAIL":        nil,
								"ID":           nil,
								"MOBILE_PHONE": nil,
								"NAME":         nil,
								"READ":         nil,
								"TITLE":        nil,
								"UPDATED":      nil,
								"USER_ID":      nil,
							},
						},
					},
				},
			},
			want: []domain.ChangelogRow{
				{
					Id:      133,
					Op:      "u",
					Updated: "2024-10-08T08:03:10.000",
					Actor: domain.ChangelogActor{
						Id:    1,
						Name:  "admin",
						Email: "admin@energyshares.jp",
					},
					Columns: map[string]interface{}{
						"CONTENT": map[string]interface{}{
							"AFTER":  "test11",
							"BEFORE": "test10",
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshallRows(tt.args.queryRows)

			assert.NoErrorf(t, err, "unmarshallRows() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "unmarshallRows() got = %v, want %v", got, tt.want)
		})
	}
}
