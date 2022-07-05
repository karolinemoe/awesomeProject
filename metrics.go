package awesomeProject

import (
	"strings"
	"time"

	prometheus "github.com/prometheus/client_golang/prometheus"
	promMetrics "github.com/slok/go-http-metrics/metrics/prometheus"
	promMiddleware "github.com/slok/go-http-metrics/middleware"
)

// init must be used to make sure the metric is registered in the collector.
// We only need to register "our" custom metrics as the rest is handled by
// the prometheus lib and /slok/go-http-metrics
func init() {
	mustRegister(
		errorCounter,
		dbDurHistogram,
	)
}

// mustRegister tries to register the collector in a safe way. If the collector is
// already registered it assigns the already registered collector to the provided
// parameter.
//
// Panics if the already registered collector is different.
func mustRegister(cs ...prometheus.Collector) {
	for _, c := range cs {
		prometheus.MustRegister(c)
	}
}

var (
	dbDurBuckets = []float64{.001, .005, .01, .025, .05, .1}

	// Used with promHelpers.HandlerProvider to capture genral metrics for given routes
	Middleware = promMiddleware.New(promMiddleware.Config{ // nolint
		Recorder: promMetrics.NewRecorder(promMetrics.Config{}),
		Service:  "vippsnummer",
	})

	errorCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "custom_error_counter",
		Help: "Number of logged errors.",
	}, []string{"handler"})

	dbDurHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_request_duration_seconds",
		Help:    "The latency of the DB requests.",
		Buckets: dbDurBuckets,
	}, []string{"handler"})
)

// Incase given handler equals to a LogEvent, we use the first word as handler,
// to keep cardinality down.
func CountError(handler string) {
	if handler == "" {
		return
	}
	h := strings.Split(handler, ".")[0]
	errorCounter.WithLabelValues(h).Inc()
}

func ObserveDBDuration(function string, start time.Time) {
	duration := time.Since(start)
	dbDurHistogram.WithLabelValues(function).Observe(duration.Seconds())
}
