package flow

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/dag"
	"github.com/grafana/agent/pkg/flow/graphviz"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rfratto/gohcl"
	"github.com/rfratto/gohcl/hclfmt"
)

// GraphHandler returns an http.HandlerFunc that render's the flow's DAG as an
// SVG. Graphviz must be installed for this to work.
func GraphHandler(f *Flow) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		f.graphMut.RLock()
		contents := dag.MarshalDOT(f.graph)
		f.graphMut.RUnlock()

		svgBytes, err := graphviz.Dot(contents, "svg")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = io.Copy(w, bytes.NewReader(svgBytes))
	}
}

// NametableHandler returns an http.HandlerFunc that render's the flow's
// nametable as an SVG. Graphviz must be installed for this to work.
func NametableHandler(f *Flow) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		f.graphMut.RLock()
		contents := dag.MarshalDOT(&f.nametable.graph)
		f.graphMut.RUnlock()

		svgBytes, err := graphviz.Dot(contents, "svg")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = io.Copy(w, bytes.NewReader(svgBytes))
	}
}

// ConfigHandler returns an http.Handler which prints out the flow's current
// config as HCL.
func ConfigHandler(f *Flow) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		componentName := r.URL.Query().Get("component")

		f.graphMut.RLock()
		defer f.graphMut.RUnlock()

		file := hclwrite.NewFile()

		// If we're printing all components, include the root body.
		if componentName == "" {
			gohcl.EncodeIntoBody(f.root, file.Body())
			file.Body().AppendNewline()
		}

		blockSchema := component.RegistrySchema()
		content, _ := f.root.Remain.Content(blockSchema)

		// Encode the components now
		for _, block := range content.Blocks {
			b := hclwrite.NewBlock(block.Type, block.Labels)

			ref := referenceForBlock(block)
			if componentName != "" && ref.String() != componentName {
				continue
			}

			var cn *componentNode

			// Find the named component
			dag.Walk(f.graph, f.graph.Roots(), func(n dag.Node) error {
				nodeRef := n.(*componentNode).Reference()
				if nodeRef.Equals(ref) {
					cn = n.(*componentNode)
					return fmt.Errorf("done")
				}
				return nil
			})
			if cn == nil {
				errorMsg := fmt.Sprintf("could not find component %s in graph", ref)
				http.Error(w, errorMsg, http.StatusInternalServerError)
				return
			}

			cfg := cn.Config()
			if cfg == nil {
				http.Error(w, "Component %s did not return its config", http.StatusInternalServerError)
				return
			}
			gohcl.EncodeIntoBody(cfg, b.Body())

			// Optionally write output state if it's exposed by the component.
			if cs := cn.State(); cs != nil {
				b.Body().AppendUnstructuredTokens(hclwrite.Tokens{
					{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
					{Type: hclsyntax.TokenComment, Bytes: []byte("// Output:\n")},
				})

				gohcl.EncodeIntoBody(cs, b.Body())
			}

			// Debug info
			if r.URL.Query().Get("debug") == "1" {
				b.Body().AppendUnstructuredTokens(hclwrite.Tokens{
					{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
					{Type: hclsyntax.TokenComment, Bytes: []byte("// Debug info:\n")},
				})

				b.Body().AppendBlock(gohcl.EncodeAsBlock(cn.CurrentHealth(), "health"))

				var status any
				sc, _ := cn.Get().(component.StatusComponent)
				if sc != nil {
					status = sc.CurrentStatus()
				}
				if status != nil {
					b.Body().AppendBlock(gohcl.EncodeAsBlock(status, "status"))
				}
			}

			comment := fmt.Sprintf("// Component %s:\n", cn.ref.String())
			file.Body().AppendUnstructuredTokens(hclwrite.Tokens{
				{Type: hclsyntax.TokenComment, Bytes: []byte(comment)},
			})

			file.Body().AppendBlock(b)
			file.Body().AppendNewline()
		}

		toks := file.BuildTokens(nil)
		hclfmt.Format(toks)
		_, _ = toks.WriteTo(w)
	}
}
