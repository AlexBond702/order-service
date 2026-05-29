package processor

import (
	"context"
	"sync"
)

type Processor interface {
	StartAsync(ctx context.Context, wg *sync.WaitGroup)
}

//goland:noinspection GoNameStartsWithPackageName
type ProcessorFunc func(ctx context.Context, wg *sync.WaitGroup)

func (p ProcessorFunc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	p(ctx, wg)
}
