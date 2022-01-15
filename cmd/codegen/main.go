package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"strings"

	"github.com/grafana/agent/pkg/integrations/shared"

	"github.com/grafana/agent/pkg/integrations/v1/windows_exporter"

	"github.com/grafana/agent/pkg/integrations/v1/agent"
	"github.com/grafana/agent/pkg/integrations/v1/cadvisor"
	"github.com/grafana/agent/pkg/integrations/v1/consul_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/dnsmasq_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/elasticsearch_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/github_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/kafka_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/memcached_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/mongodb_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/mysqld_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/node_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/postgres_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/process_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/redis_exporter"
	"github.com/grafana/agent/pkg/integrations/v1/statsd_exporter"
)

var v1Configs = []configMeta{
	newConfigMeta(agent.Config{}, nil, shared.TypeSingleton),
	newConfigMeta(cadvisor.Config{}, cadvisor.DefaultConfig, shared.TypeSingleton),
	newConfigMeta(consul_exporter.Config{}, consul_exporter.DefaultConfig, shared.TypeMultiplex),
	newConfigMeta(dnsmasq_exporter.Config{}, dnsmasq_exporter.DefaultConfig, shared.TypeMultiplex),
	newConfigMeta(elasticsearch_exporter.Config{}, elasticsearch_exporter.DefaultConfig, shared.TypeMultiplex),
	newConfigMeta(github_exporter.Config{}, github_exporter.DefaultConfig, shared.TypeMultiplex),
	newConfigMeta(kafka_exporter.Config{}, kafka_exporter.DefaultConfig, shared.TypeMultiplex),
	newConfigMeta(memcached_exporter.Config{}, memcached_exporter.DefaultConfig, shared.TypeMultiplex),
	newConfigMeta(mongodb_exporter.Config{}, nil, shared.TypeMultiplex),
	newConfigMeta(mysqld_exporter.Config{}, mysqld_exporter.DefaultConfig, shared.TypeMultiplex),
	newConfigMeta(node_exporter.Config{}, node_exporter.DefaultConfig, shared.TypeSingleton),
	newConfigMeta(postgres_exporter.Config{}, nil, shared.TypeMultiplex),
	newConfigMeta(process_exporter.Config{}, process_exporter.DefaultConfig, shared.TypeSingleton),
	newConfigMeta(redis_exporter.Config{}, redis_exporter.DefaultConfig, shared.TypeMultiplex),
	newConfigMeta(statsd_exporter.Config{}, statsd_exporter.DefaultConfig, shared.TypeSingleton),
	newConfigMeta(windows_exporter.Config{}, windows_exporter.DefaultConfig, shared.TypeSingleton),
}

func main() {

	codegen := codeGen{}
	v1Config := codegen.createV1Config(v1Configs)
	err := ioutil.WriteFile("./pkg/integrations/v1/config.go", []byte(v1Config), fs.ModePerm)
	if err != nil {
		panic(err)
	}

	v2Config := codegen.createV2Config(v1Configs)
	err = ioutil.WriteFile("./pkg/integrations/v2/config.go", []byte(v2Config), fs.ModePerm)
	if err != nil {
		panic(err)
	}

}

func newConfigMeta(c interface{}, defaultConfig interface{}, p shared.Type) configMeta {
	configType := fmt.Sprintf("%T", c)
	packageName := strings.ReplaceAll(configType, ".Config", "")
	name := strings.ReplaceAll(configType, ".Config", "")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.Title(name)
	name = strings.ReplaceAll(name, " ", "")
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
	}
}
