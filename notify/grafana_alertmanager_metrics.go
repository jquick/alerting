package notify

import (
	"github.com/prometheus/alertmanager/api/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namesapce = "grafana"
const subsystem = "alerting"
const ActiveStateLabelValue = "active"
const InactiveStateLabelValue = "inactive"

type GrafanaAlertmanagerMetrics struct {
	Registerer prometheus.Registerer
	*metrics.Alerts
	configuredReceivers    *prometheus.GaugeVec
	configuredIntegrations *prometheus.GaugeVec
}

// NewGrafanaAlertmanagerMetrics creates a set of metrics for the Alertmanager.
func NewGrafanaAlertmanagerMetrics(r prometheus.Registerer) *GrafanaAlertmanagerMetrics {
	return &GrafanaAlertmanagerMetrics{
		Registerer: r,
		Alerts:     metrics.NewAlerts("grafana", r),
		configuredReceivers: promauto.With(r).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namesapce,
			Subsystem: subsystem,
			Name:      "alertmanager_receivers",
			Help:      "Number of configured receivers by state. It is considered active if used within a route.",
		}, []string{"org", "state"}),
		configuredIntegrations: promauto.With(r).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namesapce,
			Subsystem: subsystem,
			Name:      "alertmanager_integrations",
			Help:      "Number of configured integrations.",
		}, []string{"org", "type"}),
	}
}
