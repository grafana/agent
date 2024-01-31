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

	cmdLine, err := proc.CmdLine()
	isJava := false
	for _, c := range cmdLine {
		if strings.Contains(c, "java") {
			isJava = true
			break
		}
	}

	if !isJava {
		return nil
	}
	m[labelJava] = "true"
	return nil
}
