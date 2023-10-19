//go:build windows

package windowsevent

import (
	"context"
	"go.etcd.io/bbolt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/svc/eventlog"
)

func TestEventLogger(t *testing.T) {
	createTest(t, "")
}

func TestBookmarkStorage(t *testing.T) {
	datapath := createTest(t, "")
	dbPath := filepath.Join(datapath, "bookmark.db")
	// Lets remove the existing file and ensure it recovers correctly.
	_ = os.WriteFile(dbPath, nil, 0600)
	createTest(t, datapath)
	fbytes, err := os.ReadFile(dbPath)
	require.NoError(t, err)
	for i := 0; i < len(fbytes); i++ {
		// Set every tenth byte to zero.
		if i%10 != 0 {
			continue
		}
		fbytes[i] = 0
	}
	fbytes[28] = 0
	_ = os.WriteFile(dbPath, fbytes, 0600)
	createTest(t, datapath)
}

func TestBookmarkTransition(t *testing.T) {
	dir := createTest(t, "")
	bb, err := bbolt.Open(filepath.Join(dir, "bookmark.db"), os.ModeExclusive, nil)
	require.NoError(t, err)

	var bookmarkString string
	err = bb.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("bookmark"))
		v := b.Get([]byte(bookmarkKey))
		require.NotNil(t, v)
		nv := make([]byte, len(v))
		copy(nv, v)
		bookmarkString = string(nv)
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, bb.Close())

	xmlPath := filepath.Join(dir, "bookmark.xml")
	err = os.WriteFile(xmlPath, []byte(bookmarkString), 0744)
	require.NoError(t, err)
	createTest(t, dir)
	_, err = os.Stat(xmlPath)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func createTest(t *testing.T, dataPath string) string {
	var loggerName = "agent_test"
	//Setup Windows Event log with the log source name and logging levels
	_ = eventlog.InstallAsEventCreate(loggerName, eventlog.Info|eventlog.Warning|eventlog.Error)
	wlog, err := eventlog.Open(loggerName)
	require.NoError(t, err)
	if dataPath == "" {
		dataPath = t.TempDir()
	}
	rec := loki.NewLogsReceiver()
	c, err := New(component.Options{
		ID:       "loki.source.windowsevent.test",
		Logger:   util.TestFlowLogger(t),
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
		Labels:               map[string]string{"job": "windows"},
	})
	require.NoError(t, err)
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
	found := atomic.Bool{}
	go c.Run(ctx)
	tm := time.Now().Format(time.RFC3339Nano)
	err = wlog.Info(2, tm)
	require.NoError(t, err)

	go func() {
		for {
			select {
			case <-ctx.Done():
				// Fail!
				require.True(t, false)
			case e := <-rec.Chan():
				require.Equal(t, model.LabelValue("windows"), e.Labels["job"])
				if strings.Contains(e.Line, tm) {
					found.Store(true)
					return
				}
			}
		}
	}()

	require.Eventually(t, func() bool {
		return found.Load()
	}, 20*time.Second, 500*time.Millisecond)

	cancelFunc()

	require.Eventually(t, func() bool {
		return c.target.closed.Load()
	}, 20*time.Second, 500*time.Millisecond)
	return dataPath
}
