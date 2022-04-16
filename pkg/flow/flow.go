// Package flow implements a component graph system.
package flow

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/dag"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/rfratto/gohcl"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

// Flow is the Flow component graph system.
type Flow struct {
	log        log.Logger
	configFile string

	updates *updateQueue

	graphMut  sync.RWMutex
	graph     *dag.Graph
	nametable *nametable
	root      rootBlock
	handlers  map[string]http.Handler
}

// New creates a new Flow instance.
func New(l log.Logger, configFile string) *Flow {
	f := &Flow{
		log:        l,
		configFile: configFile,

		updates: newUpdateQueue(),

		graph:     &dag.Graph{},
		nametable: &nametable{},
		handlers:  make(map[string]http.Handler),
	}
	return f
}

// Load reads the config file and updates the system to reflect what was read.
func (f *Flow) Load() error {
	f.graphMut.Lock()
	defer f.graphMut.Unlock()

	// TODO(rfratto): this won't work yet for subsequent loads.
	//
	// Figuring out how to mutate the DAG to match the current state of the file
	// will take some thinking.

	bb, err := os.ReadFile(f.configFile)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	file, diags := hclsyntax.ParseConfig(bb, f.configFile, hcl.InitialPos)
	if diags.HasErrors() {
		return diags
	}

	var root rootBlock
	decodeDiags := gohcl.DecodeBody(file.Body, nil, &root)
	diags = diags.Extend(decodeDiags)
	if diags.HasErrors() {
		return diags
	}
	f.root = root

	blockSchema := component.RegistrySchema()
	content, remainDiags := root.Remain.Content(blockSchema)
	diags = diags.Extend(remainDiags)
	if diags.HasErrors() {
		return diags
	}

	// Construct our components and the nametable.
	for _, block := range content.Blocks {
		// Create the component and add it into our graph.
		component := newComponentNode(block)
		f.graph.Add(component)

		// Then, add the component into our nametable.
		f.nametable.Add(component)
	}

	// Second pass: iterate over all of our nodes and create edges.
	for _, node := range f.graph.Nodes() {
		var (
			component  = node.(*componentNode)
			body       = component.block.Body
			traversals = expressionsFromSyntaxBody(body.(*hclsyntax.Body))
		)
		for _, t := range traversals {
			target, lookupDiags := f.nametable.LookupTraversal(t)
			diags = diags.Extend(lookupDiags)
			if target == nil {
				continue
			}

			// Add dependency to the found node
			f.graph.AddEdge(dag.Edge{From: component, To: target})
		}
	}
	if diags.HasErrors() {
		return diags
	}

	// Wiring edges probably caused a mess. Reduce it.
	dag.Reduce(f.graph)

	funcMap := map[string]function.Function{
		"concat": stdlib.ConcatFunc,
	}

	// At this point, our DAG is completely formed and we can start to construct
	// the real components and evaluate expressions. Walk topologically in
	// dependency order.
	//
	// TODO(rfratto): should this happen as part of the run? If we moved this to
	// the run, we would need a separate type checking pass in the Load to ensure
	// that all expressions throughout the config are valid. As it is now, this
	// typechecks on its own.
	err = dag.WalkTopological(f.graph, f.graph.Leaves(), func(n dag.Node) error {
		cn := n.(*componentNode)

		directDeps := f.graph.Dependencies(cn)
		ectx, err := f.nametable.BuildEvalContext(directDeps)
		if err != nil {
			return err
		} else if ectx != nil {
			ectx.Functions = funcMap
		}

		opts := component.Options{
			ComponentID: cn.Name(),
			Logger:      log.With(f.log, "node", cn.Name()),

			// TODO(rfratto): remove hard-coded address
			HTTPAddr: "127.0.0.1:12345",

			OnStateChange: func() { f.updates.Enqueue(cn) },
		}

		if err := cn.Build(opts, ectx); err != nil {
			return err
		}

		hc, ok := cn.Get().(component.HTTPComponent)
		if ok {
			handler, err := hc.ComponentHandler()
			if err != nil {
				return err
			}
			f.handlers[opts.ComponentID] = handler
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

type rootBlock struct {
	LogLevel  string `hcl:"log_level,optional"`
	LogFormat string `hcl:"log_format,optional"`

	Body   hcl.Body `hcl:",body"`
	Remain hcl.Body `hcl:",remain"`
}

type reference []string

func (r reference) String() string {
	return strings.Join(r, ".")
}

func (r reference) Equals(other reference) bool {
	if len(r) != len(other) {
		return false
	}
	for i := 0; i < len(r); i++ {
		if r[i] != other[i] {
			return false
		}
	}
	return true
}

// expressionsFromSyntaxBody returcses through body and finds all variable
// references.
func expressionsFromSyntaxBody(body *hclsyntax.Body) []hcl.Traversal {
	var exprs []hcl.Traversal

	for _, attrib := range body.Attributes {
		exprs = append(exprs, attrib.Expr.Variables()...)
	}
	for _, block := range body.Blocks {
		exprs = append(exprs, expressionsFromSyntaxBody(block.Body)...)
	}

	return exprs
}

// Run runs f until ctx is canceled. It is invalid to call Run concurrently.
func (f *Flow) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	funcMap := map[string]function.Function{
		"concat": stdlib.ConcatFunc,
	}

	// TODO(rfratto): start/stop nodes after refresh
	var wg sync.WaitGroup
	defer wg.Wait()

	f.graphMut.Lock()
	for _, n := range f.graph.Nodes() {
		wg.Add(1)
		go func(cn *componentNode) {
			defer wg.Done()

			c := cn.Get()
			if c == nil {
				panic("component never initialized")
			}

			level.Info(f.log).Log("msg", "starting component", "component", cn.Name())
			err := c.Run(ctx)
			if err != nil {
				level.Error(f.log).Log("msg", "component exited with error", "component", cn.Name(), "err", err)
			} else {
				level.Info(f.log).Log("msg", "component stopped", "component", cn.Name())
			}
		}(n.(*componentNode))
	}
	f.graphMut.Unlock()

	for {
		cn, err := f.updates.Dequeue(ctx)
		if err != nil {
			return nil
		}

		level.Debug(f.log).Log("msg", "handling component with updated state", "component", cn.Name())

		f.graphMut.Lock()
		defer f.graphMut.Unlock()

		// Update any component which directly references cn.
		// TODO(rfratto): set health of node based on result of this?
		for _, n := range f.graph.Dependants(cn) {
			cn := n.(*componentNode)

			directDeps := f.graph.Dependencies(cn)
			ectx, err := f.nametable.BuildEvalContext(directDeps)
			if err != nil {
				level.Error(f.log).Log("msg", "failed to update node", "node", cn.Name(), "err", err)
				continue
			} else if ectx != nil {
				ectx.Functions = funcMap
			}

			if err := cn.Update(ectx); err != nil {
				level.Error(f.log).Log("msg", "failed to update node", "node", cn.Name(), "err", err)
				continue
			}
		}
	}
}

// WireRoutes injects routs into r for the Flow API.
func (f *Flow) WireRoutes(r *mux.Router) {
	r.PathPrefix("/component/{component}/").HandlerFunc(f.handleComponentRequest)
}

func (f *Flow) handleComponentRequest(rw http.ResponseWriter, r *http.Request) {
	f.graphMut.RLock()
	defer f.graphMut.RUnlock()

	componentID, ok := mux.Vars(r)["component"]
	if !ok {
		http.Error(rw, "no component variable set", http.StatusInternalServerError)
		return
	}
	prefix := fmt.Sprintf("/component/%s", componentID)

	handler, ok := f.handlers[componentID]
	if !ok {
		http.NotFound(rw, r)
		return
	}
	http.StripPrefix(prefix, handler).ServeHTTP(rw, r)
}
