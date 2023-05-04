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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type topCollector struct {
	ctx  context.Context
	base *baseCollector

	compatibleMode bool
	topologyInfo   labelsGetter
}

var ErrInvalidOrMissingTotalsEntry = fmt.Errorf("Invalid or misssing totals entry in top results")

func newTopCollector(ctx context.Context, client *mongo.Client, logger *logrus.Logger, compatible bool,
	topology labelsGetter,
) *topCollector {
	return &topCollector{
		ctx:            ctx,
		base:           newBaseCollector(client, logger),
		compatibleMode: compatible,
		topologyInfo:   topology,
	}
}

func (d *topCollector) Describe(ch chan<- *prometheus.Desc) {
	d.base.Describe(d.ctx, ch, d.collect)
}

func (d *topCollector) Collect(ch chan<- prometheus.Metric) {
	d.base.Collect(ch)
}

func (d *topCollector) collect(ch chan<- prometheus.Metric) {
	logger := d.base.logger
	client := d.base.client

	cmd := bson.D{{Key: "top", Value: "1"}}
	res := client.Database("admin").RunCommand(d.ctx, cmd)

	var m primitive.M
	if err := res.Decode(&m); err != nil {
		ch <- prometheus.NewInvalidMetric(prometheus.NewInvalidDesc(err), err)
		return
	}

	logrus.Debug("top result:")
	debugResult(logger, m)

	totals, ok := m["totals"].(primitive.M)
	if !ok {
		ch <- prometheus.NewInvalidMetric(prometheus.NewInvalidDesc(ErrInvalidOrMissingTotalsEntry),
			ErrInvalidOrMissingTotalsEntry)
	}

	/*
		      The top command will return a structure with a key named totals and it is a map
			  where the key is the collection namespace and for each collection there are per
			  collection usage statistics.
			  Example: 	rs1:SECONDARY> db.adminCommand({"top": 1});

		      {
		        "totals" : {
		                "note" : "all times in microseconds",
		                "admin.system.roles" : {
		                        "total" : {
		                                "time" : 41,
		                                "count" : 1
		                        },
		                        "readLock" : {
		                                "time" : 41,
		                                "count" : 1
		                        },
		                        "writeLock" : {
		                                "time" : 0,
		                                "count" : 0
		                        },
		                        "queries" : {
		                                "time" : 41,
		                                "count" : 1
		                        },
		                        "getmore" : {
		                                "time" : 0,
		                                "count" : 0
		                        },
		                        "insert" : {
		                                "time" : 0,
		                                "count" : 0
		                        },
		                        "update" : {
		                                "time" : 0,
		                                "count" : 0
		                        },
		                        "remove" : {
		                                "time" : 0,
		                                "count" : 0
		                        },
		                        "commands" : {
		                                "time" : 0,
		                                "count" : 0
		                        }
		                },
		                "admin.system.version" : {
		                        "total" : {
		                                "time" : 63541,
		                                "count" : 218
		                        },

		      If we pass this structure to the makeMetrics function, we will have metric names with the form of
			  prefix + namespace + metric like mongodb_top_totals_admin.system.role_readlock_count.
			  Having the namespace as part of the metric is a Prometheus anti pattern and diffucults grouping
			  metrics in Grafana. For this reason, we need to manually loop through the metric in the totals key
			  and pass the namespace as a label to the makeMetrics function.
	*/

	for namespace, metrics := range totals {
		labels := d.topologyInfo.baseLabels()
		db, coll := splitNamespace(namespace)
		labels["database"] = db
		labels["collection"] = coll

		mm, ok := metrics.(primitive.M) // ingore entries like -> "note" : "all times in microseconds"
		if !ok {
			continue
		}

		for _, metric := range makeMetrics("top", mm, labels, d.compatibleMode) {
			ch <- metric
		}
	}
}
