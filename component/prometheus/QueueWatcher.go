package prometheus

import "context"

type WALWatcher interface {
	SetWriteTo(write WriteTo, ctx context.Context)
	Start() error
}
