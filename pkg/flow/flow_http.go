package flow

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/river/encoding"
	"github.com/grafana/agent/web/api/apitypes"
)

// ComponentHandler returns an http.HandlerFunc which will delegate all requests to
// a component named by the first path segment
func (f *Flow) ComponentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		// find node with ID
		var node *controller.ComponentNode
		for _, n := range f.loader.Components() {
			if n.ID().String() == id {
				node = n
				break
			}
		}
		if node == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// TODO: potentially cache these handlers, and invalidate on component state change.
		handler := node.HTTPHandler()
		if handler == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// Remove prefix from path, so each component can handle paths from their
		// own root path.
		r.URL.Path = strings.TrimPrefix(r.URL.Path, path.Join(f.opts.HTTPPathPrefix, id))
		handler.ServeHTTP(w, r)
	}
}

// ComponentJSON returns the json representation of the flow component.
func (f *Flow) ComponentJSON(w io.Writer, ci *apitypes.ComponentInfo) error {
	f.loadMut.RLock()
	defer f.loadMut.RUnlock()

	var foundComponent *controller.ComponentNode
	for _, c := range f.loader.Components() {
		if c.ID().String() == ci.ID {
			foundComponent = c
			break
		}
	}
	if foundComponent == nil {
		return fmt.Errorf("unable to find component named %q", ci.ID)
	}

	var err error
	args, err := encoding.ConvertRiverBodyToJSON(foundComponent.Arguments())
	if err != nil {
		return err
	}
	ci.Arguments = args

	exports, err := encoding.ConvertRiverBodyToJSON(foundComponent.Exports())
	if err != nil {
		return err
	}
	ci.Exports = exports

	debugInfo, err := encoding.ConvertRiverBodyToJSON(foundComponent.DebugInfo())
	if err != nil {
		return err
	}
	ci.DebugInfo = debugInfo

	bb, err := json.Marshal(ci)
	if err != nil {
		return err
	}
	_, err = w.Write(bb)
	return err
}
