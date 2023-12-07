package relabel_script

import (
	"fmt"
	"testing"
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

var exampleTargets = []discovery.Target{
	{
		"__address__": "node01:12345",
		"job":         "observability/agent",
		"cluster":     "europe-south-1",
	},
	{
		"__address__": "node02:12345",
		"job":         "observability/loki",
		"cluster":     "europe-south-1",
	},
	{
		"__address__": "node03:12345",
		"job":         "observability/mimir",
		"cluster":     "europe-south-1",
	},
}

func TestScript(t *testing.T) {
	testCases := []struct {
		name           string
		script         string
		scriptFile     string
		targets        []discovery.Target
		expected       []discovery.Target
		expectedRunErr error
	}{
		{
			name:           "empty script",
			script:         ``,
			expectedRunErr: fmt.Errorf("script or script_file must be set"),
		},
		{
			name:           "file does not exist",
			scriptFile:     "does_not_exist.py",
			expectedRunErr: fmt.Errorf("error loading script file: open does_not_exist.py: no such file or directory"),
		},
		{
			name: "identity script with empty targets",
			script: `
				def relabel_targets(targets):
					return targets
			`,
			targets:  []discovery.Target{},
			expected: []discovery.Target{},
		},
		{
			name:           "broken script",
			script:         `int wrongLanguage() { return 0; }`,
			expectedRunErr: fmt.Errorf("error compiling script"),
		},
		{
			name: "exception in script prints nice error message",
			script: `
				def relabel_targets(targets):
					return targets[100]
			`,
			expectedRunErr: fmt.Errorf(
				"error calling relabel_targets function in script: index 100 out of range: empty list\n" +
					"Traceback (most recent call last):\n" +
					"  discovery.relabel_script.test:3:16: in relabel_targets\n" +
					"Error: index 100 out of range: empty list\n" +
					"script:\n" +
					"1: \n" +
					"2: \t\t\t\tdef relabel_targets(targets):\n" +
					"3: \t\t\t\t\treturn targets[100]\n" +
					"4: ",
			),
		},
		{
			name:           "script missing the relabel_targets function",
			script:         `print("hello world")`,
			expectedRunErr: fmt.Errorf("script does not contain a relabel_targets function"),
		},
		{
			name:           "script relabel_targets function got no args",
			script:         `def relabel_targets(): pass`,
			expectedRunErr: fmt.Errorf("the relabel_targets function must accept exactly 1 argument"),
		},
		{
			name:           "script relabel_targets function got two args",
			script:         `def relabel_targets(a, b): pass`,
			expectedRunErr: fmt.Errorf("the relabel_targets function must accept exactly 1 argument"),
		},
		{
			name:           "script relabel_targets is not a function",
			script:         `relabel_targets = 1`,
			expectedRunErr: fmt.Errorf("script must define relabel_targets as a function"),
		},
		{
			name: "return not a list",
			script: `
				def relabel_targets(targets):
					return 1
			`,
			targets:        exampleTargets,
			expectedRunErr: fmt.Errorf("relabel_targets function in script did not return a list of dictionaries"),
		},
		{
			name: "return list of non-dict",
			script: `
				def relabel_targets(targets):
					return [1, 2, 3]
			`,
			targets:  exampleTargets,
			expected: []discovery.Target{},
		},
		{
			name: "convert wrong value types to string",
			script: `
				def relabel_targets(targets):
					return [{"__address__": 1}, {"__address__": 3.14}, {"__address__": [1, 2, 3]}]
			`,
			targets: exampleTargets,
			expected: []discovery.Target{
				{"__address__": "1"}, {"__address__": "3.14"}, {"__address__": "[1, 2, 3]"},
			},
		},
		{
			name: "ignore keys with wrong type",
			script: `
				def relabel_targets(targets):
					return [{"__address__": 1, 123: 1}, {"__address__": 2, 3.14: "hello"}, {"__address__": 3, True: False}]
			`,
			targets: exampleTargets,
			expected: []discovery.Target{
				{"__address__": "1"}, {"__address__": "2"}, {"__address__": "3"},
			},
		},
		{
			name: "script trying to do dodgy things",
			script: `
				import os
				os.mkdir("/tmp/agent")
				os.rmdir("/tmp/agent")
			`,
			expectedRunErr: fmt.Errorf("error compiling script: discovery.relabel_script.test:2:7: got illegal token, want primary expression"),
		},
		{
			name: "pass through",
			script: `
				def relabel_targets(targets):
					return targets
			`,
			targets:  exampleTargets,
			expected: exampleTargets,
		},
		{
			name: "add pod and namespace",
			script: `
				def relabel_targets(targets):
					for t in targets:
						namespace, pod = t["job"].split("/")
						t["namespace"] = namespace
						t["pod"] = pod
					return targets
			`,
			targets: exampleTargets,
			expected: []discovery.Target{
				{
					"__address__": "node01:12345",
					"job":         "observability/agent",
					"cluster":     "europe-south-1",
					"namespace":   "observability",
					"pod":         "agent",
				},
				{
					"__address__": "node02:12345",
					"job":         "observability/loki",
					"cluster":     "europe-south-1",
					"namespace":   "observability",
					"pod":         "loki",
				},
				{
					"__address__": "node03:12345",
					"job":         "observability/mimir",
					"cluster":     "europe-south-1",
					"namespace":   "observability",
					"pod":         "mimir",
				},
			},
		},
		{
			name:       "simple relabel file - add pod and namespace",
			scriptFile: "testdata/simple_relabel.py",
			targets:    exampleTargets,
			expected: []discovery.Target{
				{
					"__address__": "node01:12345",
					"job":         "observability/agent",
					"cluster":     "europe-south-1",
					"namespace":   "observability",
					"pod":         "agent",
				},
				{
					"__address__": "node02:12345",
					"job":         "observability/loki",
					"cluster":     "europe-south-1",
					"namespace":   "observability",
					"pod":         "loki",
				},
				{
					"__address__": "node03:12345",
					"job":         "observability/mimir",
					"cluster":     "europe-south-1",
					"namespace":   "observability",
					"pod":         "mimir",
				},
			},
		},
		{
			name: "correlate and join targets demo",
			script: `
				def relabel_targets(targets):
					joined = {}
					for t in targets:
						if t["__address__"].startswith("mysql://"):
							key = t["__address__"].split("/")[2]
						else:
							key = t.pop("__address__", None)
							t["__host_address__"] = key

						joined[key] = joined[key] | t if key in joined else t
					return list(joined.values())
`,
			targets: []discovery.Target{
				{
					"__address__": "mysql://node01/something",
					"job":         "db/mysql",
				},
				{
					"__address__": "node01",
					"size":        "xs",
					"cluster":     "europe-south-1",
				},
				{
					"__address__": "node02",
					"size":        "xxs",
					"cluster":     "europe-middle-1",
				},
				{
					"__address__": "mysql://node02/something",
					"job":         "db/mysql",
				},
			},
			expected: []discovery.Target{
				{
					"__address__":      "mysql://node01/something",
					"__host_address__": "node01",
					"job":              "db/mysql",
					"size":             "xs",
					"cluster":          "europe-south-1",
				},
				{
					"__host_address__": "node02",
					"__address__":      "mysql://node02/something",
					"size":             "xxs",
					"cluster":          "europe-middle-1",
					"job":              "db/mysql",
				},
			},
		},
	}
	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputTargets := deepCopyTargets(tc.targets)

			args := Arguments{
				Targets:    inputTargets,
				Script:     tc.script,
				ScriptFile: tc.scriptFile,
			}

			ctrl, err := componenttest.NewControllerFromID(util.TestLogger(t), "discovery.relabel_script")
			require.NoError(t, err)

			var (
				runErr  error
				runDone = make(chan struct{})
			)
			go func() {
				runErr = ctrl.Run(componenttest.TestContext(t), args)
				close(runDone)
			}()

			// Check for error if needed
			if tc.expectedRunErr != nil {
				<-runDone
				require.Error(t, runErr)
				assert.Contains(t, runErr.Error(), tc.expectedRunErr.Error())
				return
			}

			// Otherwise, verify exports
			require.NoError(t, ctrl.WaitRunning(time.Second), "component never started")
			require.NoError(t, ctrl.WaitExports(time.Second), "component never exported anything")

			require.EventuallyWithT(t, func(t *assert.CollectT) {
				exports := ctrl.Exports().(Exports)
				assert.NotNil(t, exports)
				assert.Equal(t, tc.targets, inputTargets, "input targets were modified")
				assert.Equal(t, tc.expected, exports.Output, "expected export does not match actual export")
			}, 3*time.Second, 10*time.Millisecond, "component never reached the desired state")
		})
	}
}

