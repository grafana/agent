package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/loki/clients/pkg/promtail/api"
	"github.com/grafana/loki/clients/pkg/promtail/positions"
	"github.com/prometheus/common/model"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.file",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

const (
	pathLabel     = "__path__"
	filenameLabel = "filename"
)

// Arguments holds values which are used to configure the loki.source.file
// component.
type Arguments struct {
	Targets   []discovery.Target `river:"targets,attr,optional"`
	ForwardTo []chan api.Entry   `river:"forward_to,attr,optional"`
}

// DefaultArguments defines the default settings for loki.source.file.
var DefaultArguments = Arguments{}

// UnmarshalRiver implements river.Unmarshaler.
func (arg *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*arg = DefaultArguments

	type args Arguments
	return f((*args)(arg))
}

var (
	_ component.Component = (*Component)(nil)
)

// Component implements the loki.source.file component.
type Component struct {
	opts    component.Options
	metrics *metrics

	mut       sync.RWMutex
	args      Arguments
	handler   chan api.Entry
	receivers []chan api.Entry
	posFile   positions.Positions
	readers   map[string]Reader
}

// New creates a new loki.source.file component.
func New(o component.Options, args Arguments) (*Component, error) {
	err := os.Mkdir(o.DataPath, 0750)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	positionsFile, err := positions.New(o.Logger, positions.Config{
		SyncPeriod:        10 * time.Second,
		PositionsFile:     filepath.Join(o.DataPath, "positions.yml"),
		IgnoreInvalidYaml: false,
		ReadOnly:          false,
	})
	if err != nil {
		return nil, err
	}

	c := &Component{
		opts:    o,
		metrics: newMetrics(o.Registerer),

		handler:   make(chan api.Entry),
		receivers: args.ForwardTo,
		posFile:   positionsFile,
		readers:   make(map[string]Reader),
	}

	// Call to Update() to start readers and set receivers once at the start.
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
// TODO(@tpaschalis). Should we periodically re-check? What happens if a target
// comes alive _after_ it's been passed to us and we never receive another
// Update()? Or should it be a responsibility of the discovery component?
func (c *Component) Run(ctx context.Context) error {
	defer func() {
		level.Info(c.opts.Logger).Log("msg", "loki.source.file component shutting down, stopping readers")
		for _, r := range c.readers {
			r.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case entry := <-c.handler:
			for _, receiver := range c.receivers {
				receiver <- entry
			}
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.mut.Lock()
	defer c.mut.Unlock()
	c.args = newArgs
	c.receivers = newArgs.ForwardTo

	oldPaths := make(map[string]struct{})

	// Stop all readers and recreate them below. This avoids the issue we saw
	// with stranded wrapped handlers staying behind until they were GC'ed and
	// sending duplicate message to the global handler. It also makes sure that
	// we update everything with the new labels. Simply zeroing out the
	// c.readers map did not work correctly to shut down the wrapped handlers
	// in time.
	// TODO (@tpaschalis) We should be able to optimize this somehow and eg.
	// cache readers for paths we already know about, and whose labels have not
	// changed. Once we do that we should:
	// a) Call to c.pruneStoppedReaders to give cached but errored readers a
	// chance to restart.
	// b) Stop tailing any files that were no longer in the new targets
	// and conditionally remove their readers only by calling toStopTailing
	// and c.stopTailingAndRemovePosition.
	for p, r := range c.readers {
		oldPaths[p] = struct{}{}
		r.Stop()
	}
	c.readers = make(map[string]Reader)

	if len(newArgs.Targets) == 0 {
		level.Debug(c.opts.Logger).Log("msg", "no files targets were passed, nothing will be tailed")
		return nil
	}

	var paths []string
	for _, target := range newArgs.Targets {
		path := target[pathLabel]
		c.reportSize(path)
		paths = append(paths, path)

		var labels = make(model.LabelSet)
		for k, v := range target {
			if strings.HasPrefix(k, "__") {
				continue
			}
			labels[model.LabelName(k)] = model.LabelValue(v)
		}

		handler := api.AddLabelsMiddleware(labels).Wrap(api.NewEntryHandler(c.handler, func() {}))

		reader, err := c.startTailing(path, handler, labels)
		if err != nil {
			continue // TODO (@tpaschalis) return err maybe?
		}

		c.readers[path] = reader
	}

	// Remove from the positions file any paths that had a Reader before, but
	// are no longer in the updated set of Targets.
	for path := range missing(c.readers, oldPaths) {
		c.posFile.Remove(path)
	}

	return nil
}

// Returns the elements from set b which are missing from set a
func missing(as map[string]Reader, bs map[string]struct{}) map[string]struct{} {
	c := map[string]struct{}{}
	for a := range bs {
		if _, ok := as[a]; !ok {
			c[a] = struct{}{}
		}
	}
	return c
}

// startTailing starts and returns a reader for the given path. For most files,
// this will be a tailer implementation. If the file suffix alludes to it being
// a compressed file, then a decompressor will be started instead.
func (c *Component) startTailing(path string, handler api.EntryHandler, labels model.LabelSet) (Reader, error) {
	fi, err := os.Stat(path)
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "failed to tail file, stat failed", "error", err, "filename", path)
		c.metrics.totalBytes.DeleteLabelValues(path)
		return nil, fmt.Errorf("failed to stat path %s", path)
	}

	if fi.IsDir() {
		level.Info(c.opts.Logger).Log("msg", "failed to tail file", "error", "file is a directory", "filename", path)
		c.metrics.totalBytes.DeleteLabelValues(path)
		return nil, fmt.Errorf("failed to tail file, it was a directory %s", path)
	}

	var reader Reader
	if isCompressed(path) {
		level.Debug(c.opts.Logger).Log("msg", "reading from compressed file", "filename", path)
		decompressor, err := newDecompressor(
			c.metrics,
			c.opts.Logger,
			handler,
			c.posFile,
			path,
			"",
		)
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "failed to start decompressor", "error", err, "filename", path)
			return nil, fmt.Errorf("failed to start decompressor %s", err)
		}
		reader = decompressor
	} else {
		level.Debug(c.opts.Logger).Log("msg", "tailing new file", "filename", path)
		tailer, err := newTailer(
			c.metrics,
			c.opts.Logger,
			handler,
			c.posFile,
			path,
			"",
		)
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "failed to start tailer", "error", err, "filename", path)
			return nil, fmt.Errorf("failed to start tailer %s", err)
		}
		reader = tailer
	}

	return reader, nil
}

func (c *Component) reportSize(path string) {
	// Ask the reader to update the size if a reader exists, this keeps
	// position and size metrics in sync.
	if reader, ok := c.readers[path]; ok {
		err := reader.MarkPositionAndSize()
		if err != nil {
			level.Warn(c.opts.Logger).Log("msg", "failed to get file size from existing reader, ", "file", path, "error", err)
			return
		}
	} else {
		// Must be a new file, just directly read the size of it
		fi, err := os.Stat(path)
		if err != nil {
			return
		}
		c.metrics.totalBytes.WithLabelValues(path).Set(float64(fi.Size()))
	}
}
