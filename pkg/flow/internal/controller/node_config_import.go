package controller

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/local/file"
	importsource "github.com/grafana/agent/pkg/flow/internal/import-source"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/grafana/agent/pkg/flow/tracing"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/parser"
	"github.com/grafana/river/vm"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/atomic"
)

type ImportConfigNode struct {
	id                ComponentID
	label             string
	nodeID            string
	componentName     string
	globalID          string
	source            importsource.ImportSource
	registry          *prometheus.Registry
	importedContent   map[string]string
	OnComponentUpdate func(cn NodeWithDependants) // Informs controller that we need to reevaluate
	logger            log.Logger

	mut                sync.RWMutex
	importedContentMut sync.RWMutex
	block              *ast.BlockStmt // Current River blocks to derive config from
	lastUpdateTime     atomic.Time

	healthMut  sync.RWMutex
	evalHealth component.Health // Health of the last evaluate
	runHealth  component.Health // Health of running the component
}

var _ NodeWithDependants = (*ImportConfigNode)(nil)
var _ RunnableNode = (*ImportConfigNode)(nil)
var _ UINode = (*ComponentNode)(nil)

// NewImportConfigNode creates a new ImportConfigNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewImportConfigNode(block *ast.BlockStmt, globals ComponentGlobals, sourceType importsource.SourceType) *ImportConfigNode {
	var (
		id     = BlockComponentID(block)
		nodeID = id.String()
	)

	initHealth := component.Health{
		Health:     component.HealthTypeUnknown,
		Message:    "component created",
		UpdateTime: time.Now(),
	}
	globalID := nodeID
	if globals.ControllerID != "" {
		globalID = path.Join(globals.ControllerID, nodeID)
	}
	cn := &ImportConfigNode{
		id:                id,
		globalID:          globalID,
		label:             block.Label,
		nodeID:            BlockComponentID(block).String(),
		componentName:     block.GetBlockName(),
		importedContent:   make(map[string]string),
		OnComponentUpdate: globals.OnComponentUpdate,
		block:             block,
		evalHealth:        initHealth,
		runHealth:         initHealth,
	}
	managedOpts := getImportManagedOptions(globals, cn)
	cn.logger = managedOpts.Logger
	cn.source = importsource.CreateImportSource(sourceType, managedOpts, vm.New(block.Body))
	return cn
}

func getImportManagedOptions(globals ComponentGlobals, cn *ImportConfigNode) component.Options {
	cn.registry = prometheus.NewRegistry()
	return component.Options{
		ID:     cn.globalID,
		Logger: log.With(globals.Logger, "component", cn.globalID),
		Registerer: prometheus.WrapRegistererWith(prometheus.Labels{
			"component_id": cn.globalID,
		}, cn.registry),
		Tracer: tracing.WrapTracer(globals.TraceProvider, cn.globalID),

		DataPath: filepath.Join(globals.DataPath, cn.globalID),

		OnStateChange: cn.onContentUpdate,

		GetServiceData: func(name string) (interface{}, error) {
			return globals.GetServiceData(name)
		},
	}
}

// Evaluate implements BlockNode and updates the arguments for the managed config block
// by re-evaluating its River block with the provided scope. The managed config block
// will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated or if
// decoding to arguments fails.
func (cn *ImportConfigNode) Evaluate(scope *vm.Scope) error {
	err := cn.evaluate(scope)

	switch err {
	case nil:
		cn.setEvalHealth(component.HealthTypeHealthy, "component evaluated")
	default:
		msg := fmt.Sprintf("component evaluation failed: %s", err)
		cn.setEvalHealth(component.HealthTypeUnhealthy, msg)
	}
	return err
}

func (cn *ImportConfigNode) setEvalHealth(t component.HealthType, msg string) {
	cn.healthMut.Lock()
	defer cn.healthMut.Unlock()

	cn.evalHealth = component.Health{
		Health:     t,
		Message:    msg,
		UpdateTime: time.Now(),
	}
}

func (cn *ImportConfigNode) evaluate(scope *vm.Scope) error {
	cn.mut.Lock()
	defer cn.mut.Unlock()
	return cn.source.Evaluate(scope)
}

