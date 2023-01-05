package windowsevents

import (
	"context"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/loki/clients/pkg/promtail/targets/windows"
	"github.com/prometheus/common/model"
	"sync"
	"time"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.windowsevents",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the loki.source.windowsevents
// component.
type Arguments struct {
	Locale               int                 `river:"locale,attr,optional"`
	EventLogName         string              `river:"eventlog_name,attr,optional"`
	XPathQuery           string              `river:"xpath_query,attr,optional"`
	BookmarkPath         string              `river:"bookmark_path,attr,optional"`
	PollInterval         time.Duration       `river:"poll_interval,attr,optional"`
	ExcludeEventData     bool                `river:"exclude_event_data,attr,optional"`
	ExcludeUserdata      bool                `river:"exclude_user_data,attr,optional"`
	Labels               map[string]string   `river:"labels,attr,optional"`
	UseIncomingTimestamp bool                `river:"use_incoming_timestamp,attr,optional"`
	ForwardTo            []loki.LogsReceiver `river:"forward_to,attr"`
}

var (
	_ component.Component = (*Component)(nil)
)

// Component implements the loki.source.file component.
type Component struct {
	opts component.Options

	mut       sync.RWMutex
	args      Arguments
	target    *windows.Target
	handler   loki.LogsReceiver
	receivers []loki.LogsReceiver
}

// New creates a new loki.source.windowsevents component.
func New(o component.Options, args Arguments) (*Component, error) {

	c := &Component{
		opts:      o,
		handler:   make(loki.LogsReceiver),
		receivers: args.ForwardTo,
	}
	lblSet := model.LabelSet{}
	for k, v := range args.Labels {
		lblSet[model.LabelName(k)] = model.LabelValue(v)
	}
	winTarget, err := windows.New(o.Logger, loki.AddLabelsMiddleware(lblSet).Wrap(loki.NewEntryHandler(c.handler, func() {})), nil, nil)
	if err != nil {
		return nil, err
	}
	c.target = winTarget

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
	return nil
}
