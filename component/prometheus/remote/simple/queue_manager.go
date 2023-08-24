/*
 */
package simple

import (
	"context"
	"errors"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/grafana/agent/component/prometheus"
	"github.com/prometheus/common/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/textparse"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
)

type WriteClient interface {
	// Store stores the given samples in the remote storage.
	Store(context.Context, []byte) error
	// Name uniquely identifies the remote storage.
	Name() string
	// Endpoint is the remote read or write endpoint for the storage client.
	Endpoint() string
}

// QueueManager manages a queue of samples to be sent to the Storage
// indicated by the provided WriteClient. Implements writeTo interface
// used by WAL Watcher.
type QueueManager struct {
	logger               log.Logger
	flushDeadline        time.Duration
	cfg                  config.QueueConfig
	mcfg                 config.MetadataConfig
	sendExemplars        bool
	sendNativeHistograms bool

	clientMtx   sync.RWMutex
	storeClient WriteClient

	metrics              *queueManagerMetrics
	highestRecvTimestamp *maxTimestamp
}

// NewQueueManager builds a new QueueManager and starts a new
// WAL watcher with queue manager as the WriteTo destination.
// The WAL watcher takes the dir parameter as the base directory
// for where the WAL shall be located. Note that the full path to
// the WAL directory will be constructed as <dir>/wal.
func NewQueueManager(
	metrics *queueManagerMetrics,
	logger log.Logger,
	cfg config.QueueConfig,
	mCfg config.MetadataConfig,
	client WriteClient,
	flushDeadline time.Duration,
	highestRecvTimestamp *maxTimestamp,
	enableExemplarRemoteWrite bool,
	enableNativeHistogramRemoteWrite bool,
) *QueueManager {

	if logger == nil {
		logger = log.NewNopLogger()
	}

	t := &QueueManager{
		logger:               logger,
		flushDeadline:        flushDeadline,
		cfg:                  cfg,
		mcfg:                 mCfg,
		storeClient:          client,
		sendExemplars:        enableExemplarRemoteWrite,
		sendNativeHistograms: enableNativeHistogramRemoteWrite,
		metrics:              metrics,
		highestRecvTimestamp: highestRecvTimestamp,
	}
	return t
}

func (t *QueueManager) Name() string {
	return t.client().Name()
}

// AppendMetadata sends metadata the remote storage. Metadata is sent in batches, but is not parallelized.
func (t *QueueManager) AppendMetadata(metadata []prometheus.Metadata) bool {
	mm := make([]prompb.MetricMetadata, 0, len(metadata))
	for _, entry := range metadata {
		mm = append(mm, prompb.MetricMetadata{
			MetricFamilyName: entry.Name,
			Help:             entry.Meta.Help,
			Type:             metricTypeToMetricTypeProto(entry.Meta.Type),
			Unit:             entry.Meta.Unit,
		})
	}
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(t.cfg.BatchSendDeadline))
	defer cancel()
	numSends := int(math.Ceil(float64(len(metadata)) / float64(t.mcfg.MaxSamplesPerSend)))
	for i := 0; i < numSends; i++ {
		last := (i + 1) * t.mcfg.MaxSamplesPerSend
		if last > len(metadata) {
			last = len(metadata)
		}
		err := t.sendMetadataWithBackoff(ctx, mm[i*t.mcfg.MaxSamplesPerSend:last])
		if err != nil {
			t.metrics.failedMetadataTotal.Add(float64(last - (i * t.mcfg.MaxSamplesPerSend)))
			level.Error(t.logger).Log("msg", "non-recoverable error while sending metadata", "count", last-(i*t.mcfg.MaxSamplesPerSend), "err", err)
		}
	}
	return true
}

func (t *QueueManager) sendMetadataWithBackoff(ctx context.Context, metadata []prompb.MetricMetadata) error {
	// Build the WriteRequest with no samples.
	req, _, err := buildWriteRequest(nil, metadata)
	if err != nil {
		return err
	}

	metadataCount := len(metadata)

	attemptStore := func(try int) error {
		ctx, span := otel.Tracer("").Start(ctx, "Remote Metadata Send Batch")
		defer span.End()

		span.SetAttributes(
			attribute.Int("metadata", metadataCount),
			attribute.Int("try", try),
			attribute.String("remote_name", t.storeClient.Name()),
			attribute.String("remote_url", t.storeClient.Endpoint()),
		)

		begin := time.Now()
		err := t.storeClient.Store(ctx, req)
		t.metrics.sentBatchDuration.Observe(time.Since(begin).Seconds())

		if err != nil {
			span.RecordError(err)
			return err
		}

		return nil
	}

	retry := func() {
		t.metrics.retriedMetadataTotal.Add(float64(len(metadata)))
	}
	err = sendWriteRequestWithBackoff(ctx, t.cfg, t.logger, attemptStore, retry)
	if err != nil {
		return err
	}
	t.metrics.metadataTotal.Add(float64(len(metadata)))
	t.metrics.metadataBytesTotal.Add(float64(len(req)))
	return nil
}

