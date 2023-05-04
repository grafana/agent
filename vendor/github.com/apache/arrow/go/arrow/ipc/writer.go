// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ipc // import "github.com/apache/arrow/go/arrow/ipc"

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"math"
	"sync"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/bitutil"
	"github.com/apache/arrow/go/arrow/internal/flatbuf"
	"github.com/apache/arrow/go/arrow/memory"
	"golang.org/x/xerrors"
)

type swriter struct {
	w   io.Writer
	pos int64
}

func (w *swriter) Start() error { return nil }
func (w *swriter) Close() error {
	_, err := w.Write(kEOS[:])
	return err
}

func (w *swriter) WritePayload(p Payload) error {
	_, err := writeIPCPayload(w, p)
	if err != nil {
		return err
	}
	return nil
}

func (w *swriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.pos += int64(n)
	return n, err
}

// Writer is an Arrow stream writer.
type Writer struct {
	w io.Writer

	mem memory.Allocator
	pw  PayloadWriter

	started    bool
	schema     *arrow.Schema
	codec      flatbuf.CompressionType
	compressNP int
}

// NewWriterWithPayloadWriter constructs a writer with the provided payload writer
// instead of the default stream payload writer. This makes the writer more
// reusable such as by the Arrow Flight writer.
func NewWriterWithPayloadWriter(pw PayloadWriter, opts ...Option) *Writer {
	cfg := newConfig(opts...)
	return &Writer{
		mem:        cfg.alloc,
		pw:         pw,
		schema:     cfg.schema,
		codec:      cfg.codec,
		compressNP: cfg.compressNP,
	}
}

// NewWriter returns a writer that writes records to the provided output stream.
func NewWriter(w io.Writer, opts ...Option) *Writer {
	cfg := newConfig(opts...)
	return &Writer{
		w:      w,
		mem:    cfg.alloc,
		pw:     &swriter{w: w},
		schema: cfg.schema,
		codec:  cfg.codec,
	}
}

func (w *Writer) Close() error {
	if !w.started {
		err := w.start()
		if err != nil {
			return err
		}
	}

	if w.pw == nil {
		return nil
	}

	err := w.pw.Close()
	if err != nil {
		return xerrors.Errorf("arrow/ipc: could not close payload writer: %w", err)
	}
	w.pw = nil

	return nil
}

func (w *Writer) Write(rec array.Record) error {
	if !w.started {
		err := w.start()
		if err != nil {
			return err
		}
	}

	schema := rec.Schema()
	if schema == nil || !schema.Equal(w.schema) {
		return errInconsistentSchema
	}

	const allow64b = true
	var (
		data = Payload{msg: MessageRecordBatch}
		enc  = newRecordEncoder(w.mem, 0, kMaxNestingDepth, allow64b, w.codec, w.compressNP)
	)
	defer data.Release()

	if err := enc.Encode(&data, rec); err != nil {
		return xerrors.Errorf("arrow/ipc: could not encode record to payload: %w", err)
	}

	return w.pw.WritePayload(data)
}

func (w *Writer) start() error {
	w.started = true

	// write out schema payloads
	ps := payloadsFromSchema(w.schema, w.mem, nil)
	defer ps.Release()

	for _, data := range ps {
		err := w.pw.WritePayload(data)
		if err != nil {
			return err
		}
	}

	return nil
}

type recordEncoder struct {
	mem memory.Allocator

	fields []fieldMetadata
	meta   []bufferMetadata

	depth      int64
	start      int64
	allow64b   bool
	codec      flatbuf.CompressionType
	compressNP int
}

func newRecordEncoder(mem memory.Allocator, startOffset, maxDepth int64, allow64b bool, codec flatbuf.CompressionType, compressNP int) *recordEncoder {
	return &recordEncoder{
		mem:        mem,
		start:      startOffset,
		depth:      maxDepth,
		allow64b:   allow64b,
		codec:      codec,
		compressNP: compressNP,
	}
}

