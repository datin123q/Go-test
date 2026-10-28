package jobs

import (
	"github.com/labstack/echo-contrib/prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
)

const CKeyMetrics = "esjp-metrics"
const AutoInvestmentJobName = "auto_investment_job"
const GetVaTransactionJobName = "get_va_transaction_job"
const RefreshTokenJobName = "refresh_token_job"

// In case you add a metric here later, make sure to include it in the MetricsList method or you'll go to have a bad time.
type Metrics struct {
	JobMetric *prometheus.Metric
}

// MetricList
// Needed by echo-contrib so echo can register and collect these metrics
func (m *Metrics) MetricList() []*prometheus.Metric {
	return []*prometheus.Metric{
		// ADD EVERY METRIC HERE!
		m.JobMetric,
	}
}

func (m *Metrics) IncreaseCounter(jobName string) {
	m.JobMetric.MetricCollector.(*prom.CounterVec).With(prom.Labels{"job": jobName}).Inc()
}
