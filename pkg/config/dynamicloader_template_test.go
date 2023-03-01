//nolint:golint,goconst
package config

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/agent/pkg/util/subset"
	"gopkg.in/yaml.v2"

	_ "github.com/grafana/agent/pkg/integrations/install"
	v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/windows_exporter"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/assert"
)

func TestConfigMakerWithExporterWithTemplate(t *testing.T) {
	configStr := `
windows:
  enabled_collectors: {{ (datasource "vars").value }}
  instance: testinstance
`
	tDir := t.TempDir()
	writeFile(t, tDir, "vars.yaml", "value: banana")
	writeFile(t, tDir, "integrations-1.yml", configStr)
	fileFS := generateFilePath(tDir)

	loaderCfg := LoaderConfig{
		Sources: []Datasource{{
			Name: "vars",
			URL:  generateFilePath(filepath.Join(tDir, "vars.yaml")),
		}},
		TemplatePaths: []string{fileFS},
	}
	cmf := generateLoader(t, loaderCfg)
	configs, err := cmf.processIntegrations()
	require.NoError(t, err)
	assert.Len(t, configs, 1)
	wincfg, _ := configs[0].(v2.UpgradedConfig).LegacyConfig()
	assert.True(t, wincfg.(*windows_exporter.Config).EnabledCollectors == "banana")
}

func TestLoadingFromS3LoadingVarsLocally(t *testing.T) {
	configStr := `
windows:
  enabled_collectors: {{ (datasource "vars").value }}
  instance: testinstance
`
	tDir := t.TempDir()
	writeFile(t, tDir, "vars.yaml", "value: banana")
	u := pushFilesToFakeS3(t, "integrations-1.yml", configStr)
	s3Url := "s3://mybucket/?region=us-east-1&disableSSL=true&s3ForcePathStyle=true&endpoint=" + u.Host
	loaderCfg := LoaderConfig{
		Sources: []Datasource{{
			Name: "vars",
			URL:  fmt.Sprintf("file:///%s", filepath.Join(tDir, "vars.yaml")),
		}},
		TemplatePaths: []string{s3Url},
	}
	cmf := generateLoader(t, loaderCfg)
	cfg, err := cmf.processIntegrations()
	require.NoError(t, err)
	assert.Len(t, cfg, 1)
	oc, _ := cfg[0].(v2.UpgradedConfig).LegacyConfig()
	winCfg := oc.(*windows_exporter.Config)
	assert.True(t, winCfg.EnabledCollectors == "banana")
}

func TestLoadingFromS3LoadingVarsLocallyWithRange(t *testing.T) {
	configStr := `
windows:
  enabled_collectors: banana
  instance: testinstance
  autoscrape:
    metric_relabel_configs: {{ range (datasource "vars").value }}
    - source_labels: [__address__]
      target_label: {{ . }}
      replacement: "{{ . }}-value"
    {{ end }}
`
	tDir := t.TempDir()
	writeFile(t, tDir, "vars.yaml", "value: [banana,apple,pear]")
	u := pushFilesToFakeS3(t, "integrations-1.yml", configStr)

	s3Url := "s3://mybucket/?region=us-east-1&disableSSL=true&s3ForcePathStyle=true&endpoint=" + u.Host
	loaderCfg := LoaderConfig{
		Sources: []Datasource{{
			Name: "vars",
			URL:  fmt.Sprintf("file:///%s", filepath.Join(tDir, "vars.yaml")),
		}},
		TemplatePaths: []string{s3Url},
	}
	cmf := generateLoader(t, loaderCfg)
	cfg := &Config{}
	err := cmf.ProcessConfigs(cfg)
	require.NoError(t, err)
	assert.Len(t, cfg.Integrations.ExtraIntegrations, 1)
	_ = cfg.Integrations.setVersion(integrationsVersion2)
	expectBase := `
integrations:
  windows:
    autoscrape:
      metric_relabel_configs:
      - target_label: banana
      - target_label: apple
      - target_label: pear
`
	outBytes, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	assert.NoError(t, subset.YAMLAssert([]byte(expectBase), outBytes))
}

func writeFile(t *testing.T, directory string, path string, contents string) {
	fullpath := filepath.Join(directory, path)
	err := os.WriteFile(fullpath, []byte(contents), 0666)
	require.NoError(t, err)
}

func generateLoader(t *testing.T, lc LoaderConfig) *DynamicLoader {
	cmf, err := NewDynamicLoader()
	require.NoError(t, err)
	err = cmf.LoadConfig(lc)
	require.NoError(t, err)
	return cmf
}

func generateFilePath(directory string) string {
	if runtime.GOOS == "windows" {
		// The URL scheme needs an additional / on windows
		return fmt.Sprintf("file:///%s", directory)
	}
	return fmt.Sprintf("file://%s", directory)
}

func pushFilesToFakeS3(t *testing.T, filename string, filecontents string) *url.URL {
	t.Setenv("AWS_ANON", "true")

	backend := s3mem.New()
	faker := gofakes3.New(backend)

	srv := httptest.NewServer(faker.Server())
	_ = backend.CreateBucket("mybucket")
	t.Cleanup(srv.Close)
	_, err := backend.PutObject(
		"mybucket",
		filename,
		map[string]string{"Content-Type": "application/yaml"},
		bytes.NewBufferString(filecontents),
		int64(len(filecontents)),
	)
	assert.NoError(t, err)
	u, err := url.Parse(srv.URL)
	assert.NoError(t, err)
	return u
}
