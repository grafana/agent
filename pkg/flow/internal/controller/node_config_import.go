package controller

import (
	"fmt"

	"github.com/grafana/agent/pkg/flow/internal/dag"
)

// TODO: This struct is an empty shell for now, its implementation will come via another PR.

type ImportConfigNode struct {
	label  string
	nodeID string
}

var _ dag.Node = (*ImportConfigNode)(nil)

// NodeID implements dag.Node and returns the unique ID for this node. The
// NodeID is the string representation of the component's ID from its River
// block.
func (cn *ImportConfigNode) NodeID() string { return cn.nodeID }

// ImportedDeclares returns all declare blocks that it imported.
func (cn *ImportConfigNode) ImportedDeclares() map[string]*Declare {
	return nil
}

// GetImportedDeclareByLabel returns a declare block imported by the node.
func (cn *ImportConfigNode) GetImportedDeclareByLabel(declareLabel string) (*Declare, error) {
	return nil, fmt.Errorf("declareLabel %s not found in imported node %s", declareLabel, cn.label)
}
