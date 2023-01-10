package windowsevent

import (
	"context"
	"os"
	"path"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/loki/clients/pkg/promtail/api"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	"github.com/grafana/loki/clients/pkg/promtail/targets/windows"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.windowsevent",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

var (
	_ component.Component = (*Component)(nil)
)

// Component implements the loki.source.windowsevent component.
type Component struct {
	opts component.Options

	mut        sync.RWMutex
	args       Arguments
	target     *windows.Target
	handler    chan api.Entry
	doneCtx    context.Context
	cancelFunc context.CancelFunc
	receivers  []loki.LogsReceiver
}

func (c *Component) Chan() chan<- api.Entry {
	return c.handler
}

func (c *Component) Stop() {
	c.cancelFunc()
}

// New creates a new loki.source.windowsevent component.
func New(o component.Options, args Arguments) (*Component, error) {

	c := &Component{
		opts:      o,
		receivers: args.ForwardTo,
		handler:   make(chan api.Entry),
		args:      args,
	}

	// Call to Update() to start readers and set receivers once at the start.
	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-c.doneCtx.Done():
			return nil
		case entry := <-c.handler:
			lokiEntry := loki.Entry{
				Labels: entry.Labels,
				Entry:  entry.Entry,
			}
			for _, receiver := range c.receivers {
				receiver <- lokiEntry
			}
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.mut.Lock()
	defer c.mut.Unlock()

	// If no bookmark specified create one in the datapath.
	if newArgs.BookmarkPath == "" {
		newArgs.BookmarkPath = path.Join(c.opts.DataPath, "bookmark.xml")
	}

	// Create the bookmark file and parent folders if they dont exist.
	_, err := os.Stat(newArgs.BookmarkPath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path.Dir(newArgs.BookmarkPath), 644)
		if err != nil {
			return err
		}
		_, err = os.Create(newArgs.BookmarkPath)
		if err != nil {
			return err
		}
	}

	// Close the original target.
	if c.target != nil {
		err := c.target.Stop()
		if err != nil {
			return err
		}
	}
	c.doneCtx, c.cancelFunc = context.WithCancel(context.Background())
	winTarget, err := windows.New(c.opts.Logger, c, nil, convertConfig(newArgs))
	if err != nil {
		return err
	}
	c.target = winTarget

	c.args = newArgs
	c.receivers = newArgs.ForwardTo
	return nil
}

func convertConfig(arg Arguments) *scrapeconfig.WindowsEventsTargetConfig {
	return &scrapeconfig.WindowsEventsTargetConfig{
		Locale:               uint32(arg.Locale),
		EventlogName:         arg.EventLogName,
		Query:                arg.XPathQuery,
		UseIncomingTimestamp: arg.UseIncomingTimestamp,
		BookmarkPath:         arg.BookmarkPath,
		PollInterval:         arg.PollInterval,
		ExcludeEventData:     arg.ExcludeEventData,
		ExcludeEventMessage:  false,
		ExcludeUserData:      arg.ExcludeUserdata,
	}
}
