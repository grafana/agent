package flow

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	mod "github.com/grafana/agent/component/module"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/prometheus/client_golang/prometheus"
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

func TestIDList(t *testing.T) {
	nc := newModuleController(testModuleControllerOptions(t))
	require.Len(t, nc.ModuleIDs(), 0)

	_, err := nc.NewModule("t1", nil)
	require.NoError(t, err)
	require.Len(t, nc.ModuleIDs(), 1)

	_, err = nc.NewModule("t2", nil)
	require.NoError(t, err)
	require.Len(t, nc.ModuleIDs(), 2)
}

func TestIDCollision(t *testing.T) {
	nc := newModuleController(testModuleControllerOptions(t))
	m, err := nc.NewModule("t1", nil)
	require.NoError(t, err)
	require.NotNil(t, m)
	m, err = nc.NewModule("t1", nil)
	require.Error(t, err)
	require.Nil(t, m)
}

func TestIDRemoval(t *testing.T) {
	opts := testModuleControllerOptions(t)
	opts.ID = "test"
	nc := newModuleController(opts)
	m, err := nc.NewModule("t1", func(exports map[string]any) {})
	require.NoError(t, err)
	err = m.LoadConfig([]byte(""), nil)
	require.NoError(t, err)
	require.NotNil(t, m)
	ctx := context.Background()
	ctx, cncl := context.WithTimeout(ctx, 1*time.Second)
	defer cncl()
	m.Run(ctx)
	require.Len(t, nc.(*moduleController).modules, 0)
}

func TestIDRemovalIfFailedToLoad(t *testing.T) {
	f := New(testOptions(t))
	internalContent :=
		`test.module \"good\" { content=\"\"}`
	fullContent := "test.module \"t1\" { content = \"" + internalContent + "\" }"
	fl, err := ReadFile("test", []byte(fullContent))
	require.NoError(t, err)
	err = f.LoadFile(fl, nil)
	require.NoError(t, err)
	ctx := context.Background()
	ctx, cnc := context.WithTimeout(ctx, 600*time.Second)
	defer cnc()
	go f.Run(ctx)
	time.Sleep(5 * time.Second)
	require.Eventually(t, func() bool {
		t1 := f.loader.Components()[0].Component().(*testModule)
		return t1 != nil
	}, 10*time.Second, 100*time.Millisecond)
	t1 := f.loader.Components()[0].Component().(*testModule)
	m := t1.mc
	require.NotNil(t, m)
	internalModule := f.modules.modules["test.module.t1"]
	require.NotNil(t, internalModule)
	t2 := internalModule.f.loader.Components()[0].Component().(*testModule)
	require.NotNil(t, t2)
	err = t2.updateContent("garbage")
	require.Error(t, err)
	time.Sleep(5 * time.Second)
	// This has to be different since we are passing the string directly instead of via readfile.
	err = t2.updateContent("")
	require.NoError(t, err)
}

func testModuleControllerOptions(t *testing.T) *moduleControllerOptions {
	t.Helper()

	s, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)

	return &moduleControllerOptions{
		Logger:         s,
		DataPath:       t.TempDir(),
		Reg:            prometheus.NewRegistry(),
		ModuleRegistry: newModuleRegistry(),
	}
}

func init() {
	component.Register(component.Registration{
		Name:    "test.module",
		Args:    TestArguments{},
		Exports: mod.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			m, err := mod.NewModuleComponent(opts)
			if err != nil {
				return nil, err
			}

			return &testModule{
				mc:      m,
				content: args.(TestArguments).Content,
				opts:    opts,
				ch:      make(chan error),
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
	ch      chan error
	mc      *mod.ModuleComponent
}

func (t *testModule) Run(ctx context.Context) error {
	var err error
	if err != nil {
		return err
	}

	err = t.mc.LoadFlowContent(t.args, t.content)
	if err != nil {
		return err
	}
	go t.mc.RunFlowController(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case err = <-t.ch:
			return err
		}
	}
	return nil
}

func (t *testModule) updateContent(content string) error {
	t.content = content
	err := t.mc.LoadFlowContent(t.args, t.content)
	if err != nil {
		t.ch <- err
	}
	return err
}

func (t *testModule) Update(_ component.Arguments) error {
	return nil
}
