package memory

import (
	"bytes"
	"encoding/gob"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/component/prometheus"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/nutsdb/nutsdb/inmemory"
)

type db struct {
	mut sync.Mutex
	inm *inmemory.DB
	log *logging.Logger
}

func newDb(l *logging.Logger) *db {

	inm, _ := inmemory.Open(inmemory.DefaultOptions)
	return &db{
		inm: inm,
		log: l,
	}
}

func (d *db) getKeys(b string) []uint64 {
	keys, _ := d.inm.AllKeys(b)
	ret := make([]uint64, len(keys))
	for x, k := range keys {
		ret[x], _ = strconv.ParseUint(string(k), 10, 64)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i] < ret[j] })
	return ret
}

func (d *db) getValuesForBucketKey(b string, k uint64) [][]byte {
	items, _ := d.inm.LRange(b, strconv.FormatUint(k, 10), 0, -1)
	return items
}

func (d *db) writeRecords(bucket string, timestamp int64, data []any, ttl time.Duration) {
	d.mut.Lock()
	defer d.mut.Unlock()

	if len(data) == 0 {
		return
	}

	key := []byte(strconv.FormatInt(timestamp, 10))
	entry, err := d.inm.Get(bucket, key)
	if err != nil {
		level.Error(d.log).Log("err", err)
		return
	}
	items := make([][]byte, len(data))
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	for x, dt := range data {
		enc.Encode(dt)
		items[x] = buf.Bytes()
		buf.Reset()
	}
	// We need to set the TTL for this bucket.
	index := 0
	if entry == nil {
		index++
		err = d.inm.Put(bucket, key, items[0], uint32(ttl.Seconds()))
		if err != nil {
			level.Error(d.log).Log("err", err)
		}
	}
	for i := index; i <= len(items); i++ {
		err = d.inm.RPush(bucket, string(key), items[i])
		if err != nil {
			level.Error(d.log).Log("err", err)
		}
	}
}

func (d *db) getMetricRecords(timestamp uint64) []prometheus.Sample {
	items, err := d.inm.LRange("metrics", strconv.FormatUint(timestamp, 10), 0, -1)
	if err != nil {
		level.Error(d.log).Log("err", err)
	}
	samples := make([]prometheus.Sample, len(items))
	for x, v := range items {
		buf := bytes.NewBuffer(v)
		dec := gob.NewDecoder(buf)
		var s prometheus.Sample
		dec.Decode(s)
		samples[x] = s
	}
	return samples
}
