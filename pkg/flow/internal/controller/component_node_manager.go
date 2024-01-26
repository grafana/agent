package controller

import (
	"fmt"
	"strings"

	"github.com/grafana/river/ast"
)

// ComponentNodeManager manages component nodes.
type ComponentNodeManager struct {
	importNodes                 map[string]*ImportConfigNode
	declareNodes                map[string]*DeclareNode
	globals                     ComponentGlobals
	componentReg                ComponentRegistry
	customComponentDependencies map[string][]CustomComponentDependency
	additionalDeclareContents   map[string]string
}

// NewComponentNodeManager creates a new ComponentNodeManager.
func NewComponentNodeManager(globals ComponentGlobals, componentReg ComponentRegistry) *ComponentNodeManager {
	return &ComponentNodeManager{
		importNodes:                 map[string]*ImportConfigNode{},
		declareNodes:                map[string]*DeclareNode{},
		customComponentDependencies: map[string][]CustomComponentDependency{},
		globals:                     globals,
		componentReg:                componentReg,
	}
}

// Reload resets the state of the component node manager and stores the provided additionalDeclareContents.
func (m *ComponentNodeManager) Reload(additionalDeclareContents map[string]string) {
	m.additionalDeclareContents = additionalDeclareContents
	m.customComponentDependencies = make(map[string][]CustomComponentDependency)
	m.importNodes = map[string]*ImportConfigNode{}
	m.declareNodes = map[string]*DeclareNode{}
}

// CreateComponentNode creates a new builtin component or a new custom component.
func (m *ComponentNodeManager) CreateComponentNode(componentName string, block *ast.BlockStmt) (ComponentNode, error) {
	if m.isCustomComponent(componentName) {
		return NewCustomComponentNode(m.globals, block, m.GetCustomComponentConfig), nil
	} else {
		registration, exists := m.componentReg.Get(componentName)
		if !exists {
			return nil, fmt.Errorf("unrecognized component name %q", componentName)
		}
		return NewBuiltinComponentNode(m.globals, registration, block), nil
	}
}

// isCustomComponent searches for a declare corresponding to the given component name.
func (m *ComponentNodeManager) isCustomComponent(componentName string) bool {
	namespace := strings.Split(componentName, ".")[0]
	_, declareExists := m.declareNodes[namespace]
	_, importExists := m.importNodes[namespace]
	_, additionalDeclareContentExists := m.additionalDeclareContents[componentName]

	return declareExists || importExists || additionalDeclareContentExists
}

// FindLocalDeclareNode returns the declareNode matching the declareLabel of the provided CustomComponentNode if present.
func (m *ComponentNodeManager) FindLocalDeclareNode(cc *CustomComponentNode) (*DeclareNode, bool) {
	declareNode, exist := m.declareNodes[cc.declareLabel]
	return declareNode, exist
}

// FindImportedConfigNode returns the importNode matching the importLabel of the provided CustomComponentNode if present.
func (m *ComponentNodeManager) FindImportedConfigNode(cc *CustomComponentNode) (*ImportConfigNode, bool) {
	importNode, exist := m.importNodes[cc.importLabel]
	return importNode, exist
}

// CustomComponentConfig represents the config needed by a custom component to load.
type CustomComponentConfig struct {
	declareContent            string            // represents the corresponding declare as plain string
	additionalDeclareContents map[string]string // represents the additional declares that might be needed by the component to build custom components
}

// NewCustomComponentConfig creates a new CustomComponentConfig.
func NewCustomComponentConfig(declareContent string, additionalDeclareContents map[string]string) *CustomComponentConfig {
	return &CustomComponentConfig{
		declareContent:            declareContent,
		additionalDeclareContents: additionalDeclareContents,
	}
}

