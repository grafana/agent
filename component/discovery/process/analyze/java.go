package analyze

import (
	"strings"

	"github.com/prometheus/procfs"
)

const (
	labelJava = "__meta_process_java__"
)

func analyzeJava(input Input, a *Results) error {
	m := a.Labels
	proc, err := procfs.NewProc(int(input.PID))
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
