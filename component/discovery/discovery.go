package discovery

import (
	"context"
	"time"

	"github.com/grafana/agent/component/metrics/scrape"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

const MaxUpdateFrequency = 5 * time.Second

// RunDiscovery is a utility for consuming and forwarding target groups from a discoverer.
// It will handle collating targets (and clearing), as well as time based throttling of updates.
func RunDiscovery(ctx context.Context, d discovery.Discoverer, f func([]scrape.Target)) {
	// all targets we have seen so far
	cache := map[string]*targetgroup.Group{}

	ch := make(chan []*targetgroup.Group)
	go d.Run(ctx, ch)

	// function to convert and send targets in format scraper expects
	send := func() {
		allTargets := []scrape.Target{}
		for _, group := range cache {
			for _, target := range group.Targets {
				labels := map[string]string{}
				for k, v := range group.Labels {
					labels[string(k)] = string(v)
				}
				for k, v := range target {
					labels[string(k)] = string(v)
				}
				allTargets = append(allTargets, labels)
			}
		}
		f(allTargets)
	}

	ticker := time.NewTicker(MaxUpdateFrequency)
	// true if we have received new targets and need to send.
	haveUpdates := false
	for {
		select {
		case <-ticker.C:
			if haveUpdates {
				send()
				haveUpdates = false
			}
		case <-ctx.Done():
			send()
			return
		case groups := <-ch:
			for _, group := range groups {
				// Discoverer will send an empty target set to indicate the group (keyed by Source field)
				// should be removed
				if len(group.Targets) == 0 {
					delete(cache, group.Source)
				} else {
					cache[group.Source] = group
				}
			}
			haveUpdates = true
		}
	}
}
