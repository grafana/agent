package k8sattributes_test

import (
	"testing"

	"github.com/grafana/agent/component/otelcol/processor/k8sattributes"
	"github.com/grafana/river"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor"
	"github.com/stretchr/testify/require"
)

func Test_Extract(t *testing.T) {
	cfg := `
		auth_type = "kubeConfig"

		extract {
			label {
				from      = "pod"
				key_regex = "(.*)/(.*)"
				tag_name  = "$1.$2"
			}
	
			metadata = [
				"k8s.namespace.name",
				"k8s.job.name",
				"k8s.node.name",
			]
		}
	
		output {
			// no-op: will be overridden by test code.
		}
	`
	var args k8sattributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*k8sattributesprocessor.Config)

	authType := &otelObj.AuthType
	require.True(t, "kubeConfig" == *authType)

	extract := &otelObj.Extract
	require.Equal(t, []string{"k8s.namespace.name", "k8s.job.name", "k8s.node.name"}, extract.Metadata)
}

func Test_ExtractAnnotations(t *testing.T) {
	cfg := `
		extract {
			annotation {
				key_regex = "opentel.*"
				from      = "pod"
			}

			label {
				key_regex = "opentel.*"
				from      = "pod"
			}
	
			metadata = [
				"k8s.namespace.name",
				"k8s.job.name",
				"k8s.node.name",
			]
		}
	
		output {
			// no-op: will be overridden by test code.
		}
	`
	var args k8sattributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*k8sattributesprocessor.Config)

	extract := &otelObj.Extract
	require.Len(t, extract.Annotations, 1)
	require.Equal(t, extract.Annotations[0].KeyRegex, "opentel.*")
	require.Equal(t, extract.Annotations[0].From, "pod")

	require.Len(t, extract.Labels, 1)
	require.Equal(t, extract.Labels[0].KeyRegex, "opentel.*")
	require.Equal(t, extract.Labels[0].From, "pod")
}

func Test_FilterNodeEnvironmentVariable(t *testing.T) {
	cfg := `
		filter {
			node = env("K8S_ATTRIBUTES_TEST_HOSTNAME")
		}

		output {
			// no-op: will be overridden by test code.
		}
	`
	var args k8sattributes.Arguments
	testHostname := "test-hostname"
	t.Setenv("K8S_ATTRIBUTES_TEST_HOSTNAME", testHostname)
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*k8sattributesprocessor.Config)

	filter := &otelObj.Filter
	require.Equal(t, testHostname, filter.Node)
}

func Test_FilterNamespace(t *testing.T) {
	cfg := `
		filter {
			namespace = "mynamespace"
		}

		output {
			// no-op: will be overridden by test code.
		}
	`
	var args k8sattributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*k8sattributesprocessor.Config)

	filter := &otelObj.Filter
	require.Equal(t, "mynamespace", filter.Namespace)
}

func Test_FilterOps(t *testing.T) {
	cfg := `
		filter {
			label {
				key = "key1"
				value = "value1"
			}
			label {
				key = "key2"
				value = "value2"
				op = "not-equals"
			}
			field {
				key = "key1"
				value = "value1"
			}
			field {
				key = "key2"
				value = "value2"
				op = "not-equals"
			}
		}

		output {
			// no-op: will be overridden by test code.
		}
	`
	var args k8sattributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*k8sattributesprocessor.Config)

	filter := &otelObj.Filter

	labels := &filter.Labels
	require.Len(t, *labels, 2)
	require.Equal(t, (*labels)[0].Key, "key1")
	require.Equal(t, (*labels)[0].Value, "value1")
	require.Equal(t, (*labels)[1].Key, "key2")
	require.Equal(t, (*labels)[1].Value, "value2")
	require.Equal(t, (*labels)[1].Op, "not-equals")

	fields := &filter.Fields
	require.Len(t, *fields, 2)
	require.Equal(t, (*fields)[0].Key, "key1")
	require.Equal(t, (*fields)[0].Value, "value1")
	require.Equal(t, (*fields)[1].Key, "key2")
	require.Equal(t, (*fields)[1].Value, "value2")
	require.Equal(t, (*fields)[1].Op, "not-equals")
}

func Test_DefaultToServiceAccountAuth(t *testing.T) {
	cfg := `
		output {
			// no-op: will be overridden by test code.
		}
	`
	var args k8sattributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*k8sattributesprocessor.Config)

	authType := &otelObj.AuthType
	require.True(t, *authType == "serviceAccount") // Default value
}

func Test_PodAssociation(t *testing.T) {
	cfg := `
		pod_association {
			source {
				from = "resource_attribute"
				name = "k8s.pod.ip"
			}
		}
		pod_association {
			source {
				from = "resource_attribute"
				name = "k8s.pod.uid"
			}
		}
		pod_association {
			source {
				from = "connection"
			}
		}

		output {
			// no-op: will be overridden by test code.
		}
	`
	var args k8sattributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*k8sattributesprocessor.Config)

	associations := &otelObj.Association
	require.Len(t, *associations, 3)

	association := (*associations)[0]
	require.Len(t, association.Sources, 1)
	require.Equal(t, "resource_attribute", association.Sources[0].From)
	require.Equal(t, "k8s.pod.ip", association.Sources[0].Name)

	association = (*associations)[1]
	require.Len(t, association.Sources, 1)
	require.Equal(t, "resource_attribute", association.Sources[0].From)
	require.Equal(t, "k8s.pod.uid", association.Sources[0].Name)

	association = (*associations)[2]
	require.Len(t, association.Sources, 1)
	require.Equal(t, "connection", association.Sources[0].From)
}

func Test_PodAssociationPair(t *testing.T) {
	cfg := `
		pod_association {
			source {
				from = "resource_attribute"
				name = "k8s.pod.ip"
			}
		}
		pod_association {	
			source {
				from = "resource_attribute"
				name = "k8s.pod.uid"
			}
			source {
				from = "connection"	
			}
		}

		output {
			// no-op: will be overridden by test code.
		}
	`
	var args k8sattributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*k8sattributesprocessor.Config)

	associations := &otelObj.Association
	require.Len(t, *associations, 2)

	association := (*associations)[0]
	require.Len(t, association.Sources, 1)
	require.Equal(t, "resource_attribute", association.Sources[0].From)
	require.Equal(t, "k8s.pod.ip", association.Sources[0].Name)

	association = (*associations)[1]
	require.Len(t, association.Sources, 2)
	require.Equal(t, "resource_attribute", association.Sources[0].From)
	require.Equal(t, "k8s.pod.uid", association.Sources[0].Name)

	require.Equal(t, "connection", association.Sources[1].From)
}

func Test_Passthrough(t *testing.T) {
	cfg := `
		passthrough = true

		output {
			// no-op: will be overridden by test code.
		}
	`
	var args k8sattributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*k8sattributesprocessor.Config)

	require.True(t, otelObj.Passthrough)
}

func Test_Exclude(t *testing.T) {
	cfg := `
		exclude {
			pod {
				name = "jaeger-agent"
			}
			pod {
				name = "jaeger-collector"
			}
		}

		output {
			// no-op: will be overridden by test code.
		}
	`
	var args k8sattributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	convertedArgs, err := args.Convert()
	require.NoError(t, err)
	otelObj := (convertedArgs).(*k8sattributesprocessor.Config)

	exclude := &otelObj.Exclude
	require.Len(t, exclude.Pods, 2)
	require.Equal(t, "jaeger-agent", exclude.Pods[0].Name)
	require.Equal(t, "jaeger-collector", exclude.Pods[1].Name)
}
