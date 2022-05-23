package flow

import (
	"bytes"
	"io"
	"net/http"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/flow/internal/graphviz"
)

// GraphHandler returns an http.HandlerFunc which renders the current graph's
// DAG as an SVG. Graphviz must be installed for this function to work.
func (f *Flow) GraphHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		g := f.loader.Graph()
		dot := dag.MarshalDOT(g)

		svgBytes, err := graphviz.Dot(dot, "svg")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(w, bytes.NewReader(svgBytes))
		if err != nil {
			level.Error(f.log).Log("msg", "failed to write svg graph", "err", err)
		}
	}
}
