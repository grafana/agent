package process

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/discovery"
	gopsutil "github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/unix"
)

const (
	labelProcessID          = "__process_pid__"
	labelProcessExe         = "__meta_process_exe"
	labelProcessCwd         = "__meta_process_cwd"
	labelProcessCommandline = "__meta_process_commandline"
	labelProcessContainerID = "__container_id__"
)

type process struct {
	pid         string
	exe         string
	cwd         string
	commandline string
	containerID string
}

func (p process) String() string {
	return fmt.Sprintf("pid=%s exe=%s cwd=%s commandline=%s containerID=%s", p.pid, p.exe, p.cwd, p.commandline, p.containerID)
}

func convertProcesses(ps []process) []discovery.Target {
	var res []discovery.Target
	for _, p := range ps {
		t := convertProcess(p)
		res = append(res, t)
	}
	return res
}

func convertProcess(p process) discovery.Target {
	t := discovery.Target{
		labelProcessID:          p.pid,
		labelProcessExe:         p.exe,
		labelProcessCwd:         p.cwd,
		labelProcessCommandline: p.commandline,
	}
	if p.containerID != "" {
		t[labelProcessContainerID] = p.containerID
	}
	return t
}

func discover(l log.Logger) ([]process, error) {
	processes, err := gopsutil.Processes()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}
	res := make([]process, 0, len(processes))
	loge := func(pid int, e error) {
		if errors.Is(e, unix.ESRCH) {
			return
		}
		if errors.Is(e, os.ErrNotExist) {
			return
		}
		_ = level.Error(l).Log("msg", "failed to get process info", "err", e, "pid", pid)
	}
	for _, p := range processes {
		spid := fmt.Sprintf("%d", p.Pid)
		exe, err := p.Exe()
		if err != nil {
			loge(int(p.Pid), err)
			continue
		}
		cwd, err := p.Cwd()
		if err != nil {
			loge(int(p.Pid), err)
			continue
		}
		commandline, err := p.Cmdline()
		if err != nil {
			loge(int(p.Pid), err)
			continue
		}

		containerID, err := getLinuxProcessContainerID(l, spid)
		if err != nil {
			loge(int(p.Pid), err)
			continue
		}
		res = append(res, process{
			pid:         spid,
			exe:         exe,
			cwd:         cwd,
			commandline: commandline,
			containerID: containerID,
		})
		_ = level.Debug(l).Log("msg", "found process", "pid", p.Pid, "exe", exe, "cwd", cwd, "container_id", containerID, "commandline", commandline)
	}

	return res, nil
}

func getLinuxProcessContainerID(l log.Logger, pid string) (string, error) {
	if runtime.GOOS == "linux" {
		cgroup, err := os.Open(path.Join("/proc", pid, "cgroup"))
		if err != nil {
			return "", err
		}
		defer cgroup.Close()
		cid := getContainerIDFromCGroup(cgroup)
		if cid != "" {
			return cid, nil
		}
	}
	return "", nil
}