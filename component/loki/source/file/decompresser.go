package file

// This code is copied from loki/promtail@a8d5815510bd959a6dd8c176a5d9fd9bbfc8f8b5.
// Decompressor implements the reader interface and is used to read compressed log files.
// It uses the Go stdlib's compress/* packages for decoding.

import (
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/loki/positions"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/common/model"
	"go.uber.org/atomic"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"
)

func supportedCompressedFormats() map[string]struct{} {
	return map[string]struct{}{
		"gz":  {},
		"z":   {},
		"bz2": {},
		// TODO: add support for zip.
	}
}

type decompressor struct {
	metrics   *metrics
	logger    log.Logger
	handler   loki.EntryHandler
	positions positions.Positions

	path   string
	labels string

	posAndSizeMtx sync.Mutex
	stopOnce      sync.Once

	running *atomic.Bool
	posquit chan struct{}
	posdone chan struct{}
	done    chan struct{}

	decoder *encoding.Decoder

	position int64
	size     int64
	cfg      DecompressionConfig
}

func newDecompressor(
	metrics *metrics,
	logger log.Logger,
	handler loki.EntryHandler,
	positions positions.Positions,
	path string,
	labels string,
	encodingFormat string,
	cfg DecompressionConfig,
) (*decompressor, error) {

	logger = log.With(logger, "component", "decompressor")

	pos, err := positions.Get(path, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	var decoder *encoding.Decoder
	if encodingFormat != "" {
		level.Info(logger).Log("msg", "decompressor will decode messages", "from", encodingFormat, "to", "UTF8")
		encoder, err := ianaindex.IANA.Encoding(encodingFormat)
		if err != nil {
			return nil, fmt.Errorf("failed to get IANA encoding %s: %w", encodingFormat, err)
		}
		decoder = encoder.NewDecoder()
	}

	decompressor := &decompressor{
		metrics:   metrics,
		logger:    logger,
		handler:   loki.AddLabelsMiddleware(model.LabelSet{filenameLabel: model.LabelValue(path)}).Wrap(handler),
		positions: positions,
		path:      path,
		labels:    labels,
		running:   atomic.NewBool(false),
		posquit:   make(chan struct{}),
		posdone:   make(chan struct{}),
		done:      make(chan struct{}),
		position:  pos,
		decoder:   decoder,
		cfg:       cfg,
	}

	go decompressor.readLines()
	go decompressor.updatePosition()
	metrics.filesActive.Add(1.)
	return decompressor, nil
}

// mountReader instantiate a reader ready to be used by the decompressor.
//
// The reader implementation is selected based on the given CompressionFormat.
// If the actual file format is incorrect, the reading of the header may fail and return an error - depending on the
// implementation of the underlying compression library. In any case, when a file is corrupted, the subsequent reading
// of lines will fail.
func mountReader(f *os.File, logger log.Logger, format CompressionFormat) (reader io.Reader, err error) {
	var decompressLib string

	switch format.String() {
	case "gz":
		decompressLib = "compress/gzip"
		reader, err = gzip.NewReader(f)
	case "z":
		decompressLib = "compress/zlib"
		reader, err = zlib.NewReader(f)
	case "bz2":
		decompressLib = "bzip2"
		reader = bzip2.NewReader(f)
	}

	if err != nil && err != io.EOF {
		return nil, err
	}

	if reader == nil {
		supportedFormatsList := strings.Builder{}
		for format := range supportedCompressedFormats() {
			supportedFormatsList.WriteString(format)
		}
		return nil, fmt.Errorf("file %q has unsupported format, it has to be one of %q", f.Name(), supportedFormatsList.String())
	}

	level.Debug(logger).Log("msg", fmt.Sprintf("using %q to decompress file %q", decompressLib, f.Name()))
	return reader, nil
}

func (d *decompressor) updatePosition() {
	positionSyncPeriod := d.positions.SyncPeriod()
	positionWait := time.NewTicker(positionSyncPeriod)
	defer func() {
		positionWait.Stop()
		level.Info(d.logger).Log("msg", "position timer: exited", "path", d.path)
		close(d.posdone)
	}()

	for {
		select {
		case <-positionWait.C:
			if err := d.MarkPositionAndSize(); err != nil {
				level.Error(d.logger).Log("msg", "position timer: error getting position and/or size, stopping decompressor", "path", d.path, "error", err)
				return
			}
		case <-d.posquit:
			return
		}
	}
}

// readLines read all existing lines of the given compressed file.
//
// It first decompresses the file as a whole using a reader and then it will iterate
// over its chunks, separated by '\n'.
// During each iteration, the parsed and decoded log line is then sent to the API with the current timestamp.
func (d *decompressor) readLines() {
	level.Info(d.logger).Log("msg", "read lines routine: started", "path", d.path)
	d.running.Store(true)

	if d.cfg.InitialDelay > 0 {
		level.Info(d.logger).Log("msg", "sleeping before starting decompression", "path", d.path, "duration", d.cfg.InitialDelay.String())
		time.Sleep(d.cfg.InitialDelay)
	}

	defer func() {
		d.cleanupMetrics()
		level.Info(d.logger).Log("msg", "read lines routine finished", "path", d.path)
		close(d.done)
	}()
	entries := d.handler.Chan()

	f, err := os.Open(d.path)
	if err != nil {
		level.Error(d.logger).Log("msg", "error reading file", "path", d.path, "error", err)
		return
	}
	defer f.Close()

	r, err := mountReader(f, d.logger, d.cfg.Format)
	if err != nil {
		level.Error(d.logger).Log("msg", "error mounting new reader", "err", err)
		return
	}

	level.Info(d.logger).Log("msg", "successfully mounted reader", "path", d.path, "ext", filepath.Ext(d.path))

	bufferSize := 4096
	buffer := make([]byte, bufferSize)
	maxLoglineSize := 2000000 // 2 MB
	scanner := bufio.NewScanner(r)
	scanner.Buffer(buffer, maxLoglineSize)
	for line := 1; ; line++ {
		if !scanner.Scan() {
			break
		}

		if scannerErr := scanner.Err(); scannerErr != nil {
			if scannerErr != io.EOF {
				level.Error(d.logger).Log("msg", "error scanning", "err", scannerErr)
			}

			break
		}

		if line <= int(d.position) {
			// skip already seen lines.
			continue
		}

		text := scanner.Text()
		var finalText string
		if d.decoder != nil {
			var err error
			finalText, err = d.convertToUTF8(text)
			if err != nil {
				level.Debug(d.logger).Log("msg", "failed to convert encoding", "error", err)
				d.metrics.encodingFailures.WithLabelValues(d.path).Inc()
				finalText = fmt.Sprintf("the requested encoding conversion for this line failed in Grafana Agent: %s", err.Error())
			}
		} else {
			finalText = text
		}

		d.metrics.readLines.WithLabelValues(d.path).Inc()

		entries <- loki.Entry{
			Labels: model.LabelSet{},
			Entry: logproto.Entry{
				Timestamp: time.Now(),
				Line:      finalText,
			},
		}

		d.size = int64(unsafe.Sizeof(finalText))
		d.position++
	}
}

func (d *decompressor) MarkPositionAndSize() error {
	// Lock this update as there are 2 timers calling this routine, the sync in filetarget and the positions sync in this file.
	d.posAndSizeMtx.Lock()
	defer d.posAndSizeMtx.Unlock()

	d.metrics.totalBytes.WithLabelValues(d.path).Set(float64(d.size))
	d.metrics.readBytes.WithLabelValues(d.path).Set(float64(d.position))
	d.positions.Put(d.path, d.labels, d.position)

	return nil
}

func (d *decompressor) Stop() {
	// stop can be called by two separate threads in filetarget, to avoid a panic closing channels more than once
	// we wrap the stop in a sync.Once.
	d.stopOnce.Do(func() {
		// Shut down the position marker thread
		close(d.posquit)
		<-d.posdone

		// Save the current position before shutting down reader
		if err := d.MarkPositionAndSize(); err != nil {
			level.Error(d.logger).Log("msg", "error marking file position when stopping decompressor", "path", d.path, "error", err)
		}

		// Wait for readLines() to consume all the remaining messages and exit when the channel is closed
		<-d.done
		level.Info(d.logger).Log("msg", "stopped decompressor", "path", d.path)
		d.handler.Stop()
	})
}

func (d *decompressor) IsRunning() bool {
	return d.running.Load()
}

func (d *decompressor) convertToUTF8(text string) (string, error) {
	res, _, err := transform.String(d.decoder, text)
	if err != nil {
		return "", fmt.Errorf("failed to decode text to UTF8: %w", err)
	}

	return res, nil
}

// cleanupMetrics removes all metrics exported by this reader
func (d *decompressor) cleanupMetrics() {
	// When we stop tailing the file, un-export metrics related to the file.
	d.metrics.filesActive.Add(-1.)
	d.metrics.readLines.DeleteLabelValues(d.path)
	d.metrics.readBytes.DeleteLabelValues(d.path)
	d.metrics.totalBytes.DeleteLabelValues(d.path)
}

func (d *decompressor) Path() string {
	return d.path
}
