package process

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/phlare/process/internal/stages"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestJSONLabelsStage(t *testing.T) {
	defer goleak.VerifyNone(t)

	// The following stages will attempt to parse input lines as JSON.
	// The first stage _extract_ any fields found with the correct names:
	// Since 'source' is empty, it implies that we want to parse the log line
	// itself.
	//    log    --> output
	//    stream --> stream
	//    time   --> timestamp
	//    extra  --> extra
	//
	// The second stage will parse the 'extra' field as JSON, and extract the
	// 'user' field from the 'extra' field. If the expression value field is
	// empty, it is inferred we want to use the same name as the key.
	//    user   --> extra.user
	//
	// The third stage will set some labels from the extracted values above.
	// Again, if the value is empty, it is inferred that we want to use the
	// populate the label with extracted value of the same name.
	stg := `stage.json { 
			    expressions    = {"output" = "log", stream = "stream", timestamp = "time", "extra" = "" }
				drop_malformed = true
		    }
			stage.json {
			    expressions = { "user" = "" }
				source      = "extra"
			}
			stage.labels {
			    values = { 
				  stream = "",
				  user   = "",
				  ts     = "timestamp",
			    }
			}`

	// Unmarshal the River relabel rules into a custom struct, as we don't have
	// an easy way to refer to a loki.LogsReceiver value for the forward_to
	// argument.
	type cfg struct {
		Stages []stages.StageConfig `river:"stage,enum"`
	}
	var stagesCfg cfg
	err := river.Unmarshal([]byte(stg), &stagesCfg)
	require.NoError(t, err)

	ch1, ch2 := make(loki.LogsReceiver), make(loki.LogsReceiver)

	// Create and run the component, so that it can process and forwards logs.
	opts := component.Options{
		Logger:        util.TestFlowLogger(t),
		Registerer:    prometheus.NewRegistry(),
		OnStateChange: func(e component.Exports) {},
	}
	args := Arguments{
		ForwardTo: []loki.LogsReceiver{ch1, ch2},
		Stages:    stagesCfg.Stages,
	}

	c, err := New(opts, args)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go c.Run(ctx)

	// Send a log entry to the component's receiver.
	ts := time.Now()
	logline := `{"log":"log message\n","stream":"stderr","time":"2019-04-30T02:12:41.8443515Z","extra":"{\"user\":\"smith\"}"}`
	logEntry := loki.Entry{
		Labels: model.LabelSet{"filename": "/var/log/pods/agent/agent/1.log", "foo": "bar"},
		Entry: logproto.Entry{
			Timestamp: ts,
			Line:      logline,
		},
	}

	c.receiver <- logEntry

	wantLabelSet := model.LabelSet{
		"filename": "/var/log/pods/agent/agent/1.log",
		"foo":      "bar",
		"stream":   "stderr",
		"ts":       "2019-04-30T02:12:41.8443515Z",
		"user":     "smith",
	}

	// The log entry should be received in both channels, with the processing
	// stages correctly applied.
	for i := 0; i < 2; i++ {
		select {
		case logEntry := <-ch1:
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, logline, logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case logEntry := <-ch2:
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, logline, logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case <-time.After(5 * time.Second):
			require.FailNow(t, "failed waiting for log line")
		}
	}
}

