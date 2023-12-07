package write

import (
	"container/ring"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/prometheus/alertmanager/pkg/labels"
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
			c.logsSummary.logsRingBuffer.Value = entry
			if c.logsSummary.setEntries < c.args.EntriesToTrack {
				c.logsSummary.setEntries++
			}
			c.logsSummary.logsRingBuffer = c.logsSummary.logsRingBuffer.Next()

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
	LogsSummary      *logsSummary                 `river:"logs_summary,attr" json:"logs_summary"`
	LogsSummaryStats map[string]*logsSummaryStats `river:"recent_logs_stats,attr" json:"recent_logs_stats"`
}

func (c *Component) initializeLogsSummary() {
	c.logsSummary = &logsSummary{
		LogLabelFrequencies: make(map[string]int),
		logsRingBuffer:      ring.New(c.args.EntriesToTrack),
	}
}

type logsSummary struct {
	LogLabelFrequencies map[string]int `river:"log_label_frequencies,attr" json:"log_label_frequencies"`
	TotalLogs           int            `river:"total_log_entries,attr" json:"total_log_entries"`

	logsRingBuffer *ring.Ring
	setEntries     int
}

type logsSummaryStats struct {
	MinLength     int     `river:"min_length,attr" json:"min_length"`
	MaxLength     int     `river:"max_length,attr" json:"max_length"`
	MedianLength  int     `river:"median_length,attr" json:"median_length"`
	AverageLength float64 `river:"average_length,attr" json:"average_length"`
	NumEntries    int     `river:"num_entries,attr" json:"num_entries"`
}

func (l *logsSummary) getLogsSummaryStats() map[string]*logsSummaryStats {
	stats := make(map[string]*logsSummaryStats)
	if l.setEntries == 0 {
		return stats
	}

	entryLengths := make([]loki.Entry, 0, l.setEntries)
	l.logsRingBuffer.Do(func(v interface{}) {
		if v == nil {
			return
		}
		entryLengths = append(entryLengths, v.(loki.Entry))
	})

	// Sort the entry lengths in ascending order.
	slices.SortFunc(entryLengths, func(a, b loki.Entry) int {
		return len(a.Entry.Line) - len(b.Entry.Line)
	})

	entriesByLabels := make(map[string][]int)

	for _, v := range entryLengths {
		entriesByLabels[v.Labels.String()] = append(entriesByLabels[v.Labels.String()], len(v.Entry.Line))
	}

	for labels, entries := range entriesByLabels {
		total := 0
		for _, v := range entries {
			total += v
		}
		stats[labels] = &logsSummaryStats{
			MedianLength:  entries[len(entries)/2],
			AverageLength: float64(total) / float64(len(entries)),
			MaxLength:     entries[len(entries)-1],
			MinLength:     entries[0],
			NumEntries:    len(entries),
		}
	}
	return stats
}

func (c *Component) Handler() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("/summary", func(w http.ResponseWriter, r *http.Request) {
		di := c.DebugInfo().(*debugInfo)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(di)
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "failed to encode json", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	router.HandleFunc("/recent-logs", func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		// Construct matchers from the query params
		matchers := labels.Matchers{}
		for k, v := range params {
			// Error is ignored because errors are only produces when using
			// labels.MatchRegexp or labels.MatchNotRegexp
			m, _ := labels.NewMatcher(labels.MatchEqual, k, v[0])
			matchers = append(matchers, m)
		}

		matches := make([]loki.Entry, 0)
		c.logsSummary.logsRingBuffer.Do(func(v interface{}) {
			if v == nil {
				return
			}
			entry := v.(loki.Entry)

			// If no params are specified, return all entries.
			if len(params) == 0 {
				matches = append(matches, entry)
			} else {
				if matchers.Matches(entry.Labels) {
					matches = append(matches, entry)
				}
			}
		})

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(matches)
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "failed to encode json", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	return router
}
