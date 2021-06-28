package frontendcollector

import (
	"github.com/getsentry/sentry-go"
)

type SourceMapStore struct{}

func (store *SourceMapStore) resolveSourceLocation(frame sentry.Frame) (*sentry.Frame, error) {
	return nil, nil // @TODO
}
