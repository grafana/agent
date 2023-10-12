package client

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/grafana/loki/pkg/ingester/wal"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/prometheus/tsdb/chunks"
	"github.com/prometheus/prometheus/tsdb/record"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/dskit/backoff"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"

	lokiutil "github.com/grafana/loki/pkg/util"
)

type queueClient struct {
	metrics *Metrics
	logger  log.Logger
	cfg     Config
	client  *http.Client

	batches      map[string]*batch
	batchesMtx   sync.Mutex
	sendQueue    chan queuedBatch
	drainTimeout time.Duration

	once sync.Once
	wg   sync.WaitGroup

	externalLabels model.LabelSet

	// series cache
	series        map[chunks.HeadSeriesRef]model.LabelSet
	seriesSegment map[chunks.HeadSeriesRef]int
	seriesLock    sync.RWMutex

	// ctx is used in any upstream calls from the `client`.
	ctx                 context.Context
	cancel              context.CancelFunc
	maxStreams          int
	maxLineSize         int
	maxLineSizeTruncate bool
	quit                chan struct{}
}

func (c *queueClient) SeriesReset(segmentNum int) {
	c.seriesLock.Lock()
	defer c.seriesLock.Unlock()
	for k, v := range c.seriesSegment {
		if v <= segmentNum {
			level.Debug(c.logger).Log("msg", fmt.Sprintf("reclaiming series under segment %d", segmentNum))
			delete(c.seriesSegment, k)
			delete(c.series, k)
		}
	}
}

func (c *queueClient) StoreSeries(series []record.RefSeries, segment int) {
	c.seriesLock.Lock()
	defer c.seriesLock.Unlock()
	for _, seriesRec := range series {
		c.seriesSegment[seriesRec.Ref] = segment
		labels := lokiutil.MapToModelLabelSet(seriesRec.Labels.Map())
		c.series[seriesRec.Ref] = labels
	}
}

type QueueConfig struct {
	Size         int
	DrainTimeout time.Duration
}

func NewQueue(metrics *Metrics, cfg Config, maxStreams, maxLineSize int, maxLineSizeTruncate bool, logger log.Logger, queueConfig QueueConfig) (*queueClient, error) {
	if cfg.StreamLagLabels.String() != "" {
		return nil, fmt.Errorf("client config stream_lag_labels is deprecated and the associated metric has been removed, stream_lag_labels: %+v", cfg.StreamLagLabels.String())
	}
	return newQueueClient(metrics, cfg, maxStreams, maxLineSize, maxLineSizeTruncate, logger, queueConfig)
}

