package generator

import (
	"bytes"
	"fmt"
	"os"

	"github.com/grafana/agent/component"
	"golang.org/x/exp/slices"
)

const (
	startDelimiter = "<!-- START GENERATED COMPATIBLE COMPONENTS -->"
	endDelimiter   = "<!-- END GENERATED COMPATIBLE COMPONENTS -->"
)

func WriteCompatibleComponentsSection(componentName string) error {
	filePath := pathToComponentMarkdown(componentName)
	newSection, err := GenerateCompatibleComponentsSection(componentName)
	if err != nil {
		return err
	}
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	startMarker := startMarkerBytes()
	endMarker := endMarkerBytes()
	replacement := append(append(startMarker, []byte(newSection)...), endMarker...)

	startIndex := bytes.Index(fileContents, startMarker)
	endIndex := bytes.Index(fileContents, endMarker)
	var newFileContents []byte
	if startIndex == -1 || endIndex == -1 {
		// Append the new section to the end of the file
		newFileContents = append(fileContents, replacement...)
	} else {
		// Replace the section with the new content
		newFileContents = append(fileContents[:startIndex], replacement...)
		newFileContents = append(newFileContents, fileContents[endIndex+len(endMarker):]...)
	}

	err = os.WriteFile(filePath, newFileContents, 0644)
	return err
}

func ReadCompatibleComponentsSection(componentName string) (string, error) {
	filePath := pathToComponentMarkdown(componentName)
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	startMarker := startMarkerBytes()
	endMarker := endMarkerBytes()
	startIndex := bytes.Index(fileContents, startMarker)
	endIndex := bytes.Index(fileContents, endMarker)
	if startIndex == -1 || endIndex == -1 {
		return "", fmt.Errorf("compatible components section not found in %q", filePath)
	}

	return string(fileContents[startIndex+len(startMarker) : endIndex]), nil
}

func GenerateCompatibleComponentsSection(componentName string) (string, error) {
	c, ok := component.Get(componentName)
	if !ok {
		return "", fmt.Errorf("component %q not found", componentName)
	}
	if c.Metadata.Empty() {
		return "", nil
	}

	heading := "\n## Compatible components\n\n"
	acceptingSection := acceptingComponentsSection(componentName, c)
	outputSection := outputComponentsSection(componentName, c)

	if acceptingSection == "" && outputSection == "" {
		return "", nil
	}

	note := "\nNote that connecting some components may not be feasible or components may require further " +
		"configuration to make the connection work correctly. " +
		"Please refer to the linked documentation for more details.\n\n"

	return heading + acceptingSection + outputSection + note, nil
}

func allComponentsThat(f func(registration component.Registration) bool) []string {
	var result []string
	for _, name := range component.AllNames() {
		c, ok := component.Get(name)
		if !ok {
			continue
		}

		if f(c) {
			result = append(result, name)
		}
	}
	return result
}

func allComponentsThatOutput(dataType component.DataType) []string {
	return allComponentsThat(func(reg component.Registration) bool {
		return slices.Contains(reg.Metadata.Outputs, dataType)
	})
}

func allComponentsThatAccept(dataType component.DataType) []string {
	return allComponentsThat(func(reg component.Registration) bool {
		return slices.Contains(reg.Metadata.Accepts, dataType)
	})
}

func outputComponentsSection(name string, c component.Registration) string {
	section := ""
	for _, outputDataType := range c.Metadata.Outputs {
		if list := listOfComponentsAccepting(outputDataType); list != "" {
			section += fmt.Sprintf("- Components that accept %s:\n", outputDataType) + list
		}
	}
	if section != "" {
		section = fmt.Sprintf("`%s` can output data to the following components:\n\n", name) + section
	}
	return section
}

func acceptingComponentsSection(componentName string, c component.Registration) string {
	section := ""
	for _, acceptedDataType := range c.Metadata.Accepts {
		if list := listOfComponentsOutputting(acceptedDataType); list != "" {
			section += fmt.Sprintf("- Components that output %s:\n", acceptedDataType) + list
		}
	}
	if section != "" {
		section = fmt.Sprintf("`%s` can accept data from the following components:\n\n", componentName) + section + "\n"
	}
	return section
}

func listOfComponentsAccepting(dataType component.DataType) string {
	return listOfLinksToComponents(allComponentsThatAccept(dataType))
}

func listOfComponentsOutputting(dataType component.DataType) string {
	return listOfLinksToComponents(allComponentsThatOutput(dataType))
}

func listOfLinksToComponents(components []string) string {
	str := ""
	for _, comp := range components {
		str += fmt.Sprintf("  - [`%[1]s`]({{< relref \"../components/%[1]s.md\" >}})\n", comp)
	}
	return str
}

func pathToComponentMarkdown(name string) string {
	return fmt.Sprintf("sources/flow/reference/components/%s.md", name)
}

func endMarkerBytes() []byte {
	return []byte(endDelimiter + "\n")
}

func startMarkerBytes() []byte {
	return []byte(startDelimiter + "\n")
}
