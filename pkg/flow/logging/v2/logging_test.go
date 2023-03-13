package logging_test

import (
	"os"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/flow/logging/v2"
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