// GetCustomComponentConfig returns the custom component config for a given custom component or an error if not found.
func (m *ComponentNodeManager) GetCustomComponentConfig(cc *CustomComponentNode) (*CustomComponentConfig, error) {
	var customComponentConfig *CustomComponentConfig
	var err error
	if cc.importLabel == "" {
		customComponentConfig = m.getCustomComponentConfigFromLocalDeclares(cc)
		if customComponentConfig == nil {
			customComponentConfig = m.getCustomComponentConfigFromParent(cc)
		}
	} else {
		customComponentConfig, err = m.getCustomComponentConfigFromImportedDeclares(cc)
		if err != nil {
			return customComponentConfig, err
		}
		if customComponentConfig == nil {
			customComponentConfig = m.getCustomComponentConfigFromParent(cc)
			// Custom components that receive their config from imported declares in a parent controller can only access the imported declares coming from the same import.
			customComponentConfig.additionalDeclareContents = filterAdditionalDeclareContents(cc.importLabel, customComponentConfig.additionalDeclareContents)
		}
	}
	if customComponentConfig == nil {
		return nil, fmt.Errorf("custom component config not found for component %s", cc.componentName)
	}
	return customComponentConfig, nil
}

// getCustomComponentConfigFromLocalDeclares retrieves the config of a custom component from the local declares.
func (m *ComponentNodeManager) getCustomComponentConfigFromLocalDeclares(cc *CustomComponentNode) *CustomComponentConfig {
	node, exists := m.declareNodes[cc.declareLabel]
	if !exists {
		return nil
	}
	return NewCustomComponentConfig(node.Declare().content, m.getLocalAdditionalDeclareContents(cc.componentName))
}

// getCustomComponentConfigFromParent retrieves the config of a custom component from the parent controller.
func (m *ComponentNodeManager) getCustomComponentConfigFromParent(cc *CustomComponentNode) *CustomComponentConfig {
	declareContent, exists := m.additionalDeclareContents[cc.componentName]
	if !exists {
		return nil
	}
	return NewCustomComponentConfig(declareContent, m.additionalDeclareContents)
}

// getImportedCustomComponentConfig retrieves the config of a custom component from the imported declares.
func (m *ComponentNodeManager) getCustomComponentConfigFromImportedDeclares(cc *CustomComponentNode) (*CustomComponentConfig, error) {
	node, exists := m.importNodes[cc.importLabel]
	if !exists {
		return nil, nil
	}
	declare, err := node.GetImportedDeclareByLabel(cc.declareLabel)
	if err != nil {
		return nil, err
	}
	return NewCustomComponentConfig(declare.content, m.getImportAdditionalDeclareContents(node)), nil
}

// getImportAdditionalDeclareContents provides the additional declares that a custom component might need.
func (m *ComponentNodeManager) getImportAdditionalDeclareContents(node *ImportConfigNode) map[string]string {
	additionalDeclareContents := make(map[string]string, len(node.ImportedDeclares()))
	for importedDeclareLabel, importedDeclare := range node.ImportedDeclares() {
		additionalDeclareContents[importedDeclareLabel] = importedDeclare.content
	}
	return additionalDeclareContents
}

// getLocalAdditionalDeclareContents provides the additional declares that a custom component might need.
func (m *ComponentNodeManager) getLocalAdditionalDeclareContents(componentName string) map[string]string {
	additionalDeclareContents := make(map[string]string)
	for _, customComponentDependency := range m.customComponentDependencies[componentName] {
		if customComponentDependency.importNode != nil {
			for importedDeclareLabel, importedDeclare := range customComponentDependency.importNode.ImportedDeclares() {
				// The label of the importNode is added as a prefix to the declare label to create a scope.
				// This is useful in the scenario where a custom component of an imported declare is defined inside of a local declare.
				// In this case, this custom component should only have have access to the imported declares of its corresponding import node.
				additionalDeclareContents[customComponentDependency.importNode.label+"."+importedDeclareLabel] = importedDeclare.content
			}
		} else if customComponentDependency.declareNode != nil {
			additionalDeclareContents[customComponentDependency.declareLabel] = customComponentDependency.declareNode.Declare().content
		} else {
			additionalDeclareContents[customComponentDependency.componentName] = m.additionalDeclareContents[customComponentDependency.componentName]
		}
	}
	return additionalDeclareContents
}

