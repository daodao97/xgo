package xcode

import (
	"testing"

	"github.com/daodao97/xgo/xlog"
)

func TestCode(t *testing.T) {
	c := &Code{
		Code:     10000,
		HttpCode: 200,
		Message:  "success",
	}

	xlog.Info("code", xlog.Err(c))
}
