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

type LSOF struct {
	ProcessInfo
}

func (lsof LSOF) GetOpenFilenames(pid int) (string, error) {
	cmd := exec.Command("lsof", "-p", fmt.Sprint(pid), "-F", "n")
	stdout, err := cmd.Output()
	if err != nil {
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

// GetOpenFilenames ???
func GetOpenFilenames(pi ProcessInfo, pid int, extensions ...string) ([]string, error) {
	stdout, err := pi.GetOpenFilenames(pid)
	if err != nil {
		return nil, err
	}

	var res []string
	for _, line := range strings.Split(stdout, "\n") {
		s, ok := strings.CutPrefix(line, "n")
		if !ok {
			continue
		}
		for _, ext := range extensions {
			if strings.HasSuffix(s, ext) {
				res = append(res, s)
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
		//TODO: Split via : and add the port to res
	}

	return res, nil
}
