package process

// NOTE: This code is copied from Promtail (07cbef92268aecc0f20d1791a6df390c2df5c072) with changes kept to the minimum.

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/loki/process/stages"
	lsf "github.com/grafana/agent/component/loki/source/file"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/grafana/river"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/goleak"
)

func TestJSONLabelsStage(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"))

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

	ch1, ch2 := loki.NewLogsReceiver(), loki.NewLogsReceiver()

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

	c.receiver.Chan() <- logEntry

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
		case logEntry := <-ch1.Chan():
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, logline, logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case logEntry := <-ch2.Chan():
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, logline, logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case <-time.After(5 * time.Second):
			require.FailNow(t, "failed waiting for log line")
		}
	}
}

func TestStaticLabelsLabelAllowLabelDrop(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"))

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

	ch1, ch2 := loki.NewLogsReceiver(), loki.NewLogsReceiver()

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

	c.receiver.Chan() <- logEntry

	wantLabelSet := model.LabelSet{
		"filename": "/var/log/pods/agent/agent/1.log",
		"baz":      "bazval",
	}

	// The log entry should be received in both channels, with the processing
	// stages correctly applied.
	for i := 0; i < 2; i++ {
		select {
		case logEntry := <-ch1.Chan():
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, logline, logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case logEntry := <-ch2.Chan():
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, logline, logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case <-time.After(5 * time.Second):
			require.FailNow(t, "failed waiting for log line")
		}
	}
}

func TestRegexTimestampOutput(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"))

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

	ch1, ch2 := loki.NewLogsReceiver(), loki.NewLogsReceiver()

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

	c.receiver.Chan() <- logEntry

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
		case logEntry := <-ch1.Chan():
			require.Equal(t, wantLogline, logEntry.Line)
			require.Equal(t, wantTimestamp, logEntry.Timestamp)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case logEntry := <-ch2.Chan():
			require.Equal(t, wantLogline, logEntry.Line)
			require.Equal(t, wantTimestamp, logEntry.Timestamp)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case <-time.After(5 * time.Second):
			require.FailNow(t, "failed waiting for log line")
		}
	}
}

func TestEntrySentToTwoProcessComponents(t *testing.T) {
	// Set up two different loki.process components.
	stg1 := `
forward_to = []
stage.static_labels {
    values = { "lbl" = "foo" }
}
`
	stg2 := `
forward_to = []
stage.static_labels {
    values = { "lbl" = "bar" }
}
`

	ch1, ch2 := loki.NewLogsReceiver(), loki.NewLogsReceiver()
	var args1, args2 Arguments
	require.NoError(t, river.Unmarshal([]byte(stg1), &args1))
	require.NoError(t, river.Unmarshal([]byte(stg2), &args2))
	args1.ForwardTo = []loki.LogsReceiver{ch1}
	args2.ForwardTo = []loki.LogsReceiver{ch2}

	// Start the loki.process components.
	tc1, err := componenttest.NewControllerFromID(util.TestLogger(t), "loki.process")
	require.NoError(t, err)
	tc2, err := componenttest.NewControllerFromID(util.TestLogger(t), "loki.process")
	require.NoError(t, err)
	go func() { require.NoError(t, tc1.Run(componenttest.TestContext(t), args1)) }()
	go func() { require.NoError(t, tc2.Run(componenttest.TestContext(t), args2)) }()
	require.NoError(t, tc1.WaitExports(time.Second))
	require.NoError(t, tc2.WaitExports(time.Second))

	// Create a file to log to.
	f, err := os.CreateTemp(t.TempDir(), "example")
	require.NoError(t, err)
	defer f.Close()

	// Create and start a component that will read from that file and fan out to both components.
	ctrl, err := componenttest.NewControllerFromID(util.TestLogger(t), "loki.source.file")
	require.NoError(t, err)

	go func() {
		err := ctrl.Run(context.Background(), lsf.Arguments{
			Targets: []discovery.Target{{"__path__": f.Name(), "somelbl": "somevalue"}},
			ForwardTo: []loki.LogsReceiver{
				tc1.Exports().(Exports).Receiver,
				tc2.Exports().(Exports).Receiver,
			},
		})
		require.NoError(t, err)
	}()
	ctrl.WaitRunning(time.Minute)

	// Write a line to the file.
	_, err = f.Write([]byte("writing some text\n"))
	require.NoError(t, err)

	wantLabelSet := model.LabelSet{
		"filename": model.LabelValue(f.Name()),
		"somelbl":  "somevalue",
	}

	// The lines were received after processing by each component, with no
	// race condition between them.
	for i := 0; i < 2; i++ {
		select {
		case logEntry := <-ch1.Chan():
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, "writing some text", logEntry.Line)
			require.Equal(t, wantLabelSet.Clone().Merge(model.LabelSet{"lbl": "foo"}), logEntry.Labels)
		case logEntry := <-ch2.Chan():
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, "writing some text", logEntry.Line)
			require.Equal(t, wantLabelSet.Clone().Merge(model.LabelSet{"lbl": "bar"}), logEntry.Labels)
		case <-time.After(5 * time.Second):
			require.FailNow(t, "failed waiting for log line")
		}
	}
}

func TestDeadlockWithFrequentUpdates(t *testing.T) {
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

	ch1, ch2 := loki.NewLogsReceiver(), loki.NewLogsReceiver()

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

	var lastSend atomic.Value
	// Drain received logs
	go func() {
		for {
			select {
			case <-ch1.Chan():
				lastSend.Store(time.Now())
			case <-ch2.Chan():
				lastSend.Store(time.Now())
			}
		}
	}()

	// Continuously send entries to both channels
	go func() {
		for {
			ts := time.Now()
			logline := `{"log":"log message\n","stream":"stderr","time":"2019-04-30T02:12:41.8443515Z","extra":"{\"user\":\"smith\"}"}`
			logEntry := loki.Entry{
				Labels: model.LabelSet{"filename": "/var/log/pods/agent/agent/1.log", "foo": "bar"},
				Entry: logproto.Entry{
					Timestamp: ts,
					Line:      logline,
				},
			}
			c.receiver.Chan() <- logEntry
		}
	}()

	// Call Updates
	args1 := Arguments{
		ForwardTo: []loki.LogsReceiver{ch1},
		Stages:    stagesCfg.Stages,
	}
	args2 := Arguments{
		ForwardTo: []loki.LogsReceiver{ch2},
		Stages:    stagesCfg.Stages,
	}
	go func() {
		for {
			c.Update(args1)
			c.Update(args2)
		}
	}()

	// Run everything for a while
	time.Sleep(1 * time.Second)
	require.WithinDuration(t, time.Now(), lastSend.Load().(time.Time), 300*time.Millisecond)
}
