package logging

import (
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/internal/slogadapter"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/common/model"
)

// Logger is the logging subsystem of Flow. It supports being dynamically
// updated at runtime.
type Logger struct {
	inner io.Writer // Writer passed to New.

	level   *slog.LevelVar // Current configured level.
	format  *formatVar     // Current configured format.
	writer  *writerVar     // Current configured multiwriter (inner + write_to).
	handler *handler       // Handler which handles logs.
}

// New creates a New logger with the default log level and format.
func New(w io.Writer, o Options) (*Logger, error) {
	var (
		leveler slog.LevelVar
		format  formatVar
		writer  writerVar
	)

	l := &Logger{
		inner: w,

		level:  &leveler,
		format: &format,
		writer: &writer,
		handler: &handler{
			w:         &writer,
			leveler:   &leveler,
			formatter: &format,
		},
	}

	if err := l.Update(o); err != nil {
		return nil, err
	}
	return l, nil
}

// Handler returns a [slog.Handler]. The returned Handler remains valid if l is
// updated.
func (l *Logger) Handler() slog.Handler { return l.handler }

// Update re-configures the options used for the logger.
func (l *Logger) Update(o Options) error {
	switch o.Format {
	case FormatLogfmt, FormatJSON:
		// no-op
	default:
		return fmt.Errorf("unrecognized log format %q", o.Format)
	}

	l.level.Set(slogLevel(o.Level).Level())
	l.format.Set(o.Format)

	newWriter := l.inner
	if len(o.WriteTo) > 0 {
		newWriter = io.MultiWriter(l.inner, &lokiWriter{o.WriteTo})
	}
	l.writer.Set(newWriter)

	return nil
}

// Log implements log.Logger.
func (l *Logger) Log(kvps ...interface{}) error {
	// NOTE(rfratto): this method is a temporary shim while log/slog is still
	// being adopted throughout the codebase.
	return slogadapter.GoKit(l.handler).Log(kvps...)
}

type lokiWriter struct {
	f []loki.LogsReceiver
}

func (fw *lokiWriter) Write(p []byte) (int, error) {
	for _, receiver := range fw.f {
		select {
		case receiver.Chan() <- loki.Entry{
			Labels: model.LabelSet{"component": "agent"},
			Entry: logproto.Entry{
				Timestamp: time.Now(),
				Line:      string(p),
			},
		}:
		default:
			return 0, fmt.Errorf("lokiWriter failed to forward entry, channel was blocked")
		}
	}
	return len(p), nil
}

type formatVar struct {
	mut sync.RWMutex
	f   Format
}

func (f *formatVar) Format() Format {
	f.mut.RLock()
	defer f.mut.RUnlock()
	return f.f
}

func (f *formatVar) Set(format Format) {
	f.mut.Lock()
	defer f.mut.Unlock()
	f.f = format
}

type writerVar struct {
	mut sync.RWMutex
	w   io.Writer
}

func (w *writerVar) Set(inner io.Writer) {
	w.mut.Lock()
	defer w.mut.Unlock()
	w.w = inner
}

func (w *writerVar) Write(p []byte) (n int, err error) {
	w.mut.RLock()
	defer w.mut.RUnlock()

	if w.w == nil {
		return 0, fmt.Errorf("no writer available")
	}

	return w.w.Write(p)
}
