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

type CollectedHistogram struct {
	histogram       *HistogramMetric
	lastCollectedAt time.Time
}

// DeltaDistributionStore defines a set of functions which must be implemented in order to be used as a DeltaDistributionStore
// which accumulates DELTA histogram metrics over time
type DeltaDistributionStore interface {

	// Increment will use the incoming metricDescriptor and currentValue to either create a new entry or add the incoming
	// value to an existing entry in the underlying store
	Increment(metricDescriptor *monitoring.MetricDescriptor, currentValue *HistogramMetric)

	// ListMetrics will return all known entries in the store for a metricDescriptorName
	ListMetrics(metricDescriptorName string) map[string][]*CollectedHistogram
}

type histogramEntry struct {
	collected map[uint64]*CollectedHistogram
	mutex     *sync.RWMutex
}

type inMemoryDeltaDistributionStore struct {
	store  *sync.Map
	ttl    time.Duration
	logger log.Logger
}

// NewInMemoryDeltaDistributionStore returns an implementation of DeltaDistributionStore which is persisted in-memory
func NewInMemoryDeltaDistributionStore(logger log.Logger, ttl time.Duration) DeltaDistributionStore {
	return &inMemoryDeltaDistributionStore{
		store:  &sync.Map{},
		logger: logger,
		ttl:    ttl,
	}
}

func (s *inMemoryDeltaDistributionStore) Increment(metricDescriptor *monitoring.MetricDescriptor, currentValue *HistogramMetric) {
	if currentValue == nil {
		return
	}

	tmp, _ := s.store.LoadOrStore(metricDescriptor.Name, &histogramEntry{
		collected: map[uint64]*CollectedHistogram{},
		mutex:     &sync.RWMutex{},
	})
	entry := tmp.(*histogramEntry)

	key := toHistogramKey(currentValue)

	entry.mutex.Lock()
	defer entry.mutex.Unlock()
	existing := entry.collected[key]

	if existing == nil {
		level.Debug(s.logger).Log("msg", "Tracking new histogram", "fqName", currentValue.fqName, "key", key, "incoming_time", currentValue.reportTime)
		entry.collected[key] = &CollectedHistogram{histogram: currentValue, lastCollectedAt: time.Now()}
		return
	}

	if existing.histogram.reportTime.Before(currentValue.reportTime) {
		level.Debug(s.logger).Log("msg", "Incrementing existing histogram", "fqName", currentValue.fqName, "key", key, "last_reported_time", existing.histogram.reportTime, "incoming_time", currentValue.reportTime)
		existing.histogram = mergeHistograms(existing.histogram, currentValue)
		existing.lastCollectedAt = time.Now()
		return
	}

	level.Debug(s.logger).Log("msg", "Ignoring old sample for histogram", "fqName", currentValue.fqName, "key", key, "last_reported_time", existing.histogram.reportTime, "incoming_time", currentValue.reportTime)
}

func toHistogramKey(hist *HistogramMetric) uint64 {
	labels := make(map[string]string)
	keysCopy := append([]string{}, hist.labelKeys...)
	for i := range hist.labelKeys {
		labels[hist.labelKeys[i]] = hist.labelValues[i]
	}
	sort.Strings(keysCopy)

	var keyParts []string
	for _, k := range keysCopy {
		keyParts = append(keyParts, fmt.Sprintf("%s:%s", k, labels[k]))
	}
	hashText := fmt.Sprintf("%s|%s", hist.fqName, strings.Join(keyParts, "|"))
	h := hashNew()
	h = hashAdd(h, hashText)

	return h
}

func mergeHistograms(existing *HistogramMetric, current *HistogramMetric) *HistogramMetric {
	for key, value := range existing.buckets {
		current.buckets[key] += value
	}

	// Calculate a new mean and overall count
	mean := existing.dist.Mean
	mean += current.dist.Mean
	mean /= 2

	var count uint64
	for _, v := range current.buckets {
		count += v
	}

	current.dist.Mean = mean
	current.dist.Count = int64(count)

	return current
}

func (s *inMemoryDeltaDistributionStore) ListMetrics(metricDescriptorName string) map[string][]*CollectedHistogram {
	output := map[string][]*CollectedHistogram{}
	now := time.Now()
	ttlWindowStart := now.Add(-s.ttl)

	tmp, exists := s.store.Load(metricDescriptorName)
	if !exists {
		return output
	}
	entry := tmp.(*histogramEntry)

	entry.mutex.Lock()
	defer entry.mutex.Unlock()
	for key, collected := range entry.collected {
		//Scan and remove metrics which are outside the TTL
		if ttlWindowStart.After(collected.lastCollectedAt) {
			level.Debug(s.logger).Log("msg", "Deleting histogram entry outside of TTL", "key", key, "fqName", collected.histogram.fqName)
			delete(entry.collected, key)
			continue
		}

		metrics, exists := output[collected.histogram.fqName]
		if !exists {
			metrics = make([]*CollectedHistogram, 0)
		}
		histCopy := *collected.histogram
		outputEntry := CollectedHistogram{
			histogram:       &histCopy,
			lastCollectedAt: collected.lastCollectedAt,
		}
		output[collected.histogram.fqName] = append(metrics, &outputEntry)
	}

	return output
}
