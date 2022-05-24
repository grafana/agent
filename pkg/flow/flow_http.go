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
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rfratto/gohcl/hclfmt"
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
// recently loaded configuration file as HCL.
func (f *Flow) ConfigHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		debugInfo := r.URL.Query().Get("debug") == "1"
		_, _ = f.configBytes(w, debugInfo)
	}
}

// configBytes dumps the current state of the flow config as HCL.
func (f *Flow) configBytes(w io.Writer, debugInfo bool) (n int64, err error) {
	file := hclwrite.NewFile()

	blocks := f.loader.WriteBlocks(debugInfo)
	for _, block := range blocks {
		var id controller.ComponentID
		id = append(id, block.Type())
		id = append(id, block.Labels()...)

		comment := fmt.Sprintf("// Component %s:\n", id.String())
		file.Body().AppendUnstructuredTokens(hclwrite.Tokens{
			{Type: hclsyntax.TokenComment, Bytes: []byte(comment)},
		})

		file.Body().AppendBlock(block)
		file.Body().AppendNewline()
	}

	toks := file.BuildTokens(nil)
	hclfmt.Format(toks)

	return toks.WriteTo(w)
}
