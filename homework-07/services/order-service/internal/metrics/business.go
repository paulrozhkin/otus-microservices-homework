package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	ordersCreated = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "online_store",
		Name:      "orders_created_total",
		Help:      "Total number of finalized orders by status.",
	}, []string{"status"})
	paidOrderValue = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "online_store",
		Name:      "paid_order_value_minor_units_total",
		Help:      "Total value of paid orders in the smallest currency unit.",
	})
)

func init() {
	prometheus.MustRegister(ordersCreated, paidOrderValue)
}

func OrderFinalized(status string, price int64) {
	ordersCreated.WithLabelValues(status).Inc()
	if status == "paid" {
		paidOrderValue.Add(float64(price))
	}
}
