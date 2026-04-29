package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	Success = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "transfer_success_total",
	})
	Failure = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "transfer_failure_total",
	})
)

func Init() {
	prometheus.MustRegister(Success, Failure)
}