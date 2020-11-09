package spanlogger

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/cortexproject/cortex/pkg/util"
)

type loggerCtxMarker struct{}

var (
	loggerCtxKey = &loggerCtxMarker{}
)

// SpanLogger unifies tracing and logging to reduce repetition. Borrowed from
// https://github.com/cortexproject/cortex/blob/master/pkg/util/spanlogger/spanlogger.go
// with extensions to support custom loggers.
type SpanLogger struct {
	log.Logger
	opentracing.Span
}

// NewFromLogger makes a new SpanLogger from an existing logger.
func NewFromLogger(ctx context.Context, l log.Logger, method string, kvps ...interface{}) (*SpanLogger, context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, method)
	logger := &SpanLogger{
		Logger: log.With(util.WithContext(ctx, l), "method", method),
		Span:   span,
	}
	if len(kvps) > 0 {
		level.Debug(logger).Log(kvps...)
	}

	ctx = context.WithValue(ctx, loggerCtxKey, l)
	return logger, ctx
}

// FromContext returns a span logger using the current parent span and the
// logger in the context, attached from NewWithLogger. If there is no parent
// span or logger, the SpanLogger will only log using the fallback logger
// provided.
func FromContext(ctx context.Context, fallback log.Logger) *SpanLogger {
	logger, ok := ctx.Value(loggerCtxKey).(log.Logger)
	if !ok {
		logger = fallback
	}

	sp := opentracing.SpanFromContext(ctx)
	if sp == nil {
		sp = defaultNoopSpan
	}
	return &SpanLogger{
		Logger: util.WithContext(ctx, logger),
		Span:   sp,
	}
}

// Log implements go-kit's Logger interface; sends logs to underlying logger
// and also puts the on the spans.
func (s *SpanLogger) Log(kvps ...interface{}) error {
	s.Logger.Log(kvps...)
	fields, err := otlog.InterleavedKVToFields(kvps...)
	if err != nil {
		return err
	}
	s.Span.LogFields(fields...)
	return nil
}

// Error sets error flag and logs the error on the span, if non-nil. Returns
// the err passed in.
func (s *SpanLogger) Error(err error) error {
	if err == nil {
		return nil
	}
	ext.Error.Set(s.Span, true)
	s.Span.LogFields(otlog.Error(err))
	return err
}
