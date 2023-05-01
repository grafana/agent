package file

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func Test(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"))

	ctx, cancel := context.WithCancel(componenttest.TestContext(t))
	defer cancel()

	// Create file to log to.
	f, err := os.CreateTemp(t.TempDir(), "example")
	require.NoError(t, err)
	defer f.Close()

	ctrl, err := componenttest.NewControllerFromID(util.TestLogger(t), "loki.source.file")
	require.NoError(t, err)

	ch1, ch2 := make(chan loki.Entry), make(chan loki.Entry)

	go func() {
		err := ctrl.Run(ctx, Arguments{
			Targets: []discovery.Target{{
				"__path__": f.Name(),
				"foo":      "bar",
			}},
			ForwardTo: []loki.LogsReceiver{ch1, ch2},
		})
		require.NoError(t, err)
	}()

	ctrl.WaitRunning(time.Minute)

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

// Test that updating the component does not leak goroutines.
func TestUpdate_NoLeak(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"))

	ctx, cancel := context.WithCancel(componenttest.TestContext(t))
	defer cancel()

	// Create file to tail.
	f, err := os.CreateTemp(t.TempDir(), "example")
	require.NoError(t, err)
	defer f.Close()

	ctrl, err := componenttest.NewControllerFromID(util.TestLogger(t), "loki.source.file")
	require.NoError(t, err)

	args := Arguments{
		Targets: []discovery.Target{{
			"__path__": f.Name(),
			"foo":      "bar",
		}},
		ForwardTo: []loki.LogsReceiver{},
	}

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	ctrl.WaitRunning(time.Minute)

	// Update a bunch of times to ensure that no goroutines get leaked between
	// updates.
	for i := 0; i < 10; i++ {
		err := ctrl.Update(args)
		require.NoError(t, err)
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
	cancel()
	// Verify that positions.yml is written. NOTE: if we didn't wait for it, there would be a race condition between
	// temporary directory being cleaned up and this file being created.
	require.Eventually(
		t,
		func() bool {
			if _, err := os.Stat(filepath.Join(opts.DataPath, "positions.yml")); errors.Is(err, os.ErrNotExist) {
				return false
			}
			return true
		},
		5*time.Second,
		10*time.Millisecond,
		"expected positions.yml file to be written eventually",
	)
}
