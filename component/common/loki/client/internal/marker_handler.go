package internal

import (
	"github.com/grafana/agent/component/common/loki/wal"
	"sync"
)

type MarkerHandler interface {
	wal.Marker

	// UpdateReceivedData sends an update event to the handler, that informs that some dataUpdate, coming from a particular WAL
	// segment, has been read out of the WAL and enqueued for sending.
	UpdateReceivedData(segmentId, dataCount int)

	// UpdateSentData sends an update event to the handler, informing that some dataUpdate, coming from a particular WAL
	// segment, has been delivered, or the sender has given up on it.
	UpdateSentData(segmentId, dataCount int) // Data which was sent or given up on sending

	// Stop stops the handler, and it's async processing of receive/send dataUpdate updates.
	Stop()
}

// markerHandler implements MarkerHandler, processing data update events in an asynchronous manner, and tracking the last
// consumed segment in a file.
type markerHandler struct {
	markerFileHandler MarkerFileHandler
	lastMarkedSegment int
	dataIOUpdate      chan dataUpdate
	quit              chan struct{}
	wg                sync.WaitGroup
}

// dataUpdate is an update event that some amount of data has been read out of the WAL and enqueued, delivered or dropped.
type dataUpdate struct {
	segmentId int
	dataCount int
}

var (
	_ MarkerHandler = (*markerHandler)(nil)
)

// NewMarkerHandler creates a new markerHandler.
func NewMarkerHandler(mfh MarkerFileHandler) MarkerHandler {
	mh := &markerHandler{
		lastMarkedSegment: -1, // Segment ID last marked on disk.
		markerFileHandler: mfh,
		//TODO: What is a good size for the channel?
		dataIOUpdate: make(chan dataUpdate, 100),
		quit:         make(chan struct{}),
	}

	// Load the last marked segment from disk (if it exists).
	if lastSegment := mh.markerFileHandler.LastMarkedSegment(); lastSegment >= 0 {
		mh.lastMarkedSegment = lastSegment
	}

	mh.wg.Add(1)
	go mh.runUpdatePendingData()

	return mh
}

func (mh *markerHandler) LastMarkedSegment() int {
	return mh.markerFileHandler.LastMarkedSegment()
}

func (mh *markerHandler) UpdateReceivedData(segmentId, dataCount int) {
	mh.dataIOUpdate <- dataUpdate{
		segmentId: segmentId,
		dataCount: dataCount,
	}
}

func (mh *markerHandler) UpdateSentData(segmentId, dataCount int) {
	mh.dataIOUpdate <- dataUpdate{
		segmentId: segmentId,
		dataCount: -1 * dataCount,
	}
}

// runUpdatePendingData is assumed to run in a separate routine, asynchronously keeping track of how much data each WAL
// segment the Watcher reads from, has left to send. When a segment reaches zero, it means that is has been consumed,
// and the highest numbered one is marked as consumed.
func (mh *markerHandler) runUpdatePendingData() {
	defer mh.wg.Done()

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

func (mh *markerHandler) Stop() {
	mh.quit <- struct{}{}
	mh.wg.Wait()
}
