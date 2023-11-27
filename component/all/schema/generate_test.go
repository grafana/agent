package schema

import (
	"flag"
	"os"
	"testing"

	"github.com/grafana/agent/component"
	_ "github.com/grafana/agent/component/all"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

//go:generate go test -fix-tests

const schemaFileName = "schema.yaml"

var fixTestsFlag = flag.Bool("fix-tests", false, "update the test files with the current generated content")

func TestGenerated(t *testing.T) {
	var allComponetns []Component
	for _, name := range component.AllNames() {
		registration, ok := component.Get(name)
		require.True(t, ok, "component %q not found", name)

		var (
			args, exports []Field
			err           error
		)

		if registration.Args != nil {
			args, err = getFields(registration.Args, name)
			require.NoError(t, err)
		}

		if registration.Exports != nil {
			exports, err = getFields(registration.Exports, name)
			require.NoError(t, err)
		}

		allComponetns = append(allComponetns, Component{
			Name:      name,
			Arguments: args,
			Exports:   exports,
		})
	}

	yamlData, err := yaml.Marshal(allComponetns)
	require.NoError(t, err)

	if *fixTestsFlag {
		err := writeTestFile(yamlData)
		require.NoError(t, err)
	}

	actual, err := readTestFile()
	require.NoError(t, err)

	require.Equal(t, string(yamlData), actual, "generated schema does not match the file saved in repository")
}

func writeTestFile(data []byte) error {
	err := os.WriteFile(schemaFileName, data, 0644)
	return err
}

func readTestFile() (string, error) {
	data, err := os.ReadFile(schemaFileName)
	return string(data), err
}
