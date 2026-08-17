package metrics

import "github.com/prometheus/client_golang/prometheus"

var usersCreated = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "online_store",
	Name:      "users_created_total",
	Help:      "Total number of successfully created users.",
}, []string{"source"})

func init() {
	prometheus.MustRegister(usersCreated)
}

func UserCreated(source string) {
	usersCreated.WithLabelValues(source).Inc()
}
