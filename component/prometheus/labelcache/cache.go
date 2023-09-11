package labelcache

import (
	"arena"
	"bytes"
	"encoding/binary"
	"path"
	"sync"
	"time"
	"unsafe"

	"github.com/cockroachdb/pebble"
	"github.com/go-kit/log"
	"github.com/prometheus/prometheus/model/labels"
)

type Cache struct {
	mut                sync.RWMutex
	db                 *pebble.DB
	remoteWriteMapping *bucket
	labelToID          *bucket
	idToLabel          *bucket
	ttlToID            *bucket
	// The key is componentID + localID
	localIDtoID *bucket
	// The key is componentID + globalID
	idToLocalID *bucket

	l log.Logger

	idMut     sync.Mutex
	currentID uint64
}

func NewCache(directory string, l log.Logger) *Cache {
	db, _ := pebble.Open(path.Join(directory, "labels"), &pebble.Options{})
	c := &Cache{
		db:                 db,
		remoteWriteMapping: newBucket(db, 1, "remote write mapping", l),
		labelToID:          newBucket(db, 2, "label to id", l),
		idToLabel:          newBucket(db, 3, "id to label", l),
		ttlToID:            newBucket(db, 4, "ttl to id", l),
		// Prometheus remote write mapping
		localIDtoID: newBucket(db, 5, "component + local id to global id", l),
		idToLocalID: newBucket(db, 6, "component + global id to local id", l),
		l:           l,
	}
	c.currentID = c.getCurrentID()
	return c
}

// WriteLabels inserts/updates labels. The ttl is a minimum timeframe before deletion. If a longer ttl exists.
// It will be honored.
func (c *Cache) WriteLabels(lbls [][]labels.Label, ttl time.Duration, mem *arena.Arena) ([]uint64, error) {
	c.mut.Lock()
	defer c.mut.Unlock()

	// TODO think about if merge solves race condition.
	// Also we can do many concurrent operations has long has they dont share labels/keys.
	ts := time.Now().Unix() + int64(ttl.Seconds())
	lblBuf := makeLabelBytes(lbls, mem)
	keys, err := c.labelToID.getValues(lblBuf, mem)
	if err != nil {
		return nil, err
	}
	keys, err = c.writeNotFoundKeys(keys, lblBuf, mem)
	buf := bytes.NewBuffer(nil)
	pushUint64ByteSlice(keys, buf, mem)
	ttlBuf := make([]byte, 8)
	binary.PutVarint(ttlBuf, ts)
	err = c.ttlToID.writeValues([][]byte{ttlBuf}, [][]byte{buf.Bytes()}, mem)
	if err != nil {
		return nil, err
	}
	return makeKeys(keys, mem), nil
}

// GetIDs will retrieve labels and create any that dont exist.
func (c *Cache) GetIDs(lbls [][]labels.Label, mem *arena.Arena) ([]uint64, error) {
	c.mut.Lock()
	defer c.mut.Unlock()

	lblBuf := makeLabelBytes(lbls, mem)
	keyBytes, err := c.labelToID.getValues(lblBuf, mem)
	if err != nil {
		return nil, err
	}
	keyBytes, err = c.writeNotFoundKeys(keyBytes, lblBuf, mem)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	pushUint64ByteSlice(keyBytes, buf, mem)
	// Set a TTL for 1 hour.
	ts := time.Now().Unix() + int64(time.Hour.Seconds())
	ttlBuf := make([]byte, 8)
	binary.PutVarint(ttlBuf, ts)
	err = c.ttlToID.writeValues([][]byte{ttlBuf}, [][]byte{buf.Bytes()}, mem)
	return makeKeys(keyBytes, mem), nil

}

func (c *Cache) GetLabels(keys []uint64, mem *arena.Arena) ([]labels.Labels, error) {
	c.mut.RLock()
	defer c.mut.RUnlock()

	keyBytes := arena.MakeSlice[[]byte](mem, len(keys), len(keys))
	for i := 0; i < len(keys); i++ {
		buf := arena.MakeSlice[byte](mem, 8, 8)
		binary.PutUvarint(buf, keys[i])
		keyBytes[i] = buf
	}
	valueBytes, err := c.idToLabel.getValues(keyBytes, mem)
	if err != nil {
		return nil, err
	}

	returnLbls := arena.MakeSlice[labels.Labels](mem, len(keys), len(keys))
	for i := 0; i < len(valueBytes); i++ {
		returnLbls[i] = fetchLabels(valueBytes[i], mem)
	}
	return returnLbls, nil
}

