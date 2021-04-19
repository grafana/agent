// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package loadbalancingexporter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

var _ resolver = (*dnsResolver)(nil)

const (
	defaultResInterval = 5 * time.Second
	defaultResTimeout  = time.Second
)

var errNoHostname = errors.New("no hostname specified to resolve the backends")

type dnsResolver struct {
	logger *zap.Logger

	hostname    string
	port        string
	resolver    netResolver
	resInterval time.Duration
	resTimeout  time.Duration

	endpoints         []string
	onChangeCallbacks []func([]string)

	stopCh             chan (struct{})
	updateLock         sync.Mutex
	shutdownWg         sync.WaitGroup
	changeCallbackLock sync.RWMutex
}

type netResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

func newDNSResolver(logger *zap.Logger, hostname string, port string) (*dnsResolver, error) {
	if len(hostname) == 0 {
		return nil, errNoHostname
	}

	return &dnsResolver{
		logger:      logger,
		hostname:    hostname,
		port:        port,
		resolver:    &net.Resolver{},
		resInterval: defaultResInterval,
		resTimeout:  defaultResTimeout,
		stopCh:      make(chan struct{}),
	}, nil
}

func (r *dnsResolver) start(ctx context.Context) error {
	if _, err := r.resolve(ctx); err != nil {
		return err
	}

	go r.periodicallyResolve()

	return nil
}

func (r *dnsResolver) shutdown(ctx context.Context) error {
	r.changeCallbackLock.Lock()
	r.onChangeCallbacks = nil
	r.changeCallbackLock.Unlock()

	close(r.stopCh)
	r.shutdownWg.Wait()
	return nil
}

func (r *dnsResolver) periodicallyResolve() {
	ticker := time.NewTicker(r.resInterval)

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), r.resTimeout)
			if _, err := r.resolve(ctx); err != nil {
				r.logger.Warn("failed to resolve", zap.Error(err))
			}
			cancel()
		case <-r.stopCh:
			return
		}
	}
}

func (r *dnsResolver) resolve(ctx context.Context) ([]string, error) {
	r.shutdownWg.Add(1)
	defer r.shutdownWg.Done()

	// the context to use for all metrics in this function
	mCtx, _ := tag.New(ctx, tag.Upsert(tag.MustNewKey("resolver"), "dns"))

	addrs, err := r.resolver.LookupIPAddr(ctx, r.hostname)
	if err != nil {
		failedCtx, _ := tag.New(mCtx, tag.Upsert(tag.MustNewKey("success"), "false"))
		stats.Record(failedCtx, mNumResolutions.M(1))
		return nil, err
	}

	// from this point, we don't fail anymore
	successCtx, _ := tag.New(mCtx, tag.Upsert(tag.MustNewKey("success"), "true"))
	stats.Record(successCtx, mNumResolutions.M(1))

	var backends []string
	for _, ip := range addrs {
		var backend string
		if ip.IP.To4() != nil {
			backend = ip.String()
		} else {
			// it's an IPv6 address
			backend = fmt.Sprintf("[%s]", ip.String())
		}

		// if a port is specified in the configuration, add it
		if r.port != "" {
			backend = fmt.Sprintf("%s:%s", backend, r.port)
		}

		backends = append(backends, backend)
	}

	// keep it always in the same order
	sort.Strings(backends)

	if equalStringSlice(r.endpoints, backends) {
		return r.endpoints, nil
	}

	// the list has changed!
	r.updateLock.Lock()
	r.endpoints = backends
	r.updateLock.Unlock()
	stats.Record(mCtx, mNumBackends.M(int64(len(backends))))

	// propagate the change
	r.changeCallbackLock.RLock()
	for _, callback := range r.onChangeCallbacks {
		callback(r.endpoints)
	}
	r.changeCallbackLock.RUnlock()

	return r.endpoints, nil
}

func (r *dnsResolver) onChange(f func([]string)) {
	r.changeCallbackLock.Lock()
	defer r.changeCallbackLock.Unlock()
	r.onChangeCallbacks = append(r.onChangeCallbacks, f)
}

func equalStringSlice(source, candidate []string) bool {
	if len(source) != len(candidate) {
		return false
	}
	for i := range source {
		if source[i] != candidate[i] {
			return false
		}
	}

	return true
}
