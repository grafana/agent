package internal

import (
	"github.com/grafana/agent/component/common/loki/wal"
	"sort"
	"sync"
	"time"
)

const (
	minimumMarkLimiter = time.Second
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

	maxSegmentAge time.Duration
	// markupLimiter will limit the number of times the Mark routine is executed, cleaning up and marking consumed segments.
	// We need to limit this since the algo takes N log N in a potentially unbounded number
	markupLimiter *time.Ticker
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
func NewMarkerHandler(mfh MarkerFileHandler, maxSegmentAge time.Duration) MarkerHandler {
	// By taking 2% as the limiter amount, we will execute the find last marked segment
	// every 2% of the maximum segment agent, which for time.Hour, is ~1min. This will allow
	// just a minute slip between the time the marker should have been advanced, and the time it actually is.
	limiterValue := maxSegmentAge / 50

	// protect against a zero value
	if limiterValue < minimumMarkLimiter {
		limiterValue = minimumMarkLimiter
	}

	mh := &markerHandler{
		lastMarkedSegment: -1, // Segment ID last marked on disk.
		markerFileHandler: mfh,
		//TODO: What is a good size for the channel?
		dataIOUpdate: make(chan dataUpdate, 100),
		quit:         make(chan struct{}),

		maxSegmentAge: maxSegmentAge,
		markupLimiter: time.NewTicker(limiterValue),
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

		// if limiter has fired, it means we can run the markup routine. If not go back to select block above
		select {
		case <-mh.markupLimiter.C:
			markableSegment := FindMarkableSegment(segmentDataCount, mh.maxSegmentAge)
			if markableSegment > mh.lastMarkedSegment {
				mh.markerFileHandler.MarkSegment(markableSegment)
				mh.lastMarkedSegment = markableSegment
			}
		default:
		}
	}
}

func (mh *markerHandler) Stop() {
	mh.markupLimiter.Stop()
	mh.quit <- struct{}{}
	mh.wg.Wait()
}

// FindMarkableSegment finds, given the summary of data updates received, and a threshold on how much time can pass for
// a segment that hasn't received updates to be considered as "live", the segment that should be marked as last consumed.
// The algorithm will find the highest numbered segment that is considered as "consumed", with its all predecessors
// "consumed" as well.
//
// A consumed segment is one with data count of zero, meaning that there's no data left in flight for it, or it hasn't
// received any updates for tooOldThreshold time.
//
// Also, while reviewing the data items in segmentDataCount, those who are consumed will be deleted to clean up space.
//
// This algorithm runs in O(N log N), being N the size of segmentDataCount, and allocates O(N) memory.
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
