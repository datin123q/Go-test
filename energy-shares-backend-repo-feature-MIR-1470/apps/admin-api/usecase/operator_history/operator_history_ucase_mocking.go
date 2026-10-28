package operator_history

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var (
	dt, _ = time.Parse(time.RFC3339, "2025-01-03T12:13:33.234758+09:00")
	dtStr = dt.Format(time.DateTime)

	operatorHistory1 = domain.OperatorHistoryRowResponse{
		Date: dtStr,
		Actor: domain.OperatorHistoryActor{
			Id:    1,
			Name:  "admin",
			Email: "admin@energyshares.jp",
		},
		Operation: "commoncode.LOG_CONTENT_TYPE.LOGIN",
		IP:        "::1",
		Log: domain.OperatorHistoryLog{
			LogType:   "request",
			RequestID: "e72c9dd2-c844-11ef-8b8e-f6f0d9467449",
			RemoteIP:  "::1",
			Host:      "localhost:8080",
			Path:      "/api/v1/admin-users/login",
			URI:       "/api/v1/admin-users/login",
			Method:    "POST",
			Referer:   "",
			UserAgent: "HTTPie/3.2.4",
			Status:    200,
			Latency:   "51.242708ms",
			BytesIn:   "74",
			BytesOut:  563,
			Date:      dt,
		},
	}
)
