// Copyright 2022 The Prometheus Authors
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

package collectors

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"google.golang.org/api/monitoring/v3"
)

type CollectedMetric struct {
	metric          *ConstMetric
	lastCollectedAt time.Time
}

// DeltaCounterStore defines a set of functions which must be implemented in order to be used as a DeltaCounterStore
// which accumulates DELTA Counter metrics over time
type DeltaCounterStore interface {

	// Increment will use the incoming metricDescriptor and currentValue to either create a new entry or add the incoming
	// value to an existing entry in the underlying store
	Increment(metricDescriptor *monitoring.MetricDescriptor, currentValue *ConstMetric)

	// ListMetrics will return all known entries in the store for a metricDescriptorName
	ListMetrics(metricDescriptorName string) map[string][]*CollectedMetric
}

type metricEntry struct {
	collected map[uint64]*CollectedMetric
	mutex     *sync.RWMutex
}

type inMemoryDeltaCounterStore struct {
	store  *sync.Map
	ttl    time.Duration
	logger log.Logger
}

// NewInMemoryDeltaCounterStore returns an implementation of DeltaCounterStore which is persisted in-memory
func NewInMemoryDeltaCounterStore(logger log.Logger, ttl time.Duration) DeltaCounterStore {
	return &inMemoryDeltaCounterStore{
		store:  &sync.Map{},
		logger: logger,
		ttl:    ttl,
	}
}

func (s *inMemoryDeltaCounterStore) Increment(metricDescriptor *monitoring.MetricDescriptor, currentValue *ConstMetric) {
	if currentValue == nil {
		return
	}

	tmp, _ := s.store.LoadOrStore(metricDescriptor.Name, &metricEntry{
		collected: map[uint64]*CollectedMetric{},
		mutex:     &sync.RWMutex{},
	})
	entry := tmp.(*metricEntry)

	key := toCounterKey(currentValue)

	entry.mutex.Lock()
	defer entry.mutex.Unlock()
	existing := entry.collected[key]

	if existing == nil {
		level.Debug(s.logger).Log("msg", "Tracking new counter", "fqName", currentValue.fqName, "key", key, "current_value", currentValue.value, "incoming_time", currentValue.reportTime)
		entry.collected[key] = &CollectedMetric{currentValue, time.Now()}
		return
	}

	if existing.metric.reportTime.Before(currentValue.reportTime) {
		level.Debug(s.logger).Log("msg", "Incrementing existing counter", "fqName", currentValue.fqName, "key", key, "current_value", existing.metric.value, "adding", currentValue.value, "last_reported_time", entry.collected[key].metric.reportTime, "incoming_time", currentValue.reportTime)
		currentValue.value = currentValue.value + existing.metric.value
		existing.metric = currentValue
		existing.lastCollectedAt = time.Now()
		return
	}

	level.Debug(s.logger).Log("msg", "Ignoring old sample for counter", "fqName", currentValue.fqName, "key", key, "last_reported_time", existing.metric.reportTime, "incoming_time", currentValue.reportTime)
}

func toCounterKey(c *ConstMetric) uint64 {
	labels := make(map[string]string)
	keysCopy := append([]string{}, c.labelKeys...)
	for i := range c.labelKeys {
		labels[c.labelKeys[i]] = c.labelValues[i]
	}
	sort.Strings(keysCopy)

	var keyParts []string
	for _, k := range keysCopy {
		keyParts = append(keyParts, fmt.Sprintf("%s:%s", k, labels[k]))
	}
	hashText := fmt.Sprintf("%s|%s", c.fqName, strings.Join(keyParts, "|"))
	h := hashNew()
	h = hashAdd(h, hashText)

	return h
}

func (s *inMemoryDeltaCounterStore) ListMetrics(metricDescriptorName string) map[string][]*CollectedMetric {
	output := map[string][]*CollectedMetric{}
	now := time.Now()
	ttlWindowStart := now.Add(-s.ttl)

	tmp, exists := s.store.Load(metricDescriptorName)
	if !exists {
		return output
	}
	entry := tmp.(*metricEntry)

	entry.mutex.Lock()
	defer entry.mutex.Unlock()
	for key, collected := range entry.collected {
		//Scan and remove metrics which are outside the TTL
		if ttlWindowStart.After(collected.lastCollectedAt) {
			level.Debug(s.logger).Log("msg", "Deleting counter entry outside of TTL", "key", key, "fqName", collected.metric.fqName)
			delete(entry.collected, key)
			continue
		}

		metrics, exists := output[collected.metric.fqName]
		if !exists {
			metrics = make([]*CollectedMetric, 0)
		}
		metricCopy := *collected.metric
		outputEntry := CollectedMetric{
			metric:          &metricCopy,
			lastCollectedAt: collected.lastCollectedAt,
		}
		output[collected.metric.fqName] = append(metrics, &outputEntry)
	}

	return output
}