func TestScriptWithParsing(t *testing.T) {
	// Note script is not full "dedented" - has leading indentation.
	script := `
		def relabel_targets(targets):
			for t in targets:
				print('got target: ', t)
			return targets
`
	cfg := `
targets = [
	{
		__address__ = "127.0.0.1:12345",
		namespace = "agent",
		pod = "agent",
	},
	{
		__address__ = "127.0.0.1:8888",
		namespace = "loki",
		pod = "loki",
	},
]
script = ` + "`" + script + "`" + `
	`
	var args Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	ctrl, err := componenttest.NewControllerFromID(util.TestLogger(t), "discovery.relabel_script")
	require.NoError(t, err)

	go func() { require.NoError(t, ctrl.Run(componenttest.TestContext(t), args)) }()

	require.NoError(t, ctrl.WaitRunning(time.Second), "component never started")
	require.NoError(t, ctrl.WaitExports(time.Second), "component never exported anything")

	require.EventuallyWithT(t, func(t *assert.CollectT) {
		exports := ctrl.Exports().(Exports)
		assert.NotNil(t, exports)
		assert.Contains(t, exports.Output, discovery.Target{
			"__address__": "127.0.0.1:12345",
			"namespace":   "agent",
			"pod":         "agent",
		})
		assert.Contains(t, exports.Output, discovery.Target{
			"__address__": "127.0.0.1:8888",
			"namespace":   "loki",
			"pod":         "loki",
		})
	}, 3*time.Second, 10*time.Millisecond, "component never reached the desired state")
}

func deepCopyTargets(targets []discovery.Target) []discovery.Target {
	inputCopy := make([]discovery.Target, len(targets))
	for i, tg := range targets {
		tCopy := make(discovery.Target, len(tg))
		maps.Copy(tCopy, tg)
		inputCopy[i] = tCopy
	}
	return inputCopy
}
