package errors

import (
	"fmt"
)

// WithContext is an error associated with a logging context string (e.g. `job="foo", instance="bar"`). It is formatted
// as:
//
//	fmt.Sprintf("[%s] %s", Context(), RawError())
type WithContext interface {
	error

	Context() string
	RawError() string
}

// withContext implements WithContext.
type withContext struct {
	context string
	err     string
}

// New creates a new WithContext.
func New(context, err string) WithContext {
	return &withContext{context, err}
}

// Errorf formats according to a format specifier and returns a new WithContext.
func Errorf(context, format string, a ...interface{}) WithContext {
	return &withContext{context, fmt.Sprintf(format, a...)}
}

// Wrap returns a WithContext wrapping err. If err is nil, it returns nil. If err is a WithContext, it is returned
// unchanged.
func Wrap(context string, err error) WithContext {
	if err == nil {
		return nil
	}
	if w, ok := err.(WithContext); ok {
		return w
	}
	return &withContext{context, err.Error()}
}

// Wrapf returns a WithContext that prepends a formatted message to err.Error(). If err is nil, it returns nil. If err
// is a WithContext, the returned WithContext will have the message prepended but the same context as err (presumed to
// be more specific).
func Wrapf(context string, err error, format string, a ...interface{}) WithContext {
	if err == nil {
		return nil
	}
	prefix := format
	if len(a) > 0 {
		prefix = fmt.Sprintf(format, a...)
	}
	if w, ok := err.(WithContext); ok {
		return &withContext{w.Context(), prefix + ": " + w.RawError()}
	}
	return &withContext{context, prefix + ": " + err.Error()}
}

// Error implements error.
func (w *withContext) Error() string {
	if len(w.context) == 0 {
		return w.err
	}
	return "[" + w.context + "] " + w.err
}

// Context implements WithContext.
func (w *withContext) Context() string {
	return w.context
}

// RawError implements WithContext.
func (w *withContext) RawError() string {
	return w.err
}
