module github.com/grafana/agent

go 1.16

require (
	contrib.go.opencensus.io/exporter/prometheus v0.3.0
	github.com/cortexproject/cortex v1.8.2-0.20210428155238-d382e1d80eaf
	github.com/drone/envsubst v1.0.2
	github.com/fatih/structs v1.1.0
	github.com/go-kit/kit v0.10.0
	github.com/go-logfmt/logfmt v0.5.0
	github.com/go-logr/logr v0.4.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/google/dnsmasq_exporter v0.0.0-00010101000000-000000000000
	github.com/google/go-jsonnet v0.17.0
	github.com/gorilla/mux v1.8.0
	github.com/grafana/loki v1.6.2-0.20210429132126-d88f3996eaa2
	github.com/hashicorp/consul/api v1.8.1
	github.com/hashicorp/go-getter v1.5.3
	github.com/jsternberg/zap-logfmt v1.2.0
	github.com/justwatchcom/elasticsearch_exporter v1.1.0
	github.com/miekg/dns v1.1.41
	github.com/ncabatoff/process-exporter v0.7.5
	github.com/oklog/run v1.1.0
	github.com/olekukonko/tablewriter v0.0.2
	github.com/oliver006/redis_exporter v1.15.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/loadbalancingexporter v0.29.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanmetricsprocessor v0.29.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor v0.29.0
	github.com/opentracing-contrib/go-grpc v0.0.0-20210225150812-73cb765af46e
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus-community/windows_exporter v0.0.0-00010101000000-000000000000
	github.com/prometheus-operator/prometheus-operator v0.47.0
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.47.0
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.29.0
	github.com/prometheus/consul_exporter v0.7.2-0.20210127095228-584c6de19f23
	github.com/prometheus/memcached_exporter v0.8.0
	github.com/prometheus/mysqld_exporter v0.0.0-00010101000000-000000000000
	github.com/prometheus/node_exporter v1.0.1
	github.com/prometheus/procfs v0.6.1-0.20210313121648-b565fefb1664
	github.com/prometheus/prometheus v1.8.2-0.20210621150501-ff58416a0b02
	github.com/prometheus/statsd_exporter v0.21.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/weaveworks/common v0.0.0-20210419092856-009d1eebd624
	github.com/wrouesnel/postgres_exporter v0.0.0-00010101000000-000000000000
	go.opencensus.io v0.23.0
	go.opentelemetry.io/collector v0.29.0
	go.uber.org/atomic v1.8.0
	go.uber.org/zap v1.17.0
	golang.org/x/sys v0.0.0-20210611083646-a4fc73990273
	google.golang.org/grpc v1.38.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.9.0-beta.5
	sigs.k8s.io/yaml v1.2.0
)

// Replace directives from Prometheus
replace k8s.io/klog => github.com/simonpasquier/klog-gokit v0.1.0

// Replace directives from Cortex
replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
	github.com/gocql/gocql => github.com/grafana/gocql v0.0.0-20200605141915-ba5dc39ece85
	github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.0
)

// Replace directives from Loki
replace (
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v36.2.0+incompatible
	github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
	github.com/hpcloud/tail => github.com/grafana/tail v0.0.0-20201004203643-7aa4e4a91f03
	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/client-go => k8s.io/client-go v0.21.0
)

replace github.com/prometheus/prometheus => github.com/grafana/prometheus v1.8.2-0.20210608193638-7b78de4ccffc

replace gopkg.in/yaml.v2 => github.com/rfratto/go-yaml v0.0.0-20200521142311-984fc90c8a04

// TODO(rfratto): remove forks when changes are merged upstream
replace (
	github.com/google/dnsmasq_exporter => github.com/grafana/dnsmasq_exporter v0.2.1-0.20201029182940-e5169b835a23
	github.com/ncabatoff/process-exporter => github.com/grafana/process-exporter v0.7.3-0.20210106202358-831154072e2a
	github.com/prometheus-community/windows_exporter => github.com/grafana/windows_exporter v0.15.1-0.20210325142439-9e8f66d53433
	github.com/prometheus/mysqld_exporter => github.com/grafana/mysqld_exporter v0.12.2-0.20201015182516-5ac885b2d38a
	github.com/wrouesnel/postgres_exporter => github.com/grafana/postgres_exporter v0.8.1-0.20201106170118-5eedee00c1db

)

// Required for redis_exporter, which is incompatible with v2.0.0+incompatible.
replace github.com/gomodule/redigo => github.com/gomodule/redigo v1.8.2

// Excluding fixes a conflict in test packages and allows "go mod tidy" to run.
exclude google.golang.org/grpc/examples v0.0.0-20200728065043-dfc0c05b2da9

// Used for /-/reload
replace github.com/weaveworks/common => github.com/rfratto/weaveworks-common v0.0.0-20210326192855-c95210d58ba7

// loadbalancingexporter uses non-fixed version of batchpertrace which fetches latest and causes problems
replace github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchpersignal => github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchpersignal v0.29.0

// Replacing for an internal fork that exposes internal folders
// Some funtionalities of the collector have been made internal and it's more difficult to build and configure pipelines in the newer versions.
// This is a temporary solution while a new configuration design is discussed for the collector (ref: https://github.com/open-telemetry/opentelemetry-collector/issues/3482).
replace go.opentelemetry.io/collector => github.com/grafana/opentelemetry-collector v0.4.1-0.20210625154936-5ed8906327ca

// Pin prometheus dependencies
replace (
	github.com/prometheus/common => github.com/prometheus/common v0.23.0
	github.com/prometheus/statsd_exporter => github.com/prometheus/statsd_exporter v0.18.1-0.20201124082027-8b2b4c1a2b49
)
