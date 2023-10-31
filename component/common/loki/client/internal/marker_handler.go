package internal

import (
	"github.com/grafana/agent/component/common/loki/wal"
	"sort"
	"sync"
	"time"
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

type countDataItem struct {
	count      int
	lastUpdate time.Time
}

type processDataItem struct {
	segment    int
	count      int
	lastUpdate time.Time
}

// runUpdatePendingData is assumed to run in a separate routine, asynchronously keeping track of how much data each WAL
// segment the Watcher reads from, has left to send. When a segment reaches zero, it means that is has been consumed,
// and the highest numbered one is marked as consumed.
func (mh *markerHandler) runUpdatePendingData() {
	defer mh.wg.Done()

	segmentDataCount := make(map[int]countDataItem)

	for {
		select {
		case <-mh.quit:
			return
		case update := <-mh.dataIOUpdate:
			if di, ok := segmentDataCount[update.segmentId]; ok {
				di.lastUpdate = time.Now()
				di.count += update.dataCount
			} else {
				segmentDataCount[update.segmentId] = countDataItem{
					count:      update.dataCount,
					lastUpdate: time.Now(),
				}
			}
		}

		markableSegment := FindMarkableSegment(segmentDataCount, time.Hour)
		if markableSegment > mh.lastMarkedSegment {
			mh.markerFileHandler.MarkSegment(markableSegment)
			mh.lastMarkedSegment = markableSegment
		}
	}
}

func FindMarkableSegment(segmentDataCount map[int]countDataItem, tooOldThreshold time.Duration) int {
	// N = len(segmentDataCount)
	// alloc slice, N
	orderedSegmentCounts := make([]processDataItem, 0, len(segmentDataCount))

	// convert map into slice, which already has expected capacity, N
	for seg, item := range segmentDataCount {
		orderedSegmentCounts = append(orderedSegmentCounts, processDataItem{
			segment:    seg,
			count:      item.count,
			lastUpdate: item.lastUpdate,
		})
	}

	// sort orderedSegmentCounts, N log N
	sort.Slice(orderedSegmentCounts, func(i, j int) bool {
		return orderedSegmentCounts[i].segment < orderedSegmentCounts[j].segment
	})

	var lastZero = -1
	for _, item := range orderedSegmentCounts {
		// we consider a segment as "consumed if it's data count is zero, or the lastUpdate is too old
		if item.count == 0 || time.Since(item.lastUpdate) > tooOldThreshold {
			lastZero = item.segment
			// since the segment has been consumed, clear from map
			delete(segmentDataCount, item.segment)
		} else {
			// if we find a "non consumed" segment, we exit
			break
		}
	}

	return lastZero
}

func (mh *markerHandler) Stop() {
	mh.quit <- struct{}{}
	mh.wg.Wait()
}
