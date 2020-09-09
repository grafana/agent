module github.com/grafana/agent

go 1.12

require (
	github.com/cortexproject/cortex v1.2.1-0.20200803161316-7014ff11ed70
	github.com/go-kit/kit v0.10.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/gorilla/mux v1.8.0
	github.com/grafana/loki v1.6.1
	github.com/jsternberg/zap-logfmt v1.0.0
	github.com/ncabatoff/process-exporter v0.0.0-00010101000000-000000000000
	github.com/oklog/run v1.1.0
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/opentracing-contrib/go-grpc v0.0.0-20191001143057-db30781987df
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.11.1
	github.com/prometheus/node_exporter v1.0.1
	github.com/prometheus/procfs v0.1.3
	github.com/prometheus/prometheus v1.8.2-0.20200727090838-6f296594a852
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/weaveworks/common v0.0.0-20200625145055-4b1847531bc9
	go.opentelemetry.io/collector v0.6.1
	go.uber.org/atomic v1.6.0
	go.uber.org/zap v1.14.1
	google.golang.org/grpc v1.31.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.3.0
)

// Needed for Cortex's dependencies to work properly.
replace (
	github.com/sercand/kuberesolver => github.com/sercand/kuberesolver v2.1.0+incompatible
	go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
	google.golang.org/api => google.golang.org/api v0.14.0
	google.golang.org/grpc => google.golang.org/grpc v1.25.1
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

replace github.com/prometheus/prometheus => github.com/grafana/prometheus v1.8.2-0.20200821135656-2efe42db3b77

replace gopkg.in/yaml.v2 => github.com/rfratto/go-yaml v0.0.0-20200521142311-984fc90c8a04

// TODO(rfratto): remove fork when changes are merged upstream
replace github.com/ncabatoff/process-exporter => github.com/grafana/process-exporter v0.7.3-0.20200902205007-6343dc1182cf
