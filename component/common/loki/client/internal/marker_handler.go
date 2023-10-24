package internal

import (
	"github.com/grafana/agent/component/common/loki/wal"
)

type MarkerHandler interface {
	wal.Marker

	UpdateReceivedData(segmentId, dataCount int) // Data queued for sending
	UpdateSentData(segmentId, dataCount int)     // Data which was sent or given up on sending

	Stop()
}

type markerHandler struct {
	markerFileHandler MarkerFileHandler
	lastMarkedSegment int
	dataIOUpdate      chan data
	quit              chan struct{}
}

// TODO: Rename this struct
type data struct {
	segmentId int
	dataCount int
}

var (
	_ MarkerHandler = (*markerHandler)(nil)
)

func NewMarkerHandler(mfh MarkerFileHandler) MarkerHandler {
	mh := &markerHandler{
		lastMarkedSegment: -1, // Segment ID last marked on disk.
		markerFileHandler: mfh,
		//TODO: What is a good size for the channel?
		dataIOUpdate: make(chan data, 100),
		quit:         make(chan struct{}),
	}

	// Load the last marked segment from disk (if it exists).
	if lastSegment := mh.markerFileHandler.LastMarkedSegment(); lastSegment >= 0 {
		mh.lastMarkedSegment = lastSegment
	}

	//TODO: Should this be in a separate Start() function?
	go mh.updatePendingData()

	return mh
}

func (mh *markerHandler) LastMarkedSegment() int {
	return mh.markerFileHandler.LastMarkedSegment()
}

func (mh *markerHandler) UpdateReceivedData(segmentId, dataCount int) {
	mh.dataIOUpdate <- data{
		segmentId: segmentId,
		dataCount: dataCount,
	}
}

func (mh *markerHandler) UpdateSentData(segmentId, dataCount int) {
	mh.dataIOUpdate <- data{
		segmentId: segmentId,
		dataCount: -1 * dataCount,
	}
}

func (mh *markerHandler) Stop() {
	// Firstly stop the Marker Handler, because it might want to use the Marker File Handler.
	mh.quit <- struct{}{}
}

// updatePendingData updates a counter for how much data is yet to be sent from each segment.
// "dataCount" will be added to the segment with ID "dataSegment".
func (mh *markerHandler) updatePendingData() {
	batchSegmentCount := make(map[int]int)

	for {
		select {
		case <-mh.quit:
			return
		case dataUpdate := <-mh.dataIOUpdate:
			batchSegmentCount[dataUpdate.segmentId] += dataUpdate.dataCount
		}

		markableSegment := -1
		for segment, count := range batchSegmentCount {
			// TODO: If count is less than 0, then log an error and remove the entry from the map?
			if count != 0 {
				continue
			}

			// we know (segment, 0) is in the map

			// TODO: Is it safe to assume that just because a segment is 0 inside the map,
			//      all samples from it have been processed?
			if segment > markableSegment {
				markableSegment = segment
			}

			// Clean up the pending map: the current segment has been completely
			// consumed and doesn't need to be considered for marking again.
			delete(batchSegmentCount, segment)
		}

		// NOTE: I think this could lead to a situation where some segments are skipped. For example,
		// consider the following situation
		// seg   0   1   2   3   4
		// cnt   0   0   1   0   10
		// and lastMarkedSegment being 0, then a run in here would cause the lastMarkedSegment
		// to be 3. while there's data in 2. But if there's data in 2, that means that there was a count
		// error somewhere right?
		if markableSegment > mh.lastMarkedSegment {
			mh.markerFileHandler.MarkSegment(markableSegment)
			mh.lastMarkedSegment = markableSegment
		}
	}
}
