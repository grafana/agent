package linear

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/tidwall/btree"
)

// linear is used as a format to serialize and deserialize metrics. It is NOT thread safe.
type linear struct {
	estimatedSize int
	index         int
	dict          map[string]int
	reverseDict   map[int]string
	tb            []byte
	tb64          []byte
	stringbuffer  []byte
	totalMetrics  int

	timestamps          map[int64][]*prepocessedmetric
	preprocessedMetrics map[string][]*prepocessedmetric
	// I found this created less allocations than a map.
	metricNameLabels *btree.Map[string, *btree.Set[int]]
}

type prepocessedmetric struct {
	ts     int64
	name   string
	keys   []int
	values []int
	val    float64
}

// none_index is used to represent a none value in the label dictionary.
const none_index = 0

// LinearPool is used to retrieve a linear object to use.
// Linear objects should be Reset and put back into the pool when done.
var LinearPool = sync.Pool{
	New: func() any {
		return newLinear()
	},
}

var metricPool = sync.Pool{
	New: func() any {
		return &prepocessedmetric{
			ts:     0,
			val:    0,
			keys:   make([]int, 0),
			values: make([]int, 0),
		}
	},
}

var deserializeMetrics = sync.Pool{
	New: func() any {
		return &deserializedMetric{
			ts:   0,
			val:  0,
			lbls: labels.EmptyLabels(),
		}
	},
}

func newLinear() *linear {
	return &linear{
		dict:                make(map[string]int),
		preprocessedMetrics: make(map[string][]*prepocessedmetric),
		timestamps:          make(map[int64][]*prepocessedmetric),
		reverseDict:         make(map[int]string),
		// using btree since maps are one of the largest users of allocations.
		metricNameLabels: &btree.Map[string, *btree.Set[int]]{},
		// index 0 is reserved for <NIL> label value.
		index:        1,
		tb:           make([]byte, 4),
		tb64:         make([]byte, 8),
		stringbuffer: make([]byte, 0),
	}
}

// Reset is used when reseting the linear before adding back to the pool.
func (l *linear) Reset() {
	clear(l.dict)
	for _, x := range l.preprocessedMetrics {
		for _, m := range x {
			m.values = m.values[:0]
			m.keys = m.keys[:0]
			m.ts = 0
			m.val = 0
			metricPool.Put(m)
		}
	}
	clear(l.preprocessedMetrics)
	l.metricNameLabels.Clear()
	clear(l.timestamps)
	clear(l.reverseDict)
	l.index = 1
	l.totalMetrics = 0
	l.estimatedSize = 0
}

// AddMetric is used to add a metric to the internal metrics for use with serialization.
func (l *linear) AddMetric(lbls labels.Labels, ts int64, val float64) {
	pm := metricPool.Get().(*prepocessedmetric)
	pm.ts = ts
	pm.val = val

	// Find the name and setup variables.
	// Likely we can simplify this down to just a single loop.
	for _, ll := range lbls {
		if ll.Name == "__name__" {
			pm.name = ll.Value
			if _, found := l.metricNameLabels.Get(pm.name); !found {
				l.metricNameLabels.Set(pm.name, &btree.Set[int]{})
			}
			break
		}
	}

	// Reset the lengths of the values and keys. Since they are reused.
	if cap(pm.values) < len(lbls) {
		pm.values = make([]int, len(lbls))
		pm.keys = make([]int, len(lbls))
	} else {
		pm.values = pm.values[:len(lbls)]
		pm.keys = pm.keys[:len(lbls)]
	}

	// Add all the labels.
	for x, ll := range lbls {
		nameid := l.addOrGetID(ll.Name)
		pm.values[x] = l.addOrGetID(ll.Value)
		pm.keys[x] = nameid
		item, _ := l.metricNameLabels.Get(pm.name)
		item.Insert(nameid)
	}

	// Need to create the parent metric root to hold the metrics underneath.
	if _, found := l.preprocessedMetrics[pm.name]; !found {
		l.preprocessedMetrics[pm.name] = make([]*prepocessedmetric, 0)
	}
	l.preprocessedMetrics[pm.name] = append(l.preprocessedMetrics[pm.name], pm)

	// Go ahead and add a timestamp record.
	_, found := l.timestamps[ts]
	if !found {
		l.timestamps[ts] = make([]*prepocessedmetric, 0)
	}
	l.timestamps[ts] = append(l.timestamps[ts], pm)
	l.totalMetrics++
	// 32 bytes is quick napkin overhead for a metric.
	l.estimatedSize = l.estimatedSize + 32
}

