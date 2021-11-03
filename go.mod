module github.com/grafana/agent

go 1.16

require (
	cloud.google.com/go/pubsub v1.5.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.4.0
	github.com/Shopify/sarama v1.30.0
	github.com/cortexproject/cortex v1.10.1-0.20211014125347-85c378182d0d
	github.com/davidmparrott/kafka_exporter/v2 v2.0.1
	github.com/drone/envsubst/v2 v2.0.0-20210730161058-179042472c46
	github.com/fatih/color v1.12.0 // indirect
	github.com/fatih/structs v1.1.0
	github.com/go-kit/kit v0.11.0
	github.com/go-kit/log v0.2.0
	github.com/go-logfmt/logfmt v0.5.1
	github.com/go-logr/logr v1.0.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/google/dnsmasq_exporter v0.0.0-00010101000000-000000000000
	github.com/google/go-jsonnet v0.17.0
	github.com/gorilla/mux v1.8.0
	github.com/grafana/dskit v0.0.0-20211011144203-3a88ec0b675f
	github.com/grafana/loki v1.6.2-0.20211021114919-0ae0d4da122d
	github.com/grafana/tempo v1.0.1
	github.com/hashicorp/consul/api v1.11.0
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-getter v1.5.3
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/infinityworks/github-exporter v0.0.0-20201016091012-831b72461034
	github.com/jsternberg/zap-logfmt v1.2.0
	github.com/lib/pq v1.10.1
	github.com/miekg/dns v1.1.43
	github.com/mitchellh/reflectwalk v1.0.2
	github.com/ncabatoff/process-exporter v0.7.5
	github.com/oklog/run v1.1.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/oliver006/redis_exporter v1.27.1
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/loadbalancingexporter v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanmetricsprocessor v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jaegerreceiver v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/opencensusreceiver v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zipkinreceiver v0.36.0
	github.com/opentracing-contrib/go-grpc v0.0.0-20210225150812-73cb765af46e
	github.com/opentracing-contrib/go-stdlib v1.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/percona/mongodb_exporter v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/prometheus-community/elasticsearch_exporter v1.2.1
	github.com/prometheus-community/postgres_exporter v0.10.0
	github.com/prometheus-community/windows_exporter v0.0.0-00010101000000-000000000000
	github.com/prometheus-operator/prometheus-operator v0.47.0
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.47.0
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.32.1
	github.com/prometheus/consul_exporter v0.7.2-0.20210127095228-584c6de19f23
	github.com/prometheus/memcached_exporter v0.9.0
	github.com/prometheus/mysqld_exporter v0.13.0
	github.com/prometheus/node_exporter v1.0.1
	github.com/prometheus/procfs v0.6.1-0.20210313121648-b565fefb1664
	github.com/prometheus/prometheus v1.8.2-0.20211102100715-d4c83da6d252
	github.com/prometheus/statsd_exporter v0.22.2
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/weaveworks/common v0.0.0-20211015155308-ebe5bdc2c89e
	go.opencensus.io v0.23.0
	go.opentelemetry.io/collector v0.36.0
	go.opentelemetry.io/collector/model v0.36.0
	go.opentelemetry.io/otel/metric v0.23.0
	go.opentelemetry.io/otel/trace v1.0.0-RC3
	go.uber.org/atomic v1.9.0
	go.uber.org/zap v1.19.1
	golang.org/x/sys v0.0.0-20211020174200-9d6173849985
	google.golang.org/grpc v1.41.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.9.0-beta.5
	sigs.k8s.io/yaml v1.2.0
)

// Replace directives from Prometheus
replace (
	github.com/go-logr/logr => github.com/go-logr/logr v0.4.0
	k8s.io/klog => github.com/simonpasquier/klog-gokit v0.3.0
	k8s.io/klog/v2 => github.com/rlankfo/klog-gokit/v3 v3.0.1-0.20211103030435-2602604e10dd
)

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
	github.com/prometheus/prometheus => github.com/grafana/prometheus v1.8.2-0.20211103031328-89bb32ee4ae7
	gopkg.in/yaml.v2 => github.com/rfratto/go-yaml v0.0.0-20200521142311-984fc90c8a04
	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.0
	k8s.io/client-go => k8s.io/client-go v0.21.0
)

// TODO(rfratto): remove forks when changes are merged upstream
replace (
	github.com/google/dnsmasq_exporter => github.com/grafana/dnsmasq_exporter v0.2.1-0.20211004193725-8712c75623e6
	github.com/infinityworks/github-exporter => github.com/rgeyer/github-exporter v0.0.0-20210722215637-d0cec2ee0dc8
	github.com/ncabatoff/process-exporter => github.com/grafana/process-exporter v0.7.3-0.20210106202358-831154072e2a
	github.com/percona/exporter_shared => github.com/rlankfo/exporter_shared v0.7.4-0.20211028185902-ab40c12bd34a
	github.com/percona/mongodb_exporter => github.com/grafana/mongodb_exporter v0.20.8-0.20211006135645-bef0f0239601
	github.com/percona/percona-toolkit => github.com/rlankfo/percona-toolkit v0.0.0-20211028191359-7aada1bf148f
	github.com/prometheus-community/postgres_exporter => github.com/grafana/postgres_exporter v0.8.1-0.20210722175051-db35d7c2f520
	github.com/prometheus-community/windows_exporter => github.com/grafana/windows_exporter v0.15.1-0.20211019183116-592dfa92f9fd
	github.com/prometheus/mysqld_exporter => github.com/grafana/mysqld_exporter v0.12.2-0.20201015182516-5ac885b2d38a
)

// Required for redis_exporter, which is incompatible with v2.0.0+incompatible.
replace github.com/gomodule/redigo => github.com/gomodule/redigo v1.8.2

// Excluding fixes a conflict in test packages and allows "go mod tidy" to run.
exclude google.golang.org/grpc/examples v0.0.0-20200728065043-dfc0c05b2da9

// Used for /-/reload
replace github.com/weaveworks/common => github.com/rlankfo/weaveworks-common v0.0.0-20211028201502-f0486066ff0a

// loadbalancingexporter uses non-fixed version of batchpertrace which fetches latest and causes problems
replace github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchpersignal => github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchpersignal v0.30.0

// Replacing for an internal fork that exposes internal folders
// Some funtionalities of the collector have been made internal and it's more difficult to build and configure pipelines in the newer versions.
// This is a temporary solution while a new configuration design is discussed for the collector (ref: https://github.com/open-telemetry/opentelemetry-collector/issues/3482).
replace go.opentelemetry.io/collector => github.com/grafana/opentelemetry-collector v0.4.1-0.20211008102704-a5d23001cb71

// Jaeger v1.16.0 can't be run with go@1.16 (https://github.com/jaegertracing/jaeger/issues/3268)
// Problem was fixed in https://github.com/jaegertracing/jaeger/issues/3268
// Replace can't be removed when all dependencies update to >=v1.27
replace github.com/jaegertracing/jaeger => github.com/jaegertracing/jaeger v1.27.0

// Replacement necessary for windows_exporter so that we can use gokit logging and not the old prometheus logging
replace github.com/leoluk/perflib_exporter v0.1.0 => github.com/grafana/perflib_exporter v0.1.1-0.20211013152516-e37e14fb8b0a
