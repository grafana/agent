package flow

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	mod "github.com/grafana/agent/component/module"
	"github.com/stretchr/testify/require"
)

func TestIDRemovalIfFailedToLoad(t *testing.T) {
	f := New(testOptions(t))

	fullContent := "test.fail.module \"t1\" { content = \"\" }"
	fl, err := ParseSource("test", []byte(fullContent))
	require.NoError(t, err)
	err = f.LoadSource(fl, nil)
	require.NoError(t, err)
	ctx := context.Background()
	ctx, cnc := context.WithTimeout(ctx, 600*time.Second)

	go f.Run(ctx)
	require.Eventually(t, func() bool {
		t1 := f.loader.Components()[0].Component().(*testFailModule)
		return t1 != nil
	}, 10*time.Second, 100*time.Millisecond)
	t1 := f.loader.Components()[0].Component().(*testFailModule)
	badContent :=
		`test.fail.module "int" {
content=""
fail=true
}`
	err = t1.updateContent(badContent)
	// Because we have bad content this should fail, but the ids should be removed.
	require.Error(t, err)
	goodContent :=
		`test.fail.module "int" { 
content=""
fail=false
}`
	err = t1.updateContent(goodContent)
	require.NoError(t, err)
	cnc()
}

func init() {
	component.Register(component.Registration{
		Name:    "test.fail.module",
		Args:    TestFailArguments{},
		Exports: mod.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			m, err := mod.NewModuleComponent(opts)
			if err != nil {
				return nil, err
			}
			if args.(TestFailArguments).Fail {
				return nil, fmt.Errorf("module told to fail")
			}
			err = m.LoadFlowSource(nil, args.(TestFailArguments).Content)
			if err != nil {
				return nil, err
			}
			return &testFailModule{
				mc:      m,
				content: args.(TestFailArguments).Content,
				opts:    opts,
				fail:    args.(TestFailArguments).Fail,
				ch:      make(chan error),
			}, nil
		},
	})
}

type TestFailArguments struct {
	Content string `river:"content,attr"`
	Fail    bool   `river:"fail,attr,optional"`
}

type testFailModule struct {
	content string
	opts    component.Options
	ch      chan error
	mc      *mod.ModuleComponent
	fail    bool
}

func (t *testFailModule) Run(ctx context.Context) error {
	go t.mc.RunFlowController(ctx)
	<-ctx.Done()
	return nil
}

func (t *testFailModule) updateContent(content string) error {
	t.content = content
	err := t.mc.LoadFlowSource(nil, t.content)
	return err
}

func (t *testFailModule) Update(_ component.Arguments) error {
	return nil
}
