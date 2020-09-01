package ha

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/prom/ha/configapi"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	totalCreatedConfigs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "agent_prometheus_ha_configs_created_total",
		Help: "Total number of created scraping service configs",
	})
	totalUpdatedConfigs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "agent_prometheus_ha_configs_updated_total",
		Help: "Total number of updated scraping service configs",
	})
	totalDeletedConfigs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "agent_prometheus_ha_configs_deleted_total",
		Help: "Total number of deleted scraping service configs",
	})
)

// APIHandler is a function that returns a configapi Response type
// and optionally an error.
type APIHandler func(r *http.Request) (interface{}, error)

// WireAPI injects routes into the provided mux router for the config
// management API.
func (s *Server) WireAPI(r *mux.Router) {
	// PutConfiguration below is wrapped in a nonConcurrentHandler to prevent a
	// race condition of two conflicting instance configs being applied at the same time.
	// Since applying config performs KV store-wide validations, it is possible that
	// concurrently applied configs that conflict each other may not be rejected if
	// neither conflicting config has been persisted to the KV store yet.

	listConfig := s.wrapHandler(s.ListConfigurations)
	getConfig := s.wrapHandler(s.GetConfiguration)
	putConfig := s.wrapHandler(nonConcurrentHandler(s.PutConfiguration))
	deleteConfig := s.wrapHandler(s.DeleteConfiguration)

	r.HandleFunc("/agent/api/v1/configs", listConfig).Methods("GET")
	r.HandleFunc("/agent/api/v1/configs/{name}", getConfig).Methods("GET")
	r.HandleFunc("/agent/api/v1/config/{name}", putConfig).Methods("PUT", "POST")
	r.HandleFunc("/agent/api/v1/config/{name}", deleteConfig).Methods("DELETE")

	// Debug ring page
	r.Handle("/debug/ring", s.ring)
}

// nonConcurrentHandler wraps an APIHandler in a mutex to prevent it from being
// called concurrently.
func nonConcurrentHandler(next APIHandler) APIHandler {
	var mut sync.Mutex
	return func(r *http.Request) (interface{}, error) {
		mut.Lock()
		defer mut.Unlock()
		return next(r)
	}
}

// wrapHandler is responsible for turning an APIHandler into an HTTP
// handler by wrapping responses and writing them as JSON.
func (s *Server) wrapHandler(next APIHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := next(r)
		if err != nil {
			httpErr, ok := err.(*httpError)

			if ok {
				err = configapi.WriteError(w, httpErr.StatusCode, httpErr.Err)
			} else {
				err = configapi.WriteError(w, http.StatusInternalServerError, err)
			}

			if err != nil {
				level.Error(s.logger).Log("msg", "failed writing error response to client", "err", err)
			}
			return
		}

		// Prepare data and status code to send back to the writer: if the handler
		// returned an *httpResponse, use the status code defined there and send the
		// internal data. Otherwise, assume HTTP 200 OK and marshal the raw response.
		var (
			data       = resp
			statusCode = http.StatusOK
		)
		if httpResp, ok := data.(*httpResponse); ok {
			data = httpResp.Data
			statusCode = httpResp.StatusCode
		}

		if err := configapi.WriteResponse(w, statusCode, data); err != nil {
			level.Error(s.logger).Log("msg", "failed to write valid response", "err", err)
		}
	})
}

// ListConfigurations returns a list of the named configurations or all
// configurations associated with the Prometheus agent.
func (s *Server) ListConfigurations(r *http.Request) (interface{}, error) {
	vv, err := s.kv.List(r.Context(), "")
	if err != nil {
		level.Error(s.logger).Log("msg", "failed to list keys in kv store", "err", err)
		return nil, err
	}

	return &configapi.ListConfigurationsResponse{Configs: vv}, nil
}

