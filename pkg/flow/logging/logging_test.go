package logging_test

import (
	"os"
	"testing"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/stretchr/testify/require"
)

func Example() {
	// Create a sink to send logs to. WriterSink supports options to customize
	// the logs sent to the sink.
	sink, err := logging.WriterSink(os.Stdout, logging.SinkOptions{
		Level:             logging.LevelDebug,
		Format:            logging.FormatLogfmt,
		IncludeTimestamps: false,
	})
	if err != nil {
		panic(err)
	}

	// Create a controller logger. A controller logger is a logger with no
	// component ID.
	controller := logging.New(sink)

	// Create two component loggers. The first component sends logs to the
	// controller, and the other sends logs to the first component.
	component1 := logging.New(logging.LoggerSink(controller), logging.WithComponentID("outer"))
	component2 := logging.New(logging.LoggerSink(component1), logging.WithComponentID("inner"))

	innerController := logging.New(logging.LoggerSink(component2))

	// Log some log lines.
	level.Info(controller).Log("msg", "hello from the controller!")
	level.Info(component1).Log("msg", "hello from the outer component!")
	level.Info(component2).Log("msg", "hello from the inner component!")
	level.Info(innerController).Log("msg", "hello from the inner controller!")

	// Output:
	// level=info msg="hello from the controller!"
	// component=outer level=info msg="hello from the outer component!"
	// component=outer/inner level=info msg="hello from the inner component!"
	// component=outer/inner level=info msg="hello from the inner controller!"
}

func TestFanoutWriter(t *testing.T) {
	var (
		ch1 = make(loki.LogsReceiver)    // unbuffered/blocked channel
		ch2 loki.LogsReceiver            // nil
		ch3 = make(loki.LogsReceiver, 2) // buffered channel
	)
	sink, err := logging.WriterSink(os.Stdout, logging.SinkOptions{
		Level:             logging.LevelDebug,
		Format:            logging.FormatLogfmt,
		Fanout:            []loki.LogsReceiver{ch1, ch2, ch3},
		IncludeTimestamps: false,
	})
	require.NoError(t, err)

	controller := logging.New(sink)

	level.Info(controller).Log("msg", "first")
	level.Info(controller).Log("msg", "second")

	e3 := <-ch3
	require.Equal(t, e3.Line, "level=info msg=first\n")
	e3 = <-ch3
	require.Equal(t, e3.Line, "level=info msg=second\n")

	// Read initial entry from the blocked channel
	go func() {
		e1 := <-ch1
		require.Equal(t, e1.Line, "level=info msg=first\n")
	}()
}