func (w *recordEncoder) compressBodyBuffers(p *Payload) error {
	compress := func(idx int, codec compressor) error {
		if p.body[idx] == nil || p.body[idx].Len() == 0 {
			return nil
		}
		var buf bytes.Buffer
		buf.Grow(codec.MaxCompressedLen(p.body[idx].Len()) + arrow.Int64SizeBytes)
		if err := binary.Write(&buf, binary.LittleEndian, uint64(p.body[idx].Len())); err != nil {
			return err
		}
		codec.Reset(&buf)
		if _, err := codec.Write(p.body[idx].Bytes()); err != nil {
			return err
		}
		if err := codec.Close(); err != nil {
			return err
		}
		p.body[idx] = memory.NewBufferBytes(buf.Bytes())
		return nil
	}

	if w.compressNP <= 1 {
		codec := getCompressor(w.codec)
		for idx := range p.body {
			if err := compress(idx, codec); err != nil {
				return err
			}
		}
		return nil
	}

	var (
		wg          sync.WaitGroup
		ch          = make(chan int)
		errch       = make(chan error)
		ctx, cancel = context.WithCancel(context.Background())
	)
	defer cancel()

	for i := 0; i < w.compressNP; i++ {
		go func() {
			defer wg.Done()
			codec := getCompressor(w.codec)
			for {
				select {
				case idx, ok := <-ch:
					if !ok {
						// we're done, channel is closed!
						return
					}

					if err := compress(idx, codec); err != nil {
						errch <- err
						cancel()
						return
					}
				case <-ctx.Done():
					// cancelled, return early
					return
				}
			}
		}()
	}

	for idx := range p.body {
		ch <- idx
	}

	close(ch)
	wg.Wait()
	close(errch)

	return <-errch
}

func (w *recordEncoder) Encode(p *Payload, rec array.Record) error {

	// perform depth-first traversal of the row-batch
	for i, col := range rec.Columns() {
		err := w.visit(p, col)
		if err != nil {
			return xerrors.Errorf("arrow/ipc: could not encode column %d (%q): %w", i, rec.ColumnName(i), err)
		}
	}

	if w.codec != -1 {
		w.compressBodyBuffers(p)
	}

	// position for the start of a buffer relative to the passed frame of reference.
	// may be 0 or some other position in an address space.
	offset := w.start
	w.meta = make([]bufferMetadata, len(p.body))

	// construct the metadata for the record batch header
	for i, buf := range p.body {
		var (
			size    int64
			padding int64
		)
		// the buffer might be null if we are handling zero row lengths.
		if buf != nil {
			size = int64(buf.Len())
			padding = bitutil.CeilByte64(size) - size
		}
		w.meta[i] = bufferMetadata{
			Offset: offset,
			// even though we add padding, we need the Len to be correct
			// so that decompressing works properly.
			Len: size,
		}
		offset += size + padding
	}

	p.size = offset - w.start
	if !bitutil.IsMultipleOf8(p.size) {
		panic("not aligned")
	}

	return w.encodeMetadata(p, rec.NumRows())
}

