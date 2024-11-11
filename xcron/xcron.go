package xcron

import (
	"github.com/daodao97/xgo/xlog"
	"github.com/robfig/cron/v3"
)

type Job struct {
	Spec string
	Func func()
}

type Cron struct {
	jobs []Job
	cron *cron.Cron
}

func New(jobs ...Job) *Cron {
	return &Cron{
		jobs: jobs,
		cron: cron.New(
			cron.WithLogger(newLogger()),
			cron.WithChain(cron.Recover(newLogger())),
		),
	}
}

func (c *Cron) Start() error {
	c.cron.Start()
	for _, job := range c.jobs {
		_, err := c.cron.AddFunc(job.Spec, job.Func)
		if err != nil {
			return err
		}
		xlog.Debug("add job", "spec", job.Spec)
	}
	return nil
}

func (c *Cron) Stop() {
	c.cron.Stop()
}
