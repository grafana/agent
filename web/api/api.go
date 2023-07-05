// Package api implements the HTTP API used for the Grafana Agent Flow UI.
//
// The API is internal only; it is not stable and shouldn't be relied on
// externally.
package api

import (
	"encoding/json"
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/cluster"
	"github.com/prometheus/prometheus/util/httputil"
)

// FlowAPI is a wrapper around the component API.
type FlowAPI struct {
	flow component.Provider
	node cluster.Node
}

// NewFlowAPI instantiates a new Flow API.
func NewFlowAPI(flow component.Provider, node cluster.Node) *FlowAPI {
	return &FlowAPI{flow: flow, node: node}
}

// RegisterRoutes registers all the API's routes.
func (f *FlowAPI) RegisterRoutes(urlPrefix string, r *mux.Router) {
	r.Handle(path.Join(urlPrefix, "/components"), httputil.CompressionHandler{Handler: f.listComponentsHandler()})
	r.Handle(path.Join(urlPrefix, "/components/{id}"), httputil.CompressionHandler{Handler: f.getComponentHandler()})
	r.Handle(path.Join(urlPrefix, "/peers"), httputil.CompressionHandler{Handler: f.getClusteringPeersHandler()})
}

func (f *FlowAPI) listComponentsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		components, err := f.flow.ListComponents("", component.InfoOptions{
			GetHealth: true,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		bb, err := json.Marshal(components)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(bb)
	}
}

func (f *FlowAPI) getComponentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		requestedComponent := vars["id"]

		component, err := f.flow.GetComponent(component.ID{
			ModuleID: "", // TODO(rfratto): support getting component from module.
			LocalID:  requestedComponent,
		}, component.InfoOptions{
			GetHealth:    true,
			GetArguments: true,
			GetExports:   true,
			GetDebugInfo: true,
		})
		if err != nil {
			http.NotFound(w, r)
			return
		}

		bb, err := json.Marshal(component)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(bb)
	}
}

func (f *FlowAPI) getClusteringPeersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// TODO(@tpaschalis) Detect if clustering is disabled and propagate to
		// the Typescript code (eg. via the returned status code?).
		peers := f.node.Peers()
		bb, err := json.Marshal(peers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(bb)
	}
}
