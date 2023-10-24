package internal

import (
	"fmt"
	"github.com/grafana/agent/component/common/loki/wal"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/prometheus/tsdb/fileutil"
)

type MarkerFileHandler interface {
	wal.Marker
	MarkSegment(segment int)
}

type markerFileHandler struct {
	logger log.Logger

	lastMarkedSegmentFilePath string
}

var (
	_ MarkerFileHandler = (*markerFileHandler)(nil)
)

func NewMarkerFileHandler(logger log.Logger, walDir string) (MarkerFileHandler, error) {
	markerDir := filepath.Join(walDir, "remote")
	// attempt to create dir if doesn't exist
	if err := os.MkdirAll(markerDir, 0o777); err != nil {
		return nil, fmt.Errorf("error creating segment marker folder %q: %w", markerDir, err)
	}

	mfh := &markerFileHandler{
		logger:                    logger,
		lastMarkedSegmentFilePath: filepath.Join(markerDir, "segment"),
	}

	return mfh, nil
}

// LastMarkedSegment implements wlog.Marker.
func (mfh *markerFileHandler) LastMarkedSegment() int {
	bb, err := os.ReadFile(mfh.lastMarkedSegmentFilePath)
	if os.IsNotExist(err) {
		level.Warn(mfh.logger).Log("msg", "marker segment file does not exist", "file", mfh.lastMarkedSegmentFilePath)
		return -1
	} else if err != nil {
		level.Error(mfh.logger).Log("msg", "could not access segment marker file", "file", mfh.lastMarkedSegmentFilePath, "err", err)
		return -1
	}

	savedSegment, err := strconv.Atoi(string(bb))
	if err != nil {
		level.Error(mfh.logger).Log("msg", "could not read segment marker file", "file", mfh.lastMarkedSegmentFilePath, "err", err)
		return -1
	}

	if savedSegment < 0 {
		level.Error(mfh.logger).Log("msg", "invalid segment number inside marker file", "file", mfh.lastMarkedSegmentFilePath, "segment number", savedSegment)
		return -1
	}

	return savedSegment
}

// MarkSegment implements MarkerHandler.
func (mfh *markerFileHandler) MarkSegment(segment int) {
	var (
		segmentText = strconv.Itoa(segment)
		tmp         = mfh.lastMarkedSegmentFilePath + ".tmp"
	)

	if err := os.WriteFile(tmp, []byte(segmentText), 0o666); err != nil {
		level.Error(mfh.logger).Log("msg", "could not create segment marker file", "file", tmp, "err", err)
		return
	}
	if err := fileutil.Replace(tmp, mfh.lastMarkedSegmentFilePath); err != nil {
		level.Error(mfh.logger).Log("msg", "could not replace segment marker file", "file", mfh.lastMarkedSegmentFilePath, "err", err)
		return
	}

	level.Debug(mfh.logger).Log("msg", "updated segment marker file", "file", mfh.lastMarkedSegmentFilePath, "segment", segment)
}
