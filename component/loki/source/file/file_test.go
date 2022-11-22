package file

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/loki/clients/pkg/promtail/api"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	// Create opts for component
	l, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)
	dataPath, err := os.MkdirTemp("", "loki.source.file")
	defer os.RemoveAll(dataPath) // clean up

	opts := component.Options{Logger: l, DataPath: dataPath}

	f, err := os.CreateTemp(dataPath, "example")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	defer os.Remove(f.Name()) // clean up

	ch := make(chan api.Entry)
	args := DefaultArguments
	args.Targets = []discovery.Target{{"__path__": f.Name(), "foo": "bar"}}
	args.ForwardTo = []chan api.Entry{ch}

	c, err := New(opts, args)
	require.NoError(t, err)

	go c.Run(context.Background())
	time.Sleep(100 * time.Millisecond)

	_, err = f.Write([]byte("writing some text\n"))
	require.NoError(t, err)
	wantLabelSet := model.LabelSet{
		"filename": model.LabelValue(f.Name()),
		"foo":      "bar",
	}

	select {
	case logEntry := <-ch:
		require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
		require.Equal(t, "writing some text", logEntry.Line)
		require.Equal(t, wantLabelSet, logEntry.Labels)
	case <-time.After(5 * time.Second):
		require.FailNow(t, "failed waiting for log line")
	}
}
