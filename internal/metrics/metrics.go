package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RequestsTotal    prometheus.Counter
	FilteredRequests prometheus.Counter
	EventsProcessed  prometheus.Counter
	ErrorsTotal      prometheus.Counter
	S3Writes         prometheus.Counter
	S3Errors         prometheus.Counter
)

func Init() {
	RequestsTotal = (prometheus.NewCounter(prometheus.CounterOpts{
		Name: "requests_total",
		Help: "Total HTTP requests",
	}))
	FilteredRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "filtered_requests_total",
		Help: "Filtered requests due to invalid tier",
	})
	EventsProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "events_processed_total",
		Help: "Successfully processed events",
	})
	ErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "errors_total",
		Help: "Total errors",
	})
	S3Writes = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "s3_writes_total",
		Help: "Total S3 writes",
	})
	S3Errors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "s3_errors_total",
		Help: "S3 write errors",
	})

	prometheus.MustRegister(RequestsTotal, FilteredRequests, EventsProcessed, ErrorsTotal, S3Writes, S3Errors)
}
