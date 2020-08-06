package instance

import (
	"context"

	"github.com/prometheus/prometheus/scrape"
)

// NoOpInstance implements the Instance interface in pkg/prom
// but does not do anything. Useful for tests.
type NoOpInstance struct{}

func (NoOpInstance) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (NoOpInstance) Update(_ Config) error {
	return nil
}

func (NoOpInstance) TargetsActive() map[string][]*scrape.Target {
	return nil
}
