package httpmiddleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_requests_total", Help: "Total number of HTTP requests"},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "HTTP request duration in seconds", Buckets: prometheus.DefBuckets},
		[]string{"method", "path"},
	)
	httpRequestDurationMax = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "http_request_duration_seconds_max", Help: "Maximum observed HTTP request duration in seconds since process start"},
		[]string{"method", "path"},
	)
	httpRequestDurationMaxMu sync.Mutex
	httpRequestDurationMaxes = make(map[string]float64)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, httpRequestDurationMax)
}

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start).Seconds()
		method, path := c.Request.Method, c.FullPath()
		httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(c.Writer.Status())).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)
		observeRequestDurationMax(method, path, duration)
	}
}

func observeRequestDurationMax(method, path string, duration float64) {
	key := method + " " + path
	httpRequestDurationMaxMu.Lock()
	defer httpRequestDurationMaxMu.Unlock()

	if duration > httpRequestDurationMaxes[key] {
		httpRequestDurationMaxes[key] = duration
		httpRequestDurationMax.WithLabelValues(method, path).Set(duration)
	}
}
