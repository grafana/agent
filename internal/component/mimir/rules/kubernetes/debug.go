package rules

import "fmt"

type DebugInfo struct {
	Error               string                   `river:"error,attr,optional"`
	PrometheusRules     []DebugK8sPrometheusRule `river:"prometheus_rule,block,optional"`
	MimirRuleNamespaces []DebugMimirNamespace    `river:"mimir_rule_namespace,block,optional"`
}

type DebugK8sPrometheusRule struct {
	Namespace     string `river:"namespace,attr"`
	Name          string `river:"name,attr"`
	UID           string `river:"uid,attr"`
	NumRuleGroups int    `river:"num_rule_groups,attr"`
}

type DebugMimirNamespace struct {
	Name          string `river:"name,attr"`
	NumRuleGroups int    `river:"num_rule_groups,attr"`
}

func (c *Component) DebugInfo() interface{} {
	var output DebugInfo

	currentState := c.eventProcessor.getMimirState()
	for namespace := range currentState {
		if !isManagedMimirNamespace(c.args.MimirNameSpacePrefix, namespace) {
			continue
		}

		output.MimirRuleNamespaces = append(output.MimirRuleNamespaces, DebugMimirNamespace{
			Name:          namespace,
			NumRuleGroups: len(currentState[namespace]),
		})
	}

	// This should load from the informer cache, so it shouldn't fail under normal circumstances.
	rulesByNamespace, err := c.eventProcessor.getKubernetesState()
	if err != nil {
		return DebugInfo{Error: fmt.Sprintf("failed to list rules: %v", err)}
	}

	for namespace, rules := range rulesByNamespace {
		for _, rule := range rules {
			output.PrometheusRules = append(output.PrometheusRules, DebugK8sPrometheusRule{
				Namespace:     namespace,
				Name:          rule.Name,
				UID:           string(rule.UID),
				NumRuleGroups: len(rule.Spec.Groups),
			})
		}
	}

	return output
}
