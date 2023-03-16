package kubernetes

import (
	"fmt"
	"os"
	"strings"

	"github.com/grafana/agent/pkg/autodiscovery"
	"github.com/grafana/agent/pkg/river"
)

// Config defines the Kubernetes metadata we'll look for to scrape pods for
// metrics.
type Config struct {
	Values []string `river:"values,attr,optional"`
}

// K8S is an autodiscovery mechanism for Kubernetes pods..
type K8S struct {
	values []string
}

func (k *K8S) String() string {
	return "kubernetes"
}

// New creates a new auto-discovery K8S mechanism instance.
func New() (*K8S, error) {
	bb, err := os.ReadFile("pkg/autodiscovery/kubernetes/kubernetes.river")
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = river.Unmarshal(bb, &cfg)
	if err != nil {
		return nil, err
	}

	return &K8S{
		values: cfg.Values,
	}, nil
}

// Run check whether we're running in Kubernetes, and if so, filter for pods
// which expose __meta_* labels in the prometheus_sd_configs report.
func (m *K8S) Run() (*autodiscovery.Result, error) {
	if os.Getenv("KUBERNETES_SERVICE_HOST") == "" &&
		os.Getenv("KUBERNETES_SERVICE_PORT") == "" {
		return nil, fmt.Errorf("not running in a kubernetes cluster")
	}

	res := &autodiscovery.Result{}

	var abr string
	if len(m.values) > 0 {
		abr = buildAnnotationsBasedRule(m.values)
	}
	res.RiverConfig = fmt.Sprintf(`
discovery.kubernetes "pods" {
  role = "pod"
}

discovery.relabel "pods" {
  rule {
    action = "keep"
    regex = ".*-metrics"
    source_labels = ["__meta_kubernetes_pod_container_port_name"]
  }
  %s
  rule {
    action = "drop"
    regex = "Succeeded|Failed"
    source_labels = ["__meta_kubernetes_pod_phase"]
  }
  rule {
    action = "drop"
    regex = "false"
    source_labels = ["__meta_kubernetes_pod_annotation_agent_autodiscover"]
  }
  rule {
    action = "replace"
    regex = "(.+?)(\\:\\d+)?;(\\d+)"
    replacement = "$1:$3"
    source_labels = ["__address__", "__meta_kubernetes_pod_annotation_prometheus_io_port"]
    target_label = "__address__"
  }
  rule {
    action = "replace"
    replacement = "$1"
    separator = "/"
    source_labels = ["__meta_kubernetes_namespace", "__meta_kubernetes_pod_label_name"]
    target_label = "job"
  }
  rule {
    action = "replace"
    source_labels = ["__meta_kubernetes_namespace"]
    target_label = "namespace"
  }
  rule {
    action = "replace"
    source_labels = ["__meta_kubernetes_pod_name"]
    target_label = "pod"
  }
  rule {
    action = "replace"
    source_labels = ["__meta_kubernetes_pod_container_name"]
    target_label = "container"
  }
  rule {
    action = "replace"
    separator = ":"
    source_labels = ["__meta_kubernetes_pod_name", "__meta_kubernetes_pod_container_name", "__meta_kubernetes_pod_container_port_name"]
    target_label = "instance"
  }
  targets = concat(discovery.kubernetes.pods.targets)
}
`, abr)

	res.MetricsExport = "discovery.relabel.kubernetes_pods.targets"

	return res, nil
}

func buildAnnotationsBasedRule(lbls []string) string {
	if len(lbls) == 0 {
		return ""
	}

	annotationsRuleTemplate := `
  rule {
    // Try to identify a service name to eventually form the job label. We'll
    // prefer the first of the below labels, in descending order.
    source_labels = [
      %s,
    ]
    target_label = "__autodiscover__"

    // Our in-memory string will be something like A;B;C;D;E;F, where any of the
    // letters could be replaced with a label value or be empty if the label
    // value did not exist.
    //
    // We want to match for the very first sequence of non-semicolon characters
    // which is either prefaced by zero or more semicolons, and is followed by
    // zero or more semicolons before the rest of the string.
    //
    // This is a very annoying way of being able to do conditionals, and
    // ideally we can use River expressions in the future to make this much
    // less bizarre.
    regex = ";*([^;]+);*.*"
  }
  rule {
    source_labels = ["__autodiscover__"]
	regex = ""
	action = "drop"
  }
`
	for i := range lbls {
		lbls[i] = fmt.Sprintf("%q", lbls[i])
	}
	return fmt.Sprintf(annotationsRuleTemplate, strings.Join(lbls, ","))
}
