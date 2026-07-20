package metric

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/AlexBond702/order-service/internal/app/monitor"
)

const (
	metricEventCreateOrderTotalName       = "order_create_total"
	metricEventCreateOrderTotalHelp       = "Total create orders"
	metricEventCreateDurationSecondsName  = "create_duration_seconds"
	metricEventCreateDurationSecondsHelp  = "Order create duration seconds"
	metricEventCreateAmountCentsTotalName = "created_amount_cents_total"
	metricEventCreateAmountCentsTotalHelp = "Total created order amount in cents"
	metricEventPublishFailedName          = "event_publish_failed_total"
	metricEventPublishFailedHelp          = "Total number of failed Kafka event publish attempts."
	namespace                             = "order"
	subsystem                             = "order"
	labelResult                           = "result"
	createResultSuccess                   = "success"
	createResultError                     = "error"
)

type Order struct {
	createTotal           *prometheus.CounterVec
	createDurationSeconds *prometheus.HistogramVec
	createdAmountTotal    prometheus.Counter
	eventPublishFailed    prometheus.Counter
}
type orderCreateMetric struct {
	parent  *Order
	started time.Time
}

var _ monitor.OrderMetrics = (*Order)(nil)

func NewPrometheusOrder(registerer prometheus.Registerer) (*Order, error) {
	if registerer == nil {
		registerer = prometheus.DefaultRegisterer
	}
	histogramDuration, err := histogramVec(registerer,
		metricEventCreateDurationSecondsName,
		metricEventCreateDurationSecondsHelp,
		prometheus.DefBuckets,
		labelResult,
	)
	if err != nil {
		return nil, fmt.Errorf("register %s: %w", metricEventCreateDurationSecondsName, err)
	}
	counterAmount, err := counter(
		registerer,
		metricEventCreateAmountCentsTotalName,
		metricEventCreateAmountCentsTotalHelp,
	)
	if err != nil {
		return nil, fmt.Errorf("register %s: %w", metricEventCreateAmountCentsTotalName, err)
	}
	counterPublish, err := counter(
		registerer,
		metricEventPublishFailedName,
		metricEventPublishFailedHelp,
	)
	if err != nil {
		return nil, fmt.Errorf("register %s: %w", metricEventPublishFailedName, err)
	}
	counterVecTotal, err := counterVec(registerer,
		metricEventCreateOrderTotalName,
		metricEventCreateOrderTotalHelp,
		"result",
	)
	if err != nil {
		return nil, fmt.Errorf("register %s: %w", metricEventCreateOrderTotalName, err)
	}

	return &Order{
		createTotal:           counterVecTotal,
		createdAmountTotal:    counterAmount,
		createDurationSeconds: histogramDuration,
		eventPublishFailed:    counterPublish,
	}, nil
}

func (m *Order) Create() monitor.OrderCreateMetric {
	return &orderCreateMetric{
		parent:  m,
		started: time.Now(),
	}
}

func (m *orderCreateMetric) Success(price int64) {
	labels := prometheus.Labels{
		labelResult: createResultSuccess,
	}
	m.parent.createTotal.With(labels).Inc()
	m.parent.createDurationSeconds.With(labels).
		Observe(time.Since(m.started).Seconds())
	if price > 0 {
		m.parent.createdAmountTotal.Add(float64(price))
	}
}

func (m *orderCreateMetric) Failed(_ error) {
}

func (m *orderCreateMetric) PublishFailed() {
	m.parent.eventPublishFailed.Inc()
}
