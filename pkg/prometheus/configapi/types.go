package configapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ContentType holds the content type of data that is either
// packaged with a request or the content type of data that a
// response should be formatted as.
type ContentType string

// Supported ContentType values
const (
	// From https://www.freeformatter.com/mime-types-list.html, even though
	// it's not an official standard, it's standard enough for our needs.

	ContentTypeUnknown ContentType = ""
	ContentTypeJSON    ContentType = "application/json"
	ContentTypeYAML    ContentType = "text/yaml"

	// DefaultContentType is the default content type to use if none is specified.
	DefaultContentType = ContentTypeYAML
)

// SupportedContentTypes is the full list of allowed Content-Type
// values.
var SupportedContentTypes = []ContentType{ContentTypeJSON, ContentTypeYAML}

// ContentTypeFromRequest returns the content type given an http Request,
// pulling it out of the header.
func ContentTypeFromRequest(r *http.Request) (ContentType, error) {
	v := r.Header.Get("Content-Type")
	v = strings.ToLower(v)

	if v == "" {
		return DefaultContentType, nil
	}

	for _, ty := range SupportedContentTypes {
		if string(ty) == v {
			return ty, nil
		}
	}

	supported := make([]string, 0, len(SupportedContentTypes))
	for _, ty := range SupportedContentTypes {
		supported = append(supported, string(ty))
	}
	supportedStr := strings.Join(supported, ", ")
	err := fmt.Errorf("unsupported Content-Type %s. Must be one of %s", v, supportedStr)
	return ContentTypeUnknown, err
}

// APIResponse is the base object returned for any API call.
// The Data field will be set to either nil or a value of
// another *Response type value from this package.
type APIResponse struct {
	Status string      `json:"status" yaml:"status"`
	Data   interface{} `json:"data,omitempty" yaml:"data,omitempty"`
}

func (r *APIResponse) WriteTo(w http.ResponseWriter, statusCode int) error {
	bb, err := json.Marshal(r)
	if err != nil {
		// If we fail here, we should at least write a 500 back.
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	w.WriteHeader(statusCode)
	n, err := w.Write(bb)
	if err != nil {
		return err
	} else if n != len(bb) {
		return fmt.Errorf("could not write full response. expected %d, wrote %d", len(bb), n)
	}

	return nil
}

// ErrorResponse is contained inside an APIResponse and returns
// an error string. Returned by any API call that can fail.
type ErrorResponse struct {
	Error string `yaml:"error"`
}

// ListConfigurationsResponse is contained inside an APIResponse
// and provides the list of configurations known to the KV store.
// Returned by ListConfigurations.
type ListConfigurationsResponse struct {
	// Configs is the list of configuration names.
	Configs []string `json:"configs" yaml:"configs"`
}

// GetConfigurationResponse is contained inside an APIResponse
// and provides a single configuration known to the KV store.
// Returned by GetConfiguration.
type GetConfigurationResponse struct {
	// Value is the stringified configuration. Depending on the
	// Content-Type in the request, the Value will either be
	// stringified json or stringified yaml.
	Value string `json:"value" yaml:"value"`
}

func WriteResponse(w http.ResponseWriter, statusCode int, resp interface{}) error {
	apiResp := &APIResponse{Status: "success", Data: resp}
	return apiResp.WriteTo(w, statusCode)
}

// WriteError writes an error response back to the ResponseWriter.
func WriteError(w http.ResponseWriter, statusCode int, err error) error {
	resp := &APIResponse{Status: "error", Data: &ErrorResponse{Error: err.Error()}}
	return resp.WriteTo(w, statusCode)
}
