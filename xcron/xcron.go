package xcron

import (
	"context"
	"fmt"
	"time"

	"github.com/daodao97/xgo/xlog"
	"github.com/daodao97/xgo/xutil"
	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"
)

type Job struct {
	Name           string        // 任务名称
	Spec           string        // 任务执行时间
	Func           func()        // 任务执行函数
	Immediate      bool          // 是否立即执行
	EnableDistLock bool          // 是否启用分布式锁
	LockTimeout    time.Duration // 锁超时时间，默认5分钟
	LockRetryDelay time.Duration // 获取锁失败重试延迟，默认1秒
}

type Cron struct {
	name string
	jobs []Job
	cron *cron.Cron
	rdb  redis.UniversalClient
}

type Option func(*Cron)

func WithName(name string) Option {
	return func(c *Cron) {
		c.name = name
	}
}

func WithRdb(rdb redis.UniversalClient) Option {
	return func(c *Cron) {
		c.rdb = rdb
	}
}

func WithCron(cron *cron.Cron) Option {
	return func(c *Cron) {
		c.cron = cron
	}
}

func WithJobs(jobs ...Job) Option {
	return func(c *Cron) {
		c.jobs = append(c.jobs, jobs...)
	}
}

func New2(opts ...Option) *Cron {
	c := &Cron{
		jobs: []Job{},
		cron: cron.New(
			cron.WithSeconds(),
			cron.WithLogger(NewLogger()),
			cron.WithChain(cron.Recover(NewLogger())),
		),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func New(jobs ...Job) *Cron {
	return &Cron{
		jobs: jobs,
		cron: cron.New(
			cron.WithSeconds(),
			cron.WithLogger(NewLogger()),
			cron.WithChain(cron.Recover(NewLogger())),
		),
	}
}

func NewWithCron(c *cron.Cron, jobs ...Job) *Cron {
	return &Cron{
		jobs: jobs,
		cron: c,
	}
}

func NewCron(opts ...cron.Option) *cron.Cron {
	opts = append([]cron.Option{cron.WithSeconds()}, opts...)
	opts = append(opts, cron.WithLogger(NewLogger()), cron.WithChain(cron.Recover(NewLogger())))
	return cron.New(
		opts...,
	)
}

func (c *Cron) Start() error {
	c.cron.Start()
	for _, job := range c.jobs {
		xlog.Debug("add job", xlog.String("name", job.Name), xlog.String("spec", job.Spec), xlog.Bool("dist_lock", job.EnableDistLock))

		// Use executeWithLock wrapper for job execution
		jobFunc := c.executeWithLock(job)
		_, err := c.cron.AddFunc(job.Spec, jobFunc)
		if err != nil {
			return err
		}

		if job.Immediate {
			xutil.Go(context.Background(), jobFunc)
		}
	}
	return nil
}

func (c *Cron) Stop() {
	c.cron.Stop()
}

// tryLock attempts to acquire a distributed lock using Redis SET NX EX
func (c *Cron) tryLock(ctx context.Context, lockKey string, timeout time.Duration) (bool, error) {
	if c.rdb == nil {
		return false, fmt.Errorf("redis client not available")
	}

	// Use SET with NX (only set if not exists) and EX (set expiration time)
	result := c.rdb.SetNX(ctx, lockKey, "locked", timeout)
	return result.Val(), result.Err()
}

// releaseLock releases the distributed lock
func (c *Cron) releaseLock(ctx context.Context, lockKey string) error {
	if c.rdb == nil {
		return nil
	}
	return c.rdb.Del(ctx, lockKey).Err()
}

// executeWithLock wraps job execution with distributed lock
func (c *Cron) executeWithLock(job Job) func() {
	return func() {
		if !job.EnableDistLock {
			job.Func()
			return
		}

		if c.rdb == nil {
			xlog.Warn("redis client not available, executing job without distributed lock",
				xlog.String("job", job.Name))
			job.Func()
			return
		}

		// Set default values
		lockTimeout := job.LockTimeout
		if lockTimeout == 0 {
			lockTimeout = 5 * time.Minute
		}

		retryDelay := job.LockRetryDelay
		if retryDelay == 0 {
			retryDelay = 1 * time.Second
		}

		lockKey := fmt.Sprintf("xcron:lock:%s:%s", c.name, job.Name)
		ctx := context.Background()

		// Try to obtain lock with retry
		var acquired bool
		var err error

		for i := 0; i < 3; i++ { // Maximum 3 retry attempts
			acquired, err = c.tryLock(ctx, lockKey, lockTimeout)
			if err != nil {
				xlog.Warn("error trying to acquire lock",
					xlog.String("job", job.Name),
					xlog.String("key", lockKey),
					xlog.String("error", err.Error()))
				return
			}

			if acquired {
				break
			}

			if i < 2 { // Don't sleep on the last attempt
				time.Sleep(retryDelay)
			}
		}

		if !acquired {
			xlog.Debug("failed to acquire distributed lock, job already running",
				xlog.String("job", job.Name),
				xlog.String("key", lockKey))
			return
		}

		defer func() {
			if err := c.releaseLock(ctx, lockKey); err != nil {
				xlog.Warn("failed to release lock",
					xlog.String("job", job.Name),
					xlog.String("key", lockKey),
					xlog.String("error", err.Error()))
			}
		}()

		xlog.Debug("obtained distributed lock",
			xlog.String("job", job.Name),
			xlog.String("key", lockKey))

		job.Func()
	}
}