func newQueueClient(metrics *Metrics, cfg Config, maxStreams, maxLineSize int, maxLineSizeTruncate bool, logger log.Logger, queueConfig QueueConfig) (*queueClient, error) {
	if cfg.URL.URL == nil {
		return nil, errors.New("client needs target URL")
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := &queueClient{
		logger:       log.With(logger, "component", "client", "host", cfg.URL.Host),
		cfg:          cfg,
		metrics:      metrics,
		sendQueue:    make(chan queuedBatch, queueConfig.Size), // this channel should have some buffering
		drainTimeout: queueConfig.DrainTimeout,                 // make this configurable
		quit:         make(chan struct{}),

		batches: make(map[string]*batch),

		series:        make(map[chunks.HeadSeriesRef]model.LabelSet),
		seriesSegment: make(map[chunks.HeadSeriesRef]int),

		externalLabels:      cfg.ExternalLabels.LabelSet,
		ctx:                 ctx,
		cancel:              cancel,
		maxStreams:          maxStreams,
		maxLineSize:         maxLineSize,
		maxLineSizeTruncate: maxLineSizeTruncate,
	}

	err := cfg.Client.Validate()
	if err != nil {
		return nil, err
	}

	c.client, err = config.NewClientFromConfig(cfg.Client, "GrafanaAgent", config.WithHTTP2Disabled())
	if err != nil {
		return nil, err
	}

	c.client.Timeout = cfg.Timeout

	// Initialize counters to 0 so the metrics are exported before the first
	// occurrence of incrementing to avoid missing metrics.
	for _, counter := range c.metrics.countersWithHost {
		counter.WithLabelValues(c.cfg.URL.Host).Add(0)
	}

	c.wg.Add(2)
	go c.runSendQueue()
	go c.runSendOldBatches()
	return c, nil
}

func (c *queueClient) initBatchMetrics(tenantID string) {
	// Initialize counters to 0 so the metrics are exported before the first
	// occurrence of incrementing to avoid missing metrics.
	for _, counter := range c.metrics.countersWithHostTenantReason {
		for _, reason := range Reasons {
			counter.WithLabelValues(c.cfg.URL.Host, tenantID, reason).Add(0)
		}
	}

	for _, counter := range c.metrics.countersWithHostTenant {
		counter.WithLabelValues(c.cfg.URL.Host, tenantID).Add(0)
	}
}

func (c *queueClient) AppendEntries(entries wal.RefEntries, segment int) error {
	c.seriesLock.RLock()
	l, ok := c.series[entries.Ref]
	c.seriesLock.RUnlock()
	if ok {
		for _, e := range entries.Entries {
			c.appendSingleEntry(l, e, segment)
		}
	} else {
		// TODO(thepalbi): Add metric here
		level.Debug(c.logger).Log("msg", "series for entry not found")
	}
	return nil
}

func (c *queueClient) appendSingleEntry(lbs model.LabelSet, e logproto.Entry, segment int) {
	lbs, tenantID := c.processLabels(lbs)

	// Either drop or mutate the log entry because its length is greater than maxLineSize. maxLineSize == 0 means disabled.
	if c.maxLineSize != 0 && len(e.Line) > c.maxLineSize {
		if !c.maxLineSizeTruncate {
			c.metrics.droppedEntries.WithLabelValues(c.cfg.URL.Host, tenantID, ReasonLineTooLong).Inc()
			c.metrics.droppedBytes.WithLabelValues(c.cfg.URL.Host, tenantID, ReasonLineTooLong).Add(float64(len(e.Line)))
			return
		}

		c.metrics.mutatedEntries.WithLabelValues(c.cfg.URL.Host, tenantID, ReasonLineTooLong).Inc()
		c.metrics.mutatedBytes.WithLabelValues(c.cfg.URL.Host, tenantID, ReasonLineTooLong).Add(float64(len(e.Line) - c.maxLineSize))
		e.Line = e.Line[:c.maxLineSize]
	}

	// TODO: can I make this locking more fine grained?
	c.batchesMtx.Lock()

	batch, ok := c.batches[tenantID]

	// If the batch doesn't exist yet, we create a new one with the entry
	if !ok {
		nb := newBatch(c.maxStreams)
		_ = nb.addFromWAL(lbs, e, segment)

		c.batches[tenantID] = nb
		c.batchesMtx.Unlock()

		c.initBatchMetrics(tenantID)
		return
	}

	// If adding the entry to the batch will increase the size over the max
	// size allowed, we do send the current batch and then create a new one
	if batch.sizeBytesAfter(e.Line) > c.cfg.BatchSize {
		c.enqueue(tenantID, batch)

		nb := newBatch(c.maxStreams)
		_ = nb.addFromWAL(lbs, e, segment)
		c.batches[tenantID] = nb
		c.batchesMtx.Unlock()

		return
	}

	// The max size of the batch isn't reached, so we can add the entry
	err := batch.addFromWAL(lbs, e, segment)
	c.batchesMtx.Unlock()

	if err != nil {
		level.Error(c.logger).Log("msg", "batch add err", "tenant", tenantID, "error", err)
		reason := ReasonGeneric
		if err.Error() == errMaxStreamsLimitExceeded {
			reason = ReasonStreamLimited
		}
		c.metrics.droppedBytes.WithLabelValues(c.cfg.URL.Host, tenantID, reason).Add(float64(len(e.Line)))
		c.metrics.droppedEntries.WithLabelValues(c.cfg.URL.Host, tenantID, reason).Inc()
	}
}

type queuedBatch struct {
	TenantID string
	Batch    *batch
}

func (c *queueClient) enqueue(tenantID string, b *batch) {
	c.sendQueue <- queuedBatch{
		TenantID: tenantID,
		Batch:    b,
	}
}

func (c *queueClient) runSendQueue() {
	defer c.wg.Done()

	for {
		select {
		case <-c.quit:
			return

		case qb, ok := <-c.sendQueue:
			if !ok {
				return
			}
			c.sendBatch(context.Background(), qb.TenantID, qb.Batch)
		}
	}
}

func (c *queueClient) runSendOldBatches() {
	// Given the client handles multiple batches (1 per tenant) and each batch
	// can be created at a different point in time, we look for batches whose
	// max wait time has been reached every 10 times per BatchWait, so that the
	// maximum delay we have sending batches is 10% of the max waiting time.
	// We apply a cap of 10ms to the ticker, to avoid too frequent checks in
	// case the BatchWait is very low.
	minWaitCheckFrequency := 10 * time.Millisecond
	maxWaitCheckFrequency := c.cfg.BatchWait / 10
	if maxWaitCheckFrequency < minWaitCheckFrequency {
		maxWaitCheckFrequency = minWaitCheckFrequency
	}

	maxWaitCheck := time.NewTicker(maxWaitCheckFrequency)

	// pablo: maybe this should be moved out
	defer func() {
		maxWaitCheck.Stop()
		c.wg.Done()
	}()

	batchesToFlush := make([]queuedBatch, 0)

	for {
		select {
		case <-c.quit:
			return

		case <-maxWaitCheck.C:
			c.batchesMtx.Lock()
			// Send all batches whose max wait time has been reached
			for tenantID, b := range c.batches {
				if b.age() < c.cfg.BatchWait {
					continue
				}

				// add to batches to flush, so we can enqueue them later and release the batches lock
				// as early as possible
				batchesToFlush = append(batchesToFlush, queuedBatch{
					TenantID: tenantID,
					Batch:    b,
				})

				// deleting assuming that since the batch expired the wait time, it
				// hasn't been written for some time
				delete(c.batches, tenantID)
			}

			c.batchesMtx.Unlock()

			// enqueue batches that were marked as too old
			for _, qb := range batchesToFlush {
				c.sendQueue <- qb
			}

			batchesToFlush = batchesToFlush[:] // renew slide
		}
	}
}

func (c *queueClient) drain(ctx context.Context) {
	// auxiliary go-routine will enqueue partial batches to use the same sending routine
	// this helps constraint the whole drain under a single timeout
	go func() {
		c.batchesMtx.Lock()
		defer c.batchesMtx.Unlock()

		for tenantID, batch := range c.batches {
			select {
			case <-ctx.Done():
				return
			case c.sendQueue <- queuedBatch{
				TenantID: tenantID,
				Batch:    batch,
			}:
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			level.Warn(c.logger).Log("msg", "drain send queue exceeded context timeout")
			return
		case qe := <-c.sendQueue:
			c.sendBatch(ctx, qe.TenantID, qe.Batch)
		default:
			// we want a default case for when the sendQueue channel has been completely drained
			level.Debug(c.logger).Log("msg", "sendQueue drain complete. No batches left in channel")
			return
		}
	}
}

func (c *queueClient) sendBatch(ctx context.Context, tenantID string, batch *batch) {
	buf, entriesCount, err := batch.encode()
	if err != nil {
		level.Error(c.logger).Log("msg", "error encoding batch", "error", err)
		return
	}
	bufBytes := float64(len(buf))
	c.metrics.encodedBytes.WithLabelValues(c.cfg.URL.Host).Add(bufBytes)

	backoff := backoff.New(c.ctx, c.cfg.BackoffConfig)
	var status int
	for {
		start := time.Now()
		// send uses `timeout` internally, so `context.Background` is good enough.
		status, err = c.send(ctx, tenantID, buf)

		c.metrics.requestDuration.WithLabelValues(strconv.Itoa(status), c.cfg.URL.Host).Observe(time.Since(start).Seconds())

		// Immediately drop rate limited batches to avoid HOL blocking for other tenants not experiencing throttling
		if c.cfg.DropRateLimitedBatches && batchIsRateLimited(status) {
			level.Warn(c.logger).Log("msg", "dropping batch due to rate limiting applied at ingester")
			c.metrics.droppedBytes.WithLabelValues(c.cfg.URL.Host, tenantID, ReasonRateLimited).Add(bufBytes)
			c.metrics.droppedEntries.WithLabelValues(c.cfg.URL.Host, tenantID, ReasonRateLimited).Add(float64(entriesCount))
			return
		}

		if err == nil {
			c.metrics.sentBytes.WithLabelValues(c.cfg.URL.Host).Add(bufBytes)
			c.metrics.sentEntries.WithLabelValues(c.cfg.URL.Host).Add(float64(entriesCount))

			return
		}

		// Only retry 429s, 500s and connection-level errors.
		if status > 0 && !batchIsRateLimited(status) && status/100 != 5 {
			break
		}

		level.Warn(c.logger).Log("msg", "error sending batch, will retry", "status", status, "tenant", tenantID, "error", err)
		c.metrics.batchRetries.WithLabelValues(c.cfg.URL.Host, tenantID).Inc()
		backoff.Wait()

		// Make sure it sends at least once before checking for retry.
		if !backoff.Ongoing() {
			break
		}
	}

	if err != nil {
		level.Error(c.logger).Log("msg", "final error sending batch", "status", status, "tenant", tenantID, "error", err)
		// If the reason for the last retry error was rate limiting, count the drops as such, even if the previous errors
		// were for a different reason
		dropReason := ReasonGeneric
		if batchIsRateLimited(status) {
			dropReason = ReasonRateLimited
		}
		c.metrics.droppedBytes.WithLabelValues(c.cfg.URL.Host, tenantID, dropReason).Add(bufBytes)
		c.metrics.droppedEntries.WithLabelValues(c.cfg.URL.Host, tenantID, dropReason).Add(float64(entriesCount))
	}
}

func (c *queueClient) send(ctx context.Context, tenantID string, buf []byte) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()
	req, err := http.NewRequest("POST", c.cfg.URL.String(), bytes.NewReader(buf))
	if err != nil {
		return -1, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", UserAgent)

	// If the tenant ID is not empty promtail is running in multi-tenant mode, so
	// we should send it to Loki
	if tenantID != "" {
		req.Header.Set("X-Scope-OrgID", tenantID)
	}

	// Add custom headers on request
	if len(c.cfg.Headers) > 0 {
		for k, v := range c.cfg.Headers {
			if req.Header.Get(k) == "" {
				req.Header.Add(k, v)
			} else {
				level.Warn(c.logger).Log("msg", "custom header key already exists, skipping", "key", k)
			}
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return -1, err
	}
	defer lokiutil.LogError("closing response body", resp.Body.Close)

	if resp.StatusCode/100 != 2 {
		scanner := bufio.NewScanner(io.LimitReader(resp.Body, maxErrMsgLen))
		line := ""
		if scanner.Scan() {
			line = scanner.Text()
		}
		err = fmt.Errorf("server returned HTTP status %s (%d): %s", resp.Status, resp.StatusCode, line)
	}
	return resp.StatusCode, err
}

func (c *queueClient) getTenantID(labels model.LabelSet) string {
	// Check if it has been overridden while processing the pipeline stages
	if value, ok := labels[ReservedLabelTenantID]; ok {
		return string(value)
	}

	// Check if has been specified in the config
	if c.cfg.TenantID != "" {
		return c.cfg.TenantID
	}

	// Defaults to an empty string, which means the X-Scope-OrgID header
	// will not be sent
	return ""
}

// Stop the client.
func (c *queueClient) Stop() {
	// first close main queue routine
	close(c.quit)
	c.wg.Wait()

	// drain with timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.drainTimeout)
	defer cancel()
	c.drain(ctx)

	close(c.sendQueue)
}

// StopNow stops the client without retries or draining the send queue
func (c *queueClient) StopNow() {
	// cancel will stop retrying http requests.
	c.cancel()
	close(c.quit)
	close(c.sendQueue)
	c.wg.Wait()
}

func (c *queueClient) processLabels(lbs model.LabelSet) (model.LabelSet, string) {
	if len(c.externalLabels) > 0 {
		lbs = c.externalLabels.Merge(lbs)
	}
	tenantID := c.getTenantID(lbs)
	return lbs, tenantID
}
