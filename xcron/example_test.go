package xcron_test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/daodao97/xgo/xcron"
	"github.com/daodao97/xgo/xredis"
)

func TestExampleDistributedLock(t *testing.T) {
	xredis.Inits([]xredis.Options{
		{
			Name: "default",
			Addr: "localhost:6379",
		},
	})

	var counter int64

	// 先测试一个不需要分布式锁的任务，确保基本功能正常
	simpleJob := xcron.Job{
		Name: "simple-test-job",
		Spec: "*/1 * * * * *", // 每秒执行一次
		Func: func() {
			atomic.AddInt64(&counter, 1)
			fmt.Printf("Simple job executed, counter: %d\n", atomic.LoadInt64(&counter))
		},
	}

	fmt.Println("Testing simple job without distributed lock...")
	simpleCron := xcron.New2(xcron.WithName("simple"), xcron.WithJobs(simpleJob))
	err := simpleCron.Start()
	if err != nil {
		t.Fatalf("Failed to start simple cron: %v", err)
	}
	time.Sleep(3 * time.Second)
	simpleCron.Stop()
	fmt.Printf("Simple job counter: %d\n", atomic.LoadInt64(&counter))

	// 重置计数器
	atomic.StoreInt64(&counter, 0)

	// 创建一个需要分布式锁的任务
	job := xcron.Job{
		Name:           "test-distributed-job",
		Spec:           "*/1 * * * * *", // 每秒执行一次
		EnableDistLock: true,
		LockTimeout:    30 * time.Second,
		LockRetryDelay: 500 * time.Millisecond,
		Func: func() {
			atomic.AddInt64(&counter, 1)
			fmt.Printf("Job executed, counter: %d\n", atomic.LoadInt64(&counter))
			time.Sleep(2 * time.Second) // 模拟任务执行时间
		},
	}

	// 创建两个 Cron 实例模拟分布式环境
	// 使用 New2 来确保 Redis 客户端被初始化
	cron1 := xcron.New2(xcron.WithName("instance1"), xcron.WithJobs(job), xcron.WithRdb(xredis.Get()))
	cron2 := xcron.New2(xcron.WithName("instance2"), xcron.WithJobs(job), xcron.WithRdb(xredis.Get()))

	// 启动两个实例
	err1 := cron1.Start()
	if err1 != nil {
		t.Fatalf("Failed to start cron1: %v", err1)
	}
	err2 := cron2.Start()
	if err2 != nil {
		t.Fatalf("Failed to start cron2: %v", err2)
	}

	// 运行10秒观察结果
	time.Sleep(10 * time.Second)

	cron1.Stop()
	cron2.Stop()

	// 由于分布式锁，counter 应该远小于20（如果没有锁，两个实例各执行10次）
	fmt.Printf("Final counter: %d (expected less than 20 due to distributed locking)\n", atomic.LoadInt64(&counter))
}

func TestExample(t *testing.T) {
	job := xcron.Job{
		Name: "test-distributed-job",
		Spec: "*/1 * * * * *", // 每秒执行一次
		Func: func() {
			fmt.Println("test job executed")
		},
	}

	cron := xcron.New2(xcron.WithName("test"), xcron.WithJobs(job))
	err := cron.Start()
	if err != nil {
		t.Fatalf("Failed to start cron: %v", err)
	}
	time.Sleep(10 * time.Second)
	cron.Stop()
}
