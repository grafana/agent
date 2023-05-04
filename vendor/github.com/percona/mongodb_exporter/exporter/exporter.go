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

// Package exporter implements the collectors and metrics handlers.
package exporter

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/percona/exporter_shared"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exporter holds Exporter methods and attributes.
type Exporter struct {
	path                  string
	client                *mongo.Client
	clientMu              sync.Mutex
	logger                *logrus.Logger
	opts                  *Opts
	webListenAddress      string
	lock                  *sync.Mutex
	totalCollectionsCount int
}

// Opts holds new exporter options.
type Opts struct {
	// Only get stats for the collections matching this list of namespaces.
	// Example: db1.col1,db.col1
	CollStatsNamespaces    []string
	CollStatsLimit         int
	CompatibleMode         bool
	DirectConnect          bool
	DisableDefaultRegistry bool
	DiscoveringMode        bool
	GlobalConnPool         bool

	CollectAll             bool
	EnableDBStats          bool
	EnableDiagnosticData   bool
	EnableReplicasetStatus bool
	EnableTopMetrics       bool
	EnableIndexStats       bool
	EnableCollStats        bool

	EnableOverrideDescendingIndex bool

	IndexStatsCollections []string
	Logger                *logrus.Logger
	Path                  string
	URI                   string
	WebListenAddress      string
}

var (
	errCannotHandleType   = fmt.Errorf("don't know how to handle data type")
	errUnexpectedDataType = fmt.Errorf("unexpected data type")
)

const (
	defaultCacheSize = 1000
)

// New connects to the database and returns a new Exporter instance.
func New(opts *Opts) *Exporter {
	if opts == nil {
		opts = new(Opts)
	}

	if opts.Logger == nil {
		opts.Logger = logrus.New()
	}

	ctx := context.Background()

	exp := &Exporter{
		path:                  opts.Path,
		logger:                opts.Logger,
		opts:                  opts,
		webListenAddress:      opts.WebListenAddress,
		lock:                  &sync.Mutex{},
		totalCollectionsCount: -1, // Not calculated yet. waiting the db connection.
	}
	// Try initial connect. Connection will be retried with every scrape.
	go func() {
		_, err := exp.getClient(ctx)
		if err != nil {
			exp.logger.Errorf("Cannot connect to MongoDB: %v", err)
		}
	}()

	return exp
}

func (e *Exporter) getTotalCollectionsCount() int {
	e.lock.Lock()
	defer e.lock.Unlock()

	return e.totalCollectionsCount
}

func (e *Exporter) makeRegistry(ctx context.Context, client *mongo.Client, topologyInfo labelsGetter, requestOpts Opts) *prometheus.Registry {
	registry := prometheus.NewRegistry()

	gc := newGeneralCollector(ctx, client, e.opts.Logger)
	registry.MustRegister(gc)

	if client == nil {
		return registry
	}

	nodeType, err := getNodeType(ctx, client)
	if err != nil {
		e.logger.Errorf("Cannot get node type to check if this is a mongos: %s", err)
	}

	// Enable collectors like collstats and indexstats depending on the number of collections
	// present in the database.
	limitsOk := false
	if e.opts.CollStatsLimit <= 0 || // Unlimited
		e.getTotalCollectionsCount() <= e.opts.CollStatsLimit {
		limitsOk = true
	}

	if e.opts.CollectAll {
		if len(e.opts.CollStatsNamespaces) == 0 {
			e.opts.DiscoveringMode = true
		}
		e.opts.EnableDiagnosticData = true
		e.opts.EnableDBStats = true
		e.opts.EnableCollStats = true
		e.opts.EnableTopMetrics = true
		e.opts.EnableReplicasetStatus = true
		e.opts.EnableIndexStats = true
	}

	// If we manually set the collection names we want or auto discovery is set.
	if (len(e.opts.CollStatsNamespaces) > 0 || e.opts.DiscoveringMode) && e.opts.EnableCollStats && limitsOk && requestOpts.EnableCollStats {
		cc := newCollectionStatsCollector(ctx, client, e.opts.Logger,
			e.opts.CompatibleMode, e.opts.DiscoveringMode,
			topologyInfo, e.opts.CollStatsNamespaces)
		registry.MustRegister(cc)
	}

	// If we manually set the collection names we want or auto discovery is set.
	if (len(e.opts.IndexStatsCollections) > 0 || e.opts.DiscoveringMode) && e.opts.EnableIndexStats && limitsOk && requestOpts.EnableIndexStats {
		ic := newIndexStatsCollector(ctx, client, e.opts.Logger,
			e.opts.DiscoveringMode, e.opts.EnableOverrideDescendingIndex,
			topologyInfo, e.opts.IndexStatsCollections)
		registry.MustRegister(ic)
	}

	if e.opts.EnableDiagnosticData && requestOpts.EnableDiagnosticData {
		ddc := newDiagnosticDataCollector(ctx, client, e.opts.Logger,
			e.opts.CompatibleMode, topologyInfo)
		registry.MustRegister(ddc)
	}

	if e.opts.EnableDBStats && limitsOk && requestOpts.EnableDBStats {
		cc := newDBStatsCollector(ctx, client, e.opts.Logger,
			e.opts.CompatibleMode, topologyInfo, nil)
		registry.MustRegister(cc)
	}

	if e.opts.EnableTopMetrics && nodeType != typeMongos && limitsOk && requestOpts.EnableTopMetrics {
		tc := newTopCollector(ctx, client, e.opts.Logger,
			e.opts.CompatibleMode, topologyInfo)
		registry.MustRegister(tc)
	}

	// replSetGetStatus is not supported through mongos.
	if e.opts.EnableReplicasetStatus && nodeType != typeMongos && requestOpts.EnableReplicasetStatus {
		rsgsc := newReplicationSetStatusCollector(ctx, client, e.opts.Logger,
			e.opts.CompatibleMode, topologyInfo)
		registry.MustRegister(rsgsc)
	}

	return registry
}

