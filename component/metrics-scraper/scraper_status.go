package metricsscraper

import (
	"sort"
	"time"
)

// Status returns debug info for a metrics_scraper component.
type Status struct {
	Targets []*TargetStatus `hcl:"target,block"`
}

type TargetStatus struct {
	URL                string            `hcl:"url"`
	Health             string            `hcl:"health"`
	Labels             map[string]string `hcl:"labels,attr"`
	LastError          string            `hcl:"last_error,optional"`
	LastScrape         time.Time         `hcl:"last_scrape"`
	LastScrapeDuration time.Duration     `hcl:"last_scrape_duration,optional"`
}

func (c *Component) CurrentStatus() any {
	var tss []*TargetStatus

	// There's only ever one job, so we can ignore the key here.
	for _, stt := range c.scraper.TargetsActive() {
		for _, st := range stt {
			var lastError string
			if err := st.LastError(); err != nil {
				lastError = err.Error()
			}

			tss = append(tss, &TargetStatus{
				URL:                st.URL().String(),
				Health:             string(st.Health()),
				Labels:             st.Labels().Map(),
				LastError:          lastError,
				LastScrape:         st.LastScrape(),
				LastScrapeDuration: st.LastScrapeDuration(),
			})
		}
	}

	sort.Slice(tss, func(i, j int) bool {
		return tss[i].URL < tss[j].URL
	})

	return Status{Targets: tss}
}
