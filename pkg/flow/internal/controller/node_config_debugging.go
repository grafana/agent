package controller

import (
	"fmt"
	"net"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/vm"
)

type DebuggingConfigNode struct {
	nodeID        string
	componentName string

	mut           sync.RWMutex
	block         *ast.BlockStmt // Current River blocks to derive config from
	eval          *vm.Evaluator
	componentIds  []string
	conn          net.Conn
	serverRunning bool
	shouldDebug   bool
	stopServerCh  chan struct{}
	dataCh        chan string
}

var _ BlockNode = (*DebuggingConfigNode)(nil)

const DEBUGGING_PORT = 12340

// TODO: IMPLEMENT PROPER LOGGING

// NewDebuggingConfigNode creates a new ArgumentConfigNode from an initial ast.BlockStmt.
// The underlying config isn't applied until Evaluate is called.
func NewDebuggingConfigNode(block *ast.BlockStmt, globals ComponentGlobals) *DebuggingConfigNode {
	return &DebuggingConfigNode{
		nodeID:        BlockComponentID(block).String(),
		componentName: block.GetBlockName(),

		block: block,
		eval:  vm.New(block.Body),
	}
}

func NewDefaultDebuggingConfigNode() *DebuggingConfigNode {
	return &DebuggingConfigNode{
		nodeID:        debuggingBlockID,
		componentName: debuggingBlockID,

		block: nil,
		eval:  nil,
	}
}

type debuggingConfigBlock struct {
	ComponentIds []string `river:"components,attr,optional"`
}

// Evaluate implements BlockNode and updates the arguments for the managed config block
// by re-evaluating its River block with the provided scope. The managed config block
// will be built the first time Evaluate is called.
//
// Evaluate will return an error if the River block cannot be evaluated or if
// decoding to arguments fails.
func (cn *DebuggingConfigNode) Evaluate(scope *vm.Scope) error {
	cn.mut.Lock()
	defer cn.mut.Unlock()

	var argument debuggingConfigBlock
	if err := cn.eval.Evaluate(scope, &argument); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	cn.componentIds = argument.ComponentIds

	return nil
}

func (cn *DebuggingConfigNode) StartTCPServer() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", DEBUGGING_PORT))
	if err != nil {
		fmt.Println("failed to start server:", err)
		return
	}
	defer ln.Close()

	cn.serverRunning = true
	fmt.Println("Debug server listening on port", DEBUGGING_PORT)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				if cn.serverRunning {
					fmt.Println("Accept error:", err)
				}
				return
			}
			cn.conn = conn
		}
	}()

	for {
		select {
		case <-cn.stopServerCh:
			if cn.conn != nil {
				cn.conn.Close()
			}
			return
		case data := <-cn.dataCh:
			if cn.conn != nil {
				fmt.Fprintln(cn.conn, data)
			}
		}
	}
}

func (cn *DebuggingConfigNode) StartDebugging(componentNodes []*ComponentNode) {
	cn.setupDebugging(componentNodes)
	cn.manageServer()
}

func (cn *DebuggingConfigNode) setupDebugging(componentNodes []*ComponentNode) {
	for _, compNode := range componentNodes {
		if cn.isTargetComponent(compNode) {
			cn.hookComponentDebugStream(compNode)
		} else {
			cn.hookNoOp(compNode)
		}
	}
}

func (cn *DebuggingConfigNode) isTargetComponent(compNode *ComponentNode) bool {
	for _, target := range cn.componentIds {
		if compNode.ID().String() == target {
			return true
		}
	}
	return false
}

func (cn *DebuggingConfigNode) hookComponentDebugStream(compNode *ComponentNode) {
	if debuggable, ok := interface{}(compNode.managed).(component.DebugStream); ok {
		debuggable.HookDebugStream(func(data string) {
			cn.dataCh <- data
		})
		cn.shouldDebug = true
	} else {
		fmt.Printf("Component %s does not support debugging\n", compNode.ID().String())
	}
}

func (cn *DebuggingConfigNode) hookNoOp(compNode *ComponentNode) {
	if debuggable, ok := interface{}(compNode.managed).(component.DebugStream); ok {
		debuggable.HookDebugStream(func(data string) {
			// no-op
		})
	}
}

func (cn *DebuggingConfigNode) manageServer() {
	if cn.shouldDebug && !cn.serverRunning {
		cn.startServer()
	} else if !cn.shouldDebug && cn.serverRunning {
		cn.stopServer()
	}
}

func (cn *DebuggingConfigNode) startServer() {
	cn.stopServerCh = make(chan struct{})
	cn.dataCh = make(chan string, 1000) // Adjust buffer size as needed
	cn.serverRunning = true
	go cn.StartTCPServer()
}

func (cn *DebuggingConfigNode) stopServer() {
	close(cn.stopServerCh)
	cn.serverRunning = false
	if cn.conn != nil {
		cn.conn.Close()
	}
	close(cn.dataCh)
}

// Block implements BlockNode and returns the current block of the managed config node.
func (cn *DebuggingConfigNode) Block() *ast.BlockStmt {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return cn.block
}

// NodeID implements dag.Node and returns the unique ID for the config node.
func (cn *DebuggingConfigNode) NodeID() string { return cn.nodeID }
