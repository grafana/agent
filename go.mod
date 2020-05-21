module github.com/grafana/agent

go 1.12

require (
	github.com/cortexproject/cortex v1.0.1-0.20200409122148-163437e76cad
	github.com/go-kit/kit v0.10.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.0
	github.com/gorilla/mux v1.7.3
	github.com/oklog/run v1.1.0
	github.com/opentracing-contrib/go-grpc v0.0.0-20180928155321-4b5a12d3ff02
	github.com/opentracing/opentracing-go v1.1.1-0.20200124165624-2876d2018785
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/common v0.9.1
	github.com/prometheus/prometheus v1.8.2-0.20200213233353-b90be6f32a33
	github.com/spf13/cobra v0.0.3
	github.com/stretchr/testify v1.5.1
	github.com/weaveworks/common v0.0.0-20200310113808-2708ba4e60a4
	go.uber.org/atomic v1.6.0
	google.golang.org/grpc v1.29.0
	gopkg.in/yaml.v2 v2.2.8
)

// Needed for Cortex's dependencies to work properly.
replace (
	google.golang.org/api => google.golang.org/api v0.14.0
	google.golang.org/grpc => google.golang.org/grpc v1.25.1
)

// Replace directives from Prometheus
replace k8s.io/klog => github.com/simonpasquier/klog-gokit v0.1.0

// Replace directives from Cortex
replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v36.2.0+incompatible
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.0+incompatible
)

replace github.com/prometheus/prometheus => github.com/grafana/prometheus v1.8.2-0.20200518163447-007aa83a0a1f

replace gopkg.in/yaml.v2 => github.com/rfratto/go-yaml v0.0.0-20200521142311-984fc90c8a04