// filterAdditionalDeclareContents prevents custom components from accessing declared content out of their scope.
func filterAdditionalDeclareContents(importLabel string, additionalDeclareContents map[string]string) map[string]string {
	filteredAdditionalDeclareContents := make(map[string]string)
	for declareLabel, declareContent := range additionalDeclareContents {
		// The scope is defined by the importLabel prefix in the declareLabel of the declare block.
		if strings.HasPrefix(declareLabel, importLabel) {
			filteredAdditionalDeclareContents[strings.TrimPrefix(declareLabel, importLabel+".")] = declareContent
		}
	}
	return filteredAdditionalDeclareContents
}

// CustomComponentDependency represents a dependency that a custom component has to a declare block.
type CustomComponentDependency struct {
	componentName string
	importLabel   string
	declareLabel  string
	importNode    *ImportConfigNode
	declareNode   *DeclareNode
}

// ComputeCustomComponentDependencies retrieves and caches the dependencies that declare might have to other declares.
func (m *ComponentNodeManager) ComputeCustomComponentDependencies(declareNode *DeclareNode) ([]CustomComponentDependency, error) {
	var dependencies []CustomComponentDependency
	var err error
	// If the dependencies of the declare were already compute, retrieve them from the cache. This is useful if you have several instances of the same custom component.
	if deps, ok := m.customComponentDependencies[declareNode.label]; ok {
		dependencies = deps
	} else {
		dependencies, err = m.findCustomComponentDependencies(declareNode.Declare())
		if err != nil {
			return nil, err
		}
		m.customComponentDependencies[declareNode.label] = dependencies
	}
	return dependencies, nil
}

// findCustomComponentDependencies traverses the AST of the provided declare and collects references to known custom components.
// Panics if declare is nil.
func (m *ComponentNodeManager) findCustomComponentDependencies(declare *Declare) ([]CustomComponentDependency, error) {
	uniqueReferences := make(map[string]CustomComponentDependency)
	m.collectCustomComponentDependencies(declare.block.Body, uniqueReferences)

	references := make([]CustomComponentDependency, 0, len(uniqueReferences))
	for _, ref := range uniqueReferences {
		references = append(references, ref)
	}

	return references, nil
}

// collectCustomComponentDependencies collects recursively references to custom components through an AST body.
func (m *ComponentNodeManager) collectCustomComponentDependencies(stmts ast.Body, uniqueReferences map[string]CustomComponentDependency) {
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *ast.BlockStmt:
			componentName := strings.Join(stmt.Name, ".")
			switch componentName {
			case "declare":
				m.collectCustomComponentDependencies(stmt.Body, uniqueReferences)
			default:
				potentialImportLabel, potentialDeclareLabel := ExtractImportAndDeclareLabels(componentName)
				if declareNode, ok := m.declareNodes[potentialDeclareLabel]; ok {
					uniqueReferences[componentName] = CustomComponentDependency{componentName: componentName, importLabel: "", declareLabel: potentialDeclareLabel, declareNode: declareNode}
				} else if importNode, ok := m.importNodes[potentialImportLabel]; ok {
					uniqueReferences[componentName] = CustomComponentDependency{componentName: componentName, importLabel: potentialImportLabel, declareLabel: potentialDeclareLabel, importNode: importNode}
				} else if _, ok := m.additionalDeclareContents[componentName]; ok {
					uniqueReferences[componentName] = CustomComponentDependency{componentName: componentName, importLabel: potentialImportLabel, declareLabel: potentialDeclareLabel}
				}
			}
		}
	}
}
