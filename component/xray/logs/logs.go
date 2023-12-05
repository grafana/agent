package write

import (
	"container/ring"
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
)

func init() {
	component.Register(component.Registration{
		Name:    "xray.logs",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the xray.logs component.
type Arguments struct {
	EntriesToTrack int `river:"entries_to_track,attr,optional"`
}

var DefaultArguments = Arguments{
	EntriesToTrack: 100,
}

// SetToDefault implements river.Defaulter
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Exports holds the receiver that is used to send log entries to the
// xray.logs component.
type Exports struct {
	Receiver loki.LogsReceiver `river:"receiver,attr"`
}

var (
	_ component.Component = (*Component)(nil)
)

// Component implements the xray.logs component.
type Component struct {
	opts component.Options
	args Arguments

	mut      sync.RWMutex
	receiver loki.LogsReceiver

	logsSummary *logsSummary
}

// New creates a new xray.logs component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts: o,
		args: args,
	}

	// Create and immediately export the receiver which remains the same for
	// the component's lifetime.
	c.receiver = loki.NewLogsReceiver()
	o.OnStateChange(Exports{Receiver: c.receiver})

	c.initializeLogsSummary()

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
		case entry := <-c.receiver.Chan():
			c.mut.Lock()

			for k, v := range entry.Labels {
				c.logsSummary.LogLabelFrequencies[fmt.Sprintf("%s=%s", k, v)]++
			}
			c.logsSummary.TotalLogs++

			// Add entry to the ring buffer, overwriting the oldest entry if
			// necessary.
			c.logsSummary.LogsRingBuffer.Value = entry
			if c.logsSummary.SetEntries < c.args.EntriesToTrack {
				c.logsSummary.SetEntries++
			}
			c.logsSummary.LogsRingBuffer = c.logsSummary.LogsRingBuffer.Next()

			c.mut.Unlock()
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	return nil
}

func (c *Component) DebugInfo() interface{} {
	c.mut.RLock()
	defer c.mut.RUnlock()

	return &debugInfo{
		LogsSummary:      c.logsSummary,
		LogsSummaryStats: c.logsSummary.getLogsSummaryStats(),
	}
}

type debugInfo struct {
	LogsSummary      *logsSummary     `river:"logs_summary,attr"`
	LogsSummaryStats logsSummaryStats `river:"recent_logs_stats,attr"`
}

func (c *Component) initializeLogsSummary() {
	c.logsSummary = &logsSummary{
		LogLabelFrequencies: make(map[string]int),
		LogsRingBuffer:      ring.New(c.args.EntriesToTrack),
	}
}

type logsSummary struct {
	LogLabelFrequencies map[string]int `river:"log_label_frequencies,attr"`
	TotalLogs           int            `river:"total_log_entries,attr"`

	LogsRingBuffer *ring.Ring
	SetEntries     int
}

type logsSummaryStats struct {
	MedianLength  int     `river:"median_length,attr"`
	AverageLength float64 `river:"average_length,attr"`
	MaxLength     int     `river:"max_length,attr"`
	MinLength     int     `river:"min_length,attr"`
}

func (l *logsSummary) getLogsSummaryStats() logsSummaryStats {
	if l.SetEntries == 0 {
		return logsSummaryStats{}
	}

	entryLengths := make([]int, 0, l.SetEntries)
	l.LogsRingBuffer.Do(func(v interface{}) {
		if v == nil {
			return
		}
		entryLengths = append(entryLengths, len(v.(loki.Entry).Line))
	})

	// Sort the entry lengths in ascending order.
	slices.SortFunc(entryLengths, func(a, b int) int {
		return a - b
	})
	total := 0
	for _, v := range entryLengths {
		total += v
	}
	return logsSummaryStats{
		MedianLength:  entryLengths[len(entryLengths)/2],
		AverageLength: float64(total) / float64(len(entryLengths)),
		MaxLength:     entryLengths[len(entryLengths)-1],
		MinLength:     entryLengths[0],
	}
}
