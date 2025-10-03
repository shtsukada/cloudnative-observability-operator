package telemetry

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	reconcileTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cno_reconcile_total",
			Help: "Number of reconciliations by kind and result",
		},
		[]string{"kind", "result"},
	)

	reconcileDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cno_reconcile_duration_seconds",
			Help:    "Reconciliation duration seconds by kind",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10},
		},
		[]string{"kind"},
	)

	eventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cno_events_total",
			Help: "Number of k8s events emitted (by type)",
		},
		[]string{"type"},
	)
)

func ObserveReconcile(kind, result string, d time.Duration) {
	reconcileDuration.WithLabelValues(kind).Observe(d.Seconds())
	reconcileTotal.WithLabelValues(kind, result).Inc()
}

func IncEvent(eventType string) {
	eventsTotal.WithLabelValues(eventType).Inc()
}
