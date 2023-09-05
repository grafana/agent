package queue

import (
	"arena"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/prometheus/prometheus/storage"
	"math"
	"unsafe"
)

//var bufPool = pebble.NewArrayBufferPool()

type sample struct {
	L         []string
	TimeStamp int64
	Value     float64
}

func (s *sample) Marshal(buf *bytes.Buffer, mem *arena.Arena) {
	pushUInt16(1, buf, mem)
	pushInt64(s.TimeStamp, buf, mem)
	pushFloat64(s.Value, buf, mem)
	pushStringSlice(s.L, buf, mem)
}

func Unmarshal(s *sample, buf *bytes.Buffer, mem *arena.Arena) error {
	version := fetchUInt16(buf, mem)
	if version != 1 {
		return fmt.Errorf("incorrect version header found for version expected 1 got %d", version)
	}
	s.TimeStamp = fetchInt64(buf, mem)
	s.Value = fetchFloat64(buf, mem)
	s.L = fetchStringArray(buf, mem)
	return nil
}

func marshalSamples(samples []*sample, buf *bytes.Buffer, mem *arena.Arena) error {
	pushUInt16(uint16(len(samples)), buf, mem)
	for _, s := range samples {
		s.Marshal(buf, mem)
	}
	return nil
}

func unmarshalSamples(buf *bytes.Buffer, mem *arena.Arena) ([]*sample, error) {
	arrLength := fetchUInt16(buf, mem)
	samples := arena.MakeSlice[*sample](mem, int(arrLength), int(arrLength))
	for i := 0; i < len(samples); i++ {
		if samples[i] == nil {
			samples[i] = &sample{}
		}
	}
	samples = samples[:arrLength]
	for i := 0; i < int(arrLength); i++ {
		err := Unmarshal(samples[i], buf, mem)
		if err != nil {
			return samples, err
		}
	}
	return samples, nil
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

func pushFloat64(v float64, buf *bytes.Buffer, mem *arena.Arena) {
	tmp := arena.MakeSlice[byte](mem, 8, 8)
	binary.BigEndian.PutUint64(tmp, math.Float64bits(v))
	buf.Write(tmp)
}

func pushString(v string, buf *bytes.Buffer, mem *arena.Arena) {
	pushUInt16(uint16(len(v)), buf, mem)
	buf.WriteString(v)
}

func pushMap(v map[string]string, buf *bytes.Buffer, mem *arena.Arena) {
	pushUInt16(uint16(len(v)*2), buf, mem)
	for key, value := range v {
		pushString(key, buf, mem)
		pushString(value, buf, mem)
	}
}

func pushStringSlice(v []string, buf *bytes.Buffer, mem *arena.Arena) {
	pushUInt16(uint16(len(v)), buf, mem)
	for _, value := range v {
		pushString(value, buf, mem)
	}
}

func fetchUInt16(buf *bytes.Buffer, mem *arena.Arena) uint16 {
	tmp := arena.MakeSlice[byte](mem, 2, 2)
	buf.Read(tmp)
	ret := binary.BigEndian.Uint16(tmp)
	return ret
}

func fetchInt64(buf *bytes.Buffer, mem *arena.Arena) int64 {
	tmp := arena.MakeSlice[byte](mem, 8, 8)
	buf.Read(tmp)
	ret, _ := binary.Varint(tmp)
	return ret
}

func fetchFloat64(buf *bytes.Buffer, mem *arena.Arena) float64 {
	tmp := arena.MakeSlice[byte](mem, 8, 8)
	buf.Read(tmp)
	large := binary.BigEndian.Uint64(tmp)
	ret := math.Float64frombits(large)
	return ret

}

func fetchString(buf *bytes.Buffer, mem *arena.Arena) string {

	length := fetchUInt16(buf, mem)
	tmp := arena.MakeSlice[byte](mem, int(length), int(length))
	buf.Read(tmp)
	return toString(&tmp)
}

func fetchMap(v map[string]string, buf *bytes.Buffer, mem *arena.Arena) {
	length := fetchUInt16(buf, mem)
	var key string
	var value string

	for i := 0; uint16(i) < length/2; i++ {
		key = fetchString(buf, mem)
		value = fetchString(buf, mem)
		v[key] = value
	}
}

func fetchStringArray(buf *bytes.Buffer, mem *arena.Arena) []string {
	length := fetchUInt16(buf, mem)
	arr := arena.MakeSlice[string](mem, int(length), int(length))
	for i := 0; uint16(i) < length; i++ {
		arr[i] = fetchString(buf, mem)
	}
	return arr
}

func toString(b *[]byte) string {
	return *(*string)(unsafe.Pointer(b))
}

var _ storage.Appender = (*appender)(nil)