func TestStaticLabelsLabelAllowLabelDrop(t *testing.T) {
	defer goleak.VerifyNone(t)

	// The following stages manipulate the label set of a log entry.
	// The first stage will define a static set of labels (foo, bar, baz, qux)
	// to add to the entry along the `filename` and `dev` labels.
	// The second stage will drop the foo and bar labels.
	// The third stage will keep only a subset of the remaining labels.
	stg := `
stage.static_labels {
    values = { "foo" = "fooval", "bar" = "barval", "baz" = "bazval", "qux" = "quxval" }
}
stage.label_drop {
    values = [ "foo", "bar" ]
}
stage.label_keep {
    values = [ "foo", "baz", "filename" ]
}`

	// Unmarshal the River relabel rules into a custom struct, as we don't have
	// an easy way to refer to a loki.LogsReceiver value for the forward_to
	// argument.
	type cfg struct {
		Stages []stages.StageConfig `river:"stage,enum"`
	}
	var stagesCfg cfg
	err := river.Unmarshal([]byte(stg), &stagesCfg)
	require.NoError(t, err)

	ch1, ch2 := make(loki.LogsReceiver), make(loki.LogsReceiver)

	// Create and run the component, so that it can process and forwards logs.
	opts := component.Options{
		Logger:        util.TestFlowLogger(t),
		Registerer:    prometheus.NewRegistry(),
		OnStateChange: func(e component.Exports) {},
	}
	args := Arguments{
		ForwardTo: []loki.LogsReceiver{ch1, ch2},
		Stages:    stagesCfg.Stages,
	}

	c, err := New(opts, args)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go c.Run(ctx)

	// Send a log entry to the component's receiver.
	ts := time.Now()
	logline := `{"log":"log message\n","stream":"stderr","time":"2022-01-09T08:37:45.8233626Z"}`
	logEntry := loki.Entry{
		Labels: model.LabelSet{"filename": "/var/log/pods/agent/agent/1.log", "env": "dev"},
		Entry: logproto.Entry{
			Timestamp: ts,
			Line:      logline,
		},
	}

	c.receiver <- logEntry

	wantLabelSet := model.LabelSet{
		"filename": "/var/log/pods/agent/agent/1.log",
		"baz":      "bazval",
	}

	// The log entry should be received in both channels, with the processing
	// stages correctly applied.
	for i := 0; i < 2; i++ {
		select {
		case logEntry := <-ch1:
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, logline, logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case logEntry := <-ch2:
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, logline, logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case <-time.After(5 * time.Second):
			require.FailNow(t, "failed waiting for log line")
		}
	}
}

func TestRegexTimestampOutput(t *testing.T) {
	defer goleak.VerifyNone(t)

	// The first stage will attempt to parse the input line using a regular
	// expression with named capture groups. The three capture groups (time,
	// stream and content) will be extracted in the shared map of values.
	// Since 'source' is empty, it implies that we want to parse the log line
	// itself.
	//
	// The second stage will parse the extracted `time` value as Unix epoch
	// time and set it to the log entry timestamp.
	//
	// The third stage will set the `content` value as the message value.
	//
	// The fourth and final stage will set the `stream` value as the label.
	stg := `
stage.regex {
		expression = "^(?s)(?P<time>\\S+?) (?P<stream>stdout|stderr) (?P<content>.*)$"
}
stage.timestamp {
		source = "time"
		format = "RFC3339"
}
stage.output {
		source = "content"
}
stage.labels {
		values = { src = "stream" }
}`

	// Unmarshal the River relabel rules into a custom struct, as we don't have
	// an easy way to refer to a loki.LogsReceiver value for the forward_to
	// argument.
	type cfg struct {
		Stages []stages.StageConfig `river:"stage,enum"`
	}
	var stagesCfg cfg
	err := river.Unmarshal([]byte(stg), &stagesCfg)
	require.NoError(t, err)

	ch1, ch2 := make(loki.LogsReceiver), make(loki.LogsReceiver)

	// Create and run the component, so that it can process and forwards logs.
	opts := component.Options{
		Logger:        util.TestFlowLogger(t),
		Registerer:    prometheus.NewRegistry(),
		OnStateChange: func(e component.Exports) {},
	}
	args := Arguments{
		ForwardTo: []loki.LogsReceiver{ch1, ch2},
		Stages:    stagesCfg.Stages,
	}

	c, err := New(opts, args)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go c.Run(ctx)

	// Send a log entry to the component's receiver.
	ts := time.Now()
	logline := `2022-01-17T08:17:42-07:00 stderr somewhere, somehow, an error occurred`
	logEntry := loki.Entry{
		Labels: model.LabelSet{"filename": "/var/log/pods/agent/agent/1.log", "foo": "bar"},
		Entry: logproto.Entry{
			Timestamp: ts,
			Line:      logline,
		},
	}

	c.receiver <- logEntry

	wantLabelSet := model.LabelSet{
		"filename": "/var/log/pods/agent/agent/1.log",
		"foo":      "bar",
		"src":      "stderr",
	}
	wantTimestamp, err := time.Parse(time.RFC3339, "2022-01-17T08:17:42-07:00")
	wantLogline := `somewhere, somehow, an error occurred`
	require.NoError(t, err)

	// The log entry should be received in both channels, with the processing
	// stages correctly applied.
	for i := 0; i < 2; i++ {
		select {
		case logEntry := <-ch1:
			require.Equal(t, wantLogline, logEntry.Line)
			require.Equal(t, wantTimestamp, logEntry.Timestamp)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case logEntry := <-ch2:
			require.Equal(t, wantLogline, logEntry.Line)
			require.Equal(t, wantTimestamp, logEntry.Timestamp)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case <-time.After(5 * time.Second):
			require.FailNow(t, "failed waiting for log line")
		}
	}
}
