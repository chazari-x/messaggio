package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Prometheus struct {
	NewMessageGauge        prometheus.Gauge
	ProcessingMessageGauge prometheus.Gauge
	OkMessageCounter       prometheus.Counter
	ErrorMessageCounter    prometheus.Counter
}

func New() *Prometheus {
	pr := Prometheus{
		NewMessageGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "new_message_gauge",
			Help: "The total number of new messages",
		}),
		ProcessingMessageGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "processing_message_gauge",
			Help: "The total number of processing messages",
		}),
		OkMessageCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "ok_message_counter",
			Help: "The total number of ok messages",
		}),
		ErrorMessageCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "error_message_counter",
			Help: "The total number of error messages",
		}),
	}

	prometheus.MustRegister(pr.NewMessageGauge, pr.ProcessingMessageGauge, pr.OkMessageCounter, pr.ErrorMessageCounter)

	return &pr
}
