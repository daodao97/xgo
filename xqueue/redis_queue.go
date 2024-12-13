package xqueue

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/daodao97/xgo/xlog"
	"github.com/go-redis/redis/v8"
)

var queueContainer = make(map[string]Queue)
var queueLock sync.Mutex

func init() {
	queueLock = sync.Mutex{}
}

func AddQueue(redis *redis.Client, topic string, handler func(data string), workers int) Queue {
	queueLock.Lock()
	defer queueLock.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	queue := &RedisQueue{
		redis:   redis,
		topic:   topic,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
		workers: workers,
		jobs:    make(chan string, workers), // 初始化任务队列
	}

	queue.startWorkers() // 启动 workers
	queueContainer[topic] = queue
	return queue
}

func GetQueue(topic string) Queue {
	queueLock.Lock()
	defer queueLock.Unlock()

	return queueContainer[topic]
}

type RedisQueue struct {
	redis   *redis.Client
	queue   *redis.PubSub
	topic   string
	handler func(data string)
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	workers int         // 添加 workers 配置
	jobs    chan string // 添加任务队列
}

func (q *RedisQueue) startWorkers() {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go func(workerID int) {
			defer q.wg.Done()
			for {
				select {
				case <-q.ctx.Done():
					xlog.Debug("worker stopping", xlog.String("topic", q.topic), xlog.Int("workerID", workerID))
					return
				case msg, ok := <-q.jobs:
					if !ok {
						return
					}
					xlog.Debug("processing message",
						xlog.String("topic", q.topic),
						xlog.String("data", msg),
						xlog.Int("workerID", workerID))

					func() {
						defer func() {
							if r := recover(); r != nil {
								xlog.Error("RedisQueue recover panic in handler",
									xlog.Any("error", r),
									xlog.String("stack", string(debug.Stack())))
							}
						}()
						q.handler(msg)
					}()
				}
			}
		}(i)
	}
}

func (q *RedisQueue) Publish(data string) error {
	return q.redis.Publish(context.Background(), q.topic, data).Err()
}

func (q *RedisQueue) Subscribe() error {
	xlog.Debug("start subscribe", xlog.String("topic", q.topic), xlog.Any("workers", q.workers))
	pubsub := q.redis.Subscribe(context.Background(), q.topic)
	q.queue = pubsub
	defer pubsub.Close()

	if _, err := pubsub.Receive(context.Background()); err != nil {
		return err
	}

	ch := pubsub.Channel()
	for {
		select {
		case <-q.ctx.Done():
			xlog.Debug("stopping subscription", xlog.String("topic", q.topic))
			close(q.jobs) // 关闭任务队列
			return nil
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			q.jobs <- msg.Payload // 将消息发送到任务队列
		}
	}
}

func (q *RedisQueue) Close() error {
	queueLock.Lock()
	defer queueLock.Unlock()

	// 1. 停止接收新消息
	q.cancel()
	xlog.Debug("stopped accepting new messages", xlog.String("topic", q.topic))

	// 2. 等待当前消息处理完成
	q.wg.Wait()
	xlog.Debug("all messages processed", xlog.String("topic", q.topic))

	// 3. 从容器中删除
	delete(queueContainer, q.topic)

	// 4. 关闭订阅
	xlog.Debug("closing subscribe", xlog.String("topic", q.topic))
	if q.queue != nil {
		return q.queue.Close()
	}
	return nil
}
