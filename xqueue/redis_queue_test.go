package xqueue

import (
	"testing"
	"time"

	"github.com/daodao97/xgo/xlog"
	"github.com/go-redis/redis/v8"
)

func TestRedisQueue_Publish(t *testing.T) {

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	q := AddQueue(client, "test", func(data string) {
		xlog.Debug("received message", xlog.String("data", data))
	}, 1)

	q.Publish("test-data-" + time.Now().Format(time.RFC3339))
}