func (w *recordEncoder) visit(p *Payload, arr array.Interface) error {
	if w.depth <= 0 {
		return errMaxRecursion
	}

	if !w.allow64b && arr.Len() > math.MaxInt32 {
		return errBigArray
	}

	if arr.DataType().ID() == arrow.EXTENSION {
		arr := arr.(array.ExtensionArray)
		err := w.visit(p, arr.Storage())
		if err != nil {
			return xerrors.Errorf("failed visiting storage of for array %T: %w", arr, err)
		}
		return nil
	}

	// add all common elements
	w.fields = append(w.fields, fieldMetadata{
		Len:    int64(arr.Len()),
		Nulls:  int64(arr.NullN()),
		Offset: 0,
	})

	if arr.DataType().ID() == arrow.NULL {
		return nil
	}

	switch arr.NullN() {
	case 0:
		p.body = append(p.body, nil)
	default:
		data := arr.Data()
		bitmap := newTruncatedBitmap(w.mem, int64(data.Offset()), int64(data.Len()), data.Buffers()[0])
		p.body = append(p.body, bitmap)
	}

	switch dtype := arr.DataType().(type) {
	case *arrow.NullType:
		// ok. NullArrays are completely empty.

	case *arrow.BooleanType:
		var (
			data = arr.Data()
			bitm *memory.Buffer
		)

		if data.Len() != 0 {
			bitm = newTruncatedBitmap(w.mem, int64(data.Offset()), int64(data.Len()), data.Buffers()[1])
		}
		p.body = append(p.body, bitm)

	case arrow.FixedWidthDataType:
		data := arr.Data()
		values := data.Buffers()[1]
		arrLen := int64(arr.Len())
		typeWidth := int64(dtype.BitWidth() / 8)
		minLength := paddedLength(arrLen*typeWidth, kArrowAlignment)

		switch {
		case needTruncate(int64(data.Offset()), values, minLength):
			// non-zero offset: slice the buffer
			offset := int64(data.Offset()) * typeWidth
			// send padding if available
			len := minI64(bitutil.CeilByte64(arrLen*typeWidth), int64(values.Len())-offset)
			values = memory.NewBufferBytes(values.Bytes()[offset : offset+len])
		default:
			if values != nil {
				values.Retain()
			}
		}
		p.body = append(p.body, values)

	case *arrow.BinaryType:
		arr := arr.(*array.Binary)
		voffsets, err := w.getZeroBasedValueOffsets(arr)
		if err != nil {
			return xerrors.Errorf("could not retrieve zero-based value offsets from %T: %w", arr, err)
		}
		data := arr.Data()
		values := data.Buffers()[2]

		var totalDataBytes int64
		if voffsets != nil {
			totalDataBytes = int64(len(arr.ValueBytes()))
		}

		switch {
		case needTruncate(int64(data.Offset()), values, totalDataBytes):
			// slice data buffer to include the range we need now.
			var (
				beg = int64(arr.ValueOffset(0))
				len = minI64(paddedLength(totalDataBytes, kArrowAlignment), int64(totalDataBytes))
			)
			values = memory.NewBufferBytes(data.Buffers()[2].Bytes()[beg : beg+len])
		default:
			if values != nil {
				values.Retain()
			}
		}
		p.body = append(p.body, voffsets)
		p.body = append(p.body, values)

	case *arrow.StringType:
		arr := arr.(*array.String)
		voffsets, err := w.getZeroBasedValueOffsets(arr)
		if err != nil {
			return xerrors.Errorf("could not retrieve zero-based value offsets from %T: %w", arr, err)
		}
		data := arr.Data()
		values := data.Buffers()[2]

		var totalDataBytes int64
		if voffsets != nil {
			totalDataBytes = int64(len(arr.ValueBytes()))
		}

		switch {
		case needTruncate(int64(data.Offset()), values, totalDataBytes):
			// slice data buffer to include the range we need now.
			var (
				beg = int64(arr.ValueOffset(0))
				len = minI64(paddedLength(totalDataBytes, kArrowAlignment), int64(totalDataBytes))
			)
			values = memory.NewBufferBytes(data.Buffers()[2].Bytes()[beg : beg+len])
		default:
			if values != nil {
				values.Retain()
			}
		}
		p.body = append(p.body, voffsets)
		p.body = append(p.body, values)

	case *arrow.StructType:
		w.depth--
		arr := arr.(*array.Struct)
		for i := 0; i < arr.NumField(); i++ {
			err := w.visit(p, arr.Field(i))
			if err != nil {
				return xerrors.Errorf("could not visit field %d of struct-array: %w", i, err)
			}
		}
		w.depth++

	case *arrow.MapType:
		arr := arr.(*array.Map)
		voffsets, err := w.getZeroBasedValueOffsets(arr)
		if err != nil {
			return xerrors.Errorf("could not retrieve zero-based value offsets for array %T: %w", arr, err)
		}
		p.body = append(p.body, voffsets)

		w.depth--
		var (
			values        = arr.ListValues()
			mustRelease   = false
			values_offset int64
			values_length int64
		)
		defer func() {
			if mustRelease {
				values.Release()
			}
		}()

		if voffsets != nil {
			values_offset = int64(arr.Offsets()[0])
			values_length = int64(arr.Offsets()[arr.Len()]) - values_offset
		}

		if len(arr.Offsets()) != 0 || values_length < int64(values.Len()) {
			// must also slice the values
			values = array.NewSlice(values, values_offset, values_length)
			mustRelease = true
		}
		err = w.visit(p, values)

		if err != nil {
			return xerrors.Errorf("could not visit list element for array %T: %w", arr, err)
		}
		w.depth++
	case *arrow.ListType:
		arr := arr.(*array.List)
		voffsets, err := w.getZeroBasedValueOffsets(arr)
		if err != nil {
			return xerrors.Errorf("could not retrieve zero-based value offsets for array %T: %w", arr, err)
		}
		p.body = append(p.body, voffsets)

		w.depth--
		var (
			values        = arr.ListValues()
			mustRelease   = false
			values_offset int64
			values_length int64
		)
		defer func() {
			if mustRelease {
				values.Release()
			}
		}()

		if voffsets != nil {
			values_offset = int64(arr.Offsets()[0])
			values_length = int64(arr.Offsets()[arr.Len()]) - values_offset
		}

		if len(arr.Offsets()) != 0 || values_length < int64(values.Len()) {
			// must also slice the values
			values = array.NewSlice(values, values_offset, values_length)
			mustRelease = true
		}
		err = w.visit(p, values)

		if err != nil {
			return xerrors.Errorf("could not visit list element for array %T: %w", arr, err)
		}
		w.depth++

	case *arrow.FixedSizeListType:
		arr := arr.(*array.FixedSizeList)

		w.depth--

		size := int64(arr.DataType().(*arrow.FixedSizeListType).Len())
		beg := int64(arr.Offset()) * size
		end := int64(arr.Offset()+arr.Len()) * size

		values := array.NewSlice(arr.ListValues(), beg, end)
		defer values.Release()

		err := w.visit(p, values)

		if err != nil {
			return xerrors.Errorf("could not visit list element for array %T: %w", arr, err)
		}
		w.depth++

	default:
		panic(xerrors.Errorf("arrow/ipc: unknown array %T (dtype=%T)", arr, dtype))
	}

	return nil
}

