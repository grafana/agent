package flowtest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/dskit/backoff"
	"rsc.io/script"
)

func createCommands(f *flow.Flow, opts *options) map[string]script.Cmd {
	cmds := script.DefaultCmds()
	registerControllerCommands(f, cmds)

	for name, cmd := range opts.cmds {
		cmds[name] = cmd
	}

	return cmds
}

func registerControllerCommands(f *flow.Flow, cmds map[string]script.Cmd) {
	cmds["controller.load"] = script.Command(
		script.CmdUsage{
			Summary: "load a file into the Flow controller",
			Args:    "file",
		},
		func(s *script.State, args ...string) (script.WaitFunc, error) {
			if len(args) != 1 {
				return nil, script.ErrUsage
			}

			bb, err := os.ReadFile(makeAbsolute(s.Getwd(), args[0]))
			if err != nil {
				return nil, err
			}
			src, err := flow.ParseSource(args[0], bb)
			if err != nil {
				return nil, fmt.Errorf("parsing %q: %w", args[0], err)
			}
			return nil, f.LoadSource(src, nil)
		},
	)

	cmds["controller.start"] = script.Command(
		script.CmdUsage{
			Summary: "start the Flow controller",
			Async:   true,
			Detail: []string{
				"This command should almost always be run in the background by appending a &, otherwise the controller will exit immediately after the command finishes'",
			},
		},
		func(s *script.State, args ...string) (script.WaitFunc, error) {
			if len(args) != 0 {
				return nil, script.ErrUsage
			}

			var wg sync.WaitGroup
			wg.Add(1)

			ctx, cancel := context.WithCancel(s.Context())
			go func() {
				defer wg.Done()
				f.Run(ctx)
			}()

			return script.WaitFunc(func(s *script.State) (stdout string, stderr string, err error) {
				cancel()
				wg.Wait()

				return "", "", nil
			}), nil
		},
	)

	cmds["assert.health"] = script.Command(
		script.CmdUsage{
			Summary: "assert a health of a component",
			Args:    "component expected_health",
		},
		func(s *script.State, args ...string) (script.WaitFunc, error) {
			if len(args) != 2 {
				return nil, script.ErrUsage
			}

			// TODO(rfratto): allow configuring via flags.
			bc := backoff.Config{
				MinBackoff: 10 * time.Millisecond,
				MaxBackoff: 1 * time.Second,
				MaxRetries: 10,
			}
			bo := backoff.New(s.Context(), bc)

			check := func() error {
				var expectedHealth component.HealthType
				if err := expectedHealth.UnmarshalText([]byte(strings.ToLower(args[1]))); err != nil {
					return err
				}

				info, err := f.GetComponent(component.ParseID(args[0]), component.InfoOptions{GetHealth: true})
				if err != nil {
					return err
				}

				if info.Health.Health != expectedHealth {
					return fmt.Errorf("expected %q, got %q", expectedHealth, info.Health.Health)
				}

				return nil
			}

			for bo.Ongoing() {
				if err := check(); err == nil {
					break
				}
				bo.Wait()
			}

			return nil, bo.Err()
		},
	)
}

func makeAbsolute(wd, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(wd, path)
}