func (c *Cache) GetOrAddLink(componentID string, localRefID uint64, lbls labels.Labels) uint64 {
	// Not a fan of the locking behavior here.
	c.mut.RLock()
	mem := arena.NewArena()
	defer mem.Free()

	bb := arena.New[bytes.Buffer](mem)
	pushString(componentID, bb, mem)
	pushUint64(localRefID, bb, mem)

	vals, _ := c.localIDtoID.getValues([][]byte{bb.Bytes()}, mem)
	if vals[0] != nil {
		id, _ := binary.Uvarint(vals[0])
		c.mut.RUnlock()
		return id
	}
	c.mut.RUnlock()

	// TODO encapsulate this writing to the two mappings.
	keys, _ := c.GetIDs([][]labels.Label{[]labels.Label(lbls)}, mem)
	kb := arena.MakeSlice[byte](mem, 8, 8)
	binary.PutUvarint(kb, keys[0])
	c.mut.Lock()
	_ = c.localIDtoID.writeValues([][]byte{bb.Bytes()}, [][]byte{kb}, mem)

	gb := arena.New[bytes.Buffer](mem)
	pushString(componentID, gb, mem)
	pushUint64(keys[0], gb, mem)
	localBuf := arena.MakeSlice[byte](mem, 8, 8)
	binary.PutUvarint(localBuf, localRefID)
	_ = c.idToLocalID.writeValues([][]byte{gb.Bytes()}, [][]byte{localBuf}, mem)
	c.mut.Unlock()
	return keys[0]
}
func (c *Cache) GetOrAddGlobalRefID(l labels.Labels) uint64 {
	mem := arena.NewArena()
	defer mem.Free()
	keys, _ := c.GetIDs([][]labels.Label{l}, mem)
	return keys[0]
}
func (c *Cache) GetGlobalRefID(componentID string, localRefID uint64) uint64 {
	mem := arena.NewArena()
	defer mem.Free()

	c.mut.Lock()
	defer c.mut.Unlock()

	bb := arena.New[bytes.Buffer](mem)
	pushString(componentID, bb, mem)
	pushUint64(localRefID, bb, mem)

	keyByte, _ := c.localIDtoID.getValues([][]byte{bb.Bytes()}, mem)
	keys := makeKeys(keyByte, mem)
	return keys[0]

}
func (c *Cache) GetLocalRefID(componentID string, globalRefID uint64) uint64 {
	mem := arena.NewArena()
	defer mem.Free()

	c.mut.Lock()
	defer c.mut.Unlock()

	bb := arena.New[bytes.Buffer](mem)
	pushString(componentID, bb, mem)
	pushUint64(globalRefID, bb, mem)

	keyByte, _ := c.idToLocalID.getValues([][]byte{bb.Bytes()}, mem)
	keys := makeKeys(keyByte, mem)
	return keys[0]

}

func makeLabelBytes(lbls [][]labels.Label, mem *arena.Arena) [][]byte {
	// Find all the existing labels.
	lblBuf := arena.MakeSlice[[]byte](mem, len(lbls), len(lbls))
	for x, l := range lbls {
		buf := arena.New[bytes.Buffer](mem)
		pushLabels(l, buf, mem)
		lblBuf[x] = buf.Bytes()
	}
	return lblBuf
}

func makeKeys(keyBytes [][]byte, mem *arena.Arena) []uint64 {
	returnIDs := arena.MakeSlice[uint64](mem, len(keyBytes), len(keyBytes))
	for x, k := range keyBytes {
		val, _ := binary.Uvarint(k)
		returnIDs[x] = val
	}
	return returnIDs
}

