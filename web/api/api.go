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

	"github.com/prometheus/prometheus/util/httputil"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/flow"
)

// FlowAPI is a wrapper around the component API.
type FlowAPI struct {
	flow *flow.Flow
}

// NewFlowAPI instantiates a new Flow API.
func NewFlowAPI(flow *flow.Flow, r *mux.Router) *FlowAPI {
	return &FlowAPI{flow: flow}
}

// RegisterRoutes registers all the API's routes.
func (f *FlowAPI) RegisterRoutes(urlPrefix string, r *mux.Router) {
	r.Handle(path.Join(urlPrefix, "/components"), httputil.CompressionHandler{Handler: f.listComponentsHandler()})
	r.Handle(path.Join(urlPrefix, "/components/{id}"), httputil.CompressionHandler{Handler: f.getComponentHandler()})
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

func (f *FlowAPI) getComponentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		infos := f.flow.ComponentInfos()
		requestedComponent := vars["id"]

		for _, info := range infos {
			if requestedComponent == info.ID {
				bb, err := f.json(info)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				_, _ = w.Write(bb)
				return
			}
		}

		http.NotFound(w, r)
	}
}

// json returns the JSON representation of c.
func (f *FlowAPI) json(c *flow.ComponentInfo) ([]byte, error) {
	var buf bytes.Buffer
	err := f.flow.ComponentJSON(&buf, c)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