func (t *QueueManager) append(ctx context.Context, samples []timeSeries) (bool, error) {
	/*
		1. Determine number of shards
		2. Queue up samples
		3. Send
	*/
	protoSamples := make([]prompb.TimeSeries, len(samples))
	t.populateTimeSeries(samples, protoSamples)
	// Simple approach that one shard can handle the load
	if t.cfg.MaxSamplesPerSend > len(protoSamples) {
		s := &shard{qm: t}
		return s.sendSamplesWithBackoff(ctx, protoSamples)
	}
	// Lets divide the work.
	// TODO remove hard coding.
	numShards := 4
	currentShard := 0
	queus := make(map[int][][]prompb.TimeSeries)
	for {
		if t.cfg.MaxSamplesPerSend > len(protoSamples) {
			queus[currentShard] = append(queus[currentShard], protoSamples)
			break
		}

		subset := protoSamples[:t.cfg.MaxSamplesPerSend]
		queus[currentShard] = append(queus[currentShard], subset)
		protoSamples = protoSamples[t.cfg.MaxSamplesPerSend:]
		currentShard = currentShard + 1
		if currentShard >= numShards {
			currentShard = 0
		}
	}
	// Now lets do the actual work.
	wg := &sync.WaitGroup{}
	wg.Add(numShards)

	overallSuccess := true
	var errMut sync.Mutex
	var overallError error
	for i := 0; i < numShards; i++ {
		go func(k int) {
			success, err := startSendingShard(t, queus[k], wg, ctx)
			if !success {
				overallSuccess = false
			}
			if err != nil {
				// TODO use a multi error
				errMut.Lock()
				overallError = err
				errMut.Unlock()
			}
		}(i)
	}
	wg.Wait()
	return overallSuccess, overallError
}

func startSendingShard(t *QueueManager, q [][]prompb.TimeSeries, wg *sync.WaitGroup, ctx context.Context) (bool, error) {
	defer wg.Done()
	s := shard{qm: t}
	// TODO reintroduce reusing the protobuf
	for _, data := range q {
		success, err := s.sendSamplesWithBackoff(ctx, data)
		if !success || err != nil {
			return false, err
		}
	}
	return true, nil
}

func (t *QueueManager) populateTimeSeries(batch []timeSeries, pendingData []prompb.TimeSeries) (int, int, int) {
	var nPendingSamples, nPendingExemplars, nPendingHistograms int
	for nPending, d := range batch {
		pendingData[nPending].Samples = pendingData[nPending].Samples[:0]
		if t.sendExemplars {
			pendingData[nPending].Exemplars = pendingData[nPending].Exemplars[:0]
		}
		if t.sendNativeHistograms {
			pendingData[nPending].Histograms = pendingData[nPending].Histograms[:0]
		}

		// Number of pending samples is limited by the fact that sendSamples (via sendSamplesWithBackoff)
		// retries endlessly, so once we reach max samples, if we can never send to the endpoint we'll
		// stop reading from the queue. This makes it safe to reference pendingSamples by index.
		pendingData[nPending].Labels = labelsToLabelsProto(d.seriesLabels, pendingData[nPending].Labels)
		switch d.sType {
		case tSample:
			pendingData[nPending].Samples = append(pendingData[nPending].Samples, prompb.Sample{
				Value:     d.value,
				Timestamp: d.timestamp,
			})
			nPendingSamples++
		case tExemplar:
			pendingData[nPending].Exemplars = append(pendingData[nPending].Exemplars, prompb.Exemplar{
				Labels:    labelsToLabelsProto(d.exemplarLabels, nil),
				Value:     d.value,
				Timestamp: d.timestamp,
			})
			nPendingExemplars++
		case tHistogram:
			pendingData[nPending].Histograms = append(pendingData[nPending].Histograms, remote.HistogramToHistogramProto(d.timestamp, d.histogram))
			nPendingHistograms++
		case tFloatHistogram:
			pendingData[nPending].Histograms = append(pendingData[nPending].Histograms, remote.FloatHistogramToHistogramProto(d.timestamp, d.floatHistogram))
			nPendingHistograms++
		}
	}
	return nPendingSamples, nPendingExemplars, nPendingHistograms
}

