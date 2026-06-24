package entity

import (
	"context"
	"time"
)

type MetricObservation struct {
	Period   time.Duration
	Callback func(ctx context.Context)
}
