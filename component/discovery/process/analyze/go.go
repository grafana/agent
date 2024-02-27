package analyze

import (
	"debug/buildinfo"
	"io"
	"regexp"
	"strings"
)

const (
	LabelGo                   = "__meta_process_go__"
	LabelGoVersion            = "__meta_process_go_version__"
	LabelGoModulePath         = "__meta_process_go_module_path__"
	LabelGoModuleVersion      = "__meta_process_go_module_version__"
	LabelGoSdk                = "__meta_process_go_sdk__"
	LabelGoSdkVersion         = "__meta_process_go_sdk_version__"
	LabelGoDeltaProf          = "__meta_process_go_godeltaprof__"
	LabelGoDeltaProfVersion   = "__meta_process_go_godeltaprof_version__"
	LabelGoBuildSettingPrefix = "__meta_process_go_build_setting_"

	goSdkModule       = "github.com/grafana/pyroscope-go"
	godeltaprofModule = "github.com/grafana/pyroscope-go/godeltaprof"
)

func analyzeGo(input Input, a *Results) error {
	m := a.Labels
	info, err := buildinfo.Read(input.File) // it reads elf second time
	if err != nil {
		if err.Error() == "not a Go executable" {
			return nil
		}
		return err
	}

	m[LabelGo] = labelValueTrue

	if info.GoVersion != "" {
		m[LabelGoVersion] = info.GoVersion
	}
	if info.Main.Path != "" {
		m[LabelGoModulePath] = info.Main.Path
	}
	if info.Main.Version != "" {
		m[LabelGoModuleVersion] = info.Main.Version
	}

	for _, setting := range info.Settings {
		k := sanitizeLabelName(setting.Key)
		m[LabelGoBuildSettingPrefix+k] = setting.Value
	}

	for _, dep := range info.Deps {
		switch dep.Path {
		case goSdkModule:
			m[LabelGoSdk] = labelValueTrue
			m[LabelGoSdkVersion] = dep.Version
		case godeltaprofModule:
			m[LabelGoDeltaProf] = labelValueTrue
			m[LabelGoDeltaProfVersion] = dep.Version
		default:
			//todo should we optionally/configurable include all deps?
			continue
		}
	}

	return io.EOF
}

var sanitizeRe = regexp.MustCompile("[^a-zA-Z0-9_]")

func sanitizeLabelName(s string) string {
	s = sanitizeRe.ReplaceAllString(s, "_")
	return strings.ToLower(s)
}
