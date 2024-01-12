package controller

import (
	"fmt"
	"strings"
)

type ModuleInfo struct {
	content            string
	moduleDependencies map[string]string
}

func getLocalModuleInfo(
	declareNodes map[string]*DeclareNode,
	moduleDependencies map[string][]ModuleReference,
	parentModuleDependencies map[string]string,
	componentName string,
	declareLabel string,
) (ModuleInfo, error) {

	var moduleInfo ModuleInfo
	var content string
	var err error

	if node, exists := declareNodes[declareLabel]; exists {
		moduleInfo.moduleDependencies, err = getLocalModuleDependencies(componentName, moduleDependencies, parentModuleDependencies)
		if err != nil {
			return moduleInfo, err
		}

		content, err = node.ModuleContent(declareLabel)
		if err != nil {
			return moduleInfo, err
		}
	} else if c, ok := parentModuleDependencies[componentName]; ok {
		content = c
		moduleInfo.moduleDependencies = parentModuleDependencies
	} else {
		return moduleInfo, fmt.Errorf("could not find the module declaration in declareNodes")
	}
	moduleInfo.content = content
	return moduleInfo, nil
}

func getLocalModuleDependencies(componentName string,
	localModuleDependencies map[string][]ModuleReference,
	parentModuleDependencies map[string]string,
) (map[string]string, error) {

	var err error
	moduleDependencies := make(map[string]string)
	for _, moduleDependency := range localModuleDependencies[componentName] {
		if moduleDependency.moduleContentProvider != nil {

			switch n := moduleDependency.moduleContentProvider.(type) {
			case *ImportConfigNode:
				for importModulePath, importModuleContent := range n.importedDeclares {
					moduleDependencies[n.label+"."+importModulePath] = importModuleContent
				}
			case *DeclareNode:
				moduleDependencies[moduleDependency.declareLabel], err = n.ModuleContent(moduleDependency.declareLabel)
				if err != nil {
					return moduleDependencies, nil
				}
			}
		} else {
			// Nested declares have access to their parents module definitions.
			if c, ok := parentModuleDependencies[moduleDependency.componentName]; ok {
				moduleDependencies[moduleDependency.componentName] = c
			} else {
				return moduleDependencies, fmt.Errorf("could not find the dependency in parentModuleDependencies")
			}
		}
	}
	return moduleDependencies, nil
}

func getImportedModuleInfo(
	importNodes map[string]*ImportConfigNode,
	parentModuleDependencies map[string]string,
	componentName string,
	declareLabel string,
	importLabel string,
) (ModuleInfo, error) {

	var moduleInfo ModuleInfo
	var content string
	var err error
	if node, exists := importNodes[importLabel]; exists {
		moduleInfo.moduleDependencies = node.importedDeclares
		content, err = node.ModuleContent(declareLabel)
		if err != nil {
			return moduleInfo, err
		}
	} else if c, ok := parentModuleDependencies[componentName]; ok {
		content = c
		moduleInfo.moduleDependencies = filterParentModuleDependencies(importLabel, parentModuleDependencies)
	} else {
		return moduleInfo, fmt.Errorf("could not find the module declaration in importNodes")
	}
	moduleInfo.content = content
	return moduleInfo, nil
}

func filterParentModuleDependencies(importLabel string, parentModuleDependencies map[string]string) map[string]string {
	filteredParentDependencies := make(map[string]string)
	for importPath, content := range parentModuleDependencies {
		if strings.HasPrefix(importPath, importLabel) {
			filteredParentDependencies[strings.TrimPrefix(importPath, importLabel+".")] = content
		}
	}
	return filteredParentDependencies
}
