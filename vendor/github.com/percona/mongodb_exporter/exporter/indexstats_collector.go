// mongodb_exporter
// Copyright (C) 2017 Percona LLC
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
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type indexstatsCollector struct {
	ctx  context.Context
	base *baseCollector

	discoveringMode         bool
	overrideDescendingIndex bool
	topologyInfo            labelsGetter

	collections []string
}

// newIndexStatsCollector creates a collector for statistics on index usage.
func newIndexStatsCollector(ctx context.Context, client *mongo.Client, logger *logrus.Logger, discovery, overrideDescendingIndex bool, topology labelsGetter, collections []string) *indexstatsCollector {
	return &indexstatsCollector{
		ctx:  ctx,
		base: newBaseCollector(client, logger),

		discoveringMode:         discovery,
		topologyInfo:            topology,
		overrideDescendingIndex: overrideDescendingIndex,

		collections: collections,
	}
}

func (d *indexstatsCollector) Describe(ch chan<- *prometheus.Desc) {
	d.base.Describe(d.ctx, ch, d.collect)
}

func (d *indexstatsCollector) Collect(ch chan<- prometheus.Metric) {
	d.base.Collect(ch)
}

func (d *indexstatsCollector) collect(ch chan<- prometheus.Metric) {
	collections := d.collections

	logger := d.base.logger
	client := d.base.client

	if d.discoveringMode {
		namespaces, err := listAllCollections(d.ctx, client, d.collections, systemDBs)
		if err != nil {
			logger.Errorf("cannot auto discover databases and collections")

			return
		}

		collections = fromMapToSlice(namespaces)
	}

	for _, dbCollection := range collections {
		parts := strings.Split(dbCollection, ".")
		if len(parts) < 2 { //nolint:gomnd
			continue
		}

		database := parts[0]
		collection := strings.Join(parts[1:], ".")

		aggregation := bson.D{
			{Key: "$indexStats", Value: bson.M{}},
		}

		cursor, err := client.Database(database).Collection(collection).Aggregate(d.ctx, mongo.Pipeline{aggregation})
		if err != nil {
			logger.Errorf("cannot get $indexStats cursor for collection %s.%s: %s", database, collection, err)

			continue
		}

		var stats []bson.M
		if err = cursor.All(d.ctx, &stats); err != nil {
			logger.Errorf("cannot get $indexStats for collection %s.%s: %s", database, collection, err)

			continue
		}

		d.base.logger.Debugf("indexStats for %s.%s", database, collection)

		debugResult(d.base.logger, stats)

		for _, metric := range stats {
			indexName := fmt.Sprintf("%s", metric["name"])
			// Override the label name
			if d.overrideDescendingIndex {
				indexName = strings.ReplaceAll(fmt.Sprintf("%s", metric["name"]), "-1", "DESC")
			}

			// prefix and labels are needed to avoid duplicated metric names since the metrics are the
			// same, for different collections.
			prefix := "indexstats"
			labels := d.topologyInfo.baseLabels()
			labels["database"] = database
			labels["collection"] = collection
			labels["key_name"] = indexName

			metrics := sanitizeMetrics(metric)
			for _, metric := range makeMetrics(prefix, metrics, labels, false) {
				ch <- metric
			}
		}
	}
}

// According to specs, we should expose only this 2 metrics. 'building' might not exist.
func sanitizeMetrics(m bson.M) bson.M {
	ops := float64(0)

	if val := walkTo(m, []string{"accesses", "ops"}); val != nil {
		if f, err := asFloat64(val); err == nil {
			ops = *f
		}
	}

	filteredMetrics := bson.M{
		"accesses": bson.M{
			"ops": ops,
		},
	}

	if val := walkTo(m, []string{"building"}); val != nil {
		if f, err := asFloat64(val); err == nil {
			filteredMetrics["building"] = *f
		}
	}

	return filteredMetrics
}

var _ prometheus.Collector = (*indexstatsCollector)(nil)
