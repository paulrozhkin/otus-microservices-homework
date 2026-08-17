package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

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
	ordersStarted = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "online_store",
		Name:      "orders_started_total",
		Help:      "Total number of newly created order sagas.",
	})
	sagaTransitions = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "online_store",
		Name:      "order_saga_transitions_total",
		Help:      "Total number of committed order saga state transitions.",
	}, []string{"from_status", "to_status"})
	sagaDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "online_store",
		Name:      "order_saga_duration_seconds",
		Help:      "Duration of an order saga from creation to a terminal state.",
		Buckets:   []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60},
	}, []string{"outcome"})
	compensations = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "online_store",
		Name:      "order_saga_compensations_total",
		Help:      "Total number of compensation steps started by the order saga.",
	}, []string{"step"})
	idempotencyRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "online_store",
		Name:      "order_idempotency_requests_total",
		Help:      "Total number of create-order requests handled by idempotency result.",
	}, []string{"result"})
)

func init() {
	prometheus.MustRegister(
		ordersCreated,
		paidOrderValue,
		ordersStarted,
		sagaTransitions,
		sagaDuration,
		compensations,
		idempotencyRequests,
	)
}

func OrderFinalized(status string, price int64) {
	ordersCreated.WithLabelValues(status).Inc()
	if status == "completed" {
		paidOrderValue.Add(float64(price))
	}
}

func OrderStarted() {
	ordersStarted.Inc()
}

func IdempotencyRequest(result string) {
	idempotencyRequests.WithLabelValues(result).Inc()
}

func SagaTransition(fromStatus, toStatus string, createdAt time.Time, price int64) {
	sagaTransitions.WithLabelValues(fromStatus, toStatus).Inc()

	switch toStatus {
	case "inventory_releasing":
		compensations.WithLabelValues("release_inventory").Inc()
	case "payment_refunding":
		compensations.WithLabelValues("refund_payment").Inc()
	case "completed", "failed":
		OrderFinalized(toStatus, price)
		sagaDuration.WithLabelValues(toStatus).Observe(time.Since(createdAt).Seconds())
	}
}
