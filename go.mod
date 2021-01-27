module github.com/grafana/agent

go 1.12

require (
	contrib.go.opencensus.io/exporter/prometheus v0.2.0
	github.com/aws/aws-sdk-go v1.35.31
	github.com/cortexproject/cortex v1.6.1-0.20210121163251-a59b811d5952
	github.com/drone/envsubst v1.0.2
	github.com/go-kit/kit v0.10.0
	github.com/go-playground/validator/v10 v10.4.0 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.3
	github.com/google/dnsmasq_exporter v0.0.0-00010101000000-000000000000
	github.com/gorilla/mux v1.8.0
	github.com/grafana/loki v1.6.2-0.20210126101127-33abab5e4203
	github.com/hashicorp/consul/api v1.7.0
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/joshdk/go-junit v0.0.0-20200702055522-6efcf4050909 // indirect
	github.com/jsternberg/zap-logfmt v1.2.0
	github.com/justwatchcom/elasticsearch_exporter v1.1.0
	github.com/miekg/dns v1.1.35
	github.com/ncabatoff/process-exporter v0.7.5
	github.com/oklog/run v1.1.0
	github.com/olekukonko/tablewriter v0.0.2
	github.com/oliver006/redis_exporter v1.15.0
	github.com/opentracing-contrib/go-grpc v0.0.0-20191001143057-db30781987df
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/common v0.15.0
	github.com/prometheus/consul_exporter v0.7.2-0.20210127095228-584c6de19f23
	github.com/prometheus/memcached_exporter v0.8.0
	github.com/prometheus/mysqld_exporter v0.0.0-00010101000000-000000000000
	github.com/prometheus/node_exporter v1.0.1
	github.com/prometheus/procfs v0.2.0
	github.com/prometheus/prometheus v1.8.2-0.20201119181812-c8f810083d3f
	github.com/prometheus/statsd_exporter v0.18.1-0.20201124082027-8b2b4c1a2b49
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/weaveworks/common v0.0.0-20210112142934-23c8d7fa6120
	github.com/wrouesnel/postgres_exporter v0.0.0-00010101000000-000000000000
	go.opencensus.io v0.22.5
	go.opentelemetry.io/collector v0.16.0
	go.uber.org/atomic v1.7.0
	go.uber.org/zap v1.16.0
	google.golang.org/grpc v1.33.2
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

// Needed for Cortex's dependencies to work properly.
replace (
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
	k8s.io/client-go => k8s.io/client-go v0.19.2
)

// Replace directives from Prometheus
replace k8s.io/klog => github.com/simonpasquier/klog-gokit v0.1.0

// Replace directives from Cortex
replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
	github.com/gocql/gocql => github.com/grafana/gocql v0.0.0-20200605141915-ba5dc39ece85
	github.com/hpcloud/tail => github.com/grafana/tail v0.0.0-20201004203643-7aa4e4a91f03
	github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.0
)

replace github.com/prometheus/prometheus => github.com/grafana/prometheus v1.8.2-0.20210111220521-c0d5de2f0ee3

replace gopkg.in/yaml.v2 => github.com/rfratto/go-yaml v0.0.0-20200521142311-984fc90c8a04

// TODO(rfratto): remove forks when changes are merged upstream
replace (
	github.com/google/dnsmasq_exporter => github.com/grafana/dnsmasq_exporter v0.2.1-0.20201029182940-e5169b835a23
	github.com/ncabatoff/process-exporter => github.com/grafana/process-exporter v0.7.3-0.20210106202358-831154072e2a
	github.com/prometheus/mysqld_exporter => github.com/grafana/mysqld_exporter v0.12.2-0.20201015182516-5ac885b2d38a
	github.com/wrouesnel/postgres_exporter => github.com/grafana/postgres_exporter v0.8.1-0.20201106170118-5eedee00c1db
)

// Required for redis_exporter, which is incompatible with v2.0.0+incompatible.
replace github.com/gomodule/redigo => github.com/gomodule/redigo v1.8.2
