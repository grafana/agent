// Package configstore abstracts the concepts of where instance files get
// retrieved.
package configstore

import (
	"context"

	"github.com/grafana/agent/pkg/prom/instance"
)

// Store is some interface to retrieving instance configurations.
type Store interface {
	// List gets the list of config names.
	List(ctx context.Context) ([]string, error)

	// Get gets an individual config by name.
	Get(ctx context.Context, key string) (instance.Config, error)

	// Put applies a new instance Config to the store.
	// If the config already exists, created will be false to indicate an
	// update.
	Put(ctx context.Context, c instance.Config) (created bool, err error)

	// Delete deletes a config from the store.
	Delete(ctx context.Context, key string) error

	// All retrieves the entire list of instance configs currently
	// in the store. A filtering "keep" function can be provided to ignore some
	// configs, which can significantly speed up the operation in some cases.
	All(ctx context.Context, keep func(key string) bool) (<-chan instance.Config, error)

	// Watch watches for new instance Configs. The entire set of known
	// instance configs is returned each time.
	//
	// All callers of Watch receive the same Channel.
	Watch() <-chan []instance.Config

	// Close closes the store.
	Close() error
}
