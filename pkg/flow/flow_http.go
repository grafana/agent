package flow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/grafana/agent/pkg/river"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/flow/internal/graphviz"
	"github.com/grafana/agent/pkg/river/token"
	"github.com/grafana/agent/pkg/river/token/builder"
)

// GraphHandler returns an http.HandlerFunc which renders the current graph's
// DAG as an SVG. Graphviz must be installed for this function to work.
func (c *Flow) GraphHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		g := c.loader.Graph()
		dot := dag.MarshalDOT(g)

		svgBytes, err := graphviz.Dot(dot, "svg")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(w, bytes.NewReader(svgBytes))
		if err != nil {
			level.Error(c.log).Log("msg", "failed to write svg graph", "err", err)
		}
	}
}

// ConfigHandler returns an http.HandlerFunc which will render the most
// recently loaded configuration file as River.
func (c *Flow) ConfigHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		debugInfo := r.URL.Query().Get("debug") == "1"

		var buf bytes.Buffer
		_, err := c.configBytes(&buf, debugInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			_, _ = io.Copy(w, &buf)
		}
	}
}

// ScopeHandler returns an http.HandlerFunc which will render the scope used
// for variable references throughout River expressions.
func (c *Flow) ScopeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		be := builder.NewExpr()
		be.SetValue(c.loader.Variables())

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
func (c *Flow) configBytes(w io.Writer, debugInfo bool) (n int64, err error) {
	file := builder.NewFile()

	blocks := c.loader.WriteBlocks(debugInfo)
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

// ComponentJSON returns the json representation of the flow component.
func (c *Flow) ComponentJSON(w io.Writer, ci *river.ComponentField) (int, error) {
	var foundComponent *controller.ComponentNode
	for _, c := range c.loader.Components() {
		if c.ID().String() == ci.ID {
			foundComponent = c
			break
		}
	}
	if foundComponent == nil {
		return 0, fmt.Errorf("unable to find component named %s", ci.ID)
	}
	field := river.ConvertComponentToJSON(
		foundComponent.ID(),
		foundComponent.Arguments(),
		foundComponent.Exports(),
		ci.References,
		ci.ReferencedBy,
		&river.Health{
			State:       ci.Health.State,
			Message:     ci.Health.Message,
			UpdatedTime: ci.Health.UpdatedTime,
		},
		"")
	bb, err := json.Marshal(field)
	if err != nil {
		return 0, err
	}
	return w.Write(bb)
}
