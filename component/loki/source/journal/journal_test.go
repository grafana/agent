//go:build linux && cgo && promtail_journal_enabled
// +build linux,cgo,promtail_journal_enabled

package journal

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-systemd/v22/journal"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestJournal(t *testing.T) {
	// Create opts for component
	l, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)
	tmp, err := os.MkdirTemp("", "testjrnl")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(tmp)
	})
	lr := make(loki.LogsReceiver)
	c, err := New(component.Options{
		ID:         "loki.source.journal.test",
		Logger:     l,
		DataPath:   tmp,
		Registerer: prometheus.DefaultRegisterer,
	}, Arguments{
		FormatAsJson: false,
		MaxAge:       7 * time.Hour,
		Path:         "",
		Receivers:    []loki.LogsReceiver{lr},
	})
	require.NoError(t, err)
	ctx := context.Background()
	ctx, cnc := context.WithTimeout(ctx, 5*time.Second)
	defer cnc()
	go c.Run(ctx)
	ts := time.Now().String()
	err = journal.Send(ts, journal.PriInfo, nil)
	require.NoError(t, err)
	found := false
	for !found {
		select {
		case <-ctx.Done():
			found = true
			// Timed out getting message
			require.True(t, false)
		case msg := <-lr:
			if strings.Contains(msg.Line, ts) {
				found = true
				break
			}
		}
	}
	require.True(t, found)
}
