package xqueue

import "github.com/daodao97/xgo/xlog"

type Queue interface {
	Publish(data string) error
	Subscribe() error
	Close() error
}

type QueueWorker struct {
	queues []Queue
}

func NewQueueWorker(queues ...Queue) *QueueWorker {
	return &QueueWorker{
		queues: queues,
	}
}

func (w *QueueWorker) Start() error {
	for _, queue := range w.queues {
		go func(q Queue) {
			if err := q.Subscribe(); err != nil {
				xlog.Error("subscribe error", xlog.Any("error", err))
			}
		}(queue)
	}

	return nil
}

func (w *QueueWorker) Stop() {
	for _, queue := range w.queues {
		queue.Close()
	}
}
