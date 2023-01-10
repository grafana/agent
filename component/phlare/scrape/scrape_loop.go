package scrape

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	commonconfig "github.com/prometheus/common/config"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/util/pool"
	"golang.org/x/net/context/ctxhttp"

	"github.com/grafana/agent/component/phlare"
	"github.com/grafana/agent/pkg/build"
)

var (
	payloadBuffers  = pool.New(1e3, 1e6, 3, func(sz int) interface{} { return make([]byte, 0, sz) })
	userAgentHeader = fmt.Sprintf("GrafanaAgent/%s", build.Version)
)

type scrapePool struct {
	config Arguments

	logger       log.Logger
	scrapeClient *http.Client
	appendable   phlare.Appendable

	mtx            sync.RWMutex
	activeTargets  map[uint64]*scrapeLoop
	droppedTargets []*Target
}

func newScrapePool(cfg Arguments, appendable phlare.Appendable, logger log.Logger) (*scrapePool, error) {
	scrapeClient, err := commonconfig.NewClientFromConfig(*cfg.HTTPClientConfig.Convert(), cfg.JobName)
	if err != nil {
		return nil, err
	}

	return &scrapePool{
		config:        cfg,
		logger:        logger,
		scrapeClient:  scrapeClient,
		appendable:    appendable,
		activeTargets: map[uint64]*scrapeLoop{},
	}, nil
}

func (tg *scrapePool) sync(groups []*targetgroup.Group) {
	tg.mtx.Lock()
	defer tg.mtx.Unlock()

	level.Info(tg.logger).Log("msg", "syncing target groups", "job", tg.config.JobName)
	var actives []*Target
	tg.droppedTargets = []*Target{}
	for _, group := range groups {
		targets, dropped, err := targetsFromGroup(group, tg.config)
		if err != nil {
			level.Error(tg.logger).Log("msg", "creating targets failed", "err", err)
			continue
		}
		for _, t := range targets {
			if t.Labels().Len() > 0 {
				actives = append(actives, t)
			}
		}
		tg.droppedTargets = append(tg.droppedTargets, dropped...)
	}

	for _, t := range actives {
		if _, ok := tg.activeTargets[t.hash()]; !ok {
			loop := newScrapeLoop(t, tg.scrapeClient, tg.appendable, tg.config.ScrapeInterval, tg.config.ScrapeTimeout, tg.logger)
			tg.activeTargets[t.hash()] = loop
			loop.start()
		} else {
			tg.activeTargets[t.hash()].SetDiscoveredLabels(t.DiscoveredLabels())
		}
	}

	// Removes inactive targets.
Outer:
	for h, t := range tg.activeTargets {
		for _, at := range actives {
			if h == at.hash() {
				continue Outer
			}
		}
		t.stop()
		delete(tg.activeTargets, h)
	}
}

func (tg *scrapePool) reload(cfg Arguments) error {
	tg.mtx.Lock()
	defer tg.mtx.Unlock()
	tg.config = cfg
	scrapeClient, err := commonconfig.NewClientFromConfig(*cfg.HTTPClientConfig.Convert(), cfg.JobName)
	if err != nil {
		return err
	}
	tg.scrapeClient = scrapeClient
	for _, t := range tg.activeTargets {
		t.reload(scrapeClient, cfg.ScrapeInterval, cfg.ScrapeTimeout)
	}
	return nil
}

func (tg *scrapePool) stop() {
	tg.mtx.Lock()
	defer tg.mtx.Unlock()

	for _, t := range tg.activeTargets {
		t.stop()
	}
}

func (tg *scrapePool) ActiveTargets() []*Target {
	tg.mtx.RLock()
	defer tg.mtx.RUnlock()
	result := make([]*Target, 0, len(tg.activeTargets))
	for _, target := range tg.activeTargets {
		result = append(result, target.Target)
	}
	return result
}

