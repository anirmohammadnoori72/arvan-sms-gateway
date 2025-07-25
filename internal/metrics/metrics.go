package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	TotalSMSRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sms_requests_total",
			Help: "Total number of SMS requests received",
		},
	)

	QueueLength = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "sms_queue_length",
			Help: "Current length of the SMS queue",
		},
	)

	KafkaMessages = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_messages_sent_total",
			Help: "Total number of messages successfully sent to Kafka",
		},
	)

	KafkaErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_messages_errors_total",
			Help: "Total number of Kafka message send errors",
		},
	)
)

func InitMetrics() {
	prometheus.MustRegister(TotalSMSRequests)
	prometheus.MustRegister(QueueLength)
	prometheus.MustRegister(KafkaMessages)
	prometheus.MustRegister(KafkaErrors)
}
