package prometheus

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/prometheus/configapi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gopkg.in/yaml.v2"
)

var (
	totalCreatedConfigs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "agent_prometheus_scraping_service_configs_created_total",
		Help: "Total number of created scraping service configs",
	})
	totalUpdatedConfigs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "agent_prometheus_scraping_service_configs_updated_total",
		Help: "Total number of updated scraping service configs",
	})
	totalDeletedConfigs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "agent_prometheus_scraping_service_configs_deleted_total",
		Help: "Total number of deleted scraping service configs",
	})
)

// APIHandler is a function that returns a configapi Response type
// and optionally an error.
type APIHandler func(r *http.Request) (interface{}, error)

// WireAPI injects routes into the provided mux router for the config
// management API.
func (a *Agent) WireAPI(r *mux.Router) {
	if !a.cfg.ServiceConfig.Enabled {
		return
	}

	listConfig := a.WrapHandler(a.ListConfigurations)
	getConfig := a.WrapHandler(a.GetConfiguration)
	putConfig := a.WrapHandler(a.PutConfiguration)
	deleteConfig := a.WrapHandler(a.DeleteConfiguration)

	r.HandleFunc("/agent/api/v1/configs", listConfig).Methods("GET")
	r.HandleFunc("/agent/api/v1/configs/{name}", getConfig).Methods("GET")
	r.HandleFunc("/agent/api/v1/config/{name}", putConfig).Methods("PUT", "POST")
	r.HandleFunc("/agent/api/v1/config/{name}", deleteConfig).Methods("DELETE")
}

// WrapHandler is responsible for turning an APIHandler into an HTTP
// handler by wrapping responses and writing them as JSON.
func (a *Agent) WrapHandler(next APIHandler) http.HandlerFunc {
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
				level.Error(a.logger).Log("msg", "failed writing error response to client", "err", err)
			}
			return
		}

		if err := configapi.WriteResponse(w, http.StatusOK, resp); err != nil {
			level.Error(a.logger).Log("msg", "failed to write valid response", "err", err)
		}
	})
}

// ListConfigurations returns a list of the named configurations or all
// configurations associated with the Prometheus agent.
func (a *Agent) ListConfigurations(r *http.Request) (interface{}, error) {
	vv, err := a.kv.List(r.Context(), "")
	if err != nil {
		level.Error(a.logger).Log("msg", "failed to list keys in kv store", "err", err)
		return nil, err
	}

	return &configapi.ListConfigurationsResponse{Configs: vv}, nil
}

// GetConfiguration returns an existing named configuration.
func (a *Agent) GetConfiguration(r *http.Request) (interface{}, error) {
	configKey := getConfigName(r)
	v, err := a.kv.Get(r.Context(), configKey)
	if err != nil {
		level.Error(a.logger).Log("msg", "error getting configuration from kv store", "err", err)
		return nil, fmt.Errorf("error getting configuration: %w", err)
	} else if v == nil {
		return nil, &httpError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("configuration %s does not exist", configKey),
		}
	}

	cfg, err := MarshalInstanceConfig(v.(*InstanceConfig))
	if err != nil {
		level.Error(a.logger).Log("msg", "error marshaling configuration", "err", err)
		return nil, err
	}

	return &configapi.GetConfigurationResponse{Value: cfg}, nil
}

// PutConfiguration creates or updates a named configuration. Completely
// overrides the previous configuration if it exists.
func (a *Agent) PutConfiguration(r *http.Request) (interface{}, error) {
	inst, err := UnmarshalInstanceConfig(r.Body)
	if err != nil {
		return nil, err
	}
	inst.Name = getConfigName(r)

	var newConfig bool
	err = a.kv.CAS(r.Context(), inst.Name, func(in interface{}) (out interface{}, retry bool, err error) {
		// The configuration is new if there's no previous value from the CAS
		newConfig = (in == nil)
		return inst, false, nil
	})

	if err == nil {
		if newConfig {
			totalCreatedConfigs.Inc()
		} else {
			totalUpdatedConfigs.Inc()
		}
	} else {
		level.Error(a.logger).Log("msg", "failed to put config", "err", err)
	}

	return nil, err
}

// DeleteConfiguration deletes an existing named configuration.
func (a *Agent) DeleteConfiguration(r *http.Request) (interface{}, error) {
	configKey := getConfigName(r)

	v, err := a.kv.Get(r.Context(), configKey)
	if err != nil {
		// Silently ignore the error; Get is just used for validation since Delete
		// will never return an error if the key doesn't exist. We'll log it anyway
		// but the user will be left unaware of there being a problem here.
		level.Error(a.logger).Log("msg", "error validating key existence for deletion", "err", err)
	} else if v == nil {
		// But if the object doesn't exist, there's nothing to delete.
		return nil, &httpError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("configuration %s does not exist", configKey),
		}
	}

	err = a.kv.Delete(r.Context(), configKey)
	if err != nil {
		level.Error(a.logger).Log("msg", "error deleting configuration from kv store", "err", err)
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

// getConfigName uses gorilla/mux's route variables to extract the
// "name" variable. If not found, getConfigName will panic.
func getConfigName(r *http.Request) string {
	vars := mux.Vars(r)
	name, _ := vars["name"]
	return name
}

// UnmarshalInstanceConfig unmarshals an instance config from a reader
// based on a provided content type.
func UnmarshalInstanceConfig(r io.Reader) (*InstanceConfig, error) {
	var cfg InstanceConfig
	err := yaml.NewDecoder(r).Decode(&cfg)
	return &cfg, err
}

// MarshalInstanceConfig marshals an instance config based on a provided
// content type.
func MarshalInstanceConfig(c *InstanceConfig) (string, error) {
	bb, err := yaml.Marshal(c)
	return string(bb), err
}
