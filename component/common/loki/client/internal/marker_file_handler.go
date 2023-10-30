package internal

import (
	"fmt"
	"github.com/grafana/agent/component/common/loki/wal"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

const (
	MarkerFolderName = "remote"
	MarkerFileName   = "segment"
)

// MarkerFileHandler is a file-backed wal.Marker, that also allows one to write to the backing store as particular
// segment number as the last one marked.
type MarkerFileHandler interface {
	wal.Marker

	// MarkSegment writes in the backing file-store that a particular segment is the last one marked.
	MarkSegment(segment int)
}

type markerFileHandler struct {
	logger                    log.Logger
	lastMarkedSegmentDir      string
	lastMarkedSegmentFilePath string
}

var (
	_ MarkerFileHandler = (*markerFileHandler)(nil)
)

// NewMarkerFileHandler creates a new markerFileHandler.
func NewMarkerFileHandler(logger log.Logger, walDir string) (MarkerFileHandler, error) {
	markerDir := filepath.Join(walDir, MarkerFolderName)
	// attempt to create dir if doesn't exist
	if err := os.MkdirAll(markerDir, 0o777); err != nil {
		return nil, fmt.Errorf("error creating segment marker folder %q: %w", markerDir, err)
	}

	mfh := &markerFileHandler{
		logger:                    logger,
		lastMarkedSegmentDir:      filepath.Join(markerDir),
		lastMarkedSegmentFilePath: filepath.Join(markerDir, MarkerFileName),
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
	encodedMarker, err := EncodeMarkerV1(uint64(segment))
	if err != nil {
		level.Error(mfh.logger).Log("msg", "failed to encode marker when marking segment", "err", err)
		return
	}

	if err := mfh.swapMarkerFile(encodedMarker); err != nil {
		level.Error(mfh.logger).Log("msg", "could not replace segment marker file", "file", mfh.lastMarkedSegmentFilePath, "err", err)
		return
	}

	level.Debug(mfh.logger).Log("msg", "updated segment marker file", "file", mfh.lastMarkedSegmentFilePath, "segment", segment)
}

func (mfh *markerFileHandler) swapMarkerFile(bs []byte) error {
	tempFile, err := os.CreateTemp(mfh.lastMarkedSegmentDir, "segment*")
	if err != nil {
		return err
	}

	// upon exit attempt to close the temporal file handler and remove the file
	defer func() {
		if err := tempFile.Close(); err != nil {
			level.Warn(mfh.logger).Log("msg", "failed to close temporary marker file handler")
			// TODO: should we try to remove the temporal file either way
			return
		}
		if err := os.Remove(tempFile.Name()); err != nil {
			level.Warn(mfh.logger).Log("msg", "failed to remove temporary marker file")
		}
	}()

	written, err := tempFile.Write(bs)
	if err != nil {
		return fmt.Errorf("failed to write to temporal marker file: %w", err)
	} else if written != len(bs) {
		return fmt.Errorf("failed to write whole marker. written %d, expected %d", written, len(bs))
	}

	if err := os.Rename(tempFile.Name(), mfh.lastMarkedSegmentFilePath); err != nil {
		return fmt.Errorf("failed to swap marker files: %w", err)
	}

	return nil
}
