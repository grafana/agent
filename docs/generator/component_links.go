package generator

import (
	"bytes"
	"fmt"
	"github.com/grafana/agent/component/metadata"
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
	meta, err := metadata.ForComponent(componentName)
	if err != nil {
		return "", err
	}
	if meta.Empty() {
		return "", nil
	}

	heading := "\n## Compatible components\n\n"
	acceptingSection := acceptingComponentsSection(componentName, meta)
	outputSection := outputComponentsSection(componentName, meta)

	if acceptingSection == "" && outputSection == "" {
		return "", nil
	}

	note := "\nNote that connecting some components may not be feasible or components may require further " +
		"configuration to make the connection work correctly. " +
		"Please refer to the linked documentation for more details.\n\n"

	return heading + acceptingSection + outputSection + note, nil
}

func allComponentsThat(f func(meta metadata.Metadata) bool) []string {
	var result []string
	for _, name := range component.AllNames() {
		meta, err := metadata.ForComponent(name)
		if err != nil {
			panic(err) // should never happen
		}

		if f(meta) {
			result = append(result, name)
		}
	}
	return result
}

func allComponentsThatOutput(dataType metadata.DataType) []string {
	return allComponentsThat(func(meta metadata.Metadata) bool {
		return slices.Contains(meta.Outputs, dataType)
	})
}

func allComponentsThatAccept(dataType metadata.DataType) []string {
	return allComponentsThat(func(meta metadata.Metadata) bool {
		return slices.Contains(meta.Accepts, dataType)
	})
}

func outputComponentsSection(name string, meta metadata.Metadata) string {
	section := ""
	for _, outputDataType := range meta.Outputs {
		if list := listOfComponentsAccepting(outputDataType); list != "" {
			section += fmt.Sprintf("- Components that accept %s:\n", outputDataType) + list
		}
	}
	if section != "" {
		section = fmt.Sprintf("`%s` can output data to the following components:\n\n", name) + section
	}
	return section
}

func acceptingComponentsSection(componentName string, meta metadata.Metadata) string {
	section := ""
	for _, acceptedDataType := range meta.Accepts {
		if list := listOfComponentsOutputting(acceptedDataType); list != "" {
			section += fmt.Sprintf("- Components that output %s:\n", acceptedDataType) + list
		}
	}
	if section != "" {
		section = fmt.Sprintf("`%s` can accept data from the following components:\n\n", componentName) + section + "\n"
	}
	return section
}

func listOfComponentsAccepting(dataType metadata.DataType) string {
	return listOfLinksToComponents(allComponentsThatAccept(dataType))
}

func listOfComponentsOutputting(dataType metadata.DataType) string {
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
