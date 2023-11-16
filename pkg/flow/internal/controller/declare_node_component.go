package controller

import (
	"fmt"
	"strings"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/vm"
)

// ComponentNode is a controller node which manages a user-defined component.
//
// ComponentNode manages the underlying component and caches its current
// arguments and exports. ComponentNode manages the arguments for the component
// from a River block.
type DeclareComponentNode struct {
	id            ComponentID
	label         string
	componentName string
	namespace     string
	nodeID        string // Cached from id.String() to avoid allocating new strings every time NodeID is called.

	mut       sync.RWMutex
	block     *ast.BlockStmt // Current River block to derive args from
	eval      *vm.Evaluator
	arguments component.Arguments // Evaluated arguments for the managed component
}

var _ BlockNode = (*DeclareComponentNode)(nil)

func NewDeclareComponentNode(globals ComponentGlobals, b *ast.BlockStmt) *DeclareComponentNode {
	var (
		id     = BlockComponentID(b)
		nodeID = id.String()
	)

	cn := &DeclareComponentNode{
		id:            id,
		label:         b.Label,
		nodeID:        nodeID,
		componentName: strings.Join(b.Name, "."),
		block:         b,
		eval:          vm.New(b.Body),
	}
	return cn
}

// ID returns the component ID of the managed component from its River block.
func (cn *DeclareComponentNode) ID() ComponentID { return cn.id }

// Label returns the label for the block or "" if none was specified.
func (cn *DeclareComponentNode) Label() string { return cn.label }

// ComponentName returns the component's type, i.e. `local.file.test` returns `local.file`.
func (cn *DeclareComponentNode) ComponentName() string { return cn.componentName }

// NodeID implements dag.Node and returns the unique ID for this node. The
// NodeID is the string representation of the component's ID from its River
// block.
func (cn *DeclareComponentNode) NodeID() string { return cn.nodeID }

// UpdateBlock updates the River block used to construct arguments for the
// managed component. The new block isn't used until the next time Evaluate is
// invoked.
//
// UpdateBlock will panic if the block does not match the component ID of the
// ComponentNode.
func (cn *DeclareComponentNode) UpdateBlock(b *ast.BlockStmt) {
	if !BlockComponentID(b).Equals(cn.id) {
		panic("UpdateBlock called with an River block with a different component ID")
	}

	cn.mut.Lock()
	defer cn.mut.Unlock()
	cn.block = b
	cn.eval = vm.New(b.Body)
}

func (cn *DeclareComponentNode) Evaluate(scope *vm.Scope) error {
	cn.mut.Lock()
	defer cn.mut.Unlock()

	var values map[string]any
	if err := cn.eval.Evaluate(scope, &values); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}
	cn.arguments = values
	return nil
}

// Arguments returns the current arguments of the managed component.
func (cn *DeclareComponentNode) Arguments() component.Arguments {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.arguments
}

// Block implements BlockNode and returns the current block of the managed component.
func (cn *DeclareComponentNode) Block() *ast.BlockStmt {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.block
}

func (cn *DeclareComponentNode) Namespace() string { return cn.namespace }

func (cn *DeclareComponentNode) SetNamespace(namespace string) { cn.namespace = namespace }
