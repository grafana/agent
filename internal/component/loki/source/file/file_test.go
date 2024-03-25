//go:build !race

package file

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/grafana/agent/internal/component/common/loki/positions"
	"github.com/grafana/loki/pkg/loghttp/push"
	"gopkg.in/yaml.v2"

	logkit "github.com/go-kit/log"
	"github.com/grafana/agent/internal/component"
	"github.com/grafana/agent/internal/component/common/loki"
	"github.com/grafana/agent/internal/component/discovery"
	"github.com/grafana/agent/internal/flow/componenttest"
	"github.com/grafana/agent/internal/static/logs"
	"github.com/grafana/agent/internal/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/goleak"
	"golang.org/x/text/encoding/unicode"
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

	ch1, ch2 := loki.NewLogsReceiver(), loki.NewLogsReceiver()

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
		case logEntry := <-ch1.Chan():
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, "writing some text", logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case logEntry := <-ch2.Chan():
			require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
			require.Equal(t, "writing some text", logEntry.Line)
			require.Equal(t, wantLabelSet, logEntry.Labels)
		case <-time.After(5 * time.Second):
			require.FailNow(t, "failed waiting for log line")
		}
	}
}

func TestFileWatch(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"))
	ctx, cancel := context.WithCancel(componenttest.TestContext(t))

	// Create file to log to.
	f, err := os.CreateTemp(t.TempDir(), "example")
	require.NoError(t, err)
	defer f.Close()

	ctrl, err := componenttest.NewControllerFromID(util.TestLogger(t), "loki.source.file")
	require.NoError(t, err)

	ch1 := loki.NewLogsReceiver()

	args := Arguments{
		Targets: []discovery.Target{{
			"__path__": f.Name(),
			"foo":      "bar",
		}},
		ForwardTo: []loki.LogsReceiver{ch1},
		FileWatch: FileWatch{
			MinPollFrequency: time.Millisecond * 500,
			MaxPollFrequency: time.Millisecond * 500,
		},
	}

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	err = ctrl.WaitRunning(time.Minute)
	require.NoError(t, err)

	timeBeforeWriting := time.Now()

	// Sleep for 600ms to miss the first poll, the next poll should be MaxPollFrequency later.
	time.Sleep(time.Millisecond * 600)

	_, err = f.Write([]byte("writing some text\n"))
	require.NoError(t, err)

	select {
	case logEntry := <-ch1.Chan():
		require.Greater(t, time.Since(timeBeforeWriting), 1*time.Second)
		require.WithinDuration(t, time.Now(), timeBeforeWriting, 2*time.Second)
		require.Equal(t, "writing some text", logEntry.Line)
	case <-time.After(5 * time.Second):
		require.FailNow(t, "failed waiting for log line")
	}

	// Shut down the component.
	cancel()

	// Wait to make sure that all go routines stopped.
	time.Sleep(args.FileWatch.MaxPollFrequency)
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

	ch1 := loki.NewLogsReceiver()
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
		case logEntry := <-ch1.Chan():
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

func TestEncoding(t *testing.T) {
	// Create opts for component
	opts := component.Options{
		Logger:        util.TestFlowLogger(t),
		Registerer:    prometheus.NewRegistry(),
		OnStateChange: func(e component.Exports) {},
		DataPath:      t.TempDir(),
	}

	// Create a file to write to and set up the component's Arguments.
	f, err := os.CreateTemp(opts.DataPath, "example")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	ch1 := loki.NewLogsReceiver()
	args := Arguments{}
	args.Targets = []discovery.Target{{"__path__": f.Name(), "lbl1": "val1"}}
	args.Encoding = "UTF-16BE"
	args.ForwardTo = []loki.LogsReceiver{ch1}

	// Create and run the component.
	c, err := New(opts, args)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go c.Run(ctx)
	require.Eventually(t, func() bool { return c.DebugInfo() != nil }, 500*time.Millisecond, 20*time.Millisecond)

	// Write a UTF-16BE encoded byte slice to the file.
	utf16Encoder := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewEncoder()
	utf16Bytes, err := utf16Encoder.Bytes([]byte("hello world!\n"))
	require.Nil(t, err)

	_, err = f.Write(utf16Bytes)
	require.Nil(t, err)

	// Make sure the log was received successfully with the correct format.
	select {
	case logEntry := <-ch1.Chan():
		require.WithinDuration(t, time.Now(), logEntry.Timestamp, 1*time.Second)
		require.Equal(t, "hello world!�", logEntry.Line)

	case <-time.After(5 * time.Second):
		require.FailNow(t, "failed waiting for log line")
	}

	// Shut down the component
	cancel()

	// Verify that positions.yml is written. NOTE: if we didn't wait for it,
	// there would be a race condition between temporary directory being
	// cleaned up and this file being created.
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

// TestFullConversion tests using static mode to tail a log, then converting and then writing new entries and ensuring
// the previously tailed data does not come through.
func TestFullConversion(t *testing.T) {
	//
	// Create a temporary file to tail
	//
	positionsDir := t.TempDir()

	tmpFile, err := os.CreateTemp(os.TempDir(), "*.log")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpFile.Name())
	})

	//
	// Listen for push requests and pass them through to a channel
	//

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, lis.Close())
	})
	var written atomic.Bool
	go func() {
		_ = http.Serve(lis, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			_, err := push.ParseRequest(logkit.NewNopLogger(), "user_id", r, nil, nil, push.ParseLokiRequest)
			require.NoError(t, err)
			_, _ = rw.Write(nil)
			written.Store(true)
		}))
	}()

	//
	// Launch Loki so it starts tailing the file and writes to our server.
	//
	cfgText := util.Untab(fmt.Sprintf(`
positions_directory: %s
configs:
- name: default
  clients:
  - url: http://%s/loki/api/v1/push
		batchwait: 50ms
		batchsize: 1
  scrape_configs:
  - job_name: system
    static_configs:
    - targets: [localhost]
      labels:
        job: test
        __path__: %s
	`, positionsDir, lis.Addr().String(), tmpFile.Name()))

	var cfg logs.Config
	dec := yaml.NewDecoder(strings.NewReader(cfgText))
	dec.SetStrict(true)
	require.NoError(t, dec.Decode(&cfg))
	require.NoError(t, cfg.ApplyDefaults())
	logger := logkit.NewNopLogger()
	l, err := logs.New(prometheus.NewRegistry(), &cfg, logger, false)
	require.NoError(t, err)
	//
	// Write a log line and wait for it to come through.
	//
	fmt.Fprintf(tmpFile, "Hello, world!\n")
	// Ensure we have written and received the data.
	require.Eventually(t, func() bool {
		return written.Load()
	}, 10*time.Second, 100*time.Millisecond)

	// Stop the tailer.
	l.Stop()
	oldPositions, err := os.ReadFile(filepath.Join(positionsDir, "default.yml"))
	require.NoError(t, err)

	legacy := positions.LegacyFile{Positions: make(map[string]string)}
	err = yaml.UnmarshalStrict(oldPositions, legacy)
	require.NoError(t, err)

	opts := component.Options{
		Logger:        util.TestFlowLogger(t),
		Registerer:    prometheus.NewRegistry(),
		OnStateChange: func(e component.Exports) {},
		DataPath:      t.TempDir(),
	}

	ch1 := loki.NewLogsReceiver()
	args := Arguments{}
	args.LegacyPositionsFile = filepath.Join(positionsDir, "default.yml")
	args.ForwardTo = []loki.LogsReceiver{ch1}
	args.Targets = []discovery.Target{
		{
			"__path__": tmpFile.Name(),
		},
	}
	c, err := New(opts, args)
	require.NoError(t, err)
	require.NotNil(t, c)
	buf, err := os.ReadFile(filepath.Join(opts.DataPath, "positions.yml"))
	require.NoError(t, err)
	newPositions := positions.File{Positions: make(map[positions.Entry]string)}
	err = yaml.UnmarshalStrict(buf, newPositions)
	require.NoError(t, err)

	require.Len(t, newPositions.Positions, 1)
	require.Len(t, legacy.Positions, 1)

	for k, v := range newPositions.Positions {
		val, found := legacy.Positions[k.Path]
		require.True(t, found)
		require.True(t, val == v)
	}

	fmt.Fprintf(tmpFile, "new thing!\n")
	ctx := context.Background()
	ctx, cncl := context.WithTimeout(ctx, 10*time.Second)
	defer cncl()
	go func() {
		runErr := c.Run(ctx)
		require.NoError(t, runErr)
	}()
	require.Eventually(t, func() bool {
		entry := <-ch1.Chan()
		// We don't want to reread the hello world so if something went wrong then the conversion didnt work.
		require.False(t, strings.Contains(entry.Line, "Hello, world!"))
		return strings.Contains(entry.Line, "new thing!")
	}, 5*time.Second, 100*time.Millisecond)
}
