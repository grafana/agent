package debugdial

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/grafana/agent-remote-config/api/gen/proto/go/agent/v1/agentv1connect"
	"github.com/grafana/agent/internal/service"
	agenthttp "github.com/grafana/agent/internal/service/http"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"

	opamp "github.com/open-telemetry/opamp-go/protobufs"
)

type Service struct {
	opts Options

	ctrl service.Controller

	mut               sync.RWMutex
	asClient          agentv1connect.AgentServiceClient
	dataPath          string
	currentConfigHash string

	serviceMetrics *prometheus.Desc

	Config   sync.Map
	Services sync.Map
}

type Options struct{}

const ServiceName = "debugdial"

func New(r prometheus.Registerer) *Service {
	s := &Service{
		Config:         sync.Map{},
		Services:       sync.Map{},
		serviceMetrics: prometheus.NewDesc("agent_debug_dial_services", "A metric identifying debug dial services and their attributes", []string{"service_name"}, nil),
	}
	_ = r.Register(s)
	return s
}

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
	return &s.Config
}

func (s *Service) Describe(m chan<- *prometheus.Desc) {
	m <- s.serviceMetrics
}

func (s *Service) Collect(m chan<- prometheus.Metric) {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.Services.Range(func(key, value any) bool {
		if time.Since(value.(time.Time)) <= 1*time.Minute {
			m <- prometheus.MustNewConstMetric(s.serviceMetrics, prometheus.GaugeValue, 1, key.(string))
		} else {
			s.Services.Delete(key)
		}
		return true
	})
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

	prefix := "/api/v0/debugdial"
	r.Handle(prefix, http.HandlerFunc(s.serveOpAMP))
	r.Handle(prefix+"/{id:.*}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		serviceID := vars["id"]
		if serviceID == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("No service specified in URL"))
			return
		}

		s.Services.Store(serviceID, time.Now())

		res, _ := s.Config.Load(serviceID)

		if res != nil {
			w.Write([]byte(res.(string)))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Service not found"))
		}
	}))

	return prefix, r
}

func opAmpError(s string, badRequest bool) *opamp.ServerToAgent {
	typ := opamp.ServerErrorResponseType_ServerErrorResponseType_Unknown
	if badRequest {
		typ = opamp.ServerErrorResponseType_ServerErrorResponseType_BadRequest
	}

	return &opamp.ServerToAgent{
		ErrorResponse: &opamp.ServerErrorResponse{
			ErrorMessage: s,
			Type:         typ,
		},
	}
}

func writeOpAmp(w http.ResponseWriter, msg *opamp.ServerToAgent) {
	responseBytes, err := proto.Marshal(msg)
	if err != nil {
		fmt.Fprintf(w, "Marshalling error: %q", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(responseBytes)
	if msg.ErrorResponse == nil {
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Service) serveOpAMP(w http.ResponseWriter, r *http.Request) {
	// if no protobuf cancel
	if r.Header.Get("content-type") != "application/x-protobuf" {
		http.Error(w, "Only \"application/x-protobuf\" content-type supported", http.StatusBadRequest)
		return
	}

	rawBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body := &opamp.AgentToServer{}

	err = proto.Unmarshal(rawBytes, body)
	if err != nil {
		writeOpAmp(w, opAmpError(err.Error(), true))
		return
	}

	serviceName := ""

	for _, att := range body.AgentDescription.GetIdentifyingAttributes() {
		if att.Key != "service.name" {
			continue
		}
		serviceName = att.Value.GetStringValue()
	}

	if serviceName == "" {
		writeOpAmp(w, opAmpError("No identifying service.name attribute specified", true))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.Services.Store(serviceName, time.Now())

	cfg, exists := s.Config.Load(serviceName)
	if !exists {
		writeOpAmp(w, opAmpError(fmt.Sprintf("No service %q found", serviceName), true))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := opamp.ServerToAgent{
		RemoteConfig: &opamp.AgentRemoteConfig{
			Config: &opamp.AgentConfigMap{
				ConfigMap: map[string]*opamp.AgentConfigFile{
					serviceName: {
						Body: []byte(cfg.(string)),
					},
				},
			},
		},
	}

	writeOpAmp(w, &response)
}
