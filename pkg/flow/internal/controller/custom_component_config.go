package controller

import (
	"fmt"
	"strings"
)

type CustomComponentConfig struct {
	declareContent            string
	additionalDeclareContents map[string]string
}

func getLocalCustomComponentConfig(
	declareNodes map[string]*DeclareNode,
	customComponentDependencies map[string][]CustomComponentDependency,
	parentDeclareContents map[string]string,
	componentName string,
	declareLabel string,
) (CustomComponentConfig, error) {

	var customComponentConfig CustomComponentConfig
	var err error

	if node, exists := declareNodes[declareLabel]; exists {
		customComponentConfig.additionalDeclareContents, err = getLocalAdditionalDeclareContents(componentName, customComponentDependencies, parentDeclareContents)
		if err != nil {
			return customComponentConfig, err
		}
		customComponentConfig.declareContent = node.Declare().Content
	} else if declareContent, ok := parentDeclareContents[componentName]; ok {
		customComponentConfig.additionalDeclareContents = parentDeclareContents
		customComponentConfig.declareContent = declareContent
	} else {
		return customComponentConfig, fmt.Errorf("could not find a corresponding declare for the custom component %s", componentName)
	}
	return customComponentConfig, nil
}

func getLocalAdditionalDeclareContents(componentName string,
	customComponentDependencies map[string][]CustomComponentDependency,
	parentDeclareContents map[string]string,
) (map[string]string, error) {

	additionalDeclareContents := make(map[string]string)
	for _, customComponentDependency := range customComponentDependencies[componentName] {
		if customComponentDependency.importNode != nil {
			for importedDeclareLabel, importedDeclare := range customComponentDependency.importNode.ImportedDeclares() {
				// The label of the importNode is added as a prefix to the declare label to create a scope.
				// This is useful in the scenario where a custom component of an imported declare is defined inside of a local declare.
				// In this case, this custom component should only have have access to the imported declares of its corresponding import node.
				additionalDeclareContents[customComponentDependency.importNode.label+"."+importedDeclareLabel] = importedDeclare.Content
			}
		} else if customComponentDependency.declareNode != nil {
			additionalDeclareContents[customComponentDependency.declareLabel] = customComponentDependency.declareNode.Declare().Content
		} else {
			// Nested declares have access to declare contents defined in their parents.
			if declareContent, ok := parentDeclareContents[customComponentDependency.componentName]; ok {
				additionalDeclareContents[customComponentDependency.componentName] = declareContent
			} else {
				return additionalDeclareContents, fmt.Errorf("could not find the required declare content %s for the custom component %s", customComponentDependency.componentName, componentName)
			}
		}
	}
	return additionalDeclareContents, nil
}

func getImportedCustomComponentConfig(
	importNodes map[string]*ImportConfigNode,
	parentDeclareContents map[string]string,
	componentName string,
	declareLabel string,
	importLabel string,
) (CustomComponentConfig, error) {

	var customComponentConfig CustomComponentConfig
	if node, exists := importNodes[importLabel]; exists {
		customComponentConfig.additionalDeclareContents = make(map[string]string, len(node.ImportedDeclares()))
		for importedDeclareLabel, importedDeclare := range node.ImportedDeclares() {
			customComponentConfig.additionalDeclareContents[importedDeclareLabel] = importedDeclare.Content
		}
		declare, err := node.GetImportedDeclareByLabel(declareLabel)
		if err != nil {
			return customComponentConfig, err
		}
		customComponentConfig.declareContent = declare.Content
	} else if declareContent, ok := parentDeclareContents[componentName]; ok {
		customComponentConfig.additionalDeclareContents = filterParentDeclareContents(importLabel, parentDeclareContents)
		customComponentConfig.declareContent = declareContent
	} else {
		return customComponentConfig, fmt.Errorf("could not find a corresponding imported declare for the custom component %s", componentName)
	}
	return customComponentConfig, nil
}

// filterParentDeclareContents prevents custom components from accessing declared content out of their scope.
func filterParentDeclareContents(importLabel string, parentDeclareContents map[string]string) map[string]string {
	filteredParentDeclareContents := make(map[string]string)
	for declareLabel, declareContent := range parentDeclareContents {
		// The scope is defined by the importLabel prefix in the declareLabel of the declare block.
		if strings.HasPrefix(declareLabel, importLabel) {
			filteredParentDeclareContents[strings.TrimPrefix(declareLabel, importLabel+".")] = declareContent
		}
	}
	return filteredParentDeclareContents
}
