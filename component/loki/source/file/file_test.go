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
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func Test(t *testing.T) {
	defer goleak.VerifyNone(t)

	// Create opts for component
	l, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)
	dataPath, err := os.MkdirTemp("", "loki.source.file")
	require.NoError(t, err)
	defer os.RemoveAll(dataPath) // clean up

	opts := component.Options{Logger: l, DataPath: dataPath}

	f, err := os.CreateTemp(dataPath, "example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())
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
