package controller

import (
	"fmt"
	"strings"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/vm"
)

// DeclareComponentNode is an instance of a module.
// Its arguments are passed to the arguments nodes of the corresponding module.
type DeclareComponentNode struct {
	id            ComponentID
	label         string
	componentName string
	namespace     string
	nodeID        string // Cached from id.String() to avoid allocating new strings every time NodeID is called.

	mut       sync.RWMutex
	block     *ast.BlockStmt // Current River block to derive args from
	eval      *vm.Evaluator
	arguments component.Arguments // Evaluated arguments for the corresponding module.
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

func (cn *DeclareComponentNode) ID() ComponentID { return cn.id }

func (cn *DeclareComponentNode) Label() string { return cn.label }

func (cn *DeclareComponentNode) ComponentName() string { return cn.componentName }

func (cn *DeclareComponentNode) NodeID() string { return cn.nodeID }

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

func (cn *DeclareComponentNode) Arguments() component.Arguments {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.arguments
}

func (cn *DeclareComponentNode) Block() *ast.BlockStmt {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.block
}

func (cn *DeclareComponentNode) Namespace() string { return cn.namespace }

func (cn *DeclareComponentNode) SetNamespace(namespace string) { cn.namespace = namespace }

func (cn *DeclareComponentNode) Clone(newID string) dag.Node {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return &DeclareComponentNode{
		id:            strings.Split(newID, "."),
		nodeID:        newID,
		componentName: cn.componentName,
		mut:           sync.RWMutex{},
		block:         cn.block,
		eval:          vm.New(cn.block.Body),
		label:         cn.label,
		namespace:     cn.namespace,
		// does not clone the arguments
	}
}
