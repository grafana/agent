package autoscrape

import (
	"fmt"

	"github.com/prometheus/prometheus/pkg/exemplar"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"
)

// failedAppender is used as the appender when an instance couldn't be found.
type failedAppender struct {
	instanceName string
}

var _ storage.Appender = (*failedAppender)(nil)

func (fa *failedAppender) Append(ref uint64, l labels.Labels, t int64, v float64) (uint64, error) {
	return 0, fmt.Errorf("no such instance %s", fa.instanceName)
}

func (fa *failedAppender) Commit() error {
	return fmt.Errorf("no such instance %s", fa.instanceName)
}

func (fa *failedAppender) Rollback() error {
	return fmt.Errorf("no such instance %s", fa.instanceName)
}

func (fa *failedAppender) AppendExemplar(ref uint64, l labels.Labels, e exemplar.Exemplar) (uint64, error) {
	return 0, fmt.Errorf("no such instance %s", fa.instanceName)
}
