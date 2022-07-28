package flow

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/flow/internal/graphviz"
	"github.com/grafana/agent/pkg/river/token"
	"github.com/grafana/agent/pkg/river/token/builder"
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

// ConfigHandler returns an http.HandlerFunc which will render the most
// recently loaded configuration file as River.
func (f *Flow) ConfigHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		debugInfo := r.URL.Query().Get("debug") == "1"

		var buf bytes.Buffer
		_, err := f.configBytes(&buf, debugInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			_, _ = io.Copy(w, &buf)
		}
	}
}

// ConfigHandler returns an http.HandlerFunc which will render the scope used
// for variable references throughout River expressions.
func (f *Flow) ScopeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		be := builder.NewExpr()
		be.SetValue(f.loader.Variables())

		var buf bytes.Buffer
		_, err := be.WriteTo(&buf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			_, _ = io.Copy(w, &buf)
		}
	}
}

// configBytes dumps the current state of the flow config as River.
func (f *Flow) configBytes(w io.Writer, debugInfo bool) (n int64, err error) {
	file := builder.NewFile()

	blocks := f.loader.WriteBlocks(debugInfo)
	for _, block := range blocks {
		var id controller.ComponentID
		id = append(id, block.Name...)
		if block.Label != "" {
			id = append(id, block.Label)
		}

		comment := fmt.Sprintf("// Component %s:", id.String())
		file.Body().AppendTokens([]builder.Token{
			{Tok: token.COMMENT, Lit: comment},
		})

		file.Body().AppendBlock(block)
		file.Body().AppendTokens([]builder.Token{
			{Tok: token.LITERAL, Lit: "\n"},
		})
	}

	return file.WriteTo(w)
}
