package analyze

import (
	"io"
	"strconv"
	"strings"

	"github.com/prometheus/procfs"
)

const (
	labelJava = "__meta_process_java__"
)

func analyzeJava(pid string, reader io.ReaderAt, m map[string]string) error {
	pidn, _ := strconv.Atoi(pid)
	proc, err := procfs.NewProc(pidn)
	if err != nil {
		return err
	}

	executable, err := proc.Executable()
	if err != nil {
		return err
	}
	if strings.HasSuffix(executable, "java") {
		m[labelJava] = "true"
	} else {
		cmdLine, err := proc.CmdLine()
		if err != nil {
			return err
		}

		for _, c := range cmdLine {
			if strings.HasPrefix(c, "java") {
				m[labelJava] = "true"
				break
			}
		}
	}

	return nil
}
