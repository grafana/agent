package relabel_script

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.relabel_script",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	// Targets contains the input 'targets' passed by a service discovery component.
	Targets []discovery.Target `river:"targets,attr"`
	// Script contains a Python (Starlark dialect) script to run that will relabel the targets.
	// The script must contain a function with the signature: `def relabel_targets(targets)` that takes a list of
	// dictionaries representing the targets and returns a new list of dictionaries for the user-modified targets.
	Script string `river:"script,attr,optional"`
	// ScriptFile contains the path to a Starlark (Python dialect) script to run that will relabel the targets. See
	// Script for details on the script format.
	ScriptFile string `river:"script_file,attr,optional"`
}

func (a Arguments) Validate() error {
	if a.Script == "" && a.ScriptFile == "" {
		return fmt.Errorf("script or script_file must be set")
	}
	if a.Script != "" && a.ScriptFile != "" {
		return fmt.Errorf("only one of script or script_file can be set")
	}
	return nil
}

type Exports struct {
	Output []discovery.Target `river:"output,attr"`
}

type Component struct {
	opts component.Options

	mut              sync.RWMutex
	currentScript    string
	thread           *starlark.Thread
	relabelTargetsFn starlark.Value
}

var _ component.Component = (*Component)(nil)

func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{opts: o}

	// Call to Update() to set the output once at the start
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	c.thread.Cancel("component exiting")
	return nil
}

func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	newArgs := args.(Arguments)

	if err := c.updateScript(newArgs); err != nil {
		return err
	}

	targets, err := c.doRelabel(newArgs.Targets)
	if err != nil {
		return err
	}

	c.opts.OnStateChange(Exports{
		Output: targets,
	})

	return nil
}

func (c *Component) doRelabel(targets []discovery.Target) ([]discovery.Target, error) {
	sTargets, err := c.toStarlarkTargets(targets)
	if err != nil {
		return nil, fmt.Errorf("error converting targets to Starlark: %w", err)
	}

	value, err := starlark.Call(c.thread, c.relabelTargetsFn, starlark.Tuple{starlark.NewList(sTargets)}, nil)
	if err != nil {
		finalErr := err
		if evalError, ok := err.(*starlark.EvalError); ok {
			finalErr = fmt.Errorf("error calling relabel_targets function in script: %w\n%v\nscript:\n%v", err, evalError.Backtrace(), numberLines(c.currentScript))
		}
		return nil, finalErr
	}

	return c.toFlowTargets(value)
}

func (c *Component) updateScript(newArgs Arguments) error {
	newScript := ""
	if newArgs.Script != "" {
		newScript = newArgs.Script
	} else if newArgs.ScriptFile != "" {
		scriptBytes, err := os.ReadFile(newArgs.ScriptFile)
		if err != nil {
			return fmt.Errorf("error loading script file: %w", err)
		}
		newScript = string(scriptBytes)
	} else {
		// Should never happen thanks to validation
		return fmt.Errorf("script or script_file must be set")
	}

	if newScript == c.currentScript {
		return nil
	}
	var opts = syntax.FileOptions{Set: true, While: true, TopLevelControl: true, Recursion: true}

	printFn := func(_ *starlark.Thread, msg string) {
		level.Info(c.opts.Logger).Log("subcomponent", "script", "msg", msg)
	}
	c.thread = &starlark.Thread{Name: c.opts.ID, Print: printFn}
	compiled, err := starlark.ExecFileOptions(&opts, c.thread, c.opts.ID, dedent(newScript), nil)
	if err != nil {
		return fmt.Errorf("error compiling script: %w\nscript:\n%s", err, numberLines(newScript))
	}

	fn, ok := compiled["relabel_targets"]
	if !ok {
		return fmt.Errorf("script does not contain a relabel_targets function: %s", newScript)
	}
	if sfn, ok := fn.(*starlark.Function); ok {
		if sfn.NumParams() != 1 {
			return fmt.Errorf("the relabel_targets function must accept exactly 1 argument: %s", newScript)
		}
	} else {
		return fmt.Errorf("script must define relabel_targets as a function: %s", newScript)
	}
	c.relabelTargetsFn = fn
	c.currentScript = newScript
	return nil
}

// dedent removes the common leading whitespace from all lines in the script. This is to allow users to indent their
// scripts for readability in their River config file, without causing syntax errors.
func dedent(script string) string {
	lines := strings.Split(script, "\n")
	emptyLinesToSkip := map[int]struct{}{}
	smallestDedent := math.MaxInt
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			emptyLinesToSkip[i] = struct{}{}
			continue
		}
		dedentCount := len(line) - len(strings.TrimLeft(line, " \t"))
		if dedentCount < smallestDedent {
			smallestDedent = dedentCount
		}
	}
	if smallestDedent == 0 {
		return script // Script is already dedented
	}
	for i, line := range lines {
		if _, ok := emptyLinesToSkip[i]; ok {
			continue
		}
		lines[i] = line[smallestDedent:]
	}
	return strings.Join(lines, "\n")
}

func (c *Component) toFlowTargets(value starlark.Value) ([]discovery.Target, error) {
	listVal, ok := value.(*starlark.List)
	if !ok {
		return nil, fmt.Errorf("relabel_targets function in script did not return "+
			"a list of dictionaries compatible with targets: %+v", value)
	}

	newTargets := make([]discovery.Target, 0, listVal.Len())

	it := listVal.Iterate()
	defer it.Done()

	var val starlark.Value
	for it.Next(&val) {
		dictVal, ok := val.(*starlark.Dict)
		if !ok {
			level.Error(c.opts.Logger).Log("msg", "skipping invalid target: relabel_targets function must return a list of dictionaries", "invalid_list_element", fmt.Sprintf("%+v", val))
			continue
		}
		newTarget := make(discovery.Target)
		for _, k := range dictVal.Keys() {
			v, _, _ := dictVal.Get(k)
			if key, ok := k.(starlark.String); ok {
				newTarget[string(key)] = strings.Trim(v.String(), "\"")
			} else {
				level.Error(c.opts.Logger).Log("msg", "skipping invalid target label: relabel_targets function must return a list of dictionaries with string keys", "invalid_key", fmt.Sprintf("%+v", k))
				continue
			}
		}
		if len(newTarget) > 0 {
			newTargets = append(newTargets, newTarget)
		}
	}

	return newTargets, nil
}

func (c *Component) toStarlarkTargets(targets []discovery.Target) ([]starlark.Value, error) {
	sTargets := make([]starlark.Value, len(targets))
	for i, target := range targets {
		st := starlark.NewDict(len(target))
		for k, v := range target {
			err := st.SetKey(starlark.String(k), starlark.String(v))
			if err != nil {
				return nil, err
			}
		}
		sTargets[i] = st
	}
	return sTargets, nil
}

func numberLines(script string) string {
	lines := strings.Split(script, "\n")
	for i := range lines {
		lines[i] = fmt.Sprintf("%d: %s", i+1, lines[i])
	}
	return strings.Join(lines, "\n")
}
