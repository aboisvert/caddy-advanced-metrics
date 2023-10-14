package advanced_metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (am *AdvancedMetricsModule) StartServer() {
	if am.prometheusPort == 0 {
		am.logger.Sugar().Infof("Using default port 6611\n")
		am.prometheusPort = 6611
	}
	port := am.prometheusPort

	am.logger.Sugar().Infof("Starting advanced metrics on port %d", port)

	// buckets := prometheus.ExponentialBuckets(0.1, 1.5, 10)

	am.requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "caddy_advanced_metrics_requests_total",
			Help: "Number of requests",
		},
		[]string{"method", "path", "status", "host"},
	)

	am.requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "caddy_advanced_metrics_request_duration_seconds",
			Help:    "Latency for HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status", "host"},
	)

	am.requestSize = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "caddy_advanced_metrics_request_size_bytes",
			Help: "Size of HTTP requests.",
		},
		[]string{"method", "path", "status", "host"},
	)
	am.responseSize = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "caddy_advanced_metrics_response_size_bytes",
			Help: "Size of HTTP responses.",
		},
		[]string{"method", "path", "status", "host"},
	)

	reg := prometheus.NewRegistry()
	reg.MustRegister(am.requestsTotal)
	reg.MustRegister(am.requestDuration)
	reg.MustRegister(am.requestSize)
	reg.MustRegister(am.responseSize)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

func (am *AdvancedMetricsHandler) HandleRequest(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	m := am.module

	// wrap the request handler
	lrw := NewLoggingResponseWriter(w)
	before := time.Now()
	err := next.ServeHTTP(lrw, r)
	after := time.Now()
	statusCode := strconv.Itoa(lrw.statusCode)

	// update stats
	if am.Counter {
		m.requestsTotal.
			WithLabelValues(r.Method, r.URL.Path, statusCode, r.Host).
			Inc()
	}

	if am.Latency {
		m.requestDuration.
			WithLabelValues(r.Method, r.URL.Path, statusCode, r.Host).
			Observe(float64((after.Sub(before).Seconds())))
	}

	// promhttp.InstrumentHandlerCounter(
	// 	am.requestsTotal,
	// 	promhttp.InstrumentHandlerDuration(
	// 		am.requestDuration,
	// 		promhttp.InstrumentHandlerRequestSize(
	// 			am.requestSize,
	// 			promhttp.InstrumentHandlerResponseSize(
	// 				am.responseSize,
	// 				http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
	// 					next.ServeHTTP(lrw, r)
	// 				}),
	// 			),
	// 		),
	// 	),
	// )

	return err
}
