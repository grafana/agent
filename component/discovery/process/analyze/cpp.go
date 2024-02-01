package analyze

import (
	"strings"
)

const (
	LabelCPP = "__meta_process_cpp__"
)

func analyzeCpp(input Input, a *Results) error {
	m := a.Labels
	libs, err := input.ElfFile.ImportedLibraries()
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
