package kafka

/*
import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/loki/source/kafka/internal/kafkatarget"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/regexp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func TestPush(t *testing.T) {
	l, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)

	opts := component.Options{Logger: l, Registerer: prometheus.NewRegistry()}

	ch1, ch2 := make(chan loki.Entry), make(chan loki.Entry)
	args := Arguments{
		KafkaListener: ListenerConfig{
			ListenAddress: address,
			ListenPort:    port,
		},
		UseIncomingTimestamp: false,
		Labels:               map[string]string{"foo": "bar"},
		ForwardTo:            []loki.LogsReceiver{ch1, ch2},
		RelabelRules:         rulesExport,
	}

	// Create and run the component.
	c, err := New(opts, args)
	require.NoError(t, err)

	go c.Run(context.Background())
	time.Sleep(200 * time.Millisecond)

	// Create a Kafka Drain Request and send it to the launched server.
	req, err := http.NewRequest(http.MethodPost, getEndpoint(c.target), strings.NewReader(testPayload))
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.StatusCode)

	// Check the received log entries
	wantLabelSet := model.LabelSet{"foo": "bar", "host": "host", "app": "kafka", "proc": "router", "log_id": "-"}
	wantLogLine := "at=info method=GET path=\"/\" host=cryptic-cliffs-27764.kafkaapp.com request_id=59da6323-2bc4-4143-8677-cc66ccfb115f fwd=\"181.167.87.140\" dyno=web.1 connect=0ms service=3ms status=200 bytes=6979 protocol=https\n"

	for i := 0; i < 2; i++ {
		select {
		case logEntry := <-ch1:
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, wantLogLine, logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case logEntry := <-ch2:
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, wantLogLine, logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case <-time.After(5 * time.Second):
			require.FailNow(t, "failed waiting for log line")
		}
	}
}

const address = "localhost"
const port = 42421
const testPayload = `270 <158>1 2022-06-13T14:52:23.622778+00:00 host kafka router - at=info method=GET path="/" host=cryptic-cliffs-27764.kafkaapp.com request_id=59da6323-2bc4-4143-8677-cc66ccfb115f fwd="181.167.87.140" dyno=web.1 connect=0ms service=3ms status=200 bytes=6979 protocol=https
`

var rulesExport = flow_relabel.Rules{
	{
		SourceLabels: []string{"__kafka_drain_host"},
		Regex:        newRegexp(),
		Action:       flow_relabel.Replace,
		Replacement:  "$1",
		TargetLabel:  "host",
	},
	{
		SourceLabels: []string{"__kafka_drain_app"},
		Regex:        newRegexp(),
		Action:       flow_relabel.Replace,
		Replacement:  "$1",
		TargetLabel:  "app",
	},
	{
		SourceLabels: []string{"__kafka_drain_proc"},
		Regex:        newRegexp(),
		Action:       flow_relabel.Replace,
		Replacement:  "$1",
		TargetLabel:  "proc",
	},
	{
		SourceLabels: []string{"__kafka_drain_log_id"},
		Regex:        newRegexp(),
		Action:       flow_relabel.Replace,
		Replacement:  "$1",
		TargetLabel:  "log_id",
	},
}

func newRegexp() flow_relabel.Regexp {
	re, err := regexp.Compile("^(?:(.*))$")
	if err != nil {
		panic(err)
	}
	return flow_relabel.Regexp{Regexp: re}
}

func getEndpoint(target *kafkatarget.KafkaTarget) string {
	return fmt.Sprintf("http://%s:%d%s", address, port, target.DrainEndpoint())
}
*/
