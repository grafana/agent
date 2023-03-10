package autodiscovery

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetOpenFilenames ???
func GetOpenFilenames(pid int, extensions ...string) ([]string, error) {
	cmd := exec.Command("lsof", "-p", fmt.Sprint(pid), "-F", "n")
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var res []string
	for _, line := range strings.Split(string(stdout), "\n") {
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
