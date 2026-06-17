package monitor

import (
	"context"
	service "github.com/AlexBond702/order-service/internal/app/module"
	"github.com/AlexBond702/order-service/internal/app/processor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"slices"
	"sync"
	"time"
)

type promProc struct {
	tasks []promProcTask
	chCb  chan func(ctx context.Context)
}

type promProcTask struct {
	NextCall int64
	Period   int64
	Callback func(ctx context.Context)
}

func NewPrometheusObserver(tasks ...service.Metered) processor.Processor {
	var promTasks []promProcTask
	factory := promauto.With(prometheus.DefaultRegisterer)

	for _, task := range tasks {
		metrics := task.ProvideMetrics(factory)
		for _, metric := range metrics {
			if metric.Callback == nil {
				continue
			}
			promTasks = append(promTasks, promProcTask{
				Period:   int64(metric.Period.Seconds()),
				Callback: metric.Callback,
			})
		}
	}
	return &promProc{
		tasks: promTasks,
		chCb:  make(chan func(ctx context.Context), 2048),
	}
}
func (p *promProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	nowUnix := time.Now().UnixNano()

	// Первый запуск всех callbacks и установка NextCall
	for i, n := 0, len(p.tasks); i < n; i++ {
		p.tasks[i].NextCall = nowUnix + p.tasks[i].Period
		p.chCb <- p.tasks[i].Callback
	}

	p.reSortTasks()

	// Запуск горутин
	const CheckTime = 5 * time.Second
	chTick := time.Tick(CheckTime)

	go processor.Loop(ctx, wg, p.chCb, p.execCallback)
	go processor.Loop(ctx, wg, chTick, p.checkUpcomingTasks)
}

// execCallback выполняет callback из канала
func (p *promProc) execCallback(
	ctx context.Context, _ *sync.WaitGroup, cb func(ctx context.Context),
) {
	cb(ctx)
}

// checkUpcomingTasks проверяет задачи, которые пора выполнить
func (p *promProc) checkUpcomingTasks(
	_ context.Context, _ *sync.WaitGroup, now time.Time,
) {
	nowUnix := now.UnixNano()

	i := 0
	for n := len(p.tasks); i < n && p.tasks[i].NextCall <= nowUnix; i++ {
		p.chCb <- p.tasks[i].Callback
		p.tasks[i].NextCall += p.tasks[i].Period
	}

	if i != 0 {
		p.reSortTasks()
	}
}

// reSortTasks сортирует задачи по времени следующего вызова
func (p *promProc) reSortTasks() {
	slices.SortFunc(p.tasks, func(a, b promProcTask) int {
		switch {
		case a.NextCall < b.NextCall:
			return -1
		case a.NextCall > b.NextCall:
			return 1
		default:
			return 0
		}
	})
}
