package autodiscovery

import (
	"fmt"
	"os/exec"
	"strings"
)

// TODO: Not sure if ProcessInfo/LSOF and the free functions should be in different files
type ProcessInfo interface {
	GetOpenFilenames(pid int) (string, error)
	GetOpenConnections(pid int) (string, error)
}

// GetOpenFilenames ???
func GetOpenFilenames(pi ProcessInfo, pid int, extensions ...string) (map[string]struct{}, error) {
	if pi == nil {
		return nil, fmt.Errorf("ProcessInfo must not be null")
	}

	stdout, err := pi.GetOpenFilenames(pid)
	if err != nil {
		return nil, err
	}

	res := make(map[string]struct{})
	for _, line := range strings.Split(stdout, "\n") {
		s, ok := strings.CutPrefix(line, "n")
		if !ok {
			continue
		}
		for _, ext := range extensions {
			if strings.HasSuffix(s, ext) {
				res[s] = struct{}{}
			}
		}
	}

	return res, nil
}

func GetOpenPorts(pi ProcessInfo, pid int) ([]string, error) {
	stdout, err := pi.GetOpenConnections(pid)
	if err != nil {
		return nil, err
	}

	var res []string
	for _, line := range strings.Split(stdout, "\n") {
		s, ok := strings.CutPrefix(line, "n")
		if !ok {
			continue
		}
		_ = s
		//TODO: Split via : and add the port to res
	}

	return res, nil
}

type LSOF struct {
	ProcessInfo
}

func (lsof LSOF) GetOpenFilenames(pid int) (string, error) {
	cmd := exec.Command("lsof", "-p", fmt.Sprint(pid), "-F", "n")
	stdout, err := cmd.Output()
	if err != nil {
		//TODO: If the process cannot be found, this error is just "exit status 1".
		// A common reason not to find the process is that we need to run the agent with root privileges.
		// Make the error message more clear? Also, somehow "ps.Processes()" finds these processes even
		// if the agent doesn't run as root.
		return "", err
	}
	return string(stdout), nil
}

func (lsof LSOF) GetOpenConnections(pid int) (string, error) {
	cmd := exec.Command("lsof", "-p", fmt.Sprint(pid), "-a", "-i", "-n", "-P", "-F", "n")
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(stdout), nil
}
