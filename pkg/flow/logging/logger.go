package logging

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/common/model"
)

// Logger implements the github.com/go-kit/log.Logger interface. It supports
// being dynamically updated at runtime.
type Logger struct {
	w io.Writer

	mut sync.RWMutex
	l   log.Logger
}

// New creates a New logger with the default log level and format.
func New(w io.Writer, o Options) (*Logger, error) {
	inner, err := buildLogger(w, o)
	if err != nil {
		return nil, err
	}

	return &Logger{w: w, l: inner}, nil
}

// Log implements log.Logger.
func (l *Logger) Log(kvps ...interface{}) error {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.l.Log(kvps...)
}

// Update re-configures the options used for the logger.
func (l *Logger) Update(o Options) error {
	newLogger, err := buildLogger(l.w, o)
	if err != nil {
		return err
	}

	l.mut.Lock()
	defer l.mut.Unlock()
	l.l = newLogger
	return nil
}

func buildLogger(w io.Writer, o Options) (log.Logger, error) {
	var l log.Logger
	var wr io.Writer
	wr = w

	if len(o.WriteTo) > 0 {
		wr = io.MultiWriter(w, &lokiWriter{o.WriteTo})
	}

	switch o.Format {
	case FormatLogfmt:
		l = log.NewLogfmtLogger(log.NewSyncWriter(wr))
	case FormatJSON:
		l = log.NewJSONLogger(log.NewSyncWriter(wr))
	default:
		return nil, fmt.Errorf("unrecognized log format %q", o.Format)
	}

	l = level.NewFilter(l, o.Level.Filter())

	l = log.With(l, "ts", log.DefaultTimestampUTC)
	return l, nil
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
