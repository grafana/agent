package metrics

import (
	"context"
	"fmt"
	"github.com/golang/snappy"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/util"
	http_service "github.com/grafana/agent/service/http"
	"github.com/grafana/agent/service/labelstore"
	prometheus_client "github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"net"
	"strings"
	"testing"
	"time"
)

func TestMetrcsGeneration(t *testing.T) {
	opts := component.Options{
		Logger:     util.TestFlowLogger(t),
		Registerer: prometheus_client.NewRegistry(),
		GetServiceData: func(name string) (interface{}, error) {
			switch name {
			case http_service.ServiceName:
				return http_service.Data{
					HTTPListenAddr:   "localhost:12345",
					MemoryListenAddr: "agent.internal:1245",
					BaseHTTPPath:     "/",
					DialFunc:         (&net.Dialer{}).DialContext,
				}, nil
			case labelstore.ServiceName:
				return labelstore.New(nil), nil
			default:
				return nil, fmt.Errorf("service %q does not exist", name)
			}
		},
	}

	s, err := NewComponent(opts, Arguments{
		NumberOfInstances: 1,
		NumberOfMetrics:   1,
		NumberOfSeries:    1,
		MetricsRefresh:    1 * time.Minute,
		ChurnPercent:      0,
	})
	require.NoError(t, err)
	ctx := context.Background()
	ctx, cncl := context.WithTimeout(ctx, 20*time.Second)
	defer cncl()
	go s.Run(ctx)
	var bb [][]byte
	require.Eventually(t, func() bool {
		bb = s.data()
		return len(bb) > 0 && len(bb[0]) > 2
	}, 10*time.Second, 100*time.Millisecond)
	require.Len(t, bb, 1)
	out, err := snappy.Decode(nil, bb[0])
	require.NoError(t, err)
	metrics := string(out)
	require.True(t, strings.Contains(metrics, "counter"))
	require.True(t, strings.Contains(metrics, "agent_metric_test_0"))

}
