package ksql

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// apiTimeoutはMIR-1327の修正前、DIコンテナが全機能に共有するresty.New()を
// そのまま(Cloneもせず)保持していたことで実質的に効いていた値(emailutils.New
// がClone().SetTimeoutしていた10秒)と同じ。修正後も打ち切り時間を変えないため、
// 同じ値を明示的に設定する(河内さん判断、2026-09-29)。
const apiTimeout = 10 * time.Second

type operatorHistoryKsqlRepository struct {
	cfg         *config.Config
	restyClient *resty.Client
	logger      factory.LoggerFactory
}

// MIR-1327: 以前は引数のrestyClient(DIコンテナが全機能に共有するresty.New())を
// そのまま(Cloneもせず)保持していたため、他の機能(email等)がClone().SetTimeout()
// すると、Clone()が内部の*http.Clientをポインタのまま引き継ぐ仕様により、この
// クライアントの打ち切り時間まで意図せず変わっていた。resty.New()で独立した
// クライアントを持ち、修正前に実際に効いていた値(10秒)を明示的に設定する。
func New(
	cfg *config.Config,
	logger factory.LoggerFactory,
) cdomain.OperatorHistoryKsqlRepository {
	return &operatorHistoryKsqlRepository{
		cfg:         cfg,
		restyClient: resty.New().SetTimeout(apiTimeout),
		logger:      logger,
	}
}

func (r operatorHistoryKsqlRepository) GetAll(ctx context.Context, options ...cdomain.OperatorHistoryKsqlQueryOption) (int64, []domain.OperatorHistoryRowResponse, error) {
	queryOpts := cdomain.OperatorHistoryKsqlQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}
	dbCfg := r.cfg.EventStream.KsqlDB

	// run ksql query
	stmt, err := makeKsqlSelectStatement(dbCfg.Resources.AccessLog, queryOpts)
	if err != nil {
		return 0, nil, errors.Wrap(err, "Failed to make ksql select statement")
	}
	restyResp, err := r.runKsqlQuery(ctx, dbCfg.Url, stmt)
	if err != nil {
		return 0, nil, errors.Wrap(err, "Failed to get changelog events")
	}

	// unmarshall response
	resp, err := unmarshallBody(restyResp.Body())
	if err != nil {
		return 0, nil, errors.Wrap(err, "Failed to unmarshall response")
	}
	rows, err := unmarshallRows(resp.Rows)
	if err != nil {
		return 0, nil, errors.Wrap(err, "Failed to unmarshall rows")
	}

	return int64(len(resp.Rows)), rows, nil
}

func (r operatorHistoryKsqlRepository) runKsqlQuery(ctx context.Context, url string, stmt string) (*resty.Response, error) {
	// MIR-1365: 以前はSetDebug(r.cfg.Debug)で設定値を見ており、debug=trueの環境では
	// ksqlへの問い合わせ内容と応答がログに出ていた。設定値での切り替えをやめ、常に出さない。
	resp, err := r.restyClient.R().
		SetContext(ctx).
		SetLogger(r.logger.Entry(ctx)).
		SetHeaders(map[string]string{
			"Accept":       "application/vnd.ksql.v1+json",
			"Content-Type": "application/json",
		}).
		SetBody(domain.KsqlRequest{
			Ksql: stmt,
			StreamsProperties: map[string]string{
				"ksql.streams.auto.offset.reset": "earliest",
			},
		}).
		Post(url + "/query")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to run ksql query")
	}
	return resp, nil
}

func makeKsqlSelectStatement(kStream string, queryOpts cdomain.OperatorHistoryKsqlQueryOptionFields) (string, error) {
	start, end, limit := queryOpts.StartDate(), queryOpts.EndDate(), queryOpts.Limit()
	stmt := "SELECT ts, actor, operation, ip, log FROM " + kStream
	var conditions []string

	// collect conditions
	if start != nil {
		conditions = append(conditions, "ts >= PARSE_TIMESTAMP('"+*start+"', 'yyyy-MM-dd HH:mm:ss', 'Asia/Tokyo')")
	}
	if end != nil {
		conditions = append(conditions, "ts <= PARSE_TIMESTAMP('"+*end+"', 'yyyy-MM-dd HH:mm:ss', 'Asia/Tokyo')")
	}

	// append where clause
	if len(conditions) > 0 {
		stmt += " WHERE " + strings.Join(conditions, " AND ")
	}

	// append limit clause
	if limit > 0 {
		stmt += " LIMIT " + strconv.Itoa(limit)
	}

	stmt += ";"
	return stmt, nil
}

func unmarshallBody(respBody []byte) (domain.QueryResponse, error) {
	var items []domain.QueryResponseItem
	err := jsonutils.Unmarshal(respBody, &items)
	if err != nil {
		return domain.QueryResponse{}, errors.Wrap(err, "Error unmarshalling ksql response")
	}

	var rows []*domain.QueryRow
	for _, item := range items {
		if item.Row != nil {
			rows = append(rows, item.Row)
		}
	}

	formatted := domain.QueryResponse{
		Header:       items[0].Header,
		Rows:         rows,
		FinalMessage: items[len(items)-1].FinalMessage,
	}
	return formatted, nil
}

func unmarshallRows(queryRows []*domain.QueryRow) ([]domain.OperatorHistoryRowResponse, error) {
	var rows []domain.OperatorHistoryRowResponse

	for _, queryRow := range queryRows {
		actor, operation, log := queryRow.Columns[1], queryRow.Columns[2], queryRow.Columns[4]

		// log
		rowLog, err := domain.OperatorHistoryLog{}.ToOperatorHistoryLog(log)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to unmarshall log")
		}

		// date
		rowDate := rowLog.Date.Format(time.DateTime)

		// actor
		rowActor, err := domain.OperatorHistoryActor{}.ToOperatorHistoryActor(actor)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to unmarshall actor")
		}

		// operation
		rowOperation := ""
		if operation != nil {
			rowOperation = operation.(string)
		}

		row := domain.OperatorHistoryRowResponse{
			Date:      rowDate,
			Actor:     rowActor,
			Operation: rowOperation,
			IP:        rowLog.RemoteIP,
			Log:       rowLog,
		}
		rows = append(rows, row)
	}

	return rows, nil
}
