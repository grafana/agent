// Package xray implements the xray service for Flow.
// This service provides debug stream APIs for components.
package xray

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-kit/log"
	"github.com/grafana/agent/service"
)

// ServiceName defines the name used for the xray service.
const ServiceName = "xray"

type Service struct {
	loadMut      sync.RWMutex
	debugStreams map[string]func(string)
}

var _ service.Service = (*Service)(nil)

// Data includes information associated with the X-Ray service.
type Data struct {
	DebugStreams map[string]string `json:"debug_streams"`
}

func New(logger log.Logger) *Service {
	return &Service{
		debugStreams: make(map[string]func(string)),
	}
}

// Data implements service.Service. It returns nil, as the otel service does
// not have any runtime data.
func (s *Service) Data() any {
	return s
}

// Definition implements service.Service.
func (*Service) Definition() service.Definition {
	return service.Definition{
		Name:       ServiceName,
		ConfigType: nil, // xray does not accept configuration
		DependsOn:  []string{},
	}
}

// Run implements service.Service.
func (s *Service) Run(ctx context.Context, host service.Host) error {
	<-ctx.Done()
	return nil
}

// Update implements service.Service.
func (*Service) Update(newConfig any) error {
	return fmt.Errorf("xray service does not support configuration")
}

func (s *Service) GetDebugStream(id string) func(string) {
	s.loadMut.RLock()
	defer s.loadMut.RUnlock()
	return s.debugStreams[id]
}

func (s *Service) SetDebugStream(id string, callback func(string)) {
	s.loadMut.Lock()
	defer s.loadMut.Unlock()

	s.debugStreams[id] = callback
}

func (s *Service) DeleteDebugStream(id string) {
	s.loadMut.Lock()
	defer s.loadMut.Unlock()

	delete(s.debugStreams, id)
}
