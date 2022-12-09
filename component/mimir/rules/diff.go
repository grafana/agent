package rules

import (
	"bytes"

	mimirClient "github.com/grafana/agent/pkg/mimir/client"

	"gopkg.in/yaml.v3"
)

type ruleGroupDiffKind string

const (
	ruleGroupDiffKindAdd    ruleGroupDiffKind = "add"
	ruleGroupDiffKindRemove ruleGroupDiffKind = "remove"
	ruleGroupDiffKindUpdate ruleGroupDiffKind = "update"
)

type ruleGroupDiff struct {
	Kind    ruleGroupDiffKind
	Actual  mimirClient.RuleGroup
	Desired mimirClient.RuleGroup
}

func diffRuleState(desired map[string][]mimirClient.RuleGroup, actual map[string][]mimirClient.RuleGroup) (map[string][]ruleGroupDiff, error) {
	seen := map[string]bool{}

	diff := make(map[string][]ruleGroupDiff)

	for namespace, desiredRuleGroups := range desired {
		seen[namespace] = true

		actualRuleGroups := actual[namespace]
		subDiff, err := diffRuleNamespaceState(desiredRuleGroups, actualRuleGroups)
		if err != nil {
			return nil, err
		}

		if len(subDiff) == 0 {
			continue
		}

		diff[namespace] = subDiff
	}

	for namespace, actualRuleGroups := range actual {
		if seen[namespace] {
			continue
		}

		subDiff, err := diffRuleNamespaceState(nil, actualRuleGroups)
		if err != nil {
			return nil, err
		}

		diff[namespace] = subDiff
	}

	return diff, nil
}

func diffRuleNamespaceState(desired []mimirClient.RuleGroup, actual []mimirClient.RuleGroup) ([]ruleGroupDiff, error) {
	var diff []ruleGroupDiff

	seenGroups := map[string]bool{}

desiredGroups:
	for _, desiredRuleGroup := range desired {
		seenGroups[desiredRuleGroup.Name] = true

		for _, actualRuleGroup := range actual {
			if desiredRuleGroup.Name == actualRuleGroup.Name {
				if equalRuleGroups(desiredRuleGroup, actualRuleGroup) {
					continue desiredGroups
				}

				diff = append(diff, ruleGroupDiff{
					Kind:    ruleGroupDiffKindUpdate,
					Actual:  actualRuleGroup,
					Desired: desiredRuleGroup,
				})
				continue desiredGroups
			}
		}

		diff = append(diff, ruleGroupDiff{
			Kind:    ruleGroupDiffKindAdd,
			Desired: desiredRuleGroup,
		})
	}

	for _, actualRuleGroup := range actual {
		if seenGroups[actualRuleGroup.Name] {
			continue
		}

		diff = append(diff, ruleGroupDiff{
			Kind:   ruleGroupDiffKindRemove,
			Actual: actualRuleGroup,
		})
	}

	return diff, nil
}

func equalRuleGroups(a, b mimirClient.RuleGroup) bool {
	aBuf, err := yaml.Marshal(a)
	if err != nil {
		return false
	}
	bBuf, err := yaml.Marshal(b)
	if err != nil {
		return false
	}

	return bytes.Equal(aBuf, bBuf)
}
