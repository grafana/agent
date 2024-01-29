package analyze

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type analyserFunc func(pid string, reader io.ReaderAt, labels map[string]string) error

func PID(logger log.Logger, pid string) (map[string]string, error) {
	m := make(map[string]string)

	procPath := filepath.Join("/proc", pid)
	exePath := filepath.Join(procPath, "exe")

	// check if executable exists
	_, err := os.Stat(exePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		// resolve path relative to mount
		// TODO:simonswine, don't think this actually needed, double check
		fmt.Println("relative to mount")
		dest, err := os.Readlink(filepath.Join(procPath, "exe"))
		if err != nil {
			return nil, err
		}

		exePath = filepath.Join(procPath, "root", dest)
	}

	// get path to executable
	f, err := os.Open(exePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	for _, a := range []analyserFunc{analyzeGo, analyzePython} {
		if err := a(pid, f, m); err == io.EOF {
			break
		} else if err != nil {
			level.Warn(logger).Log("msg", "error during", "func", "todo", "err", err)
		}
	}

	return m, nil
}
