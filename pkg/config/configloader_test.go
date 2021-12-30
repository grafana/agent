package config

import (
	"fmt"
	_ "github.com/grafana/agent/pkg/integrations/install"
	"github.com/grafana/agent/pkg/integrations/node_exporter"
	"github.com/grafana/agent/pkg/integrations/windows_exporter"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigMaker(t *testing.T) {
	configStr := `wal_directory: /tmp/wal`
	tDir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	fullpath := filepath.Join(tDir, "metrics-1.yml")
	err = ioutil.WriteFile(fullpath, []byte(configStr), fs.ModePerm)
	assert.Nil(t, err)
	fileFS := fmt.Sprintf("file://%s", tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf, err := NewConfigLoader(loaderCfg)
	assert.NoError(t, err)
	configs, err := cmf.processMetric()
	assert.Nil(t, err)
	assert.NotNil(t, configs)
	assert.Equal(t, configs.WALDir, "/tmp/wal")
	_ = os.RemoveAll(tDir)
}

func TestConfigMakerWithFakeFiles(t *testing.T) {
	configStr := `wal_directory: /tmp/wal`
	tDir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	fullpath := filepath.Join(tDir, "metrics-1.yml")
	err = ioutil.WriteFile(fullpath, []byte(configStr), fs.ModePerm)
	assert.Nil(t, err)

	fullpath = filepath.Join(tDir, "fake.yml")
	err = ioutil.WriteFile(fullpath, []byte(configStr), fs.ModePerm)
	assert.Nil(t, err)

	fileFS := fmt.Sprintf("file://%s", tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf, err := NewConfigLoader(loaderCfg)
	assert.NoError(t, err)
	configs, err := cmf.processMetric()
	assert.Nil(t, err)
	assert.NotNil(t, configs)
	assert.Equal(t, configs.WALDir, "/tmp/wal")
	_ = os.RemoveAll(tDir)
}

func TestConfigMakerWithMultipleMetrics(t *testing.T) {
	configStr := `wal_directory: /tmp/wal`
	tDir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	fullpath := filepath.Join(tDir, "metrics-1.yml")
	err = ioutil.WriteFile(fullpath, []byte(configStr), fs.ModePerm)
	assert.Nil(t, err)

	fullpath = filepath.Join(tDir, "metrics-2.yml")
	err = ioutil.WriteFile(fullpath, []byte(configStr), fs.ModePerm)
	assert.Nil(t, err)

	fileFS := fmt.Sprintf("file://%s", tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf, err := NewConfigLoader(loaderCfg)
	assert.Nil(t, err)
	_, err = cmf.processMetric()
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "multiple metrics configuration found")
	_ = os.RemoveAll(tDir)
}

func TestConfigMakerWithMetricsAndInstances(t *testing.T) {
	configStr := `wal_directory: /tmp/wal`
	tDir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	writeFile(t, tDir, "metrics-1.yml", configStr)
	writeFile(t, tDir, "metrics_instances-1.yml", "name: t1")
	writeFile(t, tDir, "metrics_instances-2.yml", "name: t2")
	fileFS := fmt.Sprintf("file://%s", tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf, err := NewConfigLoader(loaderCfg)
	assert.Nil(t, err)
	cfg, err := cmf.ProcessConfigs()
	assert.Nil(t, err)
	assert.Len(t, cfg.Metrics.Configs, 2)
	_ = os.RemoveAll(tDir)
}

func TestConfigMakerWithExporter(t *testing.T) {
	configStr := `
windows_exporter:
  enabled_collectors: one,two,three
  instance: testinstance
`
	tDir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	writeFile(t, tDir, "exporters-1.yml", configStr)
	fileFS := fmt.Sprintf("file://%s", tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf, err := NewConfigLoader(loaderCfg)
	assert.Nil(t, err)
	configs, err := cmf.processExporters()
	assert.Len(t, configs, 1)
	yamlBytes, err := yaml.Marshal(configs[0])
	assert.NoError(t, err)
	yamlString := string(yamlBytes)
	assert.True(t, strings.Contains(yamlString, "one,two,three"))
	_ = os.RemoveAll(tDir)
}

func TestConfigMakerWithMultipleExporter(t *testing.T) {
	configStr := `
windows_exporter:
  enabled_collectors: one,two,three
  instance: testinstance
node_exporter:
  enabled: false
`
	tDir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	writeFile(t, tDir, "exporters-1.yml", configStr)
	fileFS := fmt.Sprintf("file://%s", tDir)
	loaderCfg := LoaderConfig{
		Sources:       nil,
		TemplatePaths: []string{fileFS},
	}
	cmf, err := NewConfigLoader(loaderCfg)
	assert.Nil(t, err)
	configs, err := cmf.processExporters()
	assert.Len(t, configs, 2)
	for _, cfg := range configs {
		switch v := cfg.Config.(type) {
		default:
			fmt.Printf("unexpected type %T", v)
		case *windows_exporter.Config:
			assert.True(t, "one,two,three" == v.EnabledCollectors)
		case *node_exporter.Config:
			assert.True(t, false == cfg.Common.Enabled)
		}
	}
	_ = os.RemoveAll(tDir)
}

func writeFile(t *testing.T, directory string, path string, contents string) {
	fullpath := filepath.Join(directory, path)
	err := ioutil.WriteFile(fullpath, []byte(contents), fs.ModePerm)
	assert.Nil(t, err)
}

/*
func TestConfigMakerWithMultipleFiles(t *testing.T) {
	configStr := `name: bob
`
	configStr2 := `name: tommy
`
	cmf := NewComfigurator()
	tDir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	fullpath := filepath.Join(tDir, "test.yml")
	err = ioutil.WriteFile(fullpath, []byte(configStr), fs.ModePerm)

	badpath := filepath.Join(tDir, "test1.yml")
	err = ioutil.WriteFile(badpath, []byte(configStr2), fs.ModePerm)

	assert.Nil(t, err)
	fileFS := fmt.Sprintf("file://%s", tDir)
	configs, err := cmf.GenerateConfigsFromPath(fileFS, "*.yml", func() interface{} {
		return &TestConfig{}
	})
	assert.Nil(t, err)
	assert.True(t, len(configs) == 2)

	_ = os.RemoveAll(tDir)
}

func TestS3(t *testing.T) {
	backend := s3mem.New()
	faker := gofakes3.New(backend)

	srv := httptest.NewServer(faker.Server())
	backend.CreateBucket("mybucket")
	t.Cleanup(srv.Close)
	configStr := `name: bob`
	_, err := backend.PutObject(
		"mybucket",
		"test.yml",
		map[string]string{"Content-Type": "application/yaml"},
		bytes.NewBufferString(configStr),
		int64(len(configStr)),
	)

	u, err := url.Parse(srv.URL)
	os.Setenv("AWS_ANON", "true")
	defer os.Unsetenv("AWS_ANON")

	s3Url := "s3://mybucket/?region=us-east-1&disableSSL=true&s3ForcePathStyle=true&endpoint=" + u.Host
	assert.NoError(t, err)
	cmf := NewComfigurator()

	configs, err := cmf.GenerateConfigsFromPath(s3Url, "*.yml", func() interface{} {
		return &TestConfig{}
	})
	assert.Nil(t, err)
	assert.True(t, len(configs) == 1)
	found := false
	for k, v := range configs {
		if strings.HasSuffix(k, "test.yml") {
			assert.True(t, v.(*TestConfig).Name == "bob")
			found = true
		}
	}
	assert.True(t, found)
}

func TestTemplate(t *testing.T) {
	configStr := `name: {{ .Get "name" }}`
	tDir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	fullpath := filepath.Join(tDir, "test.yml")
	err = ioutil.WriteFile(fullpath, []byte(configStr), fs.ModePerm)
	assert.Nil(t, err)
	fileFS := fmt.Sprintf("file://%s", tDir)
	cmf := NewComfigurator()
	kvg := NewKVStoreGateway()
	memstore := NewMemoryStore()
	memstore.Cache["name"] = "bob"
	kvg.AddStore(memstore)
	cmf.AddKVStoreGateway(kvg)

	configs, err := cmf.GenerateConfigsFromPath(fileFS, "*.yml", func() interface{} {
		return &TestConfig{}
	})

	assert.Nil(t, err)
	assert.True(t, len(configs) == 1)
	found := false
	for k, v := range configs {
		if strings.HasSuffix(k, "test.yml") {
			assert.True(t, v.(*TestConfig).Name == "bob")
			found = true
		}
	}
	assert.True(t, found)
	_ = os.RemoveAll(tDir)
}

type TestConfig struct {
	Name string `yaml:"name"`
}*/
