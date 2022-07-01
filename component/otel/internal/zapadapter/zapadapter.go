// Package zapadapter allows go-kit/log to be used as a zap core.
package zapadapter

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New returns a new zap.Logger that logs to the provided log.Logger.
func New(l log.Logger) *zap.Logger {
	return zap.New(&loggerCore{inner: l})
}

type loggerCore struct {
	inner log.Logger
}

var _ zapcore.Core = (*loggerCore)(nil)

func (lc *loggerCore) Enabled(zapcore.Level) bool { return true }

func (lc *loggerCore) With(ff []zapcore.Field) zapcore.Core {
	enc := fieldEncoder{
		// TODO(rfratto): pool?
		fields: make([]interface{}, 0, len(ff)*2),
	}

	for _, f := range ff {
		f.AddTo(&enc)
	}

	return &loggerCore{
		inner: log.With(lc.inner, enc.fields...),
	}
}

func (lc *loggerCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(e, lc)
}

func (lc *loggerCore) Write(e zapcore.Entry, ff []zapcore.Field) error {
	enc := fieldEncoder{
		// TODO(rfratto): pool?
		fields: make([]interface{}, 0, len(ff)*2),
	}

	for _, f := range ff {
		f.AddTo(&enc)
	}

	enc.fields = append(enc.fields, "msg", e.Message)

	switch e.Level {
	case zapcore.DebugLevel:
		return level.Debug(lc.inner).Log(enc.fields...)
	case zapcore.InfoLevel:
		return level.Info(lc.inner).Log(enc.fields...)
	case zapcore.WarnLevel:
		return level.Warn(lc.inner).Log(enc.fields...)
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		// We ignore panics/fatals hwere because we really don't want components to
		// be able to do that.
		return level.Error(lc.inner).Log(enc.fields...)
	default:
		return lc.inner.Log(enc.fields...)
	}
}

func (lc *loggerCore) Sync() error {
	return nil
}

type fieldEncoder struct {
	fields []interface{}

	namespace []string
}

var _ zapcore.ObjectEncoder = (*fieldEncoder)(nil)

func (fe *fieldEncoder) keyName(k string) interface{} {
	if len(fe.namespace) == 0 {
		return k
	}
	return key(append(fe.namespace, k))
}

func (fe *fieldEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	fe.fields = append(fe.fields, fe.keyName(key), "<array>")
	return nil
}

func (fe *fieldEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	fe.fields = append(fe.fields, fe.keyName(key), "<object>")
	return nil
}

func (fe *fieldEncoder) AddBinary(key string, value []byte) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddByteString(key string, value []byte) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddBool(key string, value bool) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddComplex128(key string, value complex128) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddComplex64(key string, value complex64) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddDuration(key string, value time.Duration) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddFloat64(key string, value float64) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddFloat32(key string, value float32) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddInt(key string, value int) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddInt64(key string, value int64) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddInt32(key string, value int32) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddInt16(key string, value int16) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddInt8(key string, value int8) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddString(key, value string) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddTime(key string, value time.Time) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddUint(key string, value uint) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddUint64(key string, value uint64) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddUint32(key string, value uint32) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddUint16(key string, value uint16) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddUint8(key string, value uint8) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddUintptr(key string, value uintptr) {
	fe.fields = append(fe.fields, fe.keyName(key), value)
}

func (fe *fieldEncoder) AddReflected(key string, value interface{}) error {
	fe.fields = append(fe.fields, fe.keyName(key), value)
	return nil
}

func (fe *fieldEncoder) OpenNamespace(key string) {
	fe.namespace = append(fe.namespace, key)
}

type key []string

var _ fmt.Stringer = (key)(nil)

func (k key) String() string {
	if len(k) == 1 {
		return k[0]
	}
	return strings.Join(k, ".")
}
