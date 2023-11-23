package controller

import (
	"fmt"

	"github.com/grafana/agent/pkg/flow/internal/dag"
)

type nodeTemplates struct {
	parent    *nodeTemplates
	templates map[string]*dag.Graph
}

func NewNodeTemplates(parent *nodeTemplates) nodeTemplates {
	return nodeTemplates{parent: parent, templates: make(map[string]*dag.Graph)}
}

func (n *nodeTemplates) AddTemplate(label string, template *dag.Graph) error {
	if _, exists := n.templates[label]; exists {
		return fmt.Errorf("duplicate template key found: %s, module not added", label)
	}
	n.templates[label] = template
	return nil
}

func (n *nodeTemplates) RetrieveAvailableTemplates() (map[string]*dag.Graph, error) {
	templates := make(map[string]*dag.Graph)

	if n.parent != nil {
		parentTemplates, err := n.parent.RetrieveAvailableTemplates()
		if err != nil {
			return nil, err
		}
		for key, val := range parentTemplates {
			templates[key] = val
		}
	}

	for key, val := range n.templates {
		if _, exists := templates[key]; exists {
			return nil, fmt.Errorf("duplicate template key found: %s, it seems that the same module is declared twice in the same scope", key)
		}
		templates[key] = val
	}
	return templates, nil
}
