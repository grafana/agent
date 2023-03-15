package file

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func Test(t *testing.T) {
	defer goleak.VerifyNone(t)

	// Create opts for component
	opts := component.Options{
		Logger:        util.TestFlowLogger(t),
		Registerer:    prometheus.NewRegistry(),
		OnStateChange: func(e component.Exports) {},
		DataPath:      t.TempDir(),
	}

	f, err := os.CreateTemp(opts.DataPath, "example")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	ch1, ch2 := make(chan loki.Entry), make(chan loki.Entry)
	args := Arguments{}
	args.Targets = []discovery.Target{{"__path__": f.Name(), "foo": "bar"}}
	args.ForwardTo = []loki.LogsReceiver{ch1, ch2}

	c, err := New(opts, args)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go c.Run(ctx)
	time.Sleep(100 * time.Millisecond)

	_, err = f.Write([]byte("writing some text\n"))
	require.NoError(t, err)
	wantLabelSet := model.LabelSet{
		"filename": model.LabelValue(f.Name()),
		"foo":      "bar",
	}

	for i := 0; i < 2; i++ {
		select {
		case logEntry := <-ch1:
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, "writing some text", logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case logEntry := <-ch2:
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, "writing some text", logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case <-time.After(5 * time.Second):
			require.FailNow(t, "failed waiting for log line")
		}
	}
}

func TestTwoTargets(t *testing.T) {
	// Create opts for component
	opts := component.Options{
		Logger:        util.TestFlowLogger(t),
		Registerer:    prometheus.NewRegistry(),
		OnStateChange: func(e component.Exports) {},
		DataPath:      t.TempDir(),
	}

	f, err := os.CreateTemp(opts.DataPath, "example")
	if err != nil {
		log.Fatal(err)
	}
	f2, err := os.CreateTemp(opts.DataPath, "example2")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	defer f2.Close()

	ch1 := make(chan loki.Entry)
	args := Arguments{}
	args.Targets = []discovery.Target{
		{"__path__": f.Name(), "foo": "bar"},
		{"__path__": f2.Name(), "foo": "bar2"},
	}
	args.ForwardTo = []loki.LogsReceiver{ch1}

	c, err := New(opts, args)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go c.Run(ctx)
	time.Sleep(100 * time.Millisecond)

	_, err = f.Write([]byte("text\n"))
	require.NoError(t, err)

	_, err = f2.Write([]byte("text2\n"))
	require.NoError(t, err)

	foundF1, foundF2 := false, false
	for i := 0; i < 2; i++ {
		select {
		case logEntry := <-ch1:
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			if logEntry.Line == "text" {
				foundF1 = true
			} else if logEntry.Line == "text2" {
				foundF2 = true
			}

		case <-time.After(5 * time.Second):
			require.FailNow(t, "failed waiting for log line")
		}
	}
	require.True(t, foundF1)
	require.True(t, foundF2)
}
