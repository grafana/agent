module github.com/grafana/agent

go 1.12

require (
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/cortexproject/cortex v0.4.0
	github.com/go-kit/kit v0.10.0
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/oklog/run v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/common v0.9.1
	github.com/prometheus/procfs v0.0.11 // indirect
	github.com/prometheus/prometheus v1.8.2-0.20200106144642-d9613e5c466c
	github.com/sercand/kuberesolver v2.4.0+incompatible // indirect
	github.com/stretchr/testify v1.4.0
	github.com/weaveworks/common v0.0.0-20190822150010-afb9996716e4
	go.uber.org/atomic v1.5.0
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	google.golang.org/grpc v1.28.0 // indirect
	gopkg.in/yaml.v2 v2.2.8
)

// Replace directives from Prometheus
replace k8s.io/klog => github.com/simonpasquier/klog-gokit v0.1.0

// Replace directives from Cortex
replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v36.2.0+incompatible
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.0+incompatible
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
)

// Temporarily use a fork for memory improvements (see #5)
replace github.com/prometheus/prometheus => github.com/grafana/prometheus v1.8.2-0.20200326205120-ca4918999cce
