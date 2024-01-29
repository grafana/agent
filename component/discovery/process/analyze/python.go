package analyze

import (
	"debug/elf"
	"io"
	"strings"
)

const (
	LabelPython        = "__meta_process_python__"
	LabelPythonVersion = "__meta_process_python_version__"

	libpythonPrefix = "libpython"
)

func analyzePython(pid string, reader io.ReaderAt, m map[string]string) error {
	e, err := elf.NewFile(reader)
	if err != nil {
		return err
	}
	defer e.Close()

	libs, err := e.ImportedLibraries()
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
	m[LabelPython] = "true"
	m[LabelPythonVersion] = pythonVersion

	return nil
}