// Append queues a sample to be sent to the remote storage. Blocks until all samples are
// sent or fail.
func (t *QueueManager) Append(ctx context.Context, samples []prometheus.Sample) (bool, error) {
	pendingData := make([]timeSeries, len(samples))
	for x, k := range samples {
		pendingData[x] = timeSeries{
			seriesLabels: k.L,
			timestamp:    k.Timestamp,
			value:        k.Value,
			sType:        tSample,
		}
	}
	return t.append(ctx, pendingData)
}

func (t *QueueManager) AppendExemplars(ctx context.Context, exemplars []prometheus.Exemplar) (bool, error) {
	if !t.sendExemplars {
		return true, nil
	}
	pendingData := make([]timeSeries, len(exemplars))
	for x, k := range exemplars {
		pendingData[x] = timeSeries{
			seriesLabels:   k.L,
			timestamp:      k.Timestamp,
			value:          k.Value,
			exemplarLabels: k.L,
			sType:          tExemplar,
		}
	}
	return t.append(ctx, pendingData)
}

func (t *QueueManager) AppendHistograms(ctx context.Context, histograms []prometheus.Histogram) (bool, error) {
	if !t.sendNativeHistograms {
		return true, nil
	}
	pendingData := make([]timeSeries, len(histograms))
	for x, k := range histograms {
		pendingData[x] = timeSeries{
			seriesLabels: k.L,
			timestamp:    k.Timestamp,
			histogram:    k.Value,
			sType:        tHistogram,
		}
	}
	return t.append(ctx, pendingData)
}

func (t *QueueManager) AppendFloatHistograms(ctx context.Context, floatHistograms []prometheus.FloatHistogram) (bool, error) {
	if !t.sendNativeHistograms {
		return true, nil
	}
	pendingData := make([]timeSeries, len(floatHistograms))
	for x, k := range floatHistograms {
		pendingData[x] = timeSeries{
			seriesLabels:   k.L,
			timestamp:      k.Timestamp,
			floatHistogram: k.Value,
			sType:          tFloatHistogram,
		}
	}
	return t.append(ctx, pendingData)
}

// Start the queue manager sending samples to the remote storage.
// Does not block.
func (t *QueueManager) Start() {
	// Register and initialise some metrics.
	t.metrics.register()
	t.metrics.shardCapacity.Set(float64(t.cfg.Capacity))
	t.metrics.maxNumShards.Set(float64(t.cfg.MaxShards))
	t.metrics.minNumShards.Set(float64(t.cfg.MinShards))
	t.metrics.desiredNumShards.Set(float64(t.cfg.MinShards))
	t.metrics.maxSamplesPerSend.Set(float64(t.cfg.MaxSamplesPerSend))
}

// Stop stops sending samples to the remote storage and waits for pending
// sends to complete.
func (t *QueueManager) Stop() {
	level.Info(t.logger).Log("msg", "Stopping remote storage...")
	defer level.Info(t.logger).Log("msg", "Remote storage stopped.")

	// Wait for all QueueManager routines to end before stopping shards, metadata watcher, and WAL watcher. This
	// is to ensure we don't end up executing a reshard and shards.stop() at the same time, which
	// causes a closed channel panic.

	t.metrics.unregister()
}

// SetClient updates the client used by a queue. Used when only client specific
// fields are updated to avoid restarting the queue.
func (t *QueueManager) SetClient(c WriteClient) {
	t.clientMtx.Lock()
	t.storeClient = c
	t.clientMtx.Unlock()
}

func (t *QueueManager) client() WriteClient {
	t.clientMtx.RLock()
	defer t.clientMtx.RUnlock()
	return t.storeClient
}

type shard struct {
	qm *QueueManager
}

type timeSeries struct {
	seriesLabels   labels.Labels
	value          float64
	histogram      *histogram.Histogram
	floatHistogram *histogram.FloatHistogram
	timestamp      int64
	exemplarLabels labels.Labels
	// The type of series: sample, exemplar, or histogram.
	sType seriesType
}

type seriesType int

const (
	tSample seriesType = iota
	tExemplar
	tHistogram
	tFloatHistogram
)

