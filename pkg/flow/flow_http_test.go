package flow

import (
	"bytes"
	"testing"

	_ "github.com/grafana/agent/pkg/flow/internal/testcomponents" // Import testcomponents
	"github.com/stretchr/testify/require"
)

func Test_configBytes(t *testing.T) {
	configFile := `
		testcomponents "tick" "ticker-a" {
			frequency = "1s"
		}

		testcomponents "passthrough" "static" {
			input = "hello, world!"
		}
	`

	file, diags := ReadFile(t.Name(), []byte(configFile))
	require.NotNil(t, file)
	require.False(t, diags.HasErrors(), "Found errors when loading file")

	f, _ := newFlow(testOptions(t))

	err := f.LoadFile(file)
	require.NoError(t, err)

	var buf bytes.Buffer
	_, _ = f.configBytes(&buf, false)
	actual := buf.String()

	// Exported fields aren't reported for testcomponents.tick.ticker-a because
	// the controller isn't running, so all of its exports are the zero value and
	// get omitted from the result.
	expect :=
		`// Component testcomponents.tick.ticker-a:
testcomponents "tick" "ticker-a" {
  frequency = "1s"
}

// Component testcomponents.passthrough.static:
testcomponents "passthrough" "static" {
  input = "hello, world!"

  // Exported fields:
  output = "hello, world!"
}

`

	require.Equal(t, expect, actual)
}