func (l *linear) AddHistogram(lbls labels.Labels, h *histogram.Histogram) {
	panic("AddHistogram is not implemented yet.")
}

func (l *linear) Serialize(bb *bytes.Buffer) {
	// Write version header.
	l.addUint(bb, 1)

	// Write the timestamp
	l.addInt(bb, time.Now().Unix())

	// Write the string dictionary
	l.addUint(bb, uint32(len(l.dict)))

	// Index 0 is implicitly <NONE>
	for i := 1; i <= len(l.dict); i++ {
		// Write the string length
		l.addUint(bb, uint32(len(l.reverseDict[i])))
		// Write the string
		bb.WriteString(l.reverseDict[i])
	}

	l.addUint(bb, uint32(len(l.timestamps)))
	values := make([]int, 0)
	for ts, metrics := range l.timestamps {
		metricFamilyLabels := make([]int, 0)
		labelSet, _ := l.metricNameLabels.Get(metrics[0].name)
		labelSet.Scan(func(k int) bool {
			metricFamilyLabels = append(metricFamilyLabels, k)
			return true
		})

		sort.Ints(metricFamilyLabels)
		// Add the timestamp.
		l.addInt(bb, ts)
		// Add the number of metrics.
		l.addUint(bb, uint32(len(metrics)))
		// Add the number of labels.
		l.addUint(bb, uint32(len(metricFamilyLabels)))
		//Add label name ids.
		for i := 0; i < len(metricFamilyLabels); i++ {
			l.addUint(bb, uint32(metricFamilyLabels[i]))
		}
		// Add metrics.
		for _, series := range metrics {
			values = l.alignAndEncodeLabel(metricFamilyLabels, series.keys, series.values, values)
			for _, b := range values {
				// Add each value, none values will be inserted with a 0.
				// Since each series will have the same number of labels in the same order, we only need the values
				// from the value dictionary.
				l.addUint(bb, uint32(b))
			}
			// Add the value.
			l.addUInt64(bb, math.Float64bits(series.val))
		}
	}
}

// Deserialize takes an input buffer and converts to an array of deserializemetrics. These metrics
// should be ReleaseDeserializeMetrics and returned to the pool for resue.
func (l *linear) Deserialize(bb *bytes.Buffer, maxAgeSeconds int) ([]*deserializedMetric, error) {
	version := l.readUint(bb)
	if version != 1 {
		return nil, fmt.Errorf("unexpected version found %d while deserializing", version)
	}
	// Get the timestamp
	timestamp := l.readInt(bb)
	if time.Now().Unix()-timestamp > int64(maxAgeSeconds) {
		return nil, fmt.Errorf("wal timestamp %d is older than max age %d seconds", timestamp, maxAgeSeconds)
	}
	// Get length of the dictionary
	total := int(l.readUint(bb))
	// The plus one accounts for the none dictionary.
	dict := make([]string, total+1)
	for i := 1; i <= total; i++ {
		dict[i] = l.readString(bb)
	}
	timestampLength := l.readUint(bb)
	metrics := make([]*deserializedMetric, 0)
	for i := 0; i < int(timestampLength); i++ {
		ts := l.readInt(bb)
		metricCount := l.readUint(bb)
		metricLabelCount := l.readUint(bb)
		labelNames := make([]string, metricLabelCount)
		for j := 0; j < int(metricLabelCount); j++ {
			id := l.readUint(bb)
			name := dict[id]
			labelNames[j] = name
		}
		for j := 0; j < int(metricCount); j++ {
			dm := l.deserializeMetric(ts, bb, labelNames, metricLabelCount, dict)
			metrics = append(metrics, dm)
		}
	}
	return metrics, nil
}

