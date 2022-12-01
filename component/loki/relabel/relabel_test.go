package relabel

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	// Rename the kubernetes_(.*) labels without the suffix and remove them,
	// then set the `environment` label to the value of the namespace.
	rc := `rule {
         regex        = "kubernetes_(.*)"
         replacement  = "$1"
         action       = "labelmap"
       }
       rule {
         regex  = "kubernetes_(.*)"
         action = "labeldrop"
       }
       rule {
         source_labels = ["namespace"]
         target_label  = "environment"
         action        = "replace"
       }`

	// Unmarshal the River relabel rules into a custom struct, as we don't have
	// an easy way to refer to a loki.LogsReceiver value for the forward_to
	// argument.
	type cfg struct {
		Rcs []*flow_relabel.Config `river:"rule,block,optional"`
	}
	var relabelConfigs cfg
	err := river.Unmarshal([]byte(rc), &relabelConfigs)
	require.NoError(t, err)

	ch1, ch2 := make(loki.LogsReceiver), make(loki.LogsReceiver)

	// Create and run the component, so that it relabels and forwards logs.
	l, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)
	opts := component.Options{Logger: l, Registerer: prometheus.NewRegistry(), OnStateChange: func(e component.Exports) {}}
	args := Arguments{
		ForwardTo:      []loki.LogsReceiver{ch1, ch2},
		RelabelConfigs: relabelConfigs.Rcs,
		MaxCacheSize:   10,
	}

	c, err := New(opts, args)
	require.NoError(t, err)
	go c.Run(context.Background())

	// Send a log entry to the component's receiver.
	logEntry := loki.Entry{
		Labels: model.LabelSet{"filename": "/var/log/pods/agent/agent/1.log", "kubernetes_namespace": "dev", "kubernetes_pod_name": "agent", "foo": "bar"},
		Entry: logproto.Entry{
			Timestamp: time.Now(),
			Line:      "very important log",
		},
	}

	c.receiver <- logEntry

	wantLabelSet := model.LabelSet{
		"filename":    "/var/log/pods/agent/agent/1.log",
		"namespace":   "dev",
		"pod_name":    "agent",
		"environment": "dev",
		"foo":         "bar",
	}

	// The log entry should be received in both channels, with the relabeling
	// rules correctly applied.
	for i := 0; i < 2; i++ {
		select {
		case logEntry := <-ch1:
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, "very important log", logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case logEntry := <-ch2:
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, "very important log", logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case <-time.After(5 * time.Second):
			require.FailNow(t, "failed waiting for log line")
		}
	}

}
