package debugdial

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/grafana/agent-remote-config/api/gen/proto/go/agent/v1/agentv1connect"
	"github.com/grafana/agent/service"
	agenthttp "github.com/grafana/agent/service/http"
)

type Service struct {
	opts Options
	args Arguments

	ctrl service.Controller

	mut               sync.RWMutex
	asClient          agentv1connect.AgentServiceClient
	dataPath          string
	currentConfigHash string
}

type Arguments struct {
	Config map[string]string
}

type Options struct{}

const ServiceName = "debugdial"

// Definition returns the Definition of the Service.
// Definition must always return the same value across all
// calls.
func (s *Service) Definition() service.Definition {
	return service.Definition{
		Name:       ServiceName,
		ConfigType: nil,
		DependsOn:  []string{agenthttp.ServiceName},
	}
}

// Run starts a Service. Run must block until the provided
// context is canceled. Returning an error should be treated
// as a fatal error for the Service.
func (s *Service) Run(ctx context.Context, host service.Host) error {
	fmt.Println("DebugDial initiated")
	<-ctx.Done()
	return nil
}

// Update updates a Service at runtime. Update is never
// called if [Definition.ConfigType] is nil. newConfig will
// be the same type as ConfigType; if ConfigType is a
// pointer to a type, newConfig will be a pointer to the
// same type.
//
// Update will be called once before Run, and may be called
// while Run is active.
func (s *Service) Update(newConfig any) error {
	return errors.New("not implemented")
}

// Data returns the Data associated with a Service. Data must always return
// the same value across multiple calls, as callers are expected to be able
// to cache the result.
//
// The return result of Data must not rely on the runtime config of the
// service.
//
// Data may be invoked before Run.
func (s *Service) Data() any {
	return nil
}

// ServiceHandler returns the base route and HTTP handlers to register for
// the provided service.
//
// This method is only called for services that declare a dependency on
// the http service.
//
// The http service prioritizes longer base routes. Given two base routes of
// /foo and /foo/bar, an HTTP URL of /foo/bar/baz will be routed to the
// longer base route (/foo/bar).
func (s *Service) ServiceHandler(host service.Host) (base string, handler http.Handler) {
	r := mux.NewRouter()

	r.Handle("/api/v0/debugdial", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	}))

	return "", r
}
