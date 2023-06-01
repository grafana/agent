package sd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

var (
	cgroupContainerIDRe = regexp.MustCompile(`^.*/(?:.*-)?([0-9a-f]+)(?:\.|\s*$)`)
)

func getContainerIDFromPID(pid uint32) containerID {
	f, err := os.Open(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		cid := getContainerIDFromCGroup(line)
		if cid != "" {
			return containerID(cid)
		}
	}
	return ""
}

func getContainerIDFromCGroup(line string) string {
	matches := cgroupContainerIDRe.FindStringSubmatch(line)
	if len(matches) <= 1 {
		return ""
	}
	return matches[1]
}
