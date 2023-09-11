package labelcache

import (
	"arena"
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component/prometheus/labelcache"
	"github.com/grafana/agent/service"
	"github.com/prometheus/prometheus/model/labels"
)

const ServiceName = "labelcache"

// Options are used to configure the HTTP service. Options are constant for the
// lifetime of the HTTP service.
type Options struct {
	Logger    log.Logger // Where to send logs.
	Directory string
}

type Service struct {
	cache *labelcache.Cache
}

func New(opts Options) *Service {
	c := labelcache.NewCache(opts.Directory, opts.Logger)
	return &Service{
		cache: c,
	}

}

// Definition returns the Definition of the Service.
// Definition must always return the same value across all
// calls.
func (Service) Definition() (_ service.Definition) {
	return service.Definition{
		Name:       ServiceName,
		ConfigType: nil,
		DependsOn:  nil,
	}
}

// Run starts a Service. Run must block until the provided
// context is canceled. Returning an error should be treated
// as a fatal error for the Service.
func (c *Service) Run(ctx context.Context, host service.Host) (_ error) {
	<-ctx.Done()
	return
}

// Update updates a Service at runtime. Update is never
// called if [Definition.ConfigType] is nil. newConfig will
// be the same type as ConfigType; if ConfigType is a
// pointer to a type, newConfig will be a pointer to the
// same type.
//
// Update will be called once before Run, and may be called
// while Run is active.
func (c *Service) Update(newConfig any) (_ error) {
	return nil
}

// Data returns the Data associated with a Service. Data
// must always return the same value across multiple calls,
// as callers are expected to be able to cache the result.
//
// Data may be invoked before Run.
func (c *Service) Data() (_ any) {
	return c
}

func (c *Service) WriteLabels(lbls [][]labels.Label, ttl time.Duration, mem *arena.Arena) ([]uint64, error) {
	return c.cache.WriteLabels(lbls, ttl, mem)
}

func (c *Service) GetLabels(keys []uint64, mem *arena.Arena) ([]labels.Labels, error) {
	return c.cache.GetLabels(keys, mem)
}

func (c *Service) GetOrAddLink(componentID string, localRefID uint64, lbls labels.Labels) uint64 {
	return c.cache.GetOrAddLink(componentID, localRefID, lbls)
}
func (c *Service) GetOrAddGlobalRefID(l labels.Labels) uint64 {
	return c.cache.GetOrAddGlobalRefID(l)
}
func (c *Service) GetGlobalRefID(componentID string, localRefID uint64) uint64 {
	return c.cache.GetGlobalRefID(componentID, localRefID)
}
func (c *Service) GetLocalRefID(componentID string, globalRefID uint64) uint64 {
	return c.cache.GetLocalRefID(componentID, globalRefID)
}

type Data interface {
	WriteLabels(lbls [][]labels.Label, ttl time.Duration, mem *arena.Arena) ([]uint64, error)
	GetLabels(keys []uint64, mem *arena.Arena) ([]labels.Labels, error)
	//TODO add ttl to ref mapping
	GetOrAddLink(componentID string, localRefID uint64, lbls labels.Labels) uint64
	GetOrAddGlobalRefID(l labels.Labels) uint64
	GetGlobalRefID(componentID string, localRefID uint64) uint64
	GetLocalRefID(componentID string, globalRefID uint64) uint64
}

var _ service.Service = (*Service)(nil)
var _ Data = (*Service)(nil)