func (w *recordEncoder) getZeroBasedValueOffsets(arr array.Interface) (*memory.Buffer, error) {
	data := arr.Data()
	voffsets := data.Buffers()[1]
	if data.Offset() != 0 {
		// if we have a non-zero offset, then the value offsets do not start at
		// zero. we must a) create a new offsets array with shifted offsets and
		// b) slice the values array accordingly
		shiftedOffsets := memory.NewResizableBuffer(w.mem)
		shiftedOffsets.Resize(arrow.Int32Traits.BytesRequired(data.Len() + 1))

		dest := arrow.Int32Traits.CastFromBytes(shiftedOffsets.Bytes())
		offsets := arrow.Int32Traits.CastFromBytes(voffsets.Bytes())[data.Offset() : data.Offset()+data.Len()+1]

		startOffset := offsets[0]
		for i, o := range offsets {
			dest[i] = o - startOffset
		}
		voffsets = shiftedOffsets
	} else {
		voffsets.Retain()
	}
	if voffsets == nil || voffsets.Len() == 0 {
		return nil, nil
	}

	return voffsets, nil
}

func (w *recordEncoder) encodeMetadata(p *Payload, nrows int64) error {
	p.meta = writeRecordMessage(w.mem, nrows, p.size, w.fields, w.meta, w.codec)
	return nil
}

func newTruncatedBitmap(mem memory.Allocator, offset, length int64, input *memory.Buffer) *memory.Buffer {
	if input != nil {
		input.Retain()
		return input
	}

	minLength := paddedLength(bitutil.BytesForBits(length), kArrowAlignment)
	switch {
	case offset != 0 || minLength < int64(input.Len()):
		// with a sliced array / non-zero offset, we must copy the bitmap
		panic("not implemented") // FIXME(sbinet): writer.cc:75
	default:
		input.Retain()
		return input
	}
}

func needTruncate(offset int64, buf *memory.Buffer, minLength int64) bool {
	if buf == nil {
		return false
	}
	return offset != 0 || minLength < int64(buf.Len())
}

func minI64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
