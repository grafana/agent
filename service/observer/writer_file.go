package observer

import (
	"context"
	"os"
)

// fileAgentStateWriter writes the Agent state to a file
type fileAgentStateWriter struct {
	filepath string
}

var _ agentStateWriter = (*fileAgentStateWriter)(nil)

func (w *fileAgentStateWriter) Write(p []byte) (n int, err error) {
	f, err := os.Create(w.filepath)
	if err != nil {
		return 0, err
	}

	return f.Write(p)
}

func (w *fileAgentStateWriter) SetContext(ctx context.Context) {
	// No-op
}
