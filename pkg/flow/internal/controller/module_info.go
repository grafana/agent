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
	fullName string,
	scopedName string,
) (ModuleInfo, error) {

	var moduleInfo ModuleInfo
	var content string
	var err error

	if node, exists := declareNodes[scopedName]; exists {
		moduleInfo.moduleDependencies, err = getLocalModuleDependencies(fullName, moduleDependencies, parentModuleDependencies)
		if err != nil {
			return moduleInfo, err
		}

		content, err = node.ModuleContent(scopedName)
		if err != nil {
			return moduleInfo, err
		}
	} else if c, ok := parentModuleDependencies[fullName]; ok {
		content = c
	} else {
		return moduleInfo, fmt.Errorf("could not find the module declaration in declareNodes")
	}
	moduleInfo.content = content
	return moduleInfo, nil
}

// getLocalModuleDependencies provides the dependencies needed for nested local modules.
func getLocalModuleDependencies(fullName string,
	localModuleDependencies map[string][]ModuleReference,
	parentModuleDependencies map[string]string,
) (map[string]string, error) {

	var dependency string
	var err error
	moduleDependencies := make(map[string]string)
	for _, moduleDependency := range localModuleDependencies[fullName] {
		if moduleDependency.moduleContentProvider != nil {
			dependency, err = moduleDependency.moduleContentProvider.ModuleContent(moduleDependency.scopedName)
			if err != nil {
				return moduleDependencies, err
			}
		} else {
			if c, ok := parentModuleDependencies[moduleDependency.fullName]; ok {
				dependency = c
			} else {
				return moduleDependencies, fmt.Errorf("could not find the dependency in parentModuleDependencies")
			}
		}
		moduleDependencies[moduleDependency.fullName] = dependency
	}
	return moduleDependencies, nil
}

func getImportedModuleInfo(
	importNodes map[string]*ImportConfigNode,
	parentModuleDependencies map[string]string,
	fullName string,
	scopedName string,
	namespace string,
) (ModuleInfo, error) {

	var moduleInfo ModuleInfo
	var content string
	var err error
	if node, exists := importNodes[namespace]; exists {
		moduleInfo.moduleDependencies = getImportedModuleDependencies(node, scopedName)
		content, err = node.ModuleContent(scopedName)
		if err != nil {
			return moduleInfo, err
		}
	} else if c, ok := parentModuleDependencies[fullName]; ok {
		content = c
	} else {
		return moduleInfo, fmt.Errorf("could not find the module declaration in importNodes")
	}
	moduleInfo.content = content
	return moduleInfo, nil
}

// getImportedModuleDependencies provides the dependencies needed for nested imported modules.
// Concrete example: Let's say that we have an import "A" and we have a declare component node A.B.C.D.e "bla"
// "bla" is an instance of a declare "e" within the content imported by "D".
// Let's say that there is an instance of a declare "f" inside of "e" which is imported by "E".
// The content of this declare needs to be passed via the moduleDependencies.
// "e"'s scope is limited to everything after "D" in the chain, that's why we need to check the prefixes.
// The scopedName is B.C.D.e, the scope is B.C.D.
// "f" is currently imported via B.C.D.E, that means that it is within "e"'s scope.
// When we pass the dependency, we need to trim the scope so that "e" can resolve it properly.
func getImportedModuleDependencies(node *ImportConfigNode, scopedName string) map[string]string {
	moduleDependencies := make(map[string]string)
	lastIndex := strings.LastIndex(scopedName, ".")
	if lastIndex != -1 {
		scope := scopedName[:lastIndex]
		for importedMod, importedModContent := range node.importedContent {
			if strings.HasPrefix(importedMod, scope) {
				moduleDependencies[strings.TrimPrefix(importedMod, scope+".")] = importedModContent
			}
		}
	} else {
		// In this case the declare is only at depth 1 which corresponds to the importedContent, so we can just pass everything.
		moduleDependencies = node.importedContent
	}
	return moduleDependencies
}
