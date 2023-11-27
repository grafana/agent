package process

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/discovery"
)

const (
	labelProcessID          = "__process_pid__"
	labelProcessExe         = "__meta_process_exe"
	labelProcessCwd         = "__meta_process_cwd"
	labelProcessContainerID = "__container_id__"
)

type process struct {
	pid         string
	exe         string
	cwd         string
	containerID string
}

func discover(l log.Logger, procFS string) ([]process, error) {
	var (
		err error
		ps  []process
	)
	pids, err := os.ReadDir(procFS)
	if err != nil {
		return nil, fmt.Errorf("discovery.process: failed to read /proc: %w", err)
	}
	for _, entry := range pids {
		var (
			exe    string
			cwd    string
			cgroup *os.File
		)
		if !entry.IsDir() {
			continue
		}
		pidDir := entry
		pid := pidDir.Name()
		_, err = strconv.Atoi(pid)
		if err != nil {
			continue
		}
		exe, err = os.Readlink(path.Join(procFS, pid, "exe"))
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				_ = level.Error(l).Log("msg", "failed to read /proc/{pid}/exe", "err", err)
			}
			continue
		}
		cwd, err = os.Readlink(path.Join(procFS, pid, "cwd"))
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				_ = level.Error(l).Log("msg", "failed to read /proc/{pid}/cwd", "err", err)
			}
			continue
		}
		cgroup, err = os.Open(path.Join(procFS, pid, "cgroup"))
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				_ = level.Error(l).Log("msg", "failed to read /proc/{pid}/cgroup", "err", err)
			}
			continue
		}
		cid := getContainerIDFromCGroup(cgroup)
		ps = append(ps, process{
			pid:         pid,
			exe:         exe,
			cwd:         cwd,
			containerID: cid,
		})
		_ = cgroup.Close()
	}
	return ps, nil
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
		labelProcessID:  p.pid,
		labelProcessExe: p.exe,
		labelProcessCwd: p.cwd,
	}
	if p.containerID != "" {
		t[labelProcessContainerID] = p.containerID
	}
	return t
}