func (tg *scrapePool) DroppedTargets() []*Target {
	tg.mtx.RLock()
	defer tg.mtx.RUnlock()
	result := make([]*Target, 0, len(tg.droppedTargets))
	result = append(result, tg.droppedTargets...)
	return result
}

type scrapeLoop struct {
	*Target

	lastScrapeSize int

	scrapeClient *http.Client
	appendable   phlare.Appendable

	req               *http.Request
	logger            log.Logger
	interval, timeout time.Duration
	cancel            context.CancelFunc
	wg                sync.WaitGroup
}

func newScrapeLoop(t *Target, scrapeClient *http.Client, appendable phlare.Appendable, interval, timeout time.Duration, logger log.Logger) *scrapeLoop {
	return &scrapeLoop{
		Target:       t,
		logger:       logger,
		scrapeClient: scrapeClient,
		appendable:   appendable,
		interval:     interval,
		timeout:      timeout,
	}
}

func (t *scrapeLoop) start() {
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	t.wg.Add(1)

	go func() {
		defer func() {
			cancel()
			t.wg.Done()
		}()
		select {
		case <-time.After(t.offset(t.interval)):
		case <-ctx.Done():
			return
		}
		ticker := time.NewTicker(t.interval)
		defer ticker.Stop()

		tick := func() {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
		for ; true; tick() {
			if ctx.Err() != nil {
				return
			}
			t.scrape(ctx)
		}
	}()
}

func (t *scrapeLoop) scrape(ctx context.Context) {
	var (
		start             = time.Now()
		b                 = payloadBuffers.Get(t.lastScrapeSize).([]byte)
		buf               = bytes.NewBuffer(b)
		profileType       string
		scrapeCtx, cancel = context.WithTimeout(ctx, t.timeout)
	)
	defer cancel()

	for _, l := range t.labels {
		if l.Name == ProfileName {
			profileType = l.Value
			break
		}
	}
	err := t.fetchProfile(scrapeCtx, profileType, buf)
	t.mtx.Lock()
	defer t.mtx.Unlock()
	if err != nil {
		level.Error(t.logger).Log("msg", "fetch profile failed", "target", t.Labels().String(), "err", err)
		t.health = HealthBad
		t.lastScrapeDuration = time.Since(start)
		t.lastError = err
		t.lastScrape = start
		return
	}

	b = buf.Bytes()
	if len(b) > 0 {
		t.lastScrapeSize = len(b)
	}
	t.health = HealthGood
	t.lastScrapeDuration = time.Since(start)
	t.lastError = nil
	t.lastScrape = start
	if err := t.appendable.Appender().Append(ctx, t.labels, []*phlare.RawSample{{RawProfile: b}}); err != nil {
		level.Error(t.logger).Log("msg", "push failed", "labels", t.Labels().String(), "err", err)
	}
}

func (t *scrapeLoop) reload(scrapeClient *http.Client, interval, timeout time.Duration) {
	t.stop()
	t.scrapeClient = scrapeClient
	t.interval = interval
	t.timeout = timeout
	t.start()
}

func (t *scrapeLoop) fetchProfile(ctx context.Context, profileType string, buf io.Writer) error {
	if t.req == nil {
		req, err := http.NewRequest("GET", t.URL().String(), nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", userAgentHeader)

		t.req = req
	}

	level.Debug(t.logger).Log("msg", "scraping profile", "labels", t.Labels().String(), "url", t.req.URL.String())
	resp, err := ctxhttp.Do(ctx, t.scrapeClient, t.req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(io.TeeReader(resp.Body, buf))
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		if len(b) > 0 {
			return fmt.Errorf("server returned HTTP status (%d) %v", resp.StatusCode, string(bytes.TrimSpace(b)))
		}
		return fmt.Errorf("server returned HTTP status (%d) %v", resp.StatusCode, resp.Status)
	}

	if len(b) == 0 {
		return fmt.Errorf("empty %s profile from %s", profileType, t.req.URL.String())
	}
	return nil
}

func (t *scrapeLoop) stop() {
	t.cancel()
	t.wg.Wait()
}
