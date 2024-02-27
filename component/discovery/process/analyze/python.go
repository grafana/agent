package analyze

import (
	"strings"
)

const (
	LabelPython        = "__meta_process_python__"
	LabelPythonVersion = "__meta_process_python_version__"

	libpythonPrefix = "libpython"
)

func analyzePython(input Input, a *Results) error {
	m := a.Labels

	libs, err := input.ElfFile.ImportedLibraries()
	if err != nil {
		return err
	}

	var pythonVersion string
	for _, lib := range libs {
		if strings.HasPrefix(lib, libpythonPrefix) {
			pythonVersion = lib[len(libpythonPrefix):]
			pos := strings.Index(pythonVersion, ".so")
			if pos < 0 {
				continue
			}
			pythonVersion = pythonVersion[:pos]
			break
		}
	}
	if pythonVersion == "" {
		return nil
	}
	m[LabelPython] = labelValueTrue
	m[LabelPythonVersion] = pythonVersion

	return nil
}
