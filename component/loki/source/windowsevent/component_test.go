//go:build windows

package windowsevent

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/svc/eventlog"
)

func TestEventLogger(t *testing.T) {
	var loggerName = "agent_test"
	//Setup Windows Event log with the log source name and logging levels
	_ = eventlog.InstallAsEventCreate(loggerName, eventlog.Info|eventlog.Warning|eventlog.Error)
	wlog, err := eventlog.Open(loggerName)
	require.NoError(t, err)
	dataPath := t.TempDir()
	rec := loki.NewLogsReceiver()
	c, err := New(component.Options{
		ID:     "loki.source.windowsevent.test",
		Logger: util.TestFlowLogger(t),
		GetServiceData: func(name string) (interface{}, error) {
			if name == http_service.ServiceName {
				return http_service.Data{
					HTTPListenAddr:   "localhost:12345",
					MemoryListenAddr: "agent.internal:1245",
					BaseHTTPPath:     "/",
					DialFunc:         (&net.Dialer{}).DialContext,
				}, nil
			}

			return nil, fmt.Errorf("service %q does not exist", name)
		},
		DataPath: dataPath,
		OnStateChange: func(e component.Exports) {

		},
		Registerer: prometheus.DefaultRegisterer,
		Tracer:     nil,
	}, Arguments{
		Locale:               0,
		EventLogName:         "Application",
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
	ctx, cancelFunc := context.WithTimeout(ctx, 10*time.Second)
	found := false
	go c.Run(ctx)
	tm := time.Now().Format(time.RFC3339Nano)
	err = wlog.Info(2, tm)
	require.NoError(t, err)
	select {
	case <-ctx.Done():
		// Fail!
		require.True(t, false)
	case e := <-rec.Chan():
		if strings.Contains(e.Line, tm) {
			found = true
			break
		}
	}
	cancelFunc()
	require.True(t, found)
}
