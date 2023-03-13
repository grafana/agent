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
	component1 := logger.NewComponentLogger(logger.ControllerSink(controller), "my-component-1")
	component2 := logger.NewComponentLogger(logger.ComponentSink(component1), "my-component-2")

	// Log some log lines.
	level.Info(controller).Log("msg", "hello from controller!")
	level.Info(component1).Log("msg", "hello from component 1!")
	level.Info(component2).Log("msg", "hello from component 2!")

	// Output:
	// level=info msg="hello from controller!"
	// component=my-component-1 level=info msg="hello from component 1!"
	// component=my-component-2 level=info msg="hello from component 2!"
}