func (cn *ImportConfigNode) onContentUpdate(e component.Exports) {
	cn.importedContentMut.Lock()
	defer cn.importedContentMut.Unlock()
	fileContent := e.(file.Exports).Content.Value
	cn.importedContent = make(map[string]string)
	node, err := parser.ParseFile(cn.label, []byte(fileContent))
	if err != nil {
		level.Error(cn.logger).Log("msg", "failed to parse file on update", "err", err)
		return
	}
	for _, stmt := range node.Body {
		switch stmt := stmt.(type) {
		case *ast.BlockStmt:
			fullName := strings.Join(stmt.Name, ".")
			switch fullName {
			case "declare":
				if _, ok := cn.importedContent[stmt.Label]; ok {
					level.Error(cn.logger).Log("msg", "declare block redefined", "name", stmt.Label)
					continue
				}
				cn.importedContent[stmt.Label] = fileContent[stmt.LCurlyPos.Position().Offset+1 : stmt.RCurlyPos.Position().Offset-1]
			default:
				level.Error(cn.logger).Log("msg", "only declare blocks are allowed in a module", "forbidden", fullName)
			}
		default:
			level.Error(cn.logger).Log("msg", "only declare blocks are allowed in a module")
		}
	}
	cn.lastUpdateTime.Store(time.Now())
	cn.OnComponentUpdate(cn)
}

func (cn *ImportConfigNode) ModuleContent(module string) (string, error) {
	cn.importedContentMut.Lock()
	defer cn.importedContentMut.Unlock()
	if content, ok := cn.importedContent[module]; ok {
		return content, nil
	}
	return "", fmt.Errorf("module %s not found in imported node %s", module, cn.label)
}

// Run runs the managed component in the calling goroutine until ctx is
// canceled. Evaluate must have been called at least once without returning an
// error before calling Run.
//
// Run will immediately return ErrUnevaluated if Evaluate has never been called
// successfully. Otherwise, Run will return nil.
func (cn *ImportConfigNode) Run(ctx context.Context) error {
	cn.mut.RLock()
	managed := cn.source
	cn.mut.RUnlock()

	if managed == nil {
		return ErrUnevaluated
	}

	cn.setRunHealth(component.HealthTypeHealthy, "started component")
	err := managed.Run(ctx)

	var exitMsg string
	if err != nil {
		level.Error(cn.logger).Log("msg", "component exited with error", "err", err)
		exitMsg = fmt.Sprintf("component shut down with error: %s", err)
	} else {
		level.Info(cn.logger).Log("msg", "component exited")
		exitMsg = "component shut down normally"
	}

	cn.setRunHealth(component.HealthTypeExited, exitMsg)
	return err
}

func (cn *ImportConfigNode) setRunHealth(t component.HealthType, msg string) {
	cn.healthMut.Lock()
	defer cn.healthMut.Unlock()

	cn.runHealth = component.Health{
		Health:     t,
		Message:    msg,
		UpdateTime: time.Now(),
	}
}

func (cn *ImportConfigNode) Label() string { return cn.label }

// Block implements BlockNode and returns the current block of the managed config node.
func (cn *ImportConfigNode) Block() *ast.BlockStmt {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.block
}

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *ImportConfigNode) NodeID() string { return cn.nodeID }

// This node has no exports.
func (cn *ImportConfigNode) Exports() component.Exports {
	return nil
}

func (cn *ImportConfigNode) ID() ComponentID { return cn.id }

func (cn *ImportConfigNode) LastUpdateTime() time.Time {
	return cn.lastUpdateTime.Load()
}

// Arguments returns the current arguments of the managed component.
func (cn *ImportConfigNode) Arguments() component.Arguments {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.source.Arguments()
}

// Component returns the instance of the managed component. Component may be
// nil if the ComponentNode has not been successfully evaluated yet.
func (cn *ImportConfigNode) Component() component.Component {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.source.Component()
}

// CurrentHealth returns the current health of the ComponentNode.
//
// The health of a ComponentNode is determined by combining:
//
//  1. Health from the call to Run().
//  2. Health from the last call to Evaluate().
//  3. Health reported from the component.
func (cn *ImportConfigNode) CurrentHealth() component.Health {
	cn.healthMut.RLock()
	defer cn.healthMut.RUnlock()
	return component.LeastHealthy(cn.runHealth, cn.evalHealth, cn.source.CurrentHealth())
}

// FileComponent does not have DebugInfo
func (cn *ImportConfigNode) DebugInfo() interface{} {
	return nil
}

// This component does not manage modules.
func (cn *ImportConfigNode) ModuleIDs() []string {
	return nil
}

// BlockName returns the name of the block.
func (cn *ImportConfigNode) BlockName() string {
	return cn.componentName
}
