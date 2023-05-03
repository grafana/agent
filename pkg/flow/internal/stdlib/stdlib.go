// Package stdlib contains Flow-specific standard library functions exposed to
// River configs.
package stdlib

import (
	"encoding/json"

	"github.com/grafana/agent/component/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

// Identifiers holds a list of stdlib identifiers by name. All interface{}
// values are River-compatible values.
//
// Function identifiers are Go functions with exactly one non-error return
// value, with an optionally supported error return value as the second return
// value.
var Identifiers = map[string]interface{}{
	"discovery_target_decode": func(in string) (interface{}, error) {
		var targetGroups []*targetgroup.Group
		if err := json.Unmarshal([]byte(in), &targetGroups); err != nil {
			return nil, err
		}

		var res []discovery.Target

		for _, group := range targetGroups {
			for _, target := range group.Targets {

				// Create the output target from group and target labels. Target labels
				// should override group labels.
				outputTarget := make(discovery.Target, len(group.Labels)+len(target))
				for k, v := range group.Labels {
					outputTarget[string(k)] = string(v)
				}
				for k, v := range target {
					outputTarget[string(k)] = string(v)
				}

				res = append(res, outputTarget)
			}
		}

		return res, nil
	},
}
