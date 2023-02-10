package kafka

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/regexp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestPush(t *testing.T) {
	l, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)

	opts := component.Options{Logger: l, Registerer: prometheus.NewRegistry()}

	ch1, ch2 := make(chan loki.Entry), make(chan loki.Entry)
	args := Arguments{
		Brokers:              []string{"localhost:9092"},
		Topics:               []string{"quickstart-events1"},
		GroupID:              "promtail",
		Assignor:             "range",
		Version:              "2.2.1",
		Authentication:       KafkaAuthentication{},
		UseIncomingTimestamp: false,
		Labels:               map[string]string{"component": "loki.source.kafka", "foo": "bar"},
		ForwardTo:            []loki.LogsReceiver{ch1, ch2},
		RelabelRules:         rulesExport,
	}

	// Create and run the component.
	c, err := New(opts, args)
	require.NoError(t, err)

	go c.Run(context.Background())
	time.Sleep(200 * time.Millisecond)

	// Check the received log entries
	/*
		wantLabelSet := model.LabelSet{"component": "loki.source.kafka", "foo": "bar", "message_key": "TODO", "topic": "TODO", "partition": "TODO", "member_id": "TODO", "group_id": "TODO"}

		for i := 0; i < 2; i++ {
			select {
			case logEntry := <-ch1:
				require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
				require.Equal(t, wantLabelSet, logEntry.Labels)
			case logEntry := <-ch2:
				require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
				require.Equal(t, wantLabelSet, logEntry.Labels)
			case <-time.After(5 * time.Second):
				require.FailNow(t, "failed waiting for log line")
			}
		}
	*/
}

var rulesExport = flow_relabel.Rules{
	{
		SourceLabels: []string{"__meta_kafka_message_key"},
		Regex:        newRegexp(),
		Action:       flow_relabel.Replace,
		Replacement:  "$1",
		TargetLabel:  "message_key",
	},
	{
		SourceLabels: []string{"__meta_kafka_topic"},
		Regex:        newRegexp(),
		Action:       flow_relabel.Replace,
		Replacement:  "$1",
		TargetLabel:  "topic",
	},
	{
		SourceLabels: []string{"__meta_kafka_partition"},
		Regex:        newRegexp(),
		Action:       flow_relabel.Replace,
		Replacement:  "$1",
		TargetLabel:  "partition",
	},
	{
		SourceLabels: []string{"__meta_kafka_member_id"},
		Regex:        newRegexp(),
		Action:       flow_relabel.Replace,
		Replacement:  "$1",
		TargetLabel:  "member_id",
	},
	{
		SourceLabels: []string{"__meta_kafka_group_id"},
		Regex:        newRegexp(),
		Action:       flow_relabel.Replace,
		Replacement:  "$1",
		TargetLabel:  "group_id",
	},
}

func newRegexp() flow_relabel.Regexp {
	re, err := regexp.Compile("^(?:(.*))$")
	if err != nil {
		panic(err)
	}
	return flow_relabel.Regexp{Regexp: re}
}