func (e *Exporter) getClient(ctx context.Context) (*mongo.Client, error) {
	if e.opts.GlobalConnPool {
		// Get global client. Maybe it must be initialized first.
		// Initialization is retried with every scrape until it succeeds once.
		e.clientMu.Lock()
		defer e.clientMu.Unlock()

		// If client is already initialized, return it.
		if e.client != nil {
			return e.client, nil
		}

		client, err := connect(context.Background(), e.opts.URI, e.opts.DirectConnect)
		if err != nil {
			return nil, err
		}
		e.client = client

		return client, nil
	}

	// !e.opts.GlobalConnPool: create new client for every scrape.
	client, err := connect(ctx, e.opts.URI, e.opts.DirectConnect)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Handler returns an http.Handler that serves metrics. Can be used instead of
// run for hooking up custom HTTP servers.
func (e *Exporter) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seconds, err := strconv.Atoi(r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"))
		// To support also older ones vmagents.
		if err != nil {
			seconds = 10
		}

		var client *mongo.Client
		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(seconds)*time.Second)
		defer cancel()

		filters := r.URL.Query()["collect[]"]

		requestOpts := Opts{}

		if len(filters) == 0 {
			requestOpts = *e.opts
		}

		for _, filter := range filters {
			switch filter {
			case "diagnosticdata":
				requestOpts.EnableDiagnosticData = true
			case "replicasetstatus":
				requestOpts.EnableReplicasetStatus = true
			case "dbstats":
				requestOpts.EnableDBStats = true
			case "topmetrics":
				requestOpts.EnableTopMetrics = true
			case "indexstats":
				requestOpts.EnableIndexStats = true
			case "collstats":
				requestOpts.EnableCollStats = true
			}
		}

		client, err = e.getClient(ctx)
		if err != nil {
			e.logger.Errorf("Cannot connect to MongoDB: %v", err)
		}

		if client != nil && e.getTotalCollectionsCount() <= 0 {
			count, err := nonSystemCollectionsCount(ctx, client, nil, nil)
			if err == nil {
				e.lock.Lock()
				e.totalCollectionsCount = count
				e.lock.Unlock()
			}
		}

		// Close client after usage.
		if !e.opts.GlobalConnPool {
			defer func() {
				if client != nil {
					err := client.Disconnect(ctx)
					if err != nil {
						e.logger.Errorf("Cannot disconnect client: %v", err)
					}
				}
			}()
		}

		// Topology can change between requests, so we need to get it every time.
		ti := newTopologyInfo(ctx, client)

		registry := e.makeRegistry(ctx, client, ti, requestOpts)

		var gatherers prometheus.Gatherers

		if !e.opts.DisableDefaultRegistry {
			gatherers = append(gatherers, prometheus.DefaultGatherer)
		}
		gatherers = append(gatherers, registry)

		// Delegate http serving to Prometheus client library, which will call collector.Collect.
		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
			ErrorLog:      e.logger,
		})

		h.ServeHTTP(w, r)
	})
}

// Run starts the exporter.
func (e *Exporter) Run() {
	handler := e.Handler()
	exporter_shared.RunServer("MongoDB", e.webListenAddress, e.path, handler)
}

func connect(ctx context.Context, dsn string, directConnect bool) (*mongo.Client, error) {
	clientOpts := options.Client().ApplyURI(dsn)
	clientOpts.SetDirect(directConnect)
	clientOpts.SetAppName("mongodb_exporter")

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("invalid MongoDB options: %w", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		// Ping failed. Close background connections. Error is ignored since the ping error is more relevant.
		_ = client.Disconnect(ctx)

		return nil, fmt.Errorf("cannot connect to MongoDB: %w", err)
	}

	return client, nil
}
