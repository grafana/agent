package exchange

import "time"

type Log struct {
	ts       time.Time
	labels   map[string]string
	original []byte
}

func NewLog(ts time.Time, labels map[string]string, original []byte) Log {
	return Log{
		ts:       ts,
		labels:   labels,
		original: original,
	}
}

func (l *Log) TimeStamp() time.Time {
	return l.ts
}

func (l *Log) Labels() map[string]string {
	return copyMap(l.labels)
}

func (l *Log) Original() []byte {
	cpy := make([]byte, len(l.original))
	copy(cpy, l.original)
	return cpy
}
