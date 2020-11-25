// Copyright 2013 The Prometheus Authors
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

package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricType int

const (
	CounterMetricType MetricType = iota
	GaugeMetricType
	SummaryMetricType
	HistogramMetricType
)

type NameHash uint64

type ValueHash uint64

type LabelHash struct {
	// This is a hash over the label names
	Names NameHash
	// This is a hash over the label names + label values
	Values ValueHash
}

type MetricHolder interface{}

type VectorHolder interface {
	Delete(label prometheus.Labels) bool
}

type Vector struct {
	Holder   VectorHolder
	RefCount uint64
}

type Metric struct {
	MetricType MetricType
	// Vectors key is the hash of the label names
	Vectors map[NameHash]*Vector
	// Metrics key is a hash of the label names + label values
	Metrics map[ValueHash]*RegisteredMetric
}

type RegisteredMetric struct {
	LastRegisteredAt time.Time
	Labels           prometheus.Labels
	TTL              time.Duration
	Metric           MetricHolder
	VecKey           NameHash
}