func (c *Cache) writeNotFoundKeys(keys [][]byte, lblBuf [][]byte, mem *arena.Arena) ([][]byte, error) {
	// Since we dont know the labels we need to make declare a growable array.
	lblsToWrite := make([][]byte, 0)
	keysToWrite := make([][]byte, 0)

	// For anything without a key get a new one.
	for i := 0; i < len(keys); i++ {
		if keys[i] != nil {
			continue
		}
		keys[i] = c.getByteForNextKey(mem)
		lblsToWrite = append(lblsToWrite, lblBuf[i])
		keysToWrite = append(keysToWrite, keys[i])
	}

	err := c.labelToID.writeValues(lblsToWrite, keysToWrite, mem)
	if err != nil {
		return nil, err
	}

	err = c.idToLabel.writeValues(keysToWrite, lblsToWrite, mem)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (c *Cache) getCurrentID() uint64 {
	uintBuf := c.idToLabel.getNewestID()
	val, _ := binary.Uvarint(uintBuf)
	if val == 0 {
		val = 1
	}
	return val
}

func (c *Cache) getByteForNextKey(mem *arena.Arena) []byte {
	c.idMut.Lock()
	defer c.idMut.Unlock()
	c.currentID = c.currentID + 1
	buf := arena.MakeSlice[byte](mem, 8, 8)
	binary.PutUvarint(buf, c.currentID)
	return buf
}

func pushLabels(lbl labels.Labels, buf *bytes.Buffer, mem *arena.Arena) {
	pushUInt16(uint16(len(lbl)), buf, mem)
	for _, l := range lbl {
		pushString(l.Name, buf, mem)
		pushString(l.Value, buf, mem)
	}
}

func fetchLabels(b []byte, mem *arena.Arena) []labels.Label {
	buf := bytes.NewBuffer(b)
	count := fetchUInt16(buf, mem)
	lbls := arena.MakeSlice[labels.Label](mem, int(count), int(count))
	for i := 0; i < int(count); i++ {
		name := fetchString(buf, mem)
		value := fetchString(buf, mem)
		lbls[i] = labels.Label{Name: name, Value: value}
	}
	return lbls
}

func pushUInt16(v uint16, buf *bytes.Buffer, mem *arena.Arena) {
	tmp := arena.MakeSlice[byte](mem, 2, 2)
	binary.BigEndian.PutUint16(tmp, v)
	buf.Write(tmp)
}

func pushInt64(v int64, buf *bytes.Buffer, mem *arena.Arena) {
	tmp := arena.MakeSlice[byte](mem, 8, 8)
	binary.PutVarint(tmp, v)
	buf.Write(tmp)
}

func pushUint64(v uint64, buf *bytes.Buffer, mem *arena.Arena) {
	tmp := arena.MakeSlice[byte](mem, 8, 8)
	binary.PutUvarint(tmp, v)
	buf.Write(tmp)
}

func pushString(v string, buf *bytes.Buffer, mem *arena.Arena) {
	pushUInt16(uint16(len(v)), buf, mem)
	buf.Write(toBytes(&v))
}

func pushUint64ByteSlice(input [][]byte, buf *bytes.Buffer, mem *arena.Arena) {
	pushUint64(uint64(len(input)), buf, mem)
	for _, v := range input {
		buf.Write(v)
	}
}

func pushUInt64Slice(input []uint64, buf *bytes.Buffer, mem *arena.Arena) {
	pushUint64(uint64(len(input)), buf, mem)
	for _, v := range input {
		pushUint64(v, buf, mem)
	}
}

func fetchUInt16(buf *bytes.Buffer, mem *arena.Arena) uint16 {
	tmp := arena.MakeSlice[byte](mem, 2, 2)
	_, _ = buf.Read(tmp)
	ret := binary.BigEndian.Uint16(tmp)
	return ret
}

func fetchInt64(buf *bytes.Buffer, mem *arena.Arena) int64 {
	tmp := arena.MakeSlice[byte](mem, 8, 8)
	_, _ = buf.Read(tmp)
	ret, _ := binary.Varint(tmp)
	return ret
}

func fetchString(buf *bytes.Buffer, mem *arena.Arena) string {
	length := fetchUInt16(buf, mem)
	tmp := arena.MakeSlice[byte](mem, int(length), int(length))
	_, _ = buf.Read(tmp)
	return toString(&tmp)
}

func toBytes(s *string) []byte {
	return *(*[]byte)(unsafe.Pointer(s))
}
func toString(b *[]byte) string {
	return *(*string)(unsafe.Pointer(b))
}
