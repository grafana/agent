// Package api implements the HTTP API used for the Grafana Agent Flow UI.
//
// The API is internal only; it is not stable and shouldn't be relied on
// externally.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/flow"
)

type FlowApi struct {
	flow   *flow.Flow
	router *mux.Router
}

func NewFlowApi(flow *flow.Flow, r *mux.Router) *FlowApi {
	fa := &FlowApi{
		flow:   flow,
		router: r,
	}
	fa.SetupRoute()
	return fa
}

func (f *FlowApi) SetupRoute() {
	f.router.HandleFunc("/api/v0/web/components", f.ListComponentsHandler())
	f.router.HandleFunc("/api/v0/web/components/{id}", f.ListComponentHandler())
}

func (f *FlowApi) ListComponentsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		infos := f.flow.ComponentInfo()
		nonDetails := make([]*flow.ComponentInfo, len(infos))
		for i, ci := range infos {
			nonDetails[i] = &ci.ComponentInfo
		}
		bb, err := json.Marshal(nonDetails)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(bb)
	}
}

func (f *FlowApi) ListComponentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		infos := f.flow.ComponentInfo()
		requestedComponent := vars["id"]

		for _, info := range infos {
			if requestedComponent == info.ID {
				bb, err := f.flow.Json(info)
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
