package observer

import (
	"context"
	"io"
)

type agentStateWriter interface {
	io.Writer
	SetContext(ctx context.Context)
}
