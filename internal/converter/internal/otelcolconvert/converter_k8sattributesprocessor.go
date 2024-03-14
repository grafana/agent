package otelcolconvert

import (
	"fmt"

	"github.com/grafana/agent/internal/component/otelcol"
	"github.com/grafana/agent/internal/component/otelcol/processor/k8sattributes"
	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/common"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor"
	"go.opentelemetry.io/collector/component"
)

func init() {
	converters = append(converters, k8sAttributesProcessorConverter{})
}

type k8sAttributesProcessorConverter struct{}

func (k8sAttributesProcessorConverter) Factory() component.Factory {
	return k8sattributesprocessor.NewFactory()
}

func (k8sAttributesProcessorConverter) InputComponentName() string {
	return "otelcol.processor.k8sattributes"
}

func (k8sAttributesProcessorConverter) ConvertAndAppend(state *state, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	args := toK8SAttributesProcessor(state, id, cfg.(*k8sattributesprocessor.Config))
	block := common.NewBlockWithOverride([]string{"otelcol", "processor", "k8sattributes"}, label, args)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", stringifyInstanceID(id), stringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func toK8SAttributesProcessor(state *state, id component.InstanceID, cfg *k8sattributesprocessor.Config) *k8sattributes.Arguments {
	var (
		nextMetrics = state.Next(id, component.DataTypeMetrics)
		nextLogs    = state.Next(id, component.DataTypeLogs)
		nextTraces  = state.Next(id, component.DataTypeTraces)
	)

	return &k8sattributes.Arguments{
		AuthType:    string(cfg.AuthType),
		Passthrough: cfg.Passthrough,
		ExtractConfig: k8sattributes.ExtractConfig{
			Metadata:    cfg.Extract.Metadata,
			Annotations: toFilterExtract(cfg.Extract.Annotations),
			Labels:      toFilterExtract(cfg.Extract.Labels),
		},
		Filter: k8sattributes.FilterConfig{
			Node:      cfg.Filter.Node,
			Namespace: cfg.Filter.Namespace,
			Fields:    toFilterFields(cfg.Filter.Fields),
			Labels:    toFilterFields(cfg.Filter.Labels),
		},
		PodAssociations: toPodAssociations(cfg.Association),
		Exclude:         toExclude(cfg.Exclude),

		Output: &otelcol.ConsumerArguments{
			Metrics: toTokenizedConsumers(nextMetrics),
			Logs:    toTokenizedConsumers(nextLogs),
			Traces:  toTokenizedConsumers(nextTraces),
		},
	}
}

func toExclude(cfg k8sattributesprocessor.ExcludeConfig) k8sattributes.ExcludeConfig {
	res := k8sattributes.ExcludeConfig{
		Pods: []k8sattributes.ExcludePodConfig{},
	}

	for _, c := range cfg.Pods {
		res.Pods = append(res.Pods, k8sattributes.ExcludePodConfig{
			Name: c.Name,
		})
	}

	return res
}

func toPodAssociations(cfg []k8sattributesprocessor.PodAssociationConfig) []k8sattributes.PodAssociation {
	if len(cfg) == 0 {
		return nil
	}

	res := make([]k8sattributes.PodAssociation, 0, len(cfg))

	for i, c := range cfg {
		res = append(res, k8sattributes.PodAssociation{
			Sources: []k8sattributes.PodAssociationSource{},
		})

		for _, c2 := range c.Sources {
			res[i].Sources = append(res[i].Sources, k8sattributes.PodAssociationSource{
				From: c2.From,
				Name: c2.Name,
			})
		}
	}

	return res
}
func toFilterExtract(cfg []k8sattributesprocessor.FieldExtractConfig) []k8sattributes.FieldExtractConfig {
	if len(cfg) == 0 {
		return nil
	}

	res := make([]k8sattributes.FieldExtractConfig, 0, len(cfg))

	for _, c := range cfg {
		res = append(res, k8sattributes.FieldExtractConfig{
			TagName:  c.TagName,
			Key:      c.Key,
			KeyRegex: c.KeyRegex,
			Regex:    c.Regex,
			From:     c.From,
		})
	}

	return res
}

func toFilterFields(cfg []k8sattributesprocessor.FieldFilterConfig) []k8sattributes.FieldFilterConfig {
	if len(cfg) == 0 {
		return nil
	}

	res := make([]k8sattributes.FieldFilterConfig, 0, len(cfg))

	for _, c := range cfg {
		res = append(res, k8sattributes.FieldFilterConfig{
			Key:   c.Key,
			Value: c.Value,
			Op:    c.Op,
		})
	}

	return res
}
