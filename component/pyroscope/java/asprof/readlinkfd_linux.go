package asprof

import (
	"fmt"
	"os"
)

func readLinkFD(f *os.File) (string, error) {
	fd := f.Fd()

	path := fmt.Sprintf("/proc/self/fd/%d", fd)
	realPath, err := os.Readlink(path)
	if err != nil {
		return "", fmt.Errorf("failed to check file %s %d", path, fd)
	}
	return realPath, nil

}
