module github.com/grafana/agent

go 1.12

require (
	contrib.go.opencensus.io/exporter/prometheus v0.2.0
	github.com/cortexproject/cortex v1.4.1-0.20201022071705-85942c5703cf
	github.com/go-kit/kit v0.10.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.3
	github.com/google/dnsmasq_exporter v0.0.0-00010101000000-000000000000
	github.com/gorilla/mux v1.8.0
	github.com/grafana/loki v1.6.2-0.20201026154740-6978ee5d7387
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/jsternberg/zap-logfmt v1.2.0
	github.com/miekg/dns v1.1.31
	github.com/ncabatoff/process-exporter v0.0.0-00010101000000-000000000000
	github.com/oklog/run v1.1.0
	github.com/olekukonko/tablewriter v0.0.2
	github.com/oliver006/redis_exporter v1.12.2-0.20201026185157-c0a38f36299a
	github.com/opentracing-contrib/go-grpc v0.0.0-20191001143057-db30781987df
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.8.0
	github.com/prometheus/common v0.14.0
	github.com/prometheus/mysqld_exporter v0.0.0-00010101000000-000000000000
	github.com/prometheus/node_exporter v1.0.1
	github.com/prometheus/procfs v0.2.0
	github.com/prometheus/prometheus v1.8.2-0.20201014093524-73e2ce1bd643
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/weaveworks/common v0.0.0-20200914083218-61ffdd448099
	go.opencensus.io v0.22.4
	go.opentelemetry.io/collector v0.13.0
	go.uber.org/atomic v1.7.0
	go.uber.org/zap v1.16.0
	google.golang.org/grpc v1.32.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.3.0
)

// Needed for Cortex's dependencies to work properly.
replace (
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200520232829-54ba9589114f
	google.golang.org/api => google.golang.org/api v0.14.0
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
	k8s.io/client-go => k8s.io/client-go v0.18.5
)

// Replace directives from Prometheus
replace k8s.io/klog => github.com/simonpasquier/klog-gokit v0.1.0

// Replace directives from Cortex
replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
	github.com/gocql/gocql => github.com/grafana/gocql v0.0.0-20200605141915-ba5dc39ece85
	github.com/hpcloud/tail => github.com/grafana/tail v0.0.0-20191024143944-0b54ddf21fe7
	github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.0
)

replace github.com/prometheus/prometheus => github.com/grafana/prometheus v1.8.2-0.20201021200247-cf00050ed1e9

replace gopkg.in/yaml.v2 => github.com/rfratto/go-yaml v0.0.0-20200521142311-984fc90c8a04

// TODO(rfratto): remove forks when changes are merged upstream
replace (
	github.com/google/dnsmasq_exporter => github.com/grafana/dnsmasq_exporter v0.2.1-0.20201029182940-e5169b835a23
	github.com/ncabatoff/process-exporter => github.com/grafana/process-exporter v0.7.3-0.20200902205007-6343dc1182cf
	github.com/prometheus/mysqld_exporter => github.com/grafana/mysqld_exporter v0.12.2-0.20201015182516-5ac885b2d38a
)

// Required for redis_exporter, which is incompatible with v2.0.0+incompatible.
replace github.com/gomodule/redigo => github.com/gomodule/redigo v1.8.2
