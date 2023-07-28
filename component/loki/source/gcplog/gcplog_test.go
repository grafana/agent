package gcplog

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
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	fnet "github.com/grafana/agent/component/common/net"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/loki/source/gcplog/gcptypes"
	"github.com/grafana/agent/pkg/util"
)

// TODO (@tpaschalis) We can't test this easily as there's no way to inject
// the mock PubSub client inside the component, but we'll find a workaround.
func TestPull(t *testing.T) {}

func TestPush(t *testing.T) {
	opts := component.Options{
		Logger:        util.TestFlowLogger(t),
		Registerer:    prometheus.NewRegistry(),
		OnStateChange: func(e component.Exports) {},
	}

	ch1, ch2 := loki.NewLogsReceiver(), loki.NewLogsReceiver()
	args := Arguments{}

	port, err := freeport.GetFreePort()
	require.NoError(t, err)
	args.PushTarget = &gcptypes.PushConfig{
		Server: &fnet.ServerConfig{
			HTTP: &fnet.HTTPConfig{
				ListenAddress: "localhost",
				ListenPort:    port,
			},
			// assign random grpc port
			GRPC: &fnet.GRPCConfig{ListenPort: 0},
		},
		Labels: map[string]string{
			"foo": "bar",
		},
	}
	args.ForwardTo = []loki.LogsReceiver{ch1, ch2}
	args.RelabelRules = exportedRules

	// Create and run the component.
	c, err := New(opts, args)
	require.NoError(t, err)

	go c.Run(context.Background())
	time.Sleep(200 * time.Millisecond)

	// Create a GCP PushRequest and send it to the launched server.
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/gcp/api/v1/push", port), strings.NewReader(testPushPayload))
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.StatusCode)

	// Check the received log entries
	wantLabelSet := model.LabelSet{"foo": "bar", "message_id": "5187581549398349", "resource_type": "k8s_cluster"}
	wantLogLine := "{\"insertId\":\"4affa858-e5f2-47f7-9254-e609b5c014d0\",\"labels\":{},\"logName\":\"projects/test-project/logs/cloudaudit.googleapis.com%2Fdata_access\",\"receiveTimestamp\":\"2022-09-06T18:07:43.417714046Z\",\"resource\":{\"labels\":{\"cluster_name\":\"dev-us-central-42\",\"location\":\"us-central1\",\"project_id\":\"test-project\"},\"type\":\"k8s_cluster\"},\"timestamp\":\"2022-09-06T18:07:42.363113Z\"}\n"

	for i := 0; i < 2; i++ {
		select {
		case logEntry := <-ch1.Chan():
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, wantLogLine, logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case logEntry := <-ch2.Chan():
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, wantLogLine, logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case <-time.After(5 * time.Second):
			require.FailNow(t, "failed waiting for log line")
		}
	}
}

const testPushPayload = `
{
	"message": {
		"attributes": {
			"logging.googleapis.com/timestamp": "2022-07-25T22:19:09.903683708Z"
		},
		"data": "eyJpbnNlcnRJZCI6IjRhZmZhODU4LWU1ZjItNDdmNy05MjU0LWU2MDliNWMwMTRkMCIsImxhYmVscyI6e30sImxvZ05hbWUiOiJwcm9qZWN0cy90ZXN0LXByb2plY3QvbG9ncy9jbG91ZGF1ZGl0Lmdvb2dsZWFwaXMuY29tJTJGZGF0YV9hY2Nlc3MiLCJyZWNlaXZlVGltZXN0YW1wIjoiMjAyMi0wOS0wNlQxODowNzo0My40MTc3MTQwNDZaIiwicmVzb3VyY2UiOnsibGFiZWxzIjp7ImNsdXN0ZXJfbmFtZSI6ImRldi11cy1jZW50cmFsLTQyIiwibG9jYXRpb24iOiJ1cy1jZW50cmFsMSIsInByb2plY3RfaWQiOiJ0ZXN0LXByb2plY3QifSwidHlwZSI6Ims4c19jbHVzdGVyIn0sInRpbWVzdGFtcCI6IjIwMjItMDktMDZUMTg6MDc6NDIuMzYzMTEzWiJ9Cg==",
		"messageId": "5187581549398349",
		"message_id": "5187581549398349",
		"publishTime": "2022-07-25T22:19:15.56Z",
		"publish_time": "2022-07-25T22:19:15.56Z"
	},
	"subscription": "projects/test-project/subscriptions/test"
}`

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
