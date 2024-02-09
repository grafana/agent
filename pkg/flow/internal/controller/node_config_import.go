package controller

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/atomic"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/importsource"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/parser"
	"github.com/grafana/river/vm"
	"github.com/prometheus/client_golang/prometheus"
)

// ImportConfigNode imports declare and import blocks via a managed import source.
// The imported declare are stored in importedDeclares.
// For every imported import block, the ImportConfigNode will create ImportConfigNode children.
// The children are evaluated and ran by the parent.
// When an ImportConfigNode receives new content from its source, it updates its importedDeclares and recreates its children.
// Then an update call is propagated to the root ImportConfigNode to inform the controller for reevaluation.
type ImportConfigNode struct {
	nodeID        string
	globalID      string
	label         string
	componentName string
	globals       ComponentGlobals          // Need a copy of the globals to create other import nodes
	block         *ast.BlockStmt            // Current River blocks to derive config from
	source        importsource.ImportSource // source retrieves the module content

	OnBlockNodeUpdate func(cn BlockNode) // notifies the controller or the parent for reevaluation
	logger            log.Logger

	importChildrenUpdateChan chan struct{} // used to trigger an update of the running children

	mut                       sync.RWMutex
	importedContent           string
	importConfigNodesChildren map[string]*ImportConfigNode
	importChildrenRunning     bool
	importedDeclares          map[string]ast.Body

	inContentUpdate atomic.Bool
}

var _ RunnableNode = (*ImportConfigNode)(nil)

// NewImportConfigNode creates a new ImportConfigNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewImportConfigNode(block *ast.BlockStmt, globals ComponentGlobals, sourceType importsource.SourceType) *ImportConfigNode {
	nodeID := BlockComponentID(block).String()

	globalID := nodeID
	if globals.ControllerID != "" {
		globalID = path.Join(globals.ControllerID, nodeID)
	}

	cn := &ImportConfigNode{
		nodeID:                   nodeID,
		globalID:                 globalID,
		label:                    block.Label,
		componentName:            block.GetBlockName(),
		globals:                  globals,
		block:                    block,
		OnBlockNodeUpdate:        globals.OnBlockNodeUpdate,
		importChildrenUpdateChan: make(chan struct{}),
	}
	managedOpts := getImportManagedOptions(globals, cn)
	cn.logger = managedOpts.Logger
	cn.source = importsource.NewImportSource(sourceType, managedOpts, vm.New(block.Body), cn.onContentUpdate)
	return cn
}

func getImportManagedOptions(globals ComponentGlobals, cn *ImportConfigNode) component.Options {
	return component.Options{
		ID:     cn.globalID,
		Logger: log.With(globals.Logger, "config", cn.globalID),
		Registerer: prometheus.WrapRegistererWith(prometheus.Labels{
			"config_id": cn.globalID,
		}, prometheus.NewRegistry()),
		Tracer:   tracing.WrapTracer(globals.TraceProvider, cn.globalID),
		DataPath: filepath.Join(globals.DataPath, cn.globalID),
		GetServiceData: func(name string) (interface{}, error) {
			return globals.GetServiceData(name)
		},
	}
}

// Evaluate implements BlockNode and evaluates the import source.
func (cn *ImportConfigNode) Evaluate(scope *vm.Scope) error {
	return cn.source.Evaluate(scope)
}

// onContentUpdate is triggered every time the managed import source has new content.
func (cn *ImportConfigNode) onContentUpdate(importedContent string) {
	cn.mut.Lock()
	defer cn.mut.Unlock()

	cn.inContentUpdate.Store(true)
	defer cn.inContentUpdate.Store(false)

	// If the source sent the same content, there is no need to reload.
	if cn.importedContent == importedContent {
		return
	}

	cn.importedContent = importedContent
	cn.importedDeclares = make(map[string]ast.Body)
	cn.importConfigNodesChildren = make(map[string]*ImportConfigNode)

	parsedImportedContent, err := parser.ParseFile(cn.label, []byte(importedContent))
	if err != nil {
		level.Error(cn.logger).Log("msg", "failed to parse file on update", "err", err)
		return
	}

	// populate importedDeclares and importConfigNodesChildren
	cn.processImportedContent(parsedImportedContent)

	// evaluate the importConfigNodesChildren that have been created
	err = cn.evaluateChildren()
	if err != nil {
		level.Error(cn.logger).Log("msg", "failed to evaluate nested import", "err", err)
		return
	}

	// trigger to stop previous children from running and to start running the new ones.
	if cn.importChildrenRunning {
		cn.importChildrenUpdateChan <- struct{}{}
	}

	cn.OnBlockNodeUpdate(cn)
}

// processImportedContent processes declare and import blocks of the provided ast content.
func (cn *ImportConfigNode) processImportedContent(content *ast.File) {
	for _, stmt := range content.Body {
		switch stmt := stmt.(type) {
		case *ast.BlockStmt:
			fullName := strings.Join(stmt.Name, ".")
			switch fullName {
			case "declare":
				cn.processDeclareBlock(stmt)
			case importsource.BlockImportFile, importsource.BlockImportString: // TODO: add other import sources
				cn.processImportBlock(stmt, fullName)
			default:
				level.Error(cn.logger).Log("msg", "only declare and import blocks are allowed in a module", "forbidden", fullName)
			}
		default:
			level.Error(cn.logger).Log("msg", "only declare and import blocks are allowed in a module")
		}
	}
}

