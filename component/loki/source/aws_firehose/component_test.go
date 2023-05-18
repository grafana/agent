package aws_firehose

import (
	"context"
	"fmt"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	fnet "github.com/grafana/agent/component/common/net"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/regexp"
	"github.com/phayes/freeport"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
	"time"
)

const directPutData = `{"requestId":"a1af4300-6c09-4916-ba8f-12f336176246","timestamp":1684422829730,"records":[{"data":"eyJDSEFOR0UiOi0wLjIzLCJQUklDRSI6NC44LCJUSUNLRVJfU1lNQk9MIjoiTkdDIiwiU0VDVE9SIjoiSEVBTFRIQ0FSRSJ9"},{"data":"eyJDSEFOR0UiOjYuNzYsIlBSSUNFIjo4Mi41NiwiVElDS0VSX1NZTUJPTCI6IlNMVyIsIlNFQ1RPUiI6IkVORVJHWSJ9"},{"data":"eyJDSEFOR0UiOi01LjkyLCJQUklDRSI6MTk5LjA4LCJUSUNLRVJfU1lNQk9MIjoiSEpWIiwiU0VDVE9SIjoiRU5FUkdZIn0="}]}`

func TestComponent(t *testing.T) {
	opts := component.Options{
		ID:            "loki.source.awsfirehose",
		Logger:        util.TestFlowLogger(t),
		Registerer:    prometheus.NewRegistry(),
		OnStateChange: func(e component.Exports) {},
	}

	ch1, ch2 := make(chan loki.Entry), make(chan loki.Entry)
	args := Arguments{}

	port, err := freeport.GetFreePort()
	require.NoError(t, err)
	args.Server = &fnet.ServerConfig{
		HTTP: &fnet.HTTPConfig{
			ListenAddress: "localhost",
			ListenPort:    port,
		},
		// assign random grpc port
		GRPC: &fnet.GRPCConfig{ListenPort: 0},
	}
	args.ForwardTo = []loki.LogsReceiver{ch1, ch2}
	args.RelabelRules = exportedRules

	// Create and run the component.
	c, err := New(opts, args)
	require.NoError(t, err)

	go c.Run(context.Background())
	time.Sleep(200 * time.Millisecond)

	// Create a GCP PushRequest and send it to the launched server.
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/awsfirehose/api/v1/push", port), strings.NewReader(directPutData))
	require.NoError(t, err)

	sent := make(chan struct{}, 1)
	go func() {
		client := http.Client{}
		client.Timeout = time.Second * 5
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)
		sent <- struct{}{}
	}()

	// Check the received log entries
	//wantLabelSet := model.LabelSet{"foo": "bar", "message_id": "5187581549398349", "resource_type": "k8s_cluster"}
	wantLogLine := "{\"CHANGE\":-0.23,\"PRICE\":4.8,\"TICKER_SYMBOL\":\"NGC\",\"SECTOR\":\"HEALTHCARE\"}"

	for i := 0; i < 6; i++ {
		select {
		case logEntry := <-ch1:
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.JSONEq(t, wantLogLine, logEntry.Line)
			//require.Equal(t, wantLabelSet, logEntry.Labels)
		case logEntry := <-ch2:
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.JSONEq(t, wantLogLine, logEntry.Line)
			//require.Equal(t, wantLabelSet, logEntry.Labels)
		case <-time.After(5 * time.Second):
			require.FailNow(t, "failed waiting for log line")
		}
	}

	select {
	case <-sent:
	case <-time.After(10 * time.Second):
		require.FailNow(t, "failed waiting for routine that sent the test request")
	}

}

var exportedRules = flow_relabel.Rules{
	{
		SourceLabels: []string{"__gcp_message_id"},
		Regex:        mustNewRegexp("(.*)"),
		Action:       flow_relabel.Replace,
		Replacement:  "$1",
		TargetLabel:  "message_id",
	},
	{
		SourceLabels: []string{"__gcp_resource_type"},
		Regex:        mustNewRegexp("(.*)"),
		Action:       flow_relabel.Replace,
		Replacement:  "$1",
		TargetLabel:  "resource_type",
	},
}

func mustNewRegexp(s string) flow_relabel.Regexp {
	re, err := regexp.Compile("^(?:" + s + ")$")
	if err != nil {
		panic(err)
	}
	return flow_relabel.Regexp{Regexp: re}
}
