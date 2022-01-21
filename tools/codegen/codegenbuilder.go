package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/grafana/agent/pkg/integrations/shared"
)

type codeGen struct {
}

func (c *codeGen) generateConfigMeta() []configMeta {
	configMetas := make([]configMeta, 0)
	for _, i := range v1Configs {
		configMetas = append(configMetas, newConfigMeta(i.Config, i.DefaultConfig, i.Type, i.IsV1))
	}
	return configMetas
}

func (c *codeGen) generateV2ConfigMeta() []configMeta {
	configMetas := make([]configMeta, 0)
	for _, i := range v2Configs {
		configMetas = append(configMetas, newConfigMeta(i.Config, i.DefaultConfig, i.Type, i.IsV1))
	}
	return configMetas
}

func (c *codeGen) createV1Config() string {
	configs := c.generateConfigMeta()
	contents, err := ioutil.ReadFile(filepath.Join(".", "v1template.tmpl"))
	if err != nil {
		panic(err)
	}
	cfgTemplate, err := template.New("v1").Parse(string(contents))
	if err != nil {
		panic(err)
	}

	buffer := bytes.Buffer{}
	err = cfgTemplate.Execute(&buffer, configs)
	if err != nil {
		panic(err)
	}
	formattedFile, err := format.Source(buffer.Bytes())
	if err != nil {
		panic(err)
	}
	return string(formattedFile)
}

func (c *codeGen) createV2Config() string {
	configs := c.generateV2ConfigMeta()
	contents, err := ioutil.ReadFile(filepath.Join(".", "v2template.tmpl"))
	if err != nil {
		panic(err)
	}
	cfgTemplate, err := template.New("shared").Parse(string(contents))

	if err != nil {
		panic(err)
	}
	buffer := bytes.Buffer{}

	err = cfgTemplate.Execute(&buffer, configs)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	formattedFile, err := format.Source(buffer.Bytes())
	if err != nil {
		panic(err)
	}
	return string(formattedFile)
}

type configMeta struct {
	Name          string
	ConfigStruct  string
	DefaultConfig string
	PackageName   string
	Type          shared.Type
	IsV1          bool
	PackagePath   string
}

func newConfigMeta(c interface{}, defaultConfig interface{}, p shared.Type, isV1 bool) configMeta {
	path := reflect.TypeOf(c).PkgPath()
	// If system if v2 then c is a pointer so need to get the Elem
	if path == "" {
		path = reflect.TypeOf(c).Elem().PkgPath()
	}
	configType := fmt.Sprintf("%T", c)
	configType = strings.ReplaceAll(configType, "*", "")
	packageName := strings.ReplaceAll(configType, ".Config", "")
	name := strings.ReplaceAll(configType, ".Config", "")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.Title(name)
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "*", "")
	var dc string
	if defaultConfig != nil {
		dc = fmt.Sprintf("%T", defaultConfig)
	}

	return configMeta{
		Name:          name,
		ConfigStruct:  configType,
		DefaultConfig: dc,
		PackageName:   packageName,
		Type:          p,
		IsV1:          isV1,
		PackagePath:   path,
	}
}
