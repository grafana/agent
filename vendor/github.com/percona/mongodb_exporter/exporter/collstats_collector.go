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
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type collstatsCollector struct {
	ctx  context.Context
	base *baseCollector

	compatibleMode  bool
	discoveringMode bool
	topologyInfo    labelsGetter

	collections []string
}

// newCollectionStatsCollector creates a collector for statistics about collections.
func newCollectionStatsCollector(ctx context.Context, client *mongo.Client, logger *logrus.Logger, compatible, discovery bool, topology labelsGetter, collections []string) *collstatsCollector {
	return &collstatsCollector{
		ctx:  ctx,
		base: newBaseCollector(client, logger),

		compatibleMode:  compatible,
		discoveringMode: discovery,
		topologyInfo:    topology,

		collections: collections,
	}
}

func (d *collstatsCollector) Describe(ch chan<- *prometheus.Desc) {
	d.base.Describe(d.ctx, ch, d.collect)
}

func (d *collstatsCollector) Collect(ch chan<- prometheus.Metric) {
	d.base.Collect(ch)
}

func (d *collstatsCollector) collect(ch chan<- prometheus.Metric) {
	collections := d.collections

	client := d.base.client
	logger := d.base.logger

	if d.discoveringMode {
		namespaces, err := listAllCollections(d.ctx, client, d.collections, systemDBs)
		if err != nil {
			logger.Errorf("cannot auto discover databases and collections: %s", err.Error())

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
		collection := strings.Join(parts[1:], ".") // support collections having a .

		aggregation := bson.D{
			{
				Key: "$collStats", Value: bson.M{
					// TODO: PMM-9568 : Add support to handle histogram metrics
					"latencyStats": bson.M{"histograms": false},
					"storageStats": bson.M{"scale": 1},
				},
			},
		}
		project := bson.D{
			{
				Key: "$project", Value: bson.M{
					"storageStats.wiredTiger":   0,
					"storageStats.indexDetails": 0,
				},
			},
		}

		cursor, err := client.Database(database).Collection(collection).Aggregate(d.ctx, mongo.Pipeline{aggregation, project})
		if err != nil {
			logger.Errorf("cannot get $collstats cursor for collection %s.%s: %s", database, collection, err)

			continue
		}

		var stats []bson.M
		if err = cursor.All(d.ctx, &stats); err != nil {
			logger.Errorf("cannot get $collstats for collection %s.%s: %s", database, collection, err)

			continue
		}

		logger.Debugf("$collStats metrics for %s.%s", database, collection)
		debugResult(logger, stats)

		prefix := "collstats"
		labels := d.topologyInfo.baseLabels()
		labels["database"] = database
		labels["collection"] = collection

		for _, metrics := range stats {
			for _, metric := range makeMetrics(prefix, metrics, labels, d.compatibleMode) {
				ch <- metric
			}
		}
	}
}

func fromMapToSlice(databases map[string][]string) []string {
	var collections []string
	for db, cols := range databases {
		for _, value := range cols {
			collections = append(collections, db+"."+value)
		}
	}

	return collections
}

var _ prometheus.Collector = (*collstatsCollector)(nil)
