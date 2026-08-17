package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	accountsCreated = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "online_store",
		Name:      "billing_accounts_created_total",
		Help:      "Total number of billing accounts created.",
	})
	billingOperations = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "online_store",
		Name:      "billing_operations_total",
		Help:      "Total number of billing operations by type and result.",
	}, []string{"type", "result"})
)

func init() {
	prometheus.MustRegister(accountsCreated, billingOperations)
}

func AccountCreated() {
	accountsCreated.Inc()
}

func BillingOperation(operationType, result string) {
	billingOperations.WithLabelValues(operationType, result).Inc()
}
