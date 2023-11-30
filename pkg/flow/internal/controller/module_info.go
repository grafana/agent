package controller

import (
	"fmt"
	"strings"
)

type ModuleInfo struct {
	content           string
	moduleDefinitions map[string]string
}

func getLocalModuleInfo(
	declareNodes map[string]*DeclareNode,
	moduleReferences map[string][]ModuleReference,
	parentModuleDefinitions map[string]string,
	componentName string,
	declareLabel string,
) (ModuleInfo, error) {

	var moduleInfo ModuleInfo
	var content string
	var err error

	if node, exists := declareNodes[declareLabel]; exists {
		moduleInfo.moduleDefinitions, err = getLocalModuleDefinitions(componentName, moduleReferences, parentModuleDefinitions)
		if err != nil {
			return moduleInfo, err
		}

		content, err = node.ModuleContent()
		if err != nil {
			return moduleInfo, err
		}
	} else if c, ok := parentModuleDefinitions[componentName]; ok {
		content = c
		moduleInfo.moduleDefinitions = parentModuleDefinitions
	} else {
		return moduleInfo, fmt.Errorf("could not find a definition for the declared module %s", componentName)
	}
	moduleInfo.content = content
	return moduleInfo, nil
}

func getLocalModuleDefinitions(componentName string,
	localModuleReferences map[string][]ModuleReference,
	parentModuleDefinitions map[string]string,
) (map[string]string, error) {

	moduleReferences := make(map[string]string)
	for _, moduleDependency := range localModuleReferences[componentName] {
		if moduleDependency.importNode != nil {
			for importModulePath, importModuleContent := range moduleDependency.importNode.importedDeclares {
				moduleReferences[moduleDependency.importNode.label+"."+importModulePath] = importModuleContent
			}
		} else if moduleDependency.declareNode != nil {
			ref, err := moduleDependency.declareNode.ModuleContent()
			if err != nil {
				return moduleReferences, nil
			}
			moduleReferences[moduleDependency.declareLabel] = ref
		} else {
			// Nested declares have access to their parents module definitions.
			if c, ok := parentModuleDefinitions[moduleDependency.componentName]; ok {
				moduleReferences[moduleDependency.componentName] = c
			} else {
				return moduleReferences, fmt.Errorf("could not find the required module dependency %s for the module %s", moduleDependency.componentName, componentName)
			}
		}
	}
	return moduleReferences, nil
}

func getImportedModuleInfo(
	importNodes map[string]*ImportConfigNode,
	parentModuleDefinitions map[string]string,
	componentName string,
	declareLabel string,
	importLabel string,
) (ModuleInfo, error) {

	var moduleInfo ModuleInfo
	var content string
	var err error
	if node, exists := importNodes[importLabel]; exists {
		moduleInfo.moduleDefinitions = node.importedDeclares
		content, err = node.ModuleContent(declareLabel)
		if err != nil {
			return moduleInfo, err
		}
	} else if c, ok := parentModuleDefinitions[componentName]; ok {
		content = c
		moduleInfo.moduleDefinitions = filterParentModuleDefinitions(importLabel, parentModuleDefinitions)
	} else {
		return moduleInfo, fmt.Errorf("could not find a definition for the imported module %s", componentName)
	}
	moduleInfo.content = content
	return moduleInfo, nil
}

// filterParentModuleDefinitions prevents modules from accessing other module definitions which are not in their scope.
func filterParentModuleDefinitions(importLabel string, parentModuleDefinitions map[string]string) map[string]string {
	filteredParentModuleDefinitions := make(map[string]string)
	for importPath, content := range parentModuleDefinitions {
		// The scope is defined by the importLabel prefix in the importPath of the modules.
		if strings.HasPrefix(importPath, importLabel) {
			filteredParentModuleDefinitions[strings.TrimPrefix(importPath, importLabel+".")] = content
		}
	}
	return filteredParentModuleDefinitions
}
