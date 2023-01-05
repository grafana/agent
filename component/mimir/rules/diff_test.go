package rules

import (
	"fmt"
	"testing"

	mimirClient "github.com/grafana/agent/pkg/mimir/client"
	"github.com/prometheus/prometheus/model/rulefmt"
	"github.com/stretchr/testify/require"
)

func parseRuleGroups(t *testing.T, buf []byte) []mimirClient.RuleGroup {
	t.Helper()

	groups, errs := rulefmt.Parse(buf)
	require.Empty(t, errs)

	var result []mimirClient.RuleGroup
	for _, g := range groups.Groups {
		result = append(result, mimirClient.RuleGroup{RuleGroup: g})
	}
	return result
}

func TestDiffRuleState(t *testing.T) {
	ruleGroupsA := parseRuleGroups(t, []byte(`
groups:
- name: rule-group-a
  interval: 1m
  rules:
  - record: rule_a
    expr: 1
`))

	ruleGroupsB := parseRuleGroups(t, []byte(`
groups:
- name: rule-group-b
  interval: 1m
  rules:
  - record: rule_b
    expr: 2
`))

	ruleGroupsAModified := parseRuleGroups(t, []byte(`
groups:
- name: rule-group-a
  interval: 1m
  rules:
  - record: rule_a
    expr: 3
`))

	managedNamespace := "agent/namespace/name/12345678-1234-1234-1234-123456789012"
	unmanagedNamespace := "integrations/kubernetes"

	_ = ruleGroupsB
	_ = ruleGroupsA
	_ = ruleGroupsAModified
	_ = unmanagedNamespace

	type testCase struct {
		name     string
		desired  map[string][]mimirClient.RuleGroup
		actual   map[string][]mimirClient.RuleGroup
		expected map[string][]RuleGroupDiff
	}

	testCases := []testCase{
		{
			name:     "empty sets",
			desired:  map[string][]mimirClient.RuleGroup{},
			actual:   map[string][]mimirClient.RuleGroup{},
			expected: map[string][]RuleGroupDiff{},
		},
		{
			name: "add rule group",
			desired: map[string][]mimirClient.RuleGroup{
				managedNamespace: ruleGroupsA,
			},
			actual: map[string][]mimirClient.RuleGroup{},
			expected: map[string][]RuleGroupDiff{
				managedNamespace: {
					{
						Kind:    RuleGroupDiffKindAdd,
						Desired: ruleGroupsA[0],
					},
				},
			},
		},
		{
			name:    "remove rule group",
			desired: map[string][]mimirClient.RuleGroup{},
			actual: map[string][]mimirClient.RuleGroup{
				managedNamespace: ruleGroupsA,
			},
			expected: map[string][]RuleGroupDiff{
				managedNamespace: {
					{
						Kind:   RuleGroupDiffKindRemove,
						Actual: ruleGroupsA[0],
					},
				},
			},
		},
		{
			name: "update rule group",
			desired: map[string][]mimirClient.RuleGroup{
				managedNamespace: ruleGroupsA,
			},
			actual: map[string][]mimirClient.RuleGroup{
				managedNamespace: ruleGroupsAModified,
			},
			expected: map[string][]RuleGroupDiff{
				managedNamespace: {
					{
						Kind:    RuleGroupDiffKindUpdate,
						Desired: ruleGroupsA[0],
						Actual:  ruleGroupsAModified[0],
					},
				},
			},
		},
		{
			name: "unchanged rule groups",
			desired: map[string][]mimirClient.RuleGroup{
				managedNamespace: ruleGroupsA,
			},
			actual: map[string][]mimirClient.RuleGroup{
				managedNamespace: ruleGroupsA,
			},
			expected: map[string][]RuleGroupDiff{},
		},
		{
			name: "unmanaged namespaces",
			desired: map[string][]mimirClient.RuleGroup{
				managedNamespace: ruleGroupsA,
			},
			actual: map[string][]mimirClient.RuleGroup{
				managedNamespace:   ruleGroupsA,
				unmanagedNamespace: ruleGroupsB,
			},
			expected: map[string][]RuleGroupDiff{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := diffRuleState(tc.desired, tc.actual)
			require.NoError(t, err)
			requireEqualRuleDiffs(t, tc.expected, actual)
		})
	}
}

func requireEqualRuleDiffs(t *testing.T, expected, actual map[string][]RuleGroupDiff) {
	require.Equal(t, len(expected), len(actual))

	var summarizeDiff = func(diff RuleGroupDiff) string {
		switch diff.Kind {
		case RuleGroupDiffKindAdd:
			return fmt.Sprintf("add: %s", diff.Desired.Name)
		case RuleGroupDiffKindRemove:
			return fmt.Sprintf("remove: %s", diff.Actual.Name)
		case RuleGroupDiffKindUpdate:
			return fmt.Sprintf("update: %s", diff.Desired.Name)
		}
		panic("unreachable")
	}

	for namespace, expectedDiffs := range expected {
		actualDiffs, ok := actual[namespace]
		require.True(t, ok)

		require.Equal(t, len(expectedDiffs), len(actualDiffs))

		for i, expectedDiff := range expectedDiffs {
			actualDiff := actualDiffs[i]

			if expectedDiff.Kind != actualDiff.Kind ||
				!equalRuleGroups(expectedDiff.Desired, actualDiff.Desired) ||
				!equalRuleGroups(expectedDiff.Actual, actualDiff.Actual) {

				t.Logf("expected diff: %s", summarizeDiff(expectedDiff))
				t.Logf("actual diff: %s", summarizeDiff(actualDiff))
				t.Fail()
			}

		}
	}
}
