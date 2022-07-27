package flow

import (
	"bytes"
	"testing"

	_ "github.com/grafana/agent/pkg/flow/internal/testcomponents" // Import testcomponents
	"github.com/stretchr/testify/require"
)

func Test_configBytes(t *testing.T) {
	configFile := `
		testcomponents.tick "ticker_a" {
			frequency = "1s"
		}

		testcomponents.passthrough "static" {
			input = "hello, world!"
		}
	`

	file, err := ReadFile(t.Name(), []byte(configFile))
	require.NotNil(t, file)
	require.NoError(t, err)

	f, _ := newFlow(testOptions(t))

	err = f.LoadFile(file)
	require.NoError(t, err)

	var buf bytes.Buffer
	_, _ = f.configBytes(&buf, false)
	actual := buf.String()

	// Exported fields aren't reported for testcomponents.tick.ticker_a because
	// the controller isn't running, so all of its exports are the zero value and
	// get omitted from the result.
	expect :=
		`// Component testcomponents.tick.ticker_a:
testcomponents.tick "ticker_a" {
	frequency = "1s"
}

// Component testcomponents.passthrough.static:
testcomponents.passthrough "static" {
	input = "hello, world!"

	// Exported fields:
	output = "hello, world!"
}`

	require.Equal(t, expect, actual)
}
