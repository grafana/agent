package discovery

import (
	"context"
	"time"

	"github.com/grafana/agent/component/metrics/scrape"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

// RunDiscovery is a utility for consuming and forwarding target groups from a discoverer.
// It will handle collating targets (and clearing), as well as time based throttling of updates.
// The channel provided should be the same channel passed to the Discovery's Run method.
func RunDiscovery(ctx context.Context, ch <-chan []*targetgroup.Group, f func([]scrape.Target)) {
	cache := map[string]*targetgroup.Group{}

	dirty := false

	const maxChangeFreq = 5 * time.Second
	// this should give us 2 seconds at startup to collect some changes before sending
	var lastChange time.Time = time.Now().Add(-3 * time.Second)
	for {
		var timeChan <-chan time.Time = nil
		if dirty {
			now := time.Now()
			nextValidTime := lastChange.Add(5 * time.Second)
			if now.Unix() > nextValidTime.Unix() {
				// We are past the threshold, send change notification now
				t := []scrape.Target{}
				for _, group := range cache {
					for _, target := range group.Targets {
						m := map[string]string{}
						for k, v := range group.Labels {
							m[string(k)] = string(v)
						}
						for k, v := range target {
							m[string(k)] = string(v)
						}
						t = append(t, m)
					}
				}
				f(t)
				lastChange = now
				dirty = false
			} else {
				// else set a timer
				timeToWait := nextValidTime.Sub(now)
				timeChan = time.After(timeToWait)
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-timeChan:
			continue
		case groups := <-ch:
			for _, group := range groups {
				if len(group.Targets) == 0 {
					delete(cache, group.Source)
				} else {
					cache[group.Source] = group
					dirty = true
				}
			}
		}
	}
}
