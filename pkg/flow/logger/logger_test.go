package logger_test

import (
	"os"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/flow/logger"
)

func Example() {
	// Create a sink to send logs to. WriterSink supports options to customize
	// the logs sent to the sink.
	sink, err := logger.WriterSink(os.Stdout, logger.SinkOptions{
		Level:             logger.LevelDebug,
		Format:            logger.FormatLogfmt,
		IncludeTimestamps: false,
	})
	if err != nil {
		panic(err)
	}

	// Create a controller logger.
	controller := logger.NewControllerLogger(sink)

	// Create two component loggers. The first component sends logs to the
	// controller, and the other sends logs to the first component.
	component1 := logger.NewComponentLogger(logger.ControllerSink(controller), "outer")
	component2 := logger.NewComponentLogger(logger.ComponentSink(component1), "inner")

	innerController := logger.NewControllerLogger(logger.ComponentSink(component2))

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
