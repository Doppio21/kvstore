package server

import "github.com/prometheus/client_golang/prometheus"

type serverMetricsCollector struct {
	counter prometheus.Counter
	time    prometheus.Histogram
}

func newMetricColletor() *serverMetricsCollector {
	return &serverMetricsCollector{
		counter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "req_total",
		}),
		time: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "req_duration",
			Buckets: prometheus.DefBuckets,
		}),
	}
}

func (m *serverMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.counter.Desc()
	ch <- m.time.Desc()
}

func (m *serverMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- m.counter
	ch <- m.time
}
