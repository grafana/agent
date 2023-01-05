package rules

import (
	"bytes"

	"github.com/prometheus/prometheus/model/rulefmt"
	"gopkg.in/yaml.v3" // Used for prometheus rulefmt compatibility instead of gopkg.in/yaml.v2
)

type ruleGroupDiffKind string

const (
	ruleGroupDiffKindAdd    ruleGroupDiffKind = "add"
	ruleGroupDiffKindRemove ruleGroupDiffKind = "remove"
	ruleGroupDiffKindUpdate ruleGroupDiffKind = "update"
)

type ruleGroupDiff struct {
	Kind    ruleGroupDiffKind
	Actual  rulefmt.RuleGroup
	Desired rulefmt.RuleGroup
}

func diffRuleState(desired map[string][]rulefmt.RuleGroup, actual map[string][]rulefmt.RuleGroup) map[string][]ruleGroupDiff {
	seen := map[string]bool{}

	diff := make(map[string][]ruleGroupDiff)

	for namespace, desiredRuleGroups := range desired {
		seen[namespace] = true

		actualRuleGroups := actual[namespace]
		subDiff := diffRuleNamespaceState(desiredRuleGroups, actualRuleGroups)

		if len(subDiff) == 0 {
			continue
		}

		diff[namespace] = subDiff
	}

	for namespace, actualRuleGroups := range actual {
		if seen[namespace] {
			continue
		}

		subDiff := diffRuleNamespaceState(nil, actualRuleGroups)

		diff[namespace] = subDiff
	}

	return diff
}

func diffRuleNamespaceState(desired []rulefmt.RuleGroup, actual []rulefmt.RuleGroup) []ruleGroupDiff {
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

	return diff
}

func equalRuleGroups(a, b rulefmt.RuleGroup) bool {
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
