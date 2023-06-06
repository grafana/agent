package scrape

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/pyroscope"
	commonconfig "github.com/prometheus/common/config"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/util/pool"
	"golang.org/x/net/context/ctxhttp"

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
	appendable   pyroscope.Appendable

	mtx            sync.RWMutex
	activeTargets  map[uint64]*scrapeLoop
	droppedTargets []*Target
}

func newScrapePool(cfg Arguments, appendable pyroscope.Appendable, logger log.Logger) (*scrapePool, error) {
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
		t.stop(false)
		delete(tg.activeTargets, h)
	}
}

func (tg *scrapePool) reload(cfg Arguments) error {
	tg.mtx.Lock()
	defer tg.mtx.Unlock()

	if tg.config.ScrapeInterval == cfg.ScrapeInterval &&
		tg.config.ScrapeTimeout == cfg.ScrapeTimeout &&
		reflect.DeepEqual(tg.config.HTTPClientConfig, cfg.HTTPClientConfig) {

		tg.config = cfg
		return nil
	}
	tg.config = cfg

	scrapeClient, err := commonconfig.NewClientFromConfig(*cfg.HTTPClientConfig.Convert(), cfg.JobName)
	if err != nil {
		return err
	}
	tg.scrapeClient = scrapeClient
	for hash, t := range tg.activeTargets {
		// restart the loop with the new configuration
		t.stop(false)
		loop := newScrapeLoop(t.Target, tg.scrapeClient, tg.appendable, tg.config.ScrapeInterval, tg.config.ScrapeTimeout, tg.logger)
		tg.activeTargets[hash] = loop
		loop.start()
	}
	return nil
}

func (tg *scrapePool) stop() {
	tg.mtx.Lock()
	defer tg.mtx.Unlock()

	wg := sync.WaitGroup{}
	for _, t := range tg.activeTargets {
		wg.Add(1)
		go func(t *scrapeLoop) {
			defer wg.Done()
			t.stop(true)
		}(t)
	}
	wg.Wait()
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
	appender     pyroscope.Appender

	req               *http.Request
	logger            log.Logger
	interval, timeout time.Duration
	graceShut         chan struct{}
	once              sync.Once
	wg                sync.WaitGroup
}

func newScrapeLoop(t *Target, scrapeClient *http.Client, appendable pyroscope.Appendable, interval, timeout time.Duration, logger log.Logger) *scrapeLoop {
	return &scrapeLoop{
		Target:       t,
		logger:       logger,
		scrapeClient: scrapeClient,
		appender:     NewDeltaAppender(appendable.Appender(), t.labels),
		interval:     interval,
		timeout:      timeout,
	}
}

func (t *scrapeLoop) start() {
	t.graceShut = make(chan struct{})
	t.once = sync.Once{}
	t.wg.Add(1)

	go func() {
		defer t.wg.Done()

		select {
		case <-time.After(t.offset(t.interval)):
		case <-t.graceShut:
			return
		}
		ticker := time.NewTicker(t.interval)
		defer ticker.Stop()

		for {
			select {
			case <-t.graceShut:
				return
			case <-ticker.C:
			}
			t.scrape()
		}
	}()
}

func (t *scrapeLoop) scrape() {
	var (
		start             = time.Now()
		b                 = payloadBuffers.Get(t.lastScrapeSize).([]byte)
		buf               = bytes.NewBuffer(b)
		profileType       string
		scrapeCtx, cancel = context.WithTimeout(context.Background(), t.timeout)
	)
	defer cancel()

	for _, l := range t.labels {
		if l.Name == ProfileName {
			profileType = l.Value
			break
		}
	}
	if err := t.fetchProfile(scrapeCtx, profileType, buf); err != nil {
		level.Debug(t.logger).Log("msg", "fetch profile failed", "target", t.Labels().String(), "err", err)
		t.updateTargetStatus(start, err)
		return
	}

	b = buf.Bytes()
	if len(b) > 0 {
		t.lastScrapeSize = len(b)
	}
	if err := t.appender.Append(context.Background(), t.labels, []*pyroscope.RawSample{{RawProfile: b}}); err != nil {
		level.Error(t.logger).Log("msg", "push failed", "labels", t.Labels().String(), "err", err)
		t.updateTargetStatus(start, err)
		return
	}
	t.updateTargetStatus(start, nil)
}

func (t *scrapeLoop) updateTargetStatus(start time.Time, err error) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	if err != nil {
		t.health = HealthBad
		t.lastError = err
	} else {
		t.health = HealthGood
		t.lastError = nil
	}
	t.lastScrape = start
	t.lastScrapeDuration = time.Since(start)
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

func (t *scrapeLoop) stop(wait bool) {
	t.once.Do(func() {
		close(t.graceShut)
	})
	if wait {
		t.wg.Wait()
	}
}
