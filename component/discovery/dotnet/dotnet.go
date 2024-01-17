package dotnet

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/procfs"
	"github.com/pyroscope-io/dotnetdiag"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/process"
	"github.com/grafana/agent/pkg/flow/logging/level"
)

const LabelDotnetDiagnosticSocket = "__dotnet_diagnostic_socket__"

func init() {
	component.Register(component.Registration{
		Name:    "discovery.dotnet",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the discovery.dotnet component.
type Arguments struct {
	// Targets contains the input 'targets' passed by a service discovery component.
	Targets []discovery.Target `river:"targets,attr"`
}

// Exports holds values which are exported by the discovery.dotnet component.
type Exports struct {
	Output []discovery.Target `river:"output,attr"`
}

// Component implements the discovery.dotnet component.
type Component struct {
	opts component.Options

	mut sync.RWMutex
}

var _ component.Component = (*Component)(nil)

// New creates a new discovery.dotnet component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{opts: o}

	// Call to Update() to set the output once at the start
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

type unixSock struct {
	path  string
	inode string
}

// detectDotnetDiagnosticSocket returns the path to the dotnet diagnostic unix socket
func detectDotnetDiagnosticSockets(pid int) ([]string, error) {
	var result []string

	// small hack: the per process api procfs.Proc doesn't support reading NetUnix, so i am using the global one
	procPath := filepath.Join("/proc", strconv.Itoa(pid))
	procph, err := procfs.NewFS(procPath)
	if err != nil {
		return nil, err
	}
	netunix, err := procph.NetUNIX()
	if err != nil {
		return nil, err
	}
	sockets := map[string]*procfs.NetUNIXLine{}
	for _, sock := range netunix.Rows {
		if !strings.HasPrefix(filepath.Base(sock.Path), "dotnet-diagnostic-") {
			continue
		}
		sockets[strconv.FormatUint(sock.Inode, 10)] = sock
	}

	// now get the inodes for the fds of the process and see if they match
	procp, err := procfs.NewProc(pid)
	if err != nil {
		return nil, err
	}
	fdinfo, err := procp.FileDescriptorsInfo()
	if err != nil {
		return nil, err
	}
	for _, fd := range fdinfo {
		sock, found := sockets[fd.Ino]
		if !found {
			continue
		}
		result = append(result, filepath.Join(procPath, "root", sock.Path))
	}

	return result, nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	newArgs := args.(Arguments)

	targets := make([]discovery.Target, 0, len(newArgs.Targets))
	for _, target := range newArgs.Targets {
		pid, err := strconv.Atoi(target[process.LabelProcessID])
		_ = level.Debug(c.opts.Logger).Log("msg", "active target",
			"target", fmt.Sprintf("%+v", target),
			"pid", pid)
		if err != nil {
			_ = level.Error(c.opts.Logger).Log("msg", fmt.Sprintf("invalid label value of %s on target", process.LabelProcessID), "target", fmt.Sprintf("%v", target), "err", err)
			continue
		}

		result, err := detectDotnetDiagnosticSockets(pid)
		if err != nil {
			_ = level.Error(c.opts.Logger).Log("msg", "error detecting dotnet diagnostic socket", "err", err)
			continue
		}

		if len(result) == 0 {
			_ = level.Debug(c.opts.Logger).Log("msg", "no dotnet diagnostic socket detected", "target", fmt.Sprintf("%v", target))
			continue
		}

		for _, sock := range result {
			target[LabelDotnetDiagnosticSocket] = sock

			// gather process info
			ddc := dotnetdiag.NewClient(sock)
			info, err := ddc.ProcessInfo2()
			if err != nil {
				_ = level.Error(c.opts.Logger).Log("msg", "error creating dotnet diagnostic client", "err", err)
				continue
			}

			target["__dotnet_process_pid__"] = strconv.FormatUint(info.ProcessID, 10)
			target["__dotnet_command_line__"] = info.CommandLine
			target["__dotnet_os__"] = info.OS
			target["__dotnet_arch__"] = info.Arch
			target["__dotnet_assembly_name__"] = info.AssemblyName
			target["__dotnet_runtime_version__"] = info.RuntimeVersion

			targets = append(targets, target)
			// TODO: Remove this log line before merge
			level.Info(c.opts.Logger).Log("msg", "detected dotnet diagnostic socket", "target", fmt.Sprintf("%v", target))
		}
	}

	c.opts.OnStateChange(Exports{
		Output: targets,
	})

	return nil
}
