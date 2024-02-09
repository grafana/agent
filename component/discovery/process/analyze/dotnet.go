package analyze

import (
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/prometheus/procfs"
	"github.com/pyroscope-io/dotnetdiag"
)

const (
	LabelDotNet                 = "__meta_process_dotnet__"
	LabelDotNetArch             = "__meta_process_dotnet_arch__"
	LabelDotNetOS               = "__meta_process_dotnet_os__"
	LabelDotNetVersion          = "__meta_process_dotnet_version__"
	LabelDotNetDiagnosticSocket = "__meta_process_dotnet_diagnostic_socket__"
	LabelDotNetCommandLine      = "__meta_process_dotnet_command_line__"
	LabelDotNetAssemblyName     = "__meta_process_dotnet_assembly_name__"
)

func analyzeDotNet(input Input, a *Results) error {
	m := a.Labels
	// small hack: the per process api procfs.Proc doesn't support reading NetUnix, so i am using the global one
	procPath := filepath.Join("/proc", input.PIDs)
	procph, err := procfs.NewFS(procPath)
	if err != nil {
		return err
	}
	netunix, err := procph.NetUNIX()
	if err != nil {
		return err
	}
	sockets := map[string]*procfs.NetUNIXLine{}
	for _, sock := range netunix.Rows {
		if !strings.HasPrefix(filepath.Base(sock.Path), "dotnet-diagnostic-") {
			continue
		}
		sockets[strconv.FormatUint(sock.Inode, 10)] = sock
	}

	// now get the inodes for the fds of the process and see if they match

	procp, err := procfs.NewProc(int(input.PID))
	if err != nil {
		return err
	}
	fdinfo, err := procp.FileDescriptorsInfo()
	if err != nil {
		return err
	}
	unixSocket := ""
	for _, fd := range fdinfo {
		sock, found := sockets[fd.Ino]
		if !found {
			continue
		}
		unixSocket = filepath.Join(procPath, "root", sock.Path)
		break
	}

	// bail if no unix socket found
	if unixSocket == "" {
		return nil
	}

	// connect to the dotnet socket and retrieve metadata
	ddc := dotnetdiag.NewClient(unixSocket)
	info, err := ddc.ProcessInfo2()
	if err != nil {
		return err
	}

	m[LabelDotNet] = labelValueTrue
	m[LabelDotNetCommandLine] = info.CommandLine
	m[LabelDotNetOS] = info.OS
	m[LabelDotNetArch] = info.Arch
	m[LabelDotNetAssemblyName] = info.AssemblyName
	m[LabelDotNetVersion] = info.RuntimeVersion
	m[LabelDotNetDiagnosticSocket] = unixSocket

	return io.EOF
}
