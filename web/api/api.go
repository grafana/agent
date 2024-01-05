// Package api implements the HTTP API used for the Grafana Agent Flow UI.
//
// The API is internal only; it is not stable and shouldn't be relied on
// externally.
package api

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/service/cluster"
	"github.com/grafana/agent/service/xray"
	"github.com/prometheus/prometheus/util/httputil"
)

// FlowAPI is a wrapper around the component API.
type FlowAPI struct {
	flow    component.Provider
	cluster cluster.Cluster
	xray    *xray.Service
}

// NewFlowAPI instantiates a new Flow API.
func NewFlowAPI(flow component.Provider, cluster cluster.Cluster, xray *xray.Service) *FlowAPI {
	return &FlowAPI{flow: flow, cluster: cluster, xray: xray}
}

// RegisterRoutes registers all the API's routes.
func (f *FlowAPI) RegisterRoutes(urlPrefix string, r *mux.Router) {
	// NOTE(rfratto): {id:.+} is used in routes below to allow the
	// id to contain / characters, which is used by nested module IDs and
	// component IDs.

	r.Handle(path.Join(urlPrefix, "/modules/{moduleID:.+}/components"), httputil.CompressionHandler{Handler: f.listComponentsHandler()})
	r.Handle(path.Join(urlPrefix, "/components"), httputil.CompressionHandler{Handler: f.listComponentsHandler()})
	r.Handle(path.Join(urlPrefix, "/components/{id:.+}"), httputil.CompressionHandler{Handler: f.getComponentHandler()})
	r.Handle(path.Join(urlPrefix, "/peers"), httputil.CompressionHandler{Handler: f.getClusteringPeersHandler()})
	r.Handle(path.Join(urlPrefix, "/debugStream/{id:.+}"), f.startDebugStream())
}

func (f *FlowAPI) listComponentsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// moduleID is set from the /modules/{moduleID:.+}/components route above
		// but not from the /components route.
		var moduleID string
		if vars := mux.Vars(r); vars != nil {
			moduleID = vars["moduleID"]
		}

		components, err := f.flow.ListComponents(moduleID, component.InfoOptions{
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
		requestedComponent := component.ParseID(vars["id"])

		component, err := f.flow.GetComponent(requestedComponent, component.InfoOptions{
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
		peers := f.cluster.Peers()
		bb, err := json.Marshal(peers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(bb)
	}
}

func (f *FlowAPI) startDebugStream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		componentID := vars["id"]

		dataCh := make(chan string, 1000) // Define the size of the channel, is 1000 ok?
		ctx := r.Context()

		sampleProbParam := r.URL.Query().Get("sampleProb")
		sampleProb := 1.0
		if sampleProbParam != "" {
			var err error
			sampleProb, err = strconv.ParseFloat(sampleProbParam, 64)
			if err != nil || sampleProb < 0 || sampleProb > 1 {
				http.Error(w, "Invalid sample probability", http.StatusBadRequest)
				return
			}
		}

		f.xray.SetDebugStream(componentID, func(computeDataFunc func() string) {
			select {
			case <-ctx.Done():
				return
			default:
				if sampleProb < 1 && rand.Float64() > sampleProb {
					return
				}
				// Avoid blocking the channel when the channel is full
				select {
				case dataCh <- computeDataFunc():
				default:
				}
			}
		})
		stopStreaming := func() {
			close(dataCh)
			f.xray.DeleteDebugStream(componentID)
		}

		for {
			select {
			case data := <-dataCh:
				_, writeErr := w.Write([]byte(data + "|xray|"))
				if writeErr != nil {
					stopStreaming()
					return
				}
				w.(http.Flusher).Flush()
			case <-ctx.Done():
				stopStreaming()
				return
			}
		}
	}
}
