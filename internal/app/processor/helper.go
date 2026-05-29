package processor

import (
	"context"
	"io"
	"sync"
)

func Wrap(ctx context.Context, wg *sync.WaitGroup, cb func(ctx2 context.Context)) {
	if wg != nil {
		wg.Add(1)
	}
	go func() {
		if wg != nil {
			defer wg.Done()
		}
		select {
		case <-ctx.Done():
			return
		default:
			cb(ctx)
		}
	}()
}

func WatchForShutdown(ctx context.Context, wg *sync.WaitGroup, closer io.Closer) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		_ = closer.Close()
	}()
}
