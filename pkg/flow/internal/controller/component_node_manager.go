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
		return NewCustomComponentNode(m.globals, block, m.getCustomComponentConfig), nil
	} else {
		registration, exists := m.componentReg.Get(componentName)
		if !exists {
			return nil, fmt.Errorf("unrecognized component name %q", componentName)
		}
		return NewBuiltinComponentNode(m.globals, registration, block), nil
	}
}

// GetCustomComponentDependencies retrieves and caches the dependencies that declare might have to other declares.
func (m *ComponentNodeManager) getCustomComponentDependencies(declareNode *DeclareNode) ([]CustomComponentDependency, error) {
	var dependencies []CustomComponentDependency
	var err error
	if deps, ok := m.customComponentDependencies[declareNode.label]; ok {
		dependencies = deps
	} else {
		dependencies, err = m.FindCustomComponentDependencies(declareNode.Declare())
		if err != nil {
			return nil, err
		}
		m.customComponentDependencies[declareNode.label] = dependencies
	}
	return dependencies, nil
}

// isCustomComponent searches for a declare corresponding to the given component name.
func (m *ComponentNodeManager) isCustomComponent(componentName string) bool {
	namespace := strings.Split(componentName, ".")[0]
	_, declareExists := m.declareNodes[namespace]
	_, importExists := m.importNodes[namespace]
	_, additionalDeclareContentExists := m.additionalDeclareContents[componentName]

	return declareExists || importExists || additionalDeclareContentExists
}

// GetCorrespondingLocalDeclare returns the declareNode matching the declareLabel of the provided CustomComponentNode if present.
func (m *ComponentNodeManager) GetCorrespondingLocalDeclare(cc *CustomComponentNode) (*DeclareNode, bool) {
	declareNode, exist := m.declareNodes[cc.declareLabel]
	return declareNode, exist
}

// GetCorrespondingImportedDeclare returns the importNode matching the importLabel of the provided CustomComponentNode if present.
func (m *ComponentNodeManager) GetCorrespondingImportedDeclare(cc *CustomComponentNode) (*ImportConfigNode, bool) {
	importNode, exist := m.importNodes[cc.importLabel]
	return importNode, exist
}

// CustomComponentConfig represents the config needed by a custom component to load.
type CustomComponentConfig struct {
	declareContent            string            // represents the corresponding declare as plain string
	additionalDeclareContents map[string]string // represents the additional declare that might be needed by the component to build custom components
}

// getCustomComponentConfig returns the custom component config for a given custom component.
func (m *ComponentNodeManager) getCustomComponentConfig(cc *CustomComponentNode) (CustomComponentConfig, error) {
	var customComponentConfig CustomComponentConfig
	var found bool
	var err error
	if cc.importLabel == "" {
		customComponentConfig, found = m.getCustomComponentConfigFromLocalDeclares(cc)
		if !found {
			customComponentConfig, found = m.getCustomComponentConfigFromParent(cc)
		}
	} else {
		customComponentConfig, found, err = m.getCustomComponentConfigFromImportedDeclares(cc)
		if err != nil {
			return customComponentConfig, err
		}
		if !found {
			customComponentConfig, found = m.getCustomComponentConfigFromParent(cc)
			// Custom components that receive their config from imported declares in a parent controller can only access the imported declares coming from the same import.
			customComponentConfig.additionalDeclareContents = filterAdditionalDeclareContents(cc.importLabel, customComponentConfig.additionalDeclareContents)
		}
	}
	if !found {
		return customComponentConfig, fmt.Errorf("custom component config not found for component %s", cc.componentName)
	}
	return customComponentConfig, nil
}

// getCustomComponentConfigFromLocalDeclares retrieves the config of a custom component from the local declares.
func (m *ComponentNodeManager) getCustomComponentConfigFromLocalDeclares(cc *CustomComponentNode) (CustomComponentConfig, bool) {
	node, exists := m.declareNodes[cc.declareLabel]
	if !exists {
		return CustomComponentConfig{}, false
	}
	return CustomComponentConfig{
		declareContent:            node.Declare().content,
		additionalDeclareContents: m.getLocalAdditionalDeclareContents(cc.componentName),
	}, true
}

// getCustomComponentConfigFromParent retrieves the config of a custom component from the parent controller.
func (m *ComponentNodeManager) getCustomComponentConfigFromParent(cc *CustomComponentNode) (CustomComponentConfig, bool) {
	declareContent, exists := m.additionalDeclareContents[cc.componentName]
	if !exists {
		return CustomComponentConfig{}, false
	}
	return CustomComponentConfig{
		declareContent:            declareContent,
		additionalDeclareContents: m.additionalDeclareContents,
	}, true
}

// getImportedCustomComponentConfig retrieves the config of a custom component from the imported declares.
func (m *ComponentNodeManager) getCustomComponentConfigFromImportedDeclares(cc *CustomComponentNode) (CustomComponentConfig, bool, error) {
	node, exists := m.importNodes[cc.importLabel]
	if !exists {
		return CustomComponentConfig{}, false, nil
	}
	declare, err := node.GetImportedDeclareByLabel(cc.declareLabel)
	if err != nil {
		return CustomComponentConfig{}, false, err
	}
	return CustomComponentConfig{
		declareContent:            declare.content,
		additionalDeclareContents: m.getImportAdditionalDeclareContents(node),
	}, true, nil
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

// FindCustomComponentDependencies traverses the AST of the provided declare and collects references to known custom components.
// Panics if declare is nil.
func (m *ComponentNodeManager) FindCustomComponentDependencies(declare *Declare) ([]CustomComponentDependency, error) {
	uniqueReferences := make(map[string]CustomComponentDependency)
	m.findCustomComponentDependencies(declare.block.Body, uniqueReferences)

	references := make([]CustomComponentDependency, 0, len(uniqueReferences))
	for _, ref := range uniqueReferences {
		references = append(references, ref)
	}

	return references, nil
}

func (m *ComponentNodeManager) findCustomComponentDependencies(stmts ast.Body, uniqueReferences map[string]CustomComponentDependency) {
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *ast.BlockStmt:
			componentName := strings.Join(stmt.Name, ".")
			switch componentName {
			case "declare":
				m.findCustomComponentDependencies(stmt.Body, uniqueReferences)
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
