package agentctl

import (
	"github.com/prometheus/prometheus/tsdb/record"
	"github.com/prometheus/prometheus/tsdb/wal"
)

// walIterate iterates over the latest checkpoint in the provided WAL and all
// of the segments in the WAL and calls f for each of them.
func walIterate(w *wal.WAL, f func(r *wal.Reader) error) error {
	checkpoint, checkpointIdx, err := wal.LastCheckpoint(w.Dir())
	if err != nil && err != record.ErrNotFound {
		return err
	}

	startIdx, last, err := wal.Segments(w.Dir())
	if err != nil {
		return err
	}

	if checkpoint != "" {
		sr, err := wal.NewSegmentsReader(checkpoint)
		if err != nil {
			return err
		}
		err = f(wal.NewReader(sr))
		_ = sr.Close()
		if err != nil {
			return err
		}

		startIdx = checkpointIdx + 1
	}

	for i := startIdx; i <= last; i++ {
		s, err := wal.OpenReadSegment(wal.SegmentName(w.Dir(), i))
		if err != nil {
			return err
		}
		sr := wal.NewSegmentBufReader(s)
		err = f(wal.NewReader(sr))
		_ = sr.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