// sendSamples to the remote storage with backoff for recoverable errors.
func (s *shard) sendSamplesWithBackoff(ctx context.Context, samples []prompb.TimeSeries) (bool, error) {
	// Build the WriteRequest with no metadata.
	req, highest, err := buildWriteRequest(samples, nil)
	if err != nil {
		// Failing to build the write request is non-recoverable, since it will
		// only error if marshaling the proto to bytes fails.
		return false, err
	}

	// An anonymous function allows us to defer the completion of our per-try spans
	// without causing a memory leak, and it has the nice effect of not propagating any
	// parameters for sendSamplesWithBackoff/3.
	attemptStore := func(try int) error {
		ctx, span := otel.Tracer("").Start(ctx, "Remote Send Batch")
		defer span.End()

		span.SetAttributes(
			attribute.Int("try", try),
			attribute.String("remote_name", s.qm.storeClient.Name()),
			attribute.String("remote_url", s.qm.storeClient.Endpoint()),
		)
		begin := time.Now()
		err := s.qm.client().Store(ctx, req)
		s.qm.metrics.sentBatchDuration.Observe(time.Since(begin).Seconds())

		if err != nil {
			span.RecordError(err)
			return err
		}

		return nil
	}

	onRetry := func() {
		s.qm.metrics.retriedSamplesTotal.Add(float64(len(samples)))
	}

	err = sendWriteRequestWithBackoff(ctx, s.qm.cfg, s.qm.logger, attemptStore, onRetry)
	if errors.Is(err, context.Canceled) {
		// When there is resharding, we cancel the context for this queue, which means the data is not sent.
		// So we exit early to not update the metrics.
		return false, err
	}

	s.qm.metrics.sentBytesTotal.Add(float64(len(req)))
	s.qm.metrics.highestSentTimestamp.Set(float64(highest / 1000))

	return err == nil, err
}

func sendWriteRequestWithBackoff(ctx context.Context, cfg config.QueueConfig, l log.Logger, attempt func(int) error, onRetry func()) error {
	backoff := cfg.MinBackoff
	sleepDuration := model.Duration(0)
	try := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := attempt(try)

		if err == nil {
			return nil
		}

		// If the error is unrecoverable, we should not retry.
		var backoffErr RecoverableError
		ok := errors.As(err, &backoffErr)
		if !ok {
			return err
		}

		sleepDuration = backoff
		if backoffErr.retryAfter > 0 {
			sleepDuration = backoffErr.retryAfter
			level.Info(l).Log("msg", "Retrying after duration specified by Retry-After header", "duration", sleepDuration)
		} else if backoffErr.retryAfter < 0 {
			level.Debug(l).Log("msg", "retry-after cannot be in past, retrying using default backoff mechanism")
		}

		select {
		case <-ctx.Done():
		case <-time.After(time.Duration(sleepDuration)):
		}

		// If we make it this far, we've encountered a recoverable error and will retry.
		onRetry()
		level.Warn(l).Log("msg", "Failed to send batch, retrying", "err", err)

		backoff = sleepDuration * 2

		if backoff > cfg.MaxBackoff {
			backoff = cfg.MaxBackoff
		}

		try++
	}
}

func buildWriteRequest(samples []prompb.TimeSeries, metadata []prompb.MetricMetadata) ([]byte, int64, error) {
	var highest int64
	for _, ts := range samples {
		// At the moment we only ever append a TimeSeries with a single sample or exemplar in it.
		if len(ts.Samples) > 0 && ts.Samples[0].Timestamp > highest {
			highest = ts.Samples[0].Timestamp
		}
		if len(ts.Exemplars) > 0 && ts.Exemplars[0].Timestamp > highest {
			highest = ts.Exemplars[0].Timestamp
		}
		if len(ts.Histograms) > 0 && ts.Histograms[0].Timestamp > highest {
			highest = ts.Histograms[0].Timestamp
		}
	}

	req := &prompb.WriteRequest{
		Timeseries: samples,
		Metadata:   metadata,
	}

	pBuf := proto.NewBuffer(nil) // For convenience in tests. Not efficient.
	err := pBuf.Marshal(req)
	if err != nil {
		return nil, highest, err
	}

	compressed := snappy.Encode(nil, pBuf.Bytes())
	return compressed, highest, nil
}

// metricTypeToMetricTypeProto transforms a Prometheus metricType into prompb metricType. Since the former is a string we need to transform it to an enum.
func metricTypeToMetricTypeProto(t textparse.MetricType) prompb.MetricMetadata_MetricType {
	mt := strings.ToUpper(string(t))
	v, ok := prompb.MetricMetadata_MetricType_value[mt]
	if !ok {
		return prompb.MetricMetadata_UNKNOWN
	}

	return prompb.MetricMetadata_MetricType(v)
}

type RecoverableError struct {
	error
	retryAfter model.Duration
}

// labelsToLabelsProto transforms labels into prompb labels. The buffer slice
// will be used to avoid allocations if it is big enough to store the labels.
func labelsToLabelsProto(lbls labels.Labels, buf []prompb.Label) []prompb.Label {
	result := buf[:0]
	lbls.Range(func(l labels.Label) {
		result = append(result, prompb.Label{
			Name:  l.Name,
			Value: l.Value,
		})
	})
	return result
}
