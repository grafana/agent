package prometheus

import (
	"testing"

	"github.com/grafana/agent/service/labelcache"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/prometheus/storage"

	"context"

	"github.com/stretchr/testify/require"
)

func TestRollback(t *testing.T) {
	lc := labelcache.New(nil)
	fanout := NewFanout([]storage.Appendable{NewFanout(nil, "1", prometheus.DefaultRegisterer, lc)}, "", prometheus.DefaultRegisterer, lc)
	app := fanout.Appender(context.Background())
	err := app.Rollback()
	require.NoError(t, err)
}

func TestCommit(t *testing.T) {
	lc := labelcache.New(nil)
	fanout := NewFanout([]storage.Appendable{NewFanout(nil, "1", prometheus.DefaultRegisterer, lc)}, "", prometheus.DefaultRegisterer, lc)
	app := fanout.Appender(context.Background())
	err := app.Commit()
	require.NoError(t, err)
}
