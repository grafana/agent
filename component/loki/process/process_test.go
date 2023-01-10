package process

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/loki/process/internal/stages"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func TestJSONLabelsStage(t *testing.T) {
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
	stg := `stage {
              json {
			    expressions    = {"output" = "log", stream = "stream", timestamp = "time", "extra" = "" }
				drop_malformed = true
			  }
		    }
			stage {
			  json {
			    expressions = { "user" = "" }
				source      = "extra"
			  }
			}
			stage { 
			  labels {
			    values = { 
				  stream = "",
				  user   = "",
				  ts     = "timestamp",
			    }
		      }
			}`

	// Unmarshal the River relabel rules into a custom struct, as we don't have
	// an easy way to refer to a loki.LogsReceiver value for the forward_to
	// argument.
	type cfg struct {
		Stages []stages.StageConfig `river:"stage,block"`
	}
	var stagesCfg cfg
	err := river.Unmarshal([]byte(stg), &stagesCfg)
	require.NoError(t, err)

	ch1, ch2 := make(loki.LogsReceiver), make(loki.LogsReceiver)

	// Create and run the component, so that it can process and forwards logs.
	l, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)
	opts := component.Options{Logger: l, Registerer: prometheus.NewRegistry(), OnStateChange: func(e component.Exports) {}}
	args := Arguments{
		ForwardTo: []loki.LogsReceiver{ch1, ch2},
		Stages:    stagesCfg.Stages,
	}

	c, err := New(opts, args)
	require.NoError(t, err)
	go c.Run(context.Background())

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
	// The following stages manipulate the label set of a log entry.
	// The first stage will define a static set of labels (foo, bar, baz, qux)
	// to add to the entry along the `filename` and `dev` labels.
	// The second stage will drop the foo and bar labels.
	// The third stage will keep only a subset of the remaining labels.
	stg := `
stage {
  static_labels {
    values = { "foo" = "fooval", "bar" = "barval", "baz" = "bazval", "qux" = "quxval" }
  }
}
stage {
  labeldrop {
    values = [ "foo", "bar" ]
  }
}
stage {
  labelallow {
    values = [ "foo", "baz", "filename" ]
  }
}`

	// Unmarshal the River relabel rules into a custom struct, as we don't have
	// an easy way to refer to a loki.LogsReceiver value for the forward_to
	// argument.
	type cfg struct {
		Stages []stages.StageConfig `river:"stage,block"`
	}
	var stagesCfg cfg
	err := river.Unmarshal([]byte(stg), &stagesCfg)
	require.NoError(t, err)

	ch1, ch2 := make(loki.LogsReceiver), make(loki.LogsReceiver)

	// Create and run the component, so that it can process and forwards logs.
	l, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)
	opts := component.Options{Logger: l, Registerer: prometheus.NewRegistry(), OnStateChange: func(e component.Exports) {}}
	args := Arguments{
		ForwardTo: []loki.LogsReceiver{ch1, ch2},
		Stages:    stagesCfg.Stages,
	}

	c, err := New(opts, args)
	require.NoError(t, err)
	go c.Run(context.Background())

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
