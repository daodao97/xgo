package xcron

import (
	"log/slog"

	"github.com/daodao97/xgo/xlog"
)

func newLogger() *CronLogger {
	return &CronLogger{
		logger: xlog.GetLogger(),
	}
}

type CronLogger struct {
	logger *slog.Logger
}

func (l *CronLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, keysAndValues...)
}

func (l *CronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, xlog.Err(err))
	l.logger.Error(msg, keysAndValues...)
}
