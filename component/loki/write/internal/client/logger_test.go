package client

// This code is copied from Promtail. The client package is used to configure
// and run the clients that can send log entries to a Loki instance.

import (
	"net/url"
	"testing"
	"time"

	"github.com/grafana/agent/component/common/loki"
	cortexflag "github.com/grafana/dskit/flagext"
	"github.com/grafana/loki/pkg/logproto"
	util_log "github.com/grafana/loki/pkg/util/log"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	_, err := NewLogger(nilMetrics, nil, util_log.Logger, []Config{}...)
	require.Error(t, err)

	l, err := NewLogger(nilMetrics, nil, util_log.Logger, []Config{{URL: cortexflag.URLValue{URL: &url.URL{Host: "string"}}}}...)
	require.NoError(t, err)
	l.Chan() <- loki.Entry{Labels: model.LabelSet{"foo": "bar"}, Entry: logproto.Entry{Timestamp: time.Now(), Line: "entry"}}
	l.Stop()
}
