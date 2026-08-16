package metrics

import "github.com/prometheus/client_golang/prometheus"

var operations = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "online_store",
	Name:      "warehouse_operations_total",
	Help:      "Total number of committed warehouse operations by type and result.",
}, []string{"type", "result"})

func init() {
	prometheus.MustRegister(operations)
}

func Operation(operationType, result string) {
	operations.WithLabelValues(operationType, result).Inc()
}
