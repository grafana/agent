package aws_firehose

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/grafana/regexp"
	"github.com/phayes/freeport"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	fnet "github.com/grafana/agent/component/common/net"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/pkg/util"
)

// singleDirectPUTData contains a single record in the firehose request, that it mimics being sent with the DirectPUT API
const singleDirectPUTData = `{"requestId":"a1af4300-6c09-4916-ba8f-12f336176246","timestamp":1684422829730,"records":[{"data":"eyJDSEFOR0UiOi0wLjIzLCJQUklDRSI6NC44LCJUSUNLRVJfU1lNQk9MIjoiTkdDIiwiU0VDVE9SIjoiSEVBTFRIQ0FSRSJ9"}]}`
const expectedRecord = "{\"CHANGE\":-0.23,\"PRICE\":4.8,\"TICKER_SYMBOL\":\"NGC\",\"SECTOR\":\"HEALTHCARE\"}"

type receiver struct {
	ch       chan loki.Entry
	received []loki.Entry
}

func newReceiver(ch chan loki.Entry) *receiver {
	return &receiver{
		ch:       ch,
		received: make([]loki.Entry, 0),
	}
}

func (r *receiver) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-r.ch:
			r.received = append(r.received, e)
		}
	}
}

func TestComponent(t *testing.T) {
	opts := component.Options{
		ID:            "loki.source.awsfirehose",
		Logger:        util.TestFlowLogger(t),
		Registerer:    prometheus.NewRegistry(),
		OnStateChange: func(e component.Exports) {},
	}

	ch1, ch2 := make(chan loki.Entry), make(chan loki.Entry)
	r1, r2 := newReceiver(ch1), newReceiver(ch2)

	// call cancelReceivers to terminate them
	receiverContext, cancelReceivers := context.WithCancel(context.Background())
	go r1.run(receiverContext)
	go r2.run(receiverContext)

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

	componentCtx, cancelComponent := context.WithCancel(context.Background())
	go c.Run(componentCtx)
	defer cancelComponent()

	// small wait for server start
	time.Sleep(200 * time.Millisecond)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/awsfirehose/api/v1/push", port), strings.NewReader(singleDirectPUTData))
	require.NoError(t, err)

	// create client with timeout
	client := http.Client{
		Timeout: time.Second * 5,
	}

	res, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	require.Eventually(t, func() bool {
		return len(r1.received) == 1 && len(r2.received) == 1
	}, time.Second*10, time.Second, "timed out waiting for receivers to get all messages")

	cancelReceivers()

	// r1 and r2 should have received one entry each
	require.Len(t, r1.received, 1)
	require.Len(t, r2.received, 1)
	require.JSONEq(t, expectedRecord, r1.received[0].Line)
	require.JSONEq(t, expectedRecord, r2.received[0].Line)
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
