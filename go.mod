module github.com/grafana/agent

go 1.12

require (
	github.com/Azure/go-autorest/autorest v0.9.4 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.1 // indirect
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/cortexproject/cortex v0.4.0
	github.com/go-kit/kit v0.9.0
	github.com/prometheus/client_golang v1.3.0
	github.com/prometheus/prometheus v1.8.2-0.20200106144642-d9613e5c466c
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
