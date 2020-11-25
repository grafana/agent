// Copyright 2020 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mapper

import "time"

type mapperConfigDefaults struct {
	ObserverType        ObserverType      `yaml:"observer_type"`
	TimerType           ObserverType      `yaml:"timer_type,omitempty"` // DEPRECATED - field only present to preserve backwards compatibility in configs. Always empty
	Buckets             []float64         `yaml:"buckets"`
	Quantiles           []metricObjective `yaml:"quantiles"`
	MatchType           MatchType         `yaml:"match_type"`
	GlobDisableOrdering bool              `yaml:"glob_disable_ordering"`
	Ttl                 time.Duration     `yaml:"ttl"`
}

// UnmarshalYAML is a custom unmarshal function to allow use of deprecated config keys
// observer_type will override timer_type
func (d *mapperConfigDefaults) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type mapperConfigDefaultsAlias mapperConfigDefaults
	var tmp mapperConfigDefaultsAlias
	if err := unmarshal(&tmp); err != nil {
		return err
	}

	// Copy defaults
	d.ObserverType = tmp.ObserverType
	d.Buckets = tmp.Buckets
	d.Quantiles = tmp.Quantiles
	d.MatchType = tmp.MatchType
	d.GlobDisableOrdering = tmp.GlobDisableOrdering
	d.Ttl = tmp.Ttl

	// Use deprecated TimerType if necessary
	if tmp.ObserverType == "" {
		d.ObserverType = tmp.TimerType
	}

	return nil
}
