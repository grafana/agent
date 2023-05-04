// mongodb_exporter
// Copyright (C) 2022 Percona LLC
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package exporter

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type baseCollector struct {
	client *mongo.Client
	logger *logrus.Logger

	lock         sync.Mutex
	metricsCache []prometheus.Metric
}

// newBaseCollector creates a skeletal collector, which is used to create other collectors.
func newBaseCollector(client *mongo.Client, logger *logrus.Logger) *baseCollector {
	return &baseCollector{
		client: client,
		logger: logger,
	}
}

func (d *baseCollector) Describe(ctx context.Context, ch chan<- *prometheus.Desc, collect func(mCh chan<- prometheus.Metric)) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	d.lock.Lock()
	defer d.lock.Unlock()

	d.metricsCache = make([]prometheus.Metric, 0, defaultCacheSize)

	// This is a copy/paste of prometheus.DescribeByCollect(d, ch) with the aggreated functionality
	// to populate the metrics cache. Since on each scrape Prometheus will call Describe and inmediatelly
	// after it will call Collect, it is safe to populate the cache here.
	metrics := make(chan prometheus.Metric)
	go func() {
		collect(metrics)
		close(metrics)
	}()

	for m := range metrics {
		d.metricsCache = append(d.metricsCache, m) // populate the cache
		ch <- m.Desc()
	}
}

func (d *baseCollector) Collect(ch chan<- prometheus.Metric) {
	d.lock.Lock()
	defer d.lock.Unlock()

	for _, metric := range d.metricsCache {
		ch <- metric
	}
}
