package metricspb

import (
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

//go:generate protoc --go_out=. --go_opt=module=github.com/grafana/agent/pkg/metrics/v2/internal/metricspb --go-grpc_out=. --go-grpc_opt=module=github.com/grafana/agent/pkg/metrics/v2/internal/metricspb  ./metricspb.proto

// PrometheusGroups will convert gRPC targets into the Prometheus type.
func PrometheusGroups(in map[string]*TargetGroups) map[string][]*targetgroup.Group {
	out := make(map[string][]*targetgroup.Group, len(in))
	for k, groups := range in {
		set := make([]*targetgroup.Group, 0, len(groups.GetGroups()))

		for _, inGroup := range groups.GetGroups() {
			outTargets := make([]model.LabelSet, 0, len(inGroup.GetTargets()))
			for _, inTarget := range inGroup.GetTargets() {
				outTargets = append(outTargets, prometheusLabelset(inTarget))
			}

			outGroup := &targetgroup.Group{
				Targets: outTargets,
				Labels:  prometheusLabelset(inGroup.GetLabels()),
				Source:  inGroup.GetSource(),
			}
			set = append(set, outGroup)
		}

		out[k] = set
	}
	return out
}

func prometheusLabelset(in *LabelSet) model.LabelSet {
	inLabels := in.GetLabels()
	out := make(map[model.LabelName]model.LabelValue, len(inLabels))
	for k, v := range inLabels {
		out[model.LabelName(k)] = model.LabelValue(v)
	}
	return out
}

// ProtoGroups will convert Prometheus groups into the protobuf type.
func ProtoGroups(in map[string][]*targetgroup.Group) map[string]*TargetGroups {
	out := make(map[string]*TargetGroups, len(in))
	for k, groups := range in {
		set := make([]*TargetGroup, 0, len(groups))

		for _, inGroup := range groups {
			outTargets := make([]*LabelSet, 0, len(inGroup.Targets))
			for _, inTarget := range inGroup.Targets {
				outTargets = append(outTargets, protoLabelset(inTarget))
			}
			outGroup := &TargetGroup{
				Targets: outTargets,
				Labels:  protoLabelset(inGroup.Labels),
				Source:  inGroup.Source,
			}
			set = append(set, outGroup)
		}

		out[k] = &TargetGroups{Groups: set}
	}

	return out
}

func protoLabelset(in model.LabelSet) *LabelSet {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[string(k)] = string(v)
	}
	return &LabelSet{Labels: out}
}
