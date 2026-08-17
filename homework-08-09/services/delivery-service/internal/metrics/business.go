package metrics

import "github.com/prometheus/client_golang/prometheus"

var reservations = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "online_store",
	Name:      "delivery_reservations_total",
	Help:      "Total number of committed delivery reservation attempts by result.",
}, []string{"result"})

func init() {
	prometheus.MustRegister(reservations)
}

func Reservation(result string) {
	reservations.WithLabelValues(result).Inc()
}
