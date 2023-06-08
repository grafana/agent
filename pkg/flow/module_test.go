//go:build linux

package flow

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/cluster"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/stretchr/testify/require"
)

const loggingConfig = `
	logging {}`

const tracingConfig = `
	tracing {}`

const argumentConfig = `
	argument "username" {} 
	argument "defaulted" {
		optional = true
		default = "default_value"
	}`

const exportStringConfig = `
	export "username" {
		value = "bob"
	}`

const exportDummy = `
	export "dummy" {
		value = "bob"
	}`

func TestModule(t *testing.T) {
	tt := []struct {
		name                  string
		argumentModuleContent string
		args                  map[string]interface{}
		exportModuleContent   string
		expectedExports       []string
		expectedErrorContains string
	}{
		{
			name: "Empty Content Allowed",
		},
		{
			name:                  "Bad Module",
			argumentModuleContent: `this isn't a valid module config`,
			expectedErrorContains: `expected block label, got IDENT`,
		},
		{
			name:                  "Logging blocks not allowed in module config",
			argumentModuleContent: argumentConfig + loggingConfig,
			exportModuleContent:   exportStringConfig,
			expectedErrorContains: "logging block not allowed inside a module",
		},
		{
			name:                  "Tracing blocks not allowed in module config",
			argumentModuleContent: argumentConfig + tracingConfig,
			exportModuleContent:   exportStringConfig,
			expectedErrorContains: "tracing block not allowed inside a module",
		},
		{
			name:                  "Argument not defined in module source",
			argumentModuleContent: `argument "different_argument" {}`,
			exportModuleContent:   exportStringConfig,
			args:                  map[string]interface{}{"different_argument": "test", "username": "bad"},
			expectedErrorContains: "Provided argument \"username\" is not defined in the module",
		},

		{
			name:                  "Missing required argument",
			argumentModuleContent: argumentConfig,
			exportModuleContent:   exportStringConfig,
			expectedErrorContains: "Failed to evaluate node for config block: missing required argument \"username\" to module",
		},

		{
			name:                  "Duplicate argument config",
			argumentModuleContent: argumentConfig + argumentConfig,
			exportModuleContent:   exportStringConfig,
			expectedErrorContains: "\"argument.username\" block already declared",
		},
		{
			name:                  "Duplicate export config",
			argumentModuleContent: argumentConfig,
			exportModuleContent:   exportStringConfig + exportStringConfig,
			expectedErrorContains: "\"export.username\" block already declared",
		},
		{
			name:                "Multiple exports but none are used but still exported",
			exportModuleContent: exportStringConfig + exportDummy,
			expectedExports:     []string{"username", "dummy"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			mc := newModuleController(testModuleControllerOptions(t)).(*moduleController)

			tm := &testModule{
				content: tc.argumentModuleContent + tc.exportModuleContent,
				args:    tc.args,
				opts:    component.Options{ModuleController: mc},
			}
			ctx := context.Background()
			ctx, cnc := context.WithTimeout(ctx, 1*time.Second)
			defer cnc()
			err := tm.Run(ctx)
			if tc.expectedErrorContains == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.expectedErrorContains)
			}
			for _, e := range tc.expectedExports {
				_, found := tm.exports[e]
				require.True(t, found)
			}
		})
	}
}

func TestArgsNotInModules(t *testing.T) {
	f := New(testOptions(t))
	fl, err := ReadFile("test", []byte("argument \"arg\"{}"))
	require.NoError(t, err)
	err = f.LoadFile(fl, nil)
	require.ErrorContains(t, err, "argument blocks only allowed inside a module")
}

func TestExportsNotInModules(t *testing.T) {
	f := New(testOptions(t))
	fl, err := ReadFile("test", []byte("export \"arg\"{ value = 1}"))
	require.NoError(t, err)
	err = f.LoadFile(fl, nil)
	require.ErrorContains(t, err, "export blocks only allowed inside a module")
}

func TestExportsWhenNotUsed(t *testing.T) {
	f := New(testOptions(t))
	content := " export \\\"username\\\"  { value  = 1 } \\n export \\\"dummy\\\" { value = 2 } "
	fullContent := "test.module \"t1\" { content = \"" + content + "\" }"
	fl, err := ReadFile("test", []byte(fullContent))
	require.NoError(t, err)
	err = f.LoadFile(fl, nil)
	require.NoError(t, err)
	ctx := context.Background()
	ctx, cnc := context.WithTimeout(ctx, 1*time.Second)
	defer cnc()
	f.Run(ctx)
	exps := f.loader.Components()[0].Exports().(TestExports)
	for _, x := range []string{"username", "dummy"} {
		_, found := exps.Exports[x]
		require.True(t, found)
	}
}

func TestIDCollision(t *testing.T) {
	nc := newModuleController(&moduleControllerOptions{
		Logger:         nil,
		Tracer:         nil,
		Clusterer:      nil,
		Reg:            nil,
		DataPath:       "",
		HTTPListenAddr: "",
		HTTPPath:       "",
		DialFunc:       nil,
	})
	m, err := nc.NewModule("t1", nil)
	require.NoError(t, err)
	require.NotNil(t, m)
	m, err = nc.NewModule("t1", nil)
	require.Error(t, err)
	require.Nil(t, m)
}

func testModuleControllerOptions(t *testing.T) *moduleControllerOptions {
	t.Helper()

	s, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)

	c := &cluster.Clusterer{Node: cluster.NewLocalNode("")}

	return &moduleControllerOptions{
		Logger:    s,
		DataPath:  t.TempDir(),
		Reg:       nil,
		Clusterer: c,
	}
}

func init() {
	component.Register(component.Registration{
		Name:    "test.module",
		Args:    TestArguments{},
		Exports: TestExports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return &testModule{
				content: args.(TestArguments).Content,
				opts:    opts,
			}, nil
		},
	})
}

type TestArguments struct {
	Content string `river:"content,attr"`
}

type TestExports struct {
	Exports map[string]interface{} `river:"exports,attr"`
}

type testModule struct {
	content string
	args    map[string]interface{}
	exports map[string]interface{}
	opts    component.Options
}

func (t *testModule) Run(ctx context.Context) error {
	m, err := t.opts.ModuleController.NewModule("t1", func(exports map[string]any) {
		t.exports = exports
		if t.opts.OnStateChange == nil {
			return
		}
		t.opts.OnStateChange(TestExports{Exports: exports})
	})
	if err != nil {
		return err
	}
	err = m.LoadConfig([]byte(t.content), t.args)
	if err != nil {
		return err
	}
	m.Run(ctx)
	return nil
}

func (t *testModule) Update(_ component.Arguments) error {
	return nil
}
