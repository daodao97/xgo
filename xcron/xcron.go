package xcron

import (
	"context"

	"github.com/daodao97/xgo/xlog"
	"github.com/daodao97/xgo/xutil"
	"github.com/robfig/cron/v3"
)

type Job struct {
	Name      string // 任务名称
	Spec      string // 任务执行时间
	Func      func() // 任务执行函数
	Immediate bool   // 是否立即执行
}

type Cron struct {
	jobs []Job
	cron *cron.Cron
}

func New(jobs ...Job) *Cron {
	return &Cron{
		jobs: jobs,
		cron: cron.New(
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
	opts = append(opts, cron.WithLogger(NewLogger()), cron.WithChain(cron.Recover(NewLogger())))
	return cron.New(
		opts...,
	)
}

func (c *Cron) Start() error {
	c.cron.Start()
	for _, job := range c.jobs {
		_, err := c.cron.AddFunc(job.Spec, job.Func)
		if err != nil {
			return err
		}
		xlog.Debug("add job", xlog.String("name", job.Name), xlog.String("spec", job.Spec))
		if job.Immediate {
			xutil.Go(context.Background(), job.Func)
		}
	}
	return nil
}

func (c *Cron) Stop() {
	c.cron.Stop()
}
