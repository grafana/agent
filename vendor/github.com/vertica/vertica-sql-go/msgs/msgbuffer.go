package msgs

// Copyright (c) 2019-2022 Micro Focus or one of its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"bytes"
	"encoding/binary"
)

type msgBuffer struct {
	buf *bytes.Buffer
}

func newMsgBuffer() *msgBuffer {
	res := &msgBuffer{}
	res.buf = new(bytes.Buffer)

	return res
}

func NewMsgBufferFromBytes(b []byte) *msgBuffer {
	res := &msgBuffer{}
	res.buf = new(bytes.Buffer)
	res.buf.Write(b)

	return res
}

func (b *msgBuffer) appendUint32(i uint32) *msgBuffer {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, i)

	b.buf.Write(buf)

	return b
}

func (b *msgBuffer) appendInt32(i int32) *msgBuffer {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(i))

	b.buf.Write(buf)

	return b
}

func (b *msgBuffer) appendUint16(i uint16) *msgBuffer {
	buf := []byte{0, 0}
	binary.BigEndian.PutUint16(buf, i)

	b.buf.Write(buf)

	return b
}

func (b *msgBuffer) appendUint64(i uint64) *msgBuffer {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, i)

	b.buf.Write(buf)

	return b
}

func (b *msgBuffer) appendByte(v byte) *msgBuffer {
	b.buf.WriteByte(v)

	return b
}

func (b *msgBuffer) appendBytes(bytes []byte) *msgBuffer {
	b.buf.Write(bytes)

	return b
}

func (b *msgBuffer) appendString(stringVal string) *msgBuffer {
	b.buf.Write([]byte(stringVal))
	b.buf.Write([]byte{0})

	return b
}

func (b *msgBuffer) appendLabeledString(label, stringVal string) *msgBuffer {
	b.appendString(label)
	b.appendString(stringVal)

	return b
}

func (b *msgBuffer) bytes() []byte {
	return b.buf.Bytes()
}

func (b *msgBuffer) readString() string {
	res, _ := b.buf.ReadString(0)
	return res[:len(res)-1]
}

func (b *msgBuffer) readTaggedString() (byte, string) {

	if b.buf.Len() <= 1 {
		return 0, ""
	}

	fieldType, _ := b.buf.ReadByte()
	fieldStr := b.readString()

	return fieldType, fieldStr
}

func (b *msgBuffer) readInt16() int16 {
	return int16(b.readUint16())
}

func (b *msgBuffer) readUint16() uint16 {
	buf := []byte{0, 0}

	b.buf.Read(buf)

	return binary.BigEndian.Uint16(buf)
}

func (b *msgBuffer) readInt32() int32 {
	return int32(b.readUint32())
}

func (b *msgBuffer) readUint32() uint32 {
	buf := make([]byte, 4)

	b.buf.Read(buf)

	return binary.BigEndian.Uint32(buf)
}

func (b *msgBuffer) readInt64() int64 {
	return int64(b.readUInt64())
}

func (b *msgBuffer) readUInt64() uint64 {
	buf := make([]byte, 8)

	b.buf.Read(buf)

	return binary.BigEndian.Uint64(buf)
}

func (b *msgBuffer) readByte() byte {
	bt, _ := b.buf.ReadByte()

	return bt
}

func (b *msgBuffer) readBool() bool {
	bt, _ := b.buf.ReadByte()

	if bt == 1 {
		return true
	}

	return false
}

func (b *msgBuffer) readBytes(outBuf []byte) int {
	ct, _ := b.buf.Read(outBuf)
	return ct
}

func (b *msgBuffer) remainingBytes() int {
	return b.buf.Len()
}
