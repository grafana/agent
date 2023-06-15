package flow

import (
	"net/http"
	"path"
	"strings"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/flow/internal/controller"
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