// GetConfiguration returns an existing named configuration.
func (s *Server) GetConfiguration(r *http.Request) (interface{}, error) {
	configKey := getConfigName(r)
	v, err := s.kv.Get(r.Context(), configKey)
	if err != nil {
		level.Error(s.logger).Log("msg", "error getting configuration from kv store", "err", err)
		return nil, fmt.Errorf("error getting configuration: %w", err)
	} else if v == nil {
		return nil, &httpError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("configuration %s does not exist", configKey),
		}
	}

	cfg, err := instance.MarshalConfig(v.(*instance.Config), true)
	if err != nil {
		level.Error(s.logger).Log("msg", "error marshaling configuration", "err", err)
		return nil, err
	}

	return &configapi.GetConfigurationResponse{Value: string(cfg)}, nil
}

// PutConfiguration creates or updates a named configuration. Completely
// overrides the previous configuration if it exists.
func (s *Server) PutConfiguration(r *http.Request) (interface{}, error) {
	inst, err := instance.UnmarshalConfig(r.Body)
	if err != nil {
		return nil, err
	}
	inst.Name = getConfigName(r)

	// Validate the incoming config
	if err := inst.ApplyDefaults(s.globalConfig); err != nil {
		return nil, err
	}

	// Validate that the job names from the incoming config are unique
	if err := s.checkUnique(r.Context(), inst); err != nil {
		return nil, err
	}

	var newConfig bool
	err = s.kv.CAS(r.Context(), inst.Name, func(in interface{}) (out interface{}, retry bool, err error) {
		// The configuration is new if there's no previous value from the CAS
		newConfig = (in == nil)
		return inst, false, nil
	})

	if err != nil {
		level.Error(s.logger).Log("msg", "failed to put config", "err", err)
		return nil, err
	}

	if newConfig {
		totalCreatedConfigs.Inc()
		return &httpResponse{StatusCode: http.StatusCreated}, nil
	}

	totalUpdatedConfigs.Inc()
	return &httpResponse{StatusCode: http.StatusOK}, nil
}

// checkUnique looks at all the existing configs and ensures that no other
// config shares a job_name with the incoming config.
func (s *Server) checkUnique(ctx context.Context, cfg *instance.Config) error {
	cfgCh, err := s.AllConfigs(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// Make sure we drain the channel. This will need to be done if we are
		// returning an error.
		for range cfgCh {
		}
	}()

	newJobNames := make(map[string]struct{}, len(cfg.ScrapeConfigs))
	for _, sc := range cfg.ScrapeConfigs {
		newJobNames[sc.JobName] = struct{}{}
	}

	for otherConfig := range cfgCh {
		// Skip over the config if it's the same one we're about to apply.
		if otherConfig.Name == cfg.Name {
			continue
		}

		for _, otherScrape := range otherConfig.ScrapeConfigs {
			if _, exist := newJobNames[otherScrape.JobName]; exist {
				return &httpError{
					StatusCode: http.StatusBadRequest,
					Err:        fmt.Errorf("found multiple scrape configs with job name %q", otherScrape.JobName),
				}
			}
		}
	}

	return nil
}

// DeleteConfiguration deletes an existing named configuration.
func (s *Server) DeleteConfiguration(r *http.Request) (interface{}, error) {
	configKey := getConfigName(r)

	v, err := s.kv.Get(r.Context(), configKey)
	if err != nil {
		// Silently ignore the error; Get is just used for validation since Delete
		// will never return an error if the key doesn't exist. We'll log it anyway
		// but the user will be left unaware of there being a problem here.
		level.Error(s.logger).Log("msg", "error validating key existence for deletion", "err", err)
	} else if v == nil {
		// But if the object doesn't exist, there's nothing to delete.
		return nil, &httpError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("configuration %s does not exist", configKey),
		}
	}

	err = s.kv.Delete(r.Context(), configKey)
	if err != nil {
		level.Error(s.logger).Log("msg", "error deleting configuration from kv store", "err", err)
		return nil, fmt.Errorf("error deleting configuration: %w", err)
	}

	totalDeletedConfigs.Inc()
	return nil, err
}

type httpError struct {
	StatusCode int
	Err        error
}

func (e httpError) Error() string { return e.Err.Error() }

type httpResponse struct {
	StatusCode int
	Data       interface{}
}

// getConfigName uses gorilla/mux's route variables to extract the
// "name" variable. If not found, getConfigName will panic.
func getConfigName(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["name"]
}
