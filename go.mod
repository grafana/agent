module github.com/grafana/agent

go 1.12

require (
	github.com/Azure/go-autorest/autorest v0.9.4 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.1 // indirect
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/cortexproject/cortex v0.4.0
	github.com/go-kit/kit v0.9.0
	github.com/grafana/tail v0.0.0-20200127140945-4647d4b312f2
	github.com/oklog/run v1.0.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.3.0
	github.com/prometheus/common v0.8.0
	github.com/prometheus/prometheus v1.8.2-0.20200106144642-d9613e5c466c
	github.com/weaveworks/common v0.0.0-20190822150010-afb9996716e4
	gopkg.in/yaml.v2 v2.2.7
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

replace github.com/prometheus/prometheus => /home/robert/dev/prometheus/prometheus
