// Package api implements the HTTP API used for the Grafana Agent Flow UI.
//
// The API is internal only; it is not stable and shouldn't be relied on
// externally.
package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/web/api/types"
	"github.com/prometheus/prometheus/util/httputil"
)

// ComponentProvider provides the ability to introspect a flow components
// in a read only manner.
type ComponentProvider interface {
	ComponentInfos() []*types.ComponentInfo
	ComponentJSON(io.Writer, *types.ComponentInfo) error
}

// FlowAPI is a wrapper around the component API.
type FlowAPI struct {
	flow ComponentProvider
}

// NewFlowAPI instantiates a new Flow API.
func NewFlowAPI(flow ComponentProvider) *FlowAPI {
	return &FlowAPI{flow: flow}
}

// RegisterRoutes registers all the API's routes.
func (f *FlowAPI) RegisterRoutes(urlPrefix string, r *mux.Router) {
	r.Handle(path.Join(urlPrefix, "/components"), httputil.CompressionHandler{Handler: f.listComponentsHandler()})
	r.Handle(path.Join(urlPrefix, "/components/{id}"), httputil.CompressionHandler{Handler: f.listComponentHandler()})
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
func (f *FlowAPI) json(c *types.ComponentInfo) ([]byte, error) {
	var buf bytes.Buffer
	err := f.flow.ComponentJSON(&buf, c)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
