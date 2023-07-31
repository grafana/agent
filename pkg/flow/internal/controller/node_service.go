package controller

import (
	"context"

	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/vm"
	"github.com/grafana/agent/service"
)

// ServiceNode is a Flow DAG node which represents a running service.
type ServiceNode struct {
	host service.Host
	svc  service.Service
	def  service.Definition
}

var (
	_ BlockNode    = (*ServiceNode)(nil)
	_ RunnableNode = (*ServiceNode)(nil)
)

// NewServiceNode creates a new instance of a ServiceNode from an instance of a
// Service. The provided host is used when running the service.
func NewServiceNode(host service.Host, svc service.Service) *ServiceNode {
	return &ServiceNode{
		host: host,
		svc:  svc,
		def:  svc.Definition(),
	}
}

// Definition returns the service definition associated with the node.
func (sn *ServiceNode) Definition() service.Definition { return sn.def }

// NodeID returns the ID of the ServiceNode, which is equal to the service's
// name.
func (sn *ServiceNode) NodeID() string { return sn.def.Name }

// Block implements BlockNode. It returns nil, since ServiceNodes don't have
// associated configs.
func (sn *ServiceNode) Block() *ast.BlockStmt {
	// TODO(rfratto): support configs for services.
	return nil
}

// Evaluate implements BlockNode. It is a no-op since ServiceNodes don't have
// associated configs to evaluate.
func (sn *ServiceNode) Evaluate(scope *vm.Scope) error {
	// TODO(rfratto): support configs for services.
	return nil
}

func (sn *ServiceNode) Run(ctx context.Context) error {
	return sn.svc.Run(ctx, sn.host)
}
