package analyze

import (
	"debug/elf"
	"io"
	"strings"
)

const (
	LabelCPP = "__meta_process_cpp__"
)

func analyzeCpp(pid string, reader io.ReaderAt, m map[string]string) error {
	e, err := elf.NewFile(reader)
	if err != nil {
		return err
	}
	defer e.Close()

	libs, err := e.ImportedLibraries()
	if err != nil {
		return err
	}

	for _, lib := range libs {
		if strings.Contains(lib, "libc++") || strings.Contains(lib, "libstdc++") {
			m[LabelCPP] = "true"
			break
		}
	}

	return nil
}
