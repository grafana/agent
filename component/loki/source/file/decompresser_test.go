package file

// This code is copied from Promtail to test their decompressor implementation
// of the reader interface.

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component/common/loki"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

type noopClient struct {
	noopChan chan loki.Entry
	wg       sync.WaitGroup
	once     sync.Once
}

func (n *noopClient) Chan() chan<- loki.Entry {
	return n.noopChan
}

func (n *noopClient) Stop() {
	n.once.Do(func() { close(n.noopChan) })
}

func newNoopClient() *noopClient {
	c := &noopClient{noopChan: make(chan loki.Entry)}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for range c.noopChan {
			// noop
		}
	}()
	return c
}

// fakeClient is a fake client to be used for testing.
type fakeClient struct {
	entries  chan loki.Entry
	received []loki.Entry
	once     sync.Once
	mtx      sync.Mutex
	wg       sync.WaitGroup
	OnStop   func()
}

func newFakeClient(stop func()) *fakeClient {
	c := &fakeClient{
		OnStop:  stop,
		entries: make(chan loki.Entry),
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for e := range c.entries {
			c.mtx.Lock()
			c.received = append(c.received, e)
			c.mtx.Unlock()
		}
	}()
	return c
}

// Stop implements client.Client
func (c *fakeClient) Stop() {
	c.once.Do(func() { close(c.entries) })
	c.wg.Wait()
	c.OnStop()
}

func (c *fakeClient) Chan() chan<- loki.Entry {
	return c.entries
}

func (c *fakeClient) Received() []loki.Entry {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	cpy := make([]loki.Entry, len(c.received))
	copy(cpy, c.received)
	return cpy
}

// StopNow implements client.Client
func (c *fakeClient) StopNow() {
	c.Stop()
}

func (c *fakeClient) Name() string {
	return "fake"
}

// Clear is used to clean up the buffered received entries, so the same client can be re-used between
// test cases.
func (c *fakeClient) Clear() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.received = []loki.Entry{}
}

func BenchmarkReadlines(b *testing.B) {
	entryHandler := newNoopClient()

	scenarios := []struct {
		name string
		file string
	}{
		{
			name: "2000 lines of log .tar.gz compressed",
			file: "testdata/short-access.tar.gz",
		},
		{
			name: "100000 lines of log .gz compressed",
			file: "testdata/long-access.gz",
		},
	}

	for _, tc := range scenarios {
		b.Run(tc.name, func(b *testing.B) {
			decBase := &decompressor{
				logger:  log.NewNopLogger(),
				running: atomic.NewBool(false),
				handler: entryHandler,
				path:    tc.file,
			}

			for i := 0; i < b.N; i++ {
				newDec := decBase
				newDec.metrics = newMetrics(prometheus.NewRegistry())
				newDec.done = make(chan struct{})
				newDec.readLines()
				<-newDec.done
			}
		})
	}
}

func TestGigantiqueGunzipFile(t *testing.T) {
	file := "testdata/long-access.gz"
	handler := newFakeClient(func() {})
	defer handler.Stop()

	d := &decompressor{
		logger:  log.NewNopLogger(),
		running: atomic.NewBool(false),
		handler: handler,
		path:    file,
		done:    make(chan struct{}),
		metrics: newMetrics(prometheus.NewRegistry()),
	}

	d.readLines()

	<-d.done
	time.Sleep(time.Millisecond * 200)

	entries := handler.Received()
	require.Equal(t, 100000, len(entries))
}

// TestOnelineFiles test the supported formats for log lines that only contain 1 line.
//
// Based on our experience, this is the scenario with the most edge cases.
func TestOnelineFiles(t *testing.T) {
	fileContent, err := os.ReadFile("testdata/onelinelog.log")
	require.NoError(t, err)
	t.Run("gunzip file", func(t *testing.T) {
		file := "testdata/onelinelog.log.gz"
		handler := newFakeClient(func() {})
		defer handler.Stop()

		d := &decompressor{
			logger:  log.NewNopLogger(),
			running: atomic.NewBool(false),
			handler: handler,
			path:    file,
			done:    make(chan struct{}),
			metrics: newMetrics(prometheus.NewRegistry()),
		}

		d.readLines()

		<-d.done
		time.Sleep(time.Millisecond * 200)

		entries := handler.Received()
		require.Equal(t, 1, len(entries))
		require.Equal(t, string(fileContent), entries[0].Line)
	})

	t.Run("bzip2 file", func(t *testing.T) {
		file := "testdata/onelinelog.log.bz2"
		handler := newFakeClient(func() {})
		defer handler.Stop()

		d := &decompressor{
			logger:  log.NewNopLogger(),
			running: atomic.NewBool(false),
			handler: handler,
			path:    file,
			done:    make(chan struct{}),
			metrics: newMetrics(prometheus.NewRegistry()),
		}

		d.readLines()

		<-d.done
		time.Sleep(time.Millisecond * 200)

		entries := handler.Received()
		require.Equal(t, 1, len(entries))
		require.Equal(t, string(fileContent), entries[0].Line)
	})

	t.Run("tar.gz file", func(t *testing.T) {
		file := "testdata/onelinelog.tar.gz"
		handler := newFakeClient(func() {})
		defer handler.Stop()

		d := &decompressor{
			logger:  log.NewNopLogger(),
			running: atomic.NewBool(false),
			handler: handler,
			path:    file,
			done:    make(chan struct{}),
			metrics: newMetrics(prometheus.NewRegistry()),
		}

		d.readLines()

		<-d.done
		time.Sleep(time.Millisecond * 200)

		entries := handler.Received()
		require.Equal(t, 1, len(entries))
		firstEntry := entries[0]
		require.Contains(t, firstEntry.Line, "onelinelog.log") // contains .tar.gz headers
		require.Contains(t, firstEntry.Line, `5.202.214.160 - - [26/Jan/2019:19:45:25 +0330] "GET / HTTP/1.1" 200 30975 "https://www.zanbil.ir/" "Mozilla/5.0 (Windows NT 6.2; WOW64; rv:21.0) Gecko/20100101 Firefox/21.0" "-"`)
	})
}
