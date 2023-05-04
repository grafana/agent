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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	replicationNotEnabled        = 76
	replicationNotYetInitialized = 94
)

type replSetGetStatusCollector struct {
	ctx  context.Context
	base *baseCollector

	compatibleMode bool
	topologyInfo   labelsGetter
}

// newReplicationSetStatusCollector creates a collector for statistics on replication set.
func newReplicationSetStatusCollector(ctx context.Context, client *mongo.Client, logger *logrus.Logger, compatible bool, topology labelsGetter) *replSetGetStatusCollector {
	return &replSetGetStatusCollector{
		ctx:  ctx,
		base: newBaseCollector(client, logger),

		compatibleMode: compatible,
		topologyInfo:   topology,
	}
}

func (d *replSetGetStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	d.base.Describe(d.ctx, ch, d.collect)
}

func (d *replSetGetStatusCollector) Collect(ch chan<- prometheus.Metric) {
	d.base.Collect(ch)
}

func (d *replSetGetStatusCollector) collect(ch chan<- prometheus.Metric) {
	logger := d.base.logger
	client := d.base.client

	cmd := bson.D{{Key: "replSetGetStatus", Value: "1"}}
	res := client.Database("admin").RunCommand(d.ctx, cmd)

	var m bson.M

	if err := res.Decode(&m); err != nil {
		if e, ok := err.(mongo.CommandError); ok {
			if e.Code == replicationNotYetInitialized || e.Code == replicationNotEnabled {
				return
			}
		}
		logger.Errorf("cannot get replSetGetStatus: %s", err)

		return
	}

	logger.Debug("replSetGetStatus result:")
	debugResult(logger, m)

	for _, metric := range makeMetrics("", m, d.topologyInfo.baseLabels(), d.compatibleMode) {
		ch <- metric
	}
}

var _ prometheus.Collector = (*replSetGetStatusCollector)(nil)
