// Package flowtest provides a script-based testing framework for Flow.
//
// [TestScript] accepts a path to a txtar file, where the comment of the txtar
// file denotes the script to run. Files in the txtar archive are unpacked to a
// temporary directory as part of the test.
//
// See [rsc.io/script] for more information about the scripting language and
// available default commands.
//
// # Flow commands
//
// In addition to the default commands provided by [rsc.io/script], the
// following Flow-specific commands are available:
//
//   - `controller.load [file]`: Load a file into the Flow controller.
//   - `controller.start`: Start the Flow controller.
//   - `assert.health [component] [health]`: Assert the health of a specific component.
//
// Note that `controller.start` should almost always be run in the background
// by appending a & to the end of the command, otherwise the controller will be
// terminated immediately after the command exits.
//
// Custom commands can be provided by passing [WithExtraCommands] to
// [TestScript].
//
// # Example script
//
// The following is an example script which uses local.file:
//
//	# This file performs a basic test of loading a component and asserting its
//	# health.
//
//	controller.load main.river
//	controller.start &
//
//	assert.health local.file.example healthy
//
//	-- main.river --
//	local.file "example" {
//	  filename = "./hello.txt"
//	}
//
//	-- hello.txt --
//	Hello, world!
//
// See the testdata directory for more basic examples.
package flowtest

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/tools/txtar"
	"rsc.io/script"
)

// TestScript loads the file at filename and executes it as a script. A
// temporary working directory is created for running the script, and the
// working directory of the process is changed to the temporary directory for
// the duration of the call.
func TestScript(filename string, opts ...TestOption) error {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	// Create a temporary directory for the scope of our test to operate in.
	tmpDir, err := os.MkdirTemp("", "flowtest-*")
	if err != nil {
		return fmt.Errorf("creating temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archive, err := txtar.ParseFile(filename)
	if err != nil {
		return fmt.Errorf("loading script: %w", err)
	}

	// Create a test controller for the script to interact with. The test
	// controller's logs are buffered and printed to stderr on exit. This
	// prevents log mangling when both the script engine and the Flow controller
	// are writing logs at the same time (as the script engine will write partial
	// lines).
	var controllerLogs bytes.Buffer
	defer func() {
		fmt.Fprintln(os.Stderr, "[controller logs]")
		io.Copy(os.Stderr, &controllerLogs)
	}()
	f, err := newTestController(&controllerLogs, filepath.Join(tmpDir, "data"))
	if err != nil {
		return fmt.Errorf("creating test controller: %w", err)
	}

	// Create a state for the duration of our test, using it to unpack the txtar
	// archive into our working directory.
	//
	// Because scripts are frequently asynchronous (such as controller.start), we
	// need to close the state in a defer to make sure everything gets cleaned up
	// properly.
	s, err := script.NewState(context.Background(), tmpDir, nil)
	if err != nil {
		return fmt.Errorf("creating state: %w", err)
	}
	if err := s.ExtractFiles(archive); err != nil {
		return fmt.Errorf("extracting archive: %w", err)
	}
	defer s.CloseAndWait(os.Stderr)

	// Finally, prepare for running the script:
	//
	// 1. Change our working directory to the temporary directory. This allows
	//    components which rely on the working directory to work properly. The
	//    previous working directory is restored on exit. This step *MUST* happen
	//    here at the end before executing the engine, otherwise calls to
	//    TestScript relying on the working directory of tests will fail.
	//
	// 2. Create a new engine and register commands to scripts to use.
	//
	// 3. Execute the script.
	initWD, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		return fmt.Errorf("changing working directory: %w", err)
	}
	defer os.Chdir(initWD)

	e := script.NewEngine()
	e.Cmds = createCommands(f, &o)
	return e.Execute(s, filename, bufio.NewReader(bytes.NewReader(archive.Comment)), os.Stderr)
}

func newTestController(out io.Writer, dataDir string) (*flow.Flow, error) {
	logger, err := logging.New(out, logging.Options{
		Level:  logging.LevelDebug,
		Format: logging.FormatDefault,
	})
	if err != nil {
		return nil, fmt.Errorf("creating logger: %w", err)
	}

	tracer, err := tracing.New(tracing.Options{
		SamplingFraction: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("creating tracer: %w", err)
	}

	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}

	f := flow.New(flow.Options{
		Logger: logger,
		Tracer: tracer,
		Reg:    prometheus.NewRegistry(),

		DataPath: dataDir,
		// TODO(rfratto): services?
	})
	return f, nil
}

// TestOption modifies default behavior of testing functions.
type TestOption func(*options)

type options struct {
	cmds map[string]script.Cmd
}

// WithExtraCommands provides a list of extra commands available for scripts.
// If a key in cmds matches the name of an existing command, the existing
// command is shadowed by the command.
//
// WithExtraCommands can be passed multiple times.
func WithExtraCommands(cmds map[string]script.Cmd) TestOption {
	return func(o *options) {
		if o.cmds == nil {
			o.cmds = make(map[string]script.Cmd)
		}
		for k, cmd := range cmds {
			o.cmds[k] = cmd
		}
	}
}
