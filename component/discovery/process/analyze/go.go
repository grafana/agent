package analyze

import (
	"debug/buildinfo"
	"io"
)

const (
	LabelGo              = "__meta_process_go__"
	LabelGoVersion       = "__meta_process_go_version__"
	LabelGoModulePath    = "__meta_process_go_module_path__"
	LabelGoModuleVersion = "__meta_process_go_module_version__"
)

func analyzeGo(pid string, reader io.ReaderAt, m map[string]string) error {
	info, err := buildinfo.Read(reader)
	if err != nil {
		return err
	}

	m[LabelGo] = "true"

	if info.GoVersion != "" {
		m[LabelGoVersion] = info.GoVersion
	}

	if info.Main.Path != "" {
		m[LabelGoModulePath] = info.Main.Path
	}
	if info.Main.Version != "" {
		m[LabelGoModuleVersion] = info.Main.Version
	}

	return io.EOF
}
