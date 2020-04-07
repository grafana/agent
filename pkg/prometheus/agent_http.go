package prometheus

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/prometheus/configapi"
	"gopkg.in/yaml.v2"
)

// APIHandler is a function that returns a configapi Response type
// and optionally an error.
type APIHandler func(r *http.Request) (interface{}, error)

// WrapHandler is responsible for turning an APIHandler into an HTTP
// handler by managing responses that are written and the
// content type to write them as.
//
// WrapHandler first reads the Content-Type header and validates
// it as one of the expected values. If Content-Type is not present,
// the type used defaults to JSON. WrapHandler then invokes the
// APIHandler specified by next and formats the response based
// on whether the internal handler failed or gave a valid response.
func (a *Agent) WrapHandler(next APIHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctype, err := configapi.ContentTypeFromRequest(r)
		if err != nil {
			level.Warn(a.logger).Log("msg", "failed to parse content-type", "err", err)
			configapi.WriteError(w, http.StatusBadRequest, configapi.DefaultContentType, err)
			return
		}

		resp, err := next(r)
		if err != nil {
			httpErr, ok := err.(*httpError)
			if ok {
				configapi.WriteError(w, httpErr.StatusCode, ctype, httpErr.Err)
			} else {
				configapi.WriteError(w, http.StatusInternalServerError, ctype, err)
			}
			return
		}

		if err := configapi.WriteResponse(w, http.StatusOK, ctype, resp); err != nil {
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
	ctype, err := configapi.ContentTypeFromRequest(r)
	if err != nil {
		return nil, err
	}

	configKey := getConfigName(r)
	v, err := a.kv.Get(r.Context(), configKey)
	if err != nil {
		level.Error(a.logger).Log("msg", "error getting configuration from kv store", "err", err)
		return nil, fmt.Errorf("error getting configuration: %w", err)
	}

	cfg, err := MarshalInstanceConfig(v.(*InstanceConfig), ctype)
	if err != nil {
		level.Error(a.logger).Log("msg", "error marshaling configuration", "err", err)
		return nil, err
	}

	return &configapi.GetConfigurationResponse{Value: cfg}, nil
}

// PutConfiguration creates or updates a named configuration. Completely
// overrides the previous configuration if it exists.
func (a *Agent) PutConfiguration(r *http.Request) (interface{}, error) {
	ctype, err := configapi.ContentTypeFromRequest(r)
	if err != nil {
		return nil, err
	}

	inst, err := UnmarshalInstanceConfig(r.Body, ctype)
	if err != nil {
		return nil, err
	}
	inst.Name = getConfigName(r)

	err = a.kv.CAS(r.Context(), inst.Name, func(in interface{}) (out interface{}, retry bool, err error) {
		return inst, false, nil
	})
	return nil, err
}

// DeleteConfiguration deletes an existing named configuration.
func (a *Agent) DeleteConfiguration(r *http.Request) (interface{}, error) {
	configKey := getConfigName(r)
	ok, err := a.kv.Delete(r.Context(), configKey)
	if err != nil {
		level.Error(a.logger).Log("msg", "error deleting configuration from kv store", "err", err)
		return nil, fmt.Errorf("error deleting configuration: %w", err)
	}

	if !ok {
		err = &httpError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("configuration %s does not exist", configKey),
		}
	}
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
	name, ok := vars["name"]
	if !ok {
		panic("name not found in mux vars - the mux route is misconfigured")
	}
	return name
}

// UnmarshalInstanceConfig unmarshals an instance config from a reader
// based on a provided content type.
func UnmarshalInstanceConfig(r io.Reader, ctype configapi.ContentType) (*InstanceConfig, error) {
	var (
		cfg InstanceConfig
		err error
	)

	switch ctype {
	case configapi.ContentTypeJSON:
		err = json.NewDecoder(r).Decode(&cfg)
	case configapi.ContentTypeYAML:
		err = yaml.NewDecoder(r).Decode(&cfg)
	default:
		panic(fmt.Sprintf("unhandled content type %s", ctype))
	}

	return &cfg, err
}

// MarshalInstanceConfig marshals an instance config based on a provided
// content type.
func MarshalInstanceConfig(c *InstanceConfig, ctype configapi.ContentType) (string, error) {
	var (
		bb  []byte
		err error
	)

	switch ctype {
	case configapi.ContentTypeJSON:
		bb, err = json.Marshal(c)
	case configapi.ContentTypeYAML:
		bb, err = yaml.Marshal(c)
	default:
		panic(fmt.Sprintf("unhandled content type %s", ctype))
	}

	if err != nil {
		return "", err
	}
	return string(bb), err
}
