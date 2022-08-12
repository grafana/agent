package metrics

import (
	"math"
	"testing"
	"time"

	"github.com/prometheus/prometheus/model/labels"
	promrelabel "github.com/prometheus/prometheus/model/relabel"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

func TestRelabel(t *testing.T) {
	fm := NewFlowMetric(0, labels.FromStrings("key", "value"), 0)
	require.True(t, fm.globalRefID != 0)
	rg, _ := promrelabel.NewRegexp("(.*)")
	newfm := fm.Relabel(&promrelabel.Config{
		Replacement: "${1}_new",
		Action:      "replace",
		TargetLabel: "new",
		Regex:       rg,
	})
	require.Len(t, fm.labels, 1)
	require.True(t, fm.labels.Has("key"))

	require.Len(t, newfm.labels, 2)
	require.True(t, newfm.labels.Has("new"))
}

func TestRelabelTheSame(t *testing.T) {
	fm := NewFlowMetric(0, labels.FromStrings("key", "value"), 0)
	require.True(t, fm.globalRefID != 0)
	rg, _ := promrelabel.NewRegexp("bad")
	newfm := fm.Relabel(&promrelabel.Config{
		Replacement: "${1}_new",
		Action:      "replace",
		TargetLabel: "new",
		Regex:       rg,
	})
	require.Len(t, fm.labels, 1)
	require.True(t, fm.labels.Has("key"))
	require.Len(t, newfm.labels, 1)
	require.True(t, newfm.globalRefID == fm.globalRefID)
	require.True(t, labels.Equal(newfm.labels, fm.labels))
}

func BenchmarkWorkerReceiver(b *testing.B) {
	useWorkers = true
	for i := 0; i < b.N; i++ {
		runTest()
	}
}

func BenchmarkFuncReceiver(b *testing.B) {
	useWorkers = false
	for i := 0; i < b.N; i++ {
		runTest()
	}
}

func runTest() {
	parent := &receiverChain{
		parent:   nil,
		children: make([]*Receiver, 0),
	}
	cnt := atomic.NewInt64(0)
	finalRec := NewReceiver(func(timestamp int64, metrics []*FlowMetric) {
		cnt.Inc()
		time.Sleep(1 * time.Millisecond)
	})
	maxdepth := 4
	childrencount := 2
	expectedCount := math.Pow(float64(childrencount), float64(maxdepth))
	childs := buildReceiverPyramid(1, maxdepth, childrencount, parent, finalRec)
	for _, c := range childs {
		nr := NewReceiver(c.Receive)
		parent.children = append(parent.children, nr)
	}
	parent.Receive(time.Now().Unix(), nil)
	for {
		if cnt.Load() == int64(expectedCount) {
			break
		}
		time.Sleep(1 * time.Microsecond)
	}
}

func buildReceiverPyramid(
	currentDepth, maxDepth, childrenCount int,
	parent *receiverChain,
	finalRec *Receiver,
) []*receiverChain {

	children := make([]*receiverChain, 0)
	for i := 0; i < childrenCount; i++ {
		leaf := &receiverChain{
			parent:   parent,
			children: make([]*Receiver, 0),
		}
		children = append(children, leaf)
		if currentDepth == maxDepth {
			leaf.children = append(leaf.children, finalRec)
			continue
		} else {
			newDepth := currentDepth + 1
			leafChildren := buildReceiverPyramid(newDepth, maxDepth, childrenCount, leaf, finalRec)
			for _, lc := range leafChildren {
				rc := NewReceiver(lc.Receive)
				leaf.children = append(leaf.children, rc)
			}
		}
	}
	return children
}

type receiverChain struct {
	parent   *receiverChain
	children []*Receiver
}

func (rc *receiverChain) Receive(timestamp int64, metricArr []*FlowMetric) {
	for _, f := range rc.children {
		f.Send(timestamp, metricArr)
	}
}
