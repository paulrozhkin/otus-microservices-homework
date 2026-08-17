package metrics

import "github.com/prometheus/client_golang/prometheus"

var notificationsProcessed = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "online_store",
	Name:      "notifications_processed_total",
	Help:      "Total number of notification events successfully persisted or deduplicated.",
}, []string{"order_status"})

func init() {
	prometheus.MustRegister(notificationsProcessed)
}

func NotificationProcessed(orderStatus string) {
	notificationsProcessed.WithLabelValues(orderStatus).Inc()
}
