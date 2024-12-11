package xcron

import (
	"log/slog"

	"github.com/daodao97/xgo/xlog"
)

func newLogger() *CronLogger {
	logger := xlog.GetLogger()

	return &CronLogger{
		logger: logger,
	}
}

type CronLogger struct {
	logger *slog.Logger
}

func (l *CronLogger) Info(msg string, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, slog.String("module", "cron"))
	l.logger.Info(msg, keysAndValues...)
}

func (l *CronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, xlog.Err(err))
	keysAndValues = append(keysAndValues, slog.String("module", "cron"))
	l.logger.Error(msg, keysAndValues...)
}
