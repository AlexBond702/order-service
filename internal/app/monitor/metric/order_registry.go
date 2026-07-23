package metric

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func registerCollector(
	registerer prometheus.Registerer,
	collector prometheus.Collector,
) (prometheus.Collector, error) {
	err := registerer.Register(collector)
	if err != nil {
		var errAlreadyRegisterError prometheus.AlreadyRegisteredError
		if errors.As(err, &errAlreadyRegisterError) {
			return errAlreadyRegisterError.ExistingCollector, nil
		}
		return nil, err
	}
	return nil, nil
}

func counter(reg prometheus.Registerer, name, help string) (prometheus.Counter, error) {
	c := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	})
	existing, err := registerCollector(reg, c)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		counter, ok := existing.(prometheus.Counter)
		if ok {
			return counter, nil
		}
		return nil, fmt.Errorf(
			"metric %q already registered with incompatible collector type %T",
			name,
			existing,
		)
	}
	return c, nil
}

func counterVec(reg prometheus.Registerer, name, help string, labels ...string) (*prometheus.CounterVec, error) {
	c := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	}, labels)
	existing, err := registerCollector(reg, c)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		counterVec, ok := existing.(prometheus.CounterVec)
		if ok {
			return &counterVec, nil
		}
		return nil, fmt.Errorf(
			"metric %q already registered with incompatible collector type %T",
			name,
			existing,
		)
	}
	return c, nil
}

func histogramVec(reg prometheus.Registerer, name, help string, buckets []float64, labels ...string) (*prometheus.HistogramVec, error) {
	c := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
		Buckets:   buckets,
	}, labels)
	existing, err := registerCollector(reg, c)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		counterHis, ok := existing.(prometheus.HistogramVec)
		if ok {
			return &counterHis, nil
		}
		return nil, fmt.Errorf(
			"metric %q already registered with incompatible collector type %T",
			name,
			existing,
		)
	}
	return c, nil
}
