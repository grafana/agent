package util

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/weaveworks/common/logging"

	"github.com/weaveworks/common/server"

	"github.com/go-kit/kit/log/level"
)

func TestWindowsLogger(t *testing.T) {
	//wl, _ := NewWindowsLogger(nil)
	//level.Info(wl).Log("msg", "failed to update logger", "err")
	wl := NewWinFmtLogger(&server.Config{
		LogLevel: logging.Level{
			Logrus: logrus.InfoLevel,
			Gokit:  level.AllowAll(),
		},
	})
	level.Debug(wl).Log("msg", "debug")
	level.Info(wl).Log("msg", "info")
	level.Warn(wl).Log("msg", "warn")
	level.Error(wl).Log("msg", "error")

}
