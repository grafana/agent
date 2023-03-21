// Package string defines the module.file component.
package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/web/api"
	"github.com/prometheus/client_golang/prometheus"
)

// waitReadPeriod holds the time to wait before reading a file while the
// local.file component is running.
//
// This prevents local.file from updating too frequently and exporting partial
// writes.
const waitReadPeriod time.Duration = 30 * time.Millisecond

func init() {
	component.Register(component.Registration{
		Name:    "module.file",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the module.file
// component.
type Arguments struct {
	// Filename indicates the file to watch.
	Filename string `river:"filename,attr"`
	// Type indicates how to detect changes to the file.
	Type Detector `river:"detector,attr,optional"`
	// PollFrequency determines the frequency to check for changes when Type is
	// UpdateTypePoll.
	PollFrequency time.Duration `river:"poll_freqency,attr,optional"`
	// IsSecret marks the file as holding a secret value which should not be
	// displayed to the user.
	IsSecret bool `river:"is_secret,attr,optional"`

	// Arguments to pass into the module.
	Arguments map[string]any `river:"arguments,attr,optional"`
}

// DefaultArguments provides the default arguments for the local.file
// component.
var DefaultArguments = Arguments{
	Type:          DetectorFSNotify,
	PollFrequency: time.Minute,
}

var _ river.Unmarshaler = (*Arguments)(nil)

// UnmarshalRiver implements river.Unmarshaler.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = DefaultArguments

	type arguments Arguments
	err := f((*arguments)(a))
	if err != nil {
		return err
	}

	if a.PollFrequency < 0 {
		return errors.New("poll_freqency must be greater than 0")
	}

	return nil
}

// Exports holds values which are exported from the run module.
type Exports struct {
	// Exports exported from the running module.
	Exports map[string]any `river:"exports,attr"`
}

// Component implements the module.file component.
type Component struct {
	opts component.Options
	log  log.Logger
	ctrl *flow.Flow

	mut      sync.Mutex
	args     Arguments
	detector io.Closer
	content  rivertypes.OptionalSecret `river:"content,attr"`

	// reloadCh is a buffered channel which is written to when the watched file
	// should be reloaded by the component.
	reloadCh chan struct{}
}

var (
	_ component.Component     = (*Component)(nil)
	_ component.HTTPComponent = (*Component)(nil)
)

// New creates a new module.file component.
func New(o component.Options, args Arguments) (*Component, error) {
	// TODO(rfratto): replace these with a tracer/registry which properly
	// propagates data back to the parent.
	flowTracer, _ := tracing.New(tracing.DefaultOptions)
	flowRegistry := prometheus.NewRegistry()

	c := &Component{
		opts: o,
		log:  o.Logger,

		ctrl: flow.New(flow.Options{
			ControllerID: o.ID,
			LogSink:      logging.LoggerSink(o.Logger),
			Tracer:       flowTracer,
			Reg:          flowRegistry,

			DataPath:       o.DataPath,
			HTTPPathPrefix: o.HTTPPath,
			HTTPListenAddr: o.HTTPListenAddr,

			OnExportsChange: func(exports map[string]any) {
				o.OnStateChange(Exports{Exports: exports})
			},
		}),

		reloadCh: make(chan struct{}, 1),
	}

	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	go c.pollForChanges(ctx)
	c.ctrl.Run(ctx)
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.mut.Lock()
	defer c.mut.Unlock()
	c.args = newArgs

	// Force an immediate read of the file to report any potential errors early.
	newContent, err := c.readFile()
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	c.content = newContent

	// Each detector is dedicated to a single file path. We'll naively shut down
	// the existing detector (if any) before setting up a new one to make sure
	// the correct file is being watched in case the path changed between calls
	// to Update.
	if c.detector != nil {
		if err := c.detector.Close(); err != nil {
			level.Error(c.opts.Logger).Log("msg", "failed to shut down old detector", "err", err)
		}
		c.detector = nil
	}

	if err := c.configureDetector(); err != nil {
		return err
	}

	// The flow controller will now read the contents of the file
	f, err := flow.ReadFile(c.opts.ID, []byte(c.content.Value))
	if err != nil {
		return err
	}

	return c.ctrl.LoadFile(f, c.args.Arguments)
}

// Handler implements component.HTTPComponent.
func (c *Component) Handler() http.Handler {
	r := mux.NewRouter()

	fa := api.NewFlowAPI(c.ctrl, r)
	fa.RegisterRoutes("/", r)

	r.PathPrefix("/{id}/").Handler(c.ctrl.ComponentHandler())
	return r
}

func (c *Component) readFile() (rivertypes.OptionalSecret, error) {
	// Force a re-load of the file outside of the update detection mechanism.
	bb, err := os.ReadFile(c.args.Filename)
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "failed to read file", "path", c.opts.DataPath, "err", err)
		return rivertypes.OptionalSecret{}, err
	}

	content := rivertypes.OptionalSecret{
		IsSecret: c.args.IsSecret,
		Value:    string(bb),
	}

	return content, nil
}

// configureDetector configures the detector if one isn't set. mut must be held
// when called.
func (c *Component) configureDetector() error {
	if c.detector != nil {
		// Already have a detector; don't do anything.
		return nil
	}

	var err error

	reloadFile := func() {
		select {
		case c.reloadCh <- struct{}{}:
		default:
			// no-op: a reload is already queued so we don't need to queue a second
			// one.
		}
	}

	switch c.args.Type {
	case DetectorPoll:
		c.detector = newPoller(pollerOptions{
			Filename:      c.args.Filename,
			ReloadFile:    reloadFile,
			PollFrequency: c.args.PollFrequency,
		})
	case DetectorFSNotify:
		c.detector, err = newFSNotify(fsNotifyOptions{
			Logger:       c.opts.Logger,
			Filename:     c.args.Filename,
			ReloadFile:   reloadFile,
			PollFreqency: c.args.PollFrequency,
		})
	}

	return err
}

// Run implements component.Component.
func (c *Component) pollForChanges(ctx context.Context) error {
	defer func() {
		c.mut.Lock()
		defer c.mut.Unlock()

		if err := c.detector.Close(); err != nil {
			level.Error(c.opts.Logger).Log("msg", "failed to shut down detector", "err", err)
		}
		c.detector = nil
	}()

	// Since Run _may_ get recalled if we're told to exit but still exist in the
	// config file, we may have prematurely destroyed the detector. If no
	// detector exists, we need to recreate it for Run to work properly.
	//
	// We ignore the error (indicating the file has disappeared) so we can allow
	// the detector to inform us when it comes back.
	//
	// TODO(rfratto): this is a design wart, and can hopefully be removed in
	// future iterations.
	c.mut.Lock()
	_ = c.configureDetector()
	c.mut.Unlock()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-c.reloadCh:
			time.Sleep(waitReadPeriod)

			// We ignore the error here from readFile since readFile will log errors
			// and also report the error as the health of the component.
			var newContent rivertypes.OptionalSecret

			c.mut.Lock()
			newContent, _ = c.readFile()
			foundChange := !reflect.DeepEqual(newContent, c.content)
			c.mut.Unlock()

			if foundChange {
				c.Update(c.args)
			}
		}
	}
}