// ReleaseDeserializeMetrics is used to return any deserialized metrics to the pool.
func ReleaseDeserializeMetrics(m []*deserializedMetric) {
	for _, x := range m {
		x.lbls = x.lbls[:0]
		x.ts = 0
		x.val = 0
		deserializeMetrics.Put(x)
	}
}

func (l *linear) deserializeMetric(ts int64, bb *bytes.Buffer, names []string, lblCount uint32, dict []string) *deserializedMetric {
	dm := deserializeMetrics.Get().(*deserializedMetric)
	for i := 0; i < int(lblCount); i++ {
		id := l.readUint(bb)
		// Label is none value.
		if id == 0 {
			continue
		}
		val := dict[id]
		dm.lbls = append(dm.lbls, labels.Label{
			Name:  names[i],
			Value: val,
		})
	}
	dm.ts = ts
	dm.val = math.Float64frombits(l.readUint64(bb))
	return dm
}

type deserializedMetric struct {
	ts   int64
	val  float64
	lbls labels.Labels
}

// alignAndEncodeLabel has a lot of magic that happens. It aligns all the values of a labels for a metric to be the same across all metrics
// currently contained. Then it returns the id that each value is stored in. This means that if you have two series in the same metric family.
// test{instance="dev"} 1 and test{app="d",instance="dev",service="auth"} 2
// This will sort the labels into app,instance,service ordering. For the first series it will return
// [0,1,0] if 1 = dev, the 0 represents the none value and since it only has instance.
// the second will return
// [2,1,3]
func (l *linear) alignAndEncodeLabel(total []int, keys []int, values []int, labelRef []int) []int {
	if cap(labelRef) < len(total) {
		labelRef = make([]int, len(total))
	} else {
		labelRef = labelRef[:len(total)]
	}
	for i, s := range total {
		id := none_index
		for x, k := range keys {
			if k == s {
				id = values[x]
				break
			}
		}
		labelRef[i] = id
	}
	return labelRef
}

func (l *linear) readUint(bb *bytes.Buffer) uint32 {
	_, _ = bb.Read(l.tb)
	return binary.LittleEndian.Uint32(l.tb)
}

func (l *linear) readUint64(bb *bytes.Buffer) uint64 {
	_, _ = bb.Read(l.tb64)
	return binary.LittleEndian.Uint64(l.tb64)
}

func (l *linear) readInt(bb *bytes.Buffer) int64 {
	_, _ = bb.Read(l.tb64)
	return int64(binary.LittleEndian.Uint64(l.tb64))
}

// readString reads a string from the buffer.
func (l *linear) readString(bb *bytes.Buffer) string {
	length := l.readUint(bb)
	if cap(l.stringbuffer) < int(length) {
		l.stringbuffer = make([]byte, length)
	} else {
		l.stringbuffer = l.stringbuffer[:int(length)]
	}
	_, _ = bb.Read(l.stringbuffer)
	return string(l.stringbuffer)
}

func (l *linear) addUint(bb *bytes.Buffer, num uint32) {
	binary.LittleEndian.PutUint32(l.tb, num)
	bb.Write(l.tb)
}

func (l *linear) addInt(bb *bytes.Buffer, num int64) {
	binary.LittleEndian.PutUint64(l.tb64, uint64(num))
	bb.Write(l.tb64)
}

func (l *linear) addUInt64(bb *bytes.Buffer, num uint64) {
	binary.LittleEndian.PutUint64(l.tb64, num)
	bb.Write(l.tb64)
}

// addOrGetID adds the string to the dictionary and returns the id.
// It will also add to the estimated size.
func (l *linear) addOrGetID(name string) int {
	id, found := l.dict[name]
	if !found {
		l.dict[name] = l.index
		l.reverseDict[l.index] = name
		id = l.index
		l.index = l.index + 1
	}
	// Add 2 bytes for the length and then the length of the string itself in bytes.
	l.estimatedSize = l.estimatedSize + 4 + len(name)
	return id
}
