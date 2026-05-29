package util

import (
	"context"
	"time"
)

type CloserFunc func() error

func (c CloserFunc) Close() error {
	return c()
}

type CloserContextFunc = func(ctx context.Context) error

func NewCloserContextFunc(f CloserContextFunc, ctx context.Context, timeout time.Duration) CloserFunc {
	return func() error {
		newCtx := ctx
		if timeout > 0 {
			var cancelFunc context.CancelFunc
			newCtx, cancelFunc = context.WithTimeout(ctx, timeout)
			defer cancelFunc()
		}
		return f(newCtx)
	}
}
