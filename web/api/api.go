// Package api implements the HTTP API used for the Grafana Agent Flow UI.
//
// The API is internal only; it is not stable and shouldn't be relied on
// externally.
package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/flow"
)

// FlowAPI wraps several calls for the component health api.
type FlowAPI struct {
	flow *flow.Flow
}

// NewFlowAPI instantiates a new flow api.
func NewFlowAPI(flow *flow.Flow, r *mux.Router) *FlowAPI {
	return &FlowAPI{flow: flow}
}

// RegisterRoutes registers all the routes.
func (f *FlowAPI) RegisterRoutes(urlPrefix string, r *mux.Router) {
	r.HandleFunc(path.Join(urlPrefix, "/api/v0/web/components"), f.listComponentsHandler())
	r.HandleFunc(path.Join(urlPrefix, "/api/v0/web/components/{id}"), f.listComponentHandler())
}

func (f *FlowAPI) listComponentsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		infos := f.flow.ComponentInfos()
		bb, err := json.Marshal(infos)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(bb)
	}
}

func (f *FlowAPI) listComponentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		infos := f.flow.ComponentInfos()
		requestedComponent := vars["id"]

		for _, info := range infos {
			if requestedComponent == info.ID {
				bb, err := f.JSON(info)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				_, _ = w.Write(bb.Bytes())
				return
			}
		}

		http.NotFound(w, r)
	}
}

// JSON returns the json representation of ComponentInfoDetailed.
func (f *FlowAPI) JSON(c *flow.ComponentInfo) (bytes.Buffer, error) {
	var buf bytes.Buffer
	_, err := f.flow.ComponentJSON(&buf, c)
	if err != nil {
		return bytes.Buffer{}, err
	}
	return buf, nil
}

// Unless otherwise specified, API methods should be JSON.
//
// API methods needed:
//
// /api/v0/web/components
//
// Return list of components, where each component contains:
//   * component ID
//   * component name (metrics.remote_write)
//   * component label
//   * health info
//   * component IDs of components being referenced by this component
//   * component IDs of components referencing this component
//
// Arguments, exports, and debug info are *not* included.
//
// /api/v0/web/component/{id}
//
// Return details on a component:
//   * component name (metrics.remote_write)
//   * Arguments
//   * Exports
//   * Debug info
//   * Health info
//   * Dependencies
//   * Dependants
//
// /api/v0/web/component/{id}/raw
//
// Return raw evaluated River text for component
//
// /api/v0/web/status/build-info
//
//   Go runtime, build information (like the Prometheus page)
//
// /api/v0/web/status/flags
//
//   Command-line flags used to launch application
//
// /api/v0/web/status/config-file
//
//   Parsed config file (*not* evaluated config file)
