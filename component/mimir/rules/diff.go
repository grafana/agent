package rules

import (
	"bytes"

	mimirClient "github.com/grafana/agent/pkg/mimir/client"
	"github.com/prometheus/prometheus/model/rulefmt"

	"gopkg.in/yaml.v3"
)

type RuleGroupDiffKind string

const (
	RuleGroupDiffKindAdd    RuleGroupDiffKind = "add"
	RuleGroupDiffKindRemove RuleGroupDiffKind = "remove"
	RuleGroupDiffKindUpdate RuleGroupDiffKind = "update"
)

type RuleGroupDiff struct {
	Kind    RuleGroupDiffKind
	Actual  mimirClient.RuleGroup
	Desired mimirClient.RuleGroup
}

func diffRuleStates(desired []rulefmt.RuleGroup, actual []mimirClient.RuleGroup) ([]RuleGroupDiff, error) {
	var diff []RuleGroupDiff

	seenGroups := map[string]bool{}

desiredGroups:
	for _, desiredRuleGroup := range desired {
		mimirRuleGroup := mimirClient.RuleGroup{
			RuleGroup: desiredRuleGroup,
			// TODO: allow setting the remote write configs?
			// RWConfigs: ,
		}

		seenGroups[desiredRuleGroup.Name] = true

		for _, actualRuleGroup := range actual {
			if desiredRuleGroup.Name == actualRuleGroup.Name {
				if equalRuleGroups(desiredRuleGroup, actualRuleGroup.RuleGroup) {
					continue desiredGroups
				}

				// TODO: check if the rules are the same
				diff = append(diff, RuleGroupDiff{
					Kind:    RuleGroupDiffKindUpdate,
					Actual:  actualRuleGroup,
					Desired: mimirRuleGroup,
				})
				continue desiredGroups
			}
		}

		diff = append(diff, RuleGroupDiff{
			Kind:    RuleGroupDiffKindAdd,
			Desired: mimirRuleGroup,
		})
	}

	for _, actualRuleGroup := range actual {
		if seenGroups[actualRuleGroup.Name] {
			continue
		}

		diff = append(diff, RuleGroupDiff{
			Kind:   RuleGroupDiffKindRemove,
			Actual: actualRuleGroup,
		})
	}

	return diff, nil
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
