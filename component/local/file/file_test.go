package file_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/local/file"
	"github.com/grafana/agent/pkg/flow/hcltypes"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestFile(t *testing.T) {
	t.Run("Polling change detector", func(t *testing.T) {
		t.Parallel()
		runFileTests(t, file.DetectorPoll)
	})

	t.Run("Event change detector", func(t *testing.T) {
		t.Parallel()
		runFileTests(t, file.DetectorFSNotify)
	})
}

// runFileTests will run a suite of tests with the configured update type.
func runFileTests(t *testing.T, ut file.Detector) {
	// In our tests below, we wait some time after making changes to the file to
	// allow filesystem events to settle.
	settlePeriod := 250 * time.Millisecond

	newSuiteEnvironment := func(t *testing.T, filename string) *testEnvironment {
		err := os.WriteFile(filename, []byte("First load!"), 0644)
		require.NoError(t, err)
		time.Sleep(settlePeriod)

		te := newTestEnvironment(t, file.Arguments{
			Filename: filename,
			Type:     ut,

			// Pick a polling frequency which is fast enough so that tests finish
			// quickly but not so frequent such that Go struggles to schedule the
			// goroutines of the tests on slower machines.
			PollFrequency: 100 * time.Millisecond,
		})
		go func() {
			err := te.Run(context.Background())
			require.NoError(t, err)
		}()

		// Swallow the initial exports notification.
		require.NoError(t, te.WaitExports(time.Second))
		return te
	}

	t.Run("Updates to files are detected", func(t *testing.T) {
		t.Parallel()

		testFile := filepath.Join(t.TempDir(), "testfile")
		te := newSuiteEnvironment(t, testFile)

		// Update the file.
		require.NoError(t, os.WriteFile(testFile, []byte("New content!"), 0644))
		time.Sleep(settlePeriod)

		require.NoError(t, te.WaitExports(time.Second))
		require.Equal(t, file.Exports{
			Content: &hcltypes.OptionalSecret{
				Sensitive: false,
				Value:     "New content!",
			},
		}, te.Exports())
	})

	t.Run("Deleted and recreated files are detected", func(t *testing.T) {
		t.Parallel()

		testFile := filepath.Join(t.TempDir(), "testfile")
		te := newSuiteEnvironment(t, testFile)

		// Delete the file, then recreate it with new content.
		require.NoError(t, os.Remove(testFile))
		require.NoError(t, os.WriteFile(testFile, []byte("New content!"), 0644))
		time.Sleep(settlePeriod)

		require.NoError(t, te.WaitExports(time.Second))
		require.Equal(t, file.Exports{
			Content: &hcltypes.OptionalSecret{
				Sensitive: false,
				Value:     "New content!",
			},
		}, te.Exports())
	})
}

// TestFile_ImmediateExports validates that constructing a local.file component
// immediately exports the contents of the file.
func TestFile_ImmediateExports(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "testfile")
	err := os.WriteFile(testFile, []byte("Hello, world!"), 0644)
	require.NoError(t, err)

	te := newTestEnvironment(t, file.Arguments{
		Filename:      testFile,
		Type:          file.DetectorPoll,
		PollFrequency: 1 * time.Hour,
	})
	go func() {
		err := te.Run(context.Background())
		require.NoError(t, err)
	}()

	require.NoError(t, te.WaitExports(time.Second))
	require.Equal(t, file.Exports{
		Content: &hcltypes.OptionalSecret{
			Sensitive: false,
			Value:     "Hello, world!",
		},
	}, te.Exports())
}

// TestFile_ExistOnLoad ensures that the the configured file must exist on the
// first load of local.file.
func TestFile_ExistOnLoad(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "testfile")

	te := newTestEnvironment(t, file.Arguments{
		Filename:      testFile,
		Type:          file.DetectorPoll,
		PollFrequency: 1 * time.Hour,
	})

	expectError := fmt.Sprintf("failed to read file: open %s: no such file or directory", testFile)

	err := te.Run(canceledContext())
	require.EqualError(t, err, expectError)
}

// testEnvironment provides an environment for testing the local.file
// component.
type testEnvironment struct {
	t *testing.T

	opts component.Options
	args file.Arguments

	exportsMut sync.Mutex
	exports    file.Exports
	exportsCh  chan struct{}
}

func newTestEnvironment(t *testing.T, args file.Arguments) *testEnvironment {
	exportsCh := make(chan struct{}, 1)

	te := &testEnvironment{
		t:         t,
		args:      args,
		exportsCh: exportsCh,
	}

	te.opts = component.Options{
		ID:       "test",
		Logger:   util.TestLogger(t),
		DataPath: t.TempDir(),
		OnStateChange: func(e component.Exports) {
			te.exportsMut.Lock()
			te.exports = e.(file.Exports)
			te.exportsMut.Unlock()

			select {
			case te.exportsCh <- struct{}{}:
			default:
			}
		},
	}

	return te
}

// WaitExports blocks until new Exports are available up to the configured
// timeout.
func (te *testEnvironment) WaitExports(timeout time.Duration) error {
	select {
	case <-time.After(timeout):
		return fmt.Errorf("timed out waiting for exports")
	case <-te.exportsCh:
		return nil
	}
}

// Exports gets the most recent exports for a component.
func (te *testEnvironment) Exports() file.Exports {
	te.exportsMut.Lock()
	defer te.exportsMut.Unlock()
	return te.exports
}

// Run constructs and runs the component, blocking until the test finishes. If
// the component fails, Run returns an error.
func (te *testEnvironment) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	te.t.Cleanup(cancel)

	c, err := file.New(te.opts, te.args)
	if err != nil {
		return err
	}
	return c.Run(ctx)
}

// canceledContext creates a context which is already canceled.
func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