// processDeclareBlock stores the declare definition in the importedDeclares.
func (cn *ImportConfigNode) processDeclareBlock(stmt *ast.BlockStmt) {
	if _, ok := cn.importedDeclares[stmt.Label]; ok {
		level.Error(cn.logger).Log("msg", "declare block redefined", "name", stmt.Label)
		return
	}
	cn.importedDeclares[stmt.Label] = stmt.Body
}

// processDeclareBlock creates an ImportConfigNode child from the provided import block.
func (cn *ImportConfigNode) processImportBlock(stmt *ast.BlockStmt, fullName string) {
	sourceType := importsource.GetSourceType(fullName)
	if _, ok := cn.importConfigNodesChildren[stmt.Label]; ok {
		level.Error(cn.logger).Log("msg", "import block redefined", "name", stmt.Label)
		return
	}
	childGlobals := cn.globals
	// Children have a special OnBlockNodeUpdate function which notifies the parent when its content changes.
	childGlobals.OnBlockNodeUpdate = cn.onChildrenContentUpdate
	cn.importConfigNodesChildren[stmt.Label] = NewImportConfigNode(stmt, childGlobals, sourceType)
}

// evaluateChildren evaluates the import nodes managed by this import node.
func (cn *ImportConfigNode) evaluateChildren() error {
	for _, child := range cn.importConfigNodesChildren {
		err := child.Evaluate(&vm.Scope{
			Parent:    nil,
			Variables: make(map[string]interface{}),
		})
		if err != nil {
			return fmt.Errorf("imported node %s failed to evaluate, %v", child.label, err)
		}
	}
	return nil
}

// onChildrenContentUpdate notifies the parent that the content has been updated.
func (cn *ImportConfigNode) onChildrenContentUpdate(child BlockNode) {
	// If the node is already updating its content, it will call OnBlockNodeUpdate
	// so the notification can be ignored.
	if !cn.inContentUpdate.Load() {
		cn.OnBlockNodeUpdate(cn)
	}
}

// Run runs the managed source and the import children until ctx is
// canceled. Evaluate must have been called at least once without returning an
// error before calling Run.
//
// Run will immediately return ErrUnevaluated if Evaluate has never been called
// successfully. Otherwise, Run will return nil.
func (cn *ImportConfigNode) Run(ctx context.Context) error {
	cn.mut.Lock()
	importChildrenCount := len(cn.importConfigNodesChildren)
	cn.mut.Unlock()
	if cn.source == nil {
		return ErrUnevaluated
	}

	newCtx, cancel := context.WithCancel(ctx)
	defer cancel() // This will stop the children and the managed source.

	errChan := make(chan error, 1)

	if importChildrenCount > 0 {
		go func() {
			errChan <- cn.runChildren(newCtx)
		}()
	}

	go func() {
		errChan <- cn.source.Run(newCtx)
	}()

	err := <-errChan

	if err != nil {
		level.Error(cn.logger).Log("msg", "import exited with error", "err", err)
	} else {
		level.Info(cn.logger).Log("msg", "import exited")
	}
	return err
}

// runChildren run the import nodes managed by this import node.
// The children list can be updated by onContentUpdate. In this case we need to stop the running children and run the new set of children.
func (cn *ImportConfigNode) runChildren(parentCtx context.Context) error {
	errChildrenChan := make(chan error)
	var wg sync.WaitGroup
	var ctx context.Context
	var cancel context.CancelFunc

	startChildren := func(ctx context.Context, children map[string]*ImportConfigNode, wg *sync.WaitGroup) {
		for _, child := range children {
			wg.Add(1)
			go func(child *ImportConfigNode) {
				defer wg.Done()
				if err := child.Run(ctx); err != nil {
					errChildrenChan <- err
				}
			}(child)
		}
	}

	childrenDone := func(wg *sync.WaitGroup, doneChan chan struct{}) {
		wg.Wait()
		close(doneChan)
	}

	ctx, cancel = context.WithCancel(parentCtx)
	cn.mut.Lock()
	startChildren(ctx, cn.importConfigNodesChildren, &wg) // initial start of children
	cn.importChildrenRunning = true
	cn.mut.Unlock()

	doneChan := make(chan struct{})
	go childrenDone(&wg, doneChan) // start goroutine to cover the case when all children finish

	for {
		select {
		case <-cn.importChildrenUpdateChan:
			cancel()   // cancel all running children
			<-doneChan // wait for the children to finish

			wg = sync.WaitGroup{}
			errChildrenChan = make(chan error)

			ctx, cancel = context.WithCancel(parentCtx) // create a new context
			cn.mut.Lock()
			startChildren(ctx, cn.importConfigNodesChildren, &wg) // start the new set of children
			cn.mut.Unlock()

			doneChan = make(chan struct{})
			go childrenDone(&wg, doneChan) // start goroutine to cover the case when all children finish
		case err := <-errChildrenChan:
			// one child stopped because of an error.
			cancel()
			return err
		case <-doneChan:
			// all children stopped without error.
			cancel()
			return nil
		}
	}
}

func (cn *ImportConfigNode) Label() string { return cn.label }

// Block implements BlockNode and returns the current block of the managed config node.
func (cn *ImportConfigNode) Block() *ast.BlockStmt { return cn.block }

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *ImportConfigNode) NodeID() string { return cn.nodeID }

// ImportedDeclares returns all declare blocks that it imported.
func (cn *ImportConfigNode) ImportedDeclares() map[string]ast.Body {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.importedDeclares
}

// ImportConfigNodesChildren returns the ImportConfigNodesChildren of this ImportConfigNode.
func (cn *ImportConfigNode) ImportConfigNodesChildren() map[string]*ImportConfigNode {
	cn.mut.Lock()
	defer cn.mut.Unlock()
	return cn.importConfigNodesChildren
}
