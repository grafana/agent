//go:build windows

package windowsevent

import (
	"context"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/svc/eventlog"
	"os"
	"testing"
	"time"
)

func TestEventLogger(t *testing.T) {
	var loggerName = "agent_test"
	//Setup Windows Event log with the log source name and logging levels
	err := eventlog.InstallAsEventCreate(loggerName, eventlog.Info|eventlog.Warning|eventlog.Error)
	require.NoError(t, err)
	wlog, err := eventlog.Open(loggerName)
	require.NoError(t, err)
	tm := time.Now().Format(time.RFC3339Nano)
	err = wlog.Info(2, tm)
	require.NoError(t, err)
	l, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)
	dataPath, err := os.MkdirTemp("", "loki.source.windowsevent")
	require.NoError(t, err)
	defer os.RemoveAll(dataPath) // clean up
	rec := make(loki.LogsReceiver)
	c, err := New(component.Options{
		ID:       "loki.source.windowsevent.test",
		Logger:   l,
		DataPath: dataPath,
		OnStateChange: func(e component.Exports) {

		},
		Registerer:     prometheus.DefaultRegisterer,
		Tracer:         nil,
		HTTPListenAddr: "",
		HTTPPath:       "",
	}, Arguments{
		Locale:               0,
		EventLogName:         "agent_test",
		XPathQuery:           "*",
		BookmarkPath:         "",
		PollInterval:         10 * time.Millisecond,
		ExcludeEventData:     false,
		ExcludeUserdata:      false,
		UseIncomingTimestamp: false,
		ForwardTo:            []loki.LogsReceiver{rec},
	})
	require.NoError(t, err)
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 100*time.Millisecond)
	go c.Run(ctx)
	select {
	case <-ctx.Done():
		// Fail!
		require.True(t, false)
	case e := <-rec:
		require.True(t, tm == e.Line)
		break
	}
}
