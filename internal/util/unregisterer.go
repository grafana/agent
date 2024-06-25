package util

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus"
)

// Unregisterer is a Prometheus Registerer that can unregister all collectors
// passed to it.
type Unregisterer struct {
	wrap prometheus.Registerer
	cs   map[prometheus.Collector]struct{}
}

// WrapWithUnregisterer wraps a prometheus Registerer with capabilities to
// unregister all collectors.
func WrapWithUnregisterer(reg prometheus.Registerer) *Unregisterer {
	return &Unregisterer{
		wrap: reg,
		cs:   make(map[prometheus.Collector]struct{}),
	}
}

func describeCollector(c prometheus.Collector) string {
	var (
		descChan = make(chan *prometheus.Desc, 10)
	)
	go func() {
		c.Describe(descChan)
		close(descChan)
	}()

	descs := make([]string, 0)
	for desc := range descChan {
		descs = append(descs, desc.String())
	}

	return fmt.Sprintf("%v", descs)
}

func isUncheckedCollector(c prometheus.Collector) bool {
	var (
		descChan = make(chan *prometheus.Desc, 10)
	)
	go func() {
		c.Describe(descChan)
		close(descChan)
	}()

	i := 0
	for range descChan {
		i += 1
	}

	return i == 0
}

// Register implements prometheus.Registerer.
func (u *Unregisterer) Register(c prometheus.Collector) error {
	if u.wrap == nil {
		return nil
	}

	err := u.wrap.Register(c)
	if err != nil {
		return err
	}

	if isUncheckedCollector(c) {
		return nil
	}

	u.cs[c] = struct{}{}
	return nil
}

// MustRegister implements prometheus.Registerer.
func (u *Unregisterer) MustRegister(cs ...prometheus.Collector) {
	for _, c := range cs {
		if err := u.Register(c); err != nil {
			panic(err)
		}
	}
}

// Unregister implements prometheus.Registerer.
func (u *Unregisterer) Unregister(c prometheus.Collector) bool {
	if isUncheckedCollector(c) {
		return true
	}

	if u.wrap != nil && u.wrap.Unregister(c) {
		delete(u.cs, c)
		return true
	}
	return false
}

// UnregisterAll unregisters all collectors that were registered through the
// Registerer.
func (u *Unregisterer) UnregisterAll() error {
	var multiErr error
	for c := range u.cs {
		if !u.Unregister(c) {
			err := fmt.Errorf("failed to unregister collector %v", describeCollector(c))
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr
}
