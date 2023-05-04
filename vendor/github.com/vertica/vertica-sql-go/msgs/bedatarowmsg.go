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
	"encoding/binary"
	"fmt"
)

// BEDataRowMsg docs
type BEDataRowMsg []byte

// ColumnExtractor pulls columns out of a row
type ColumnExtractor struct {
	NumCols uint16
	data    []byte
	idx     int32
}

// Chunk provides the raw bytes for a column of data
func (c *ColumnExtractor) Chunk() []byte {
	size := int32(binary.BigEndian.Uint32(c.data[c.idx : c.idx+4]))
	c.idx += 4
	if size == -1 {
		return nil
	}
	chunk := c.data[c.idx : c.idx+size]
	c.idx += size
	return chunk
}

// CreateFromMsgBody docs
func (b *BEDataRowMsg) CreateFromMsgBody(buf *msgBuffer) (BackEndMsg, error) {
	res := BEDataRowMsg(buf.buf.Bytes())
	buf.buf.Reset()
	return &res, nil
}

// Columns provides an extractor to begin reading columns
func (b *BEDataRowMsg) Columns() ColumnExtractor {
	return ColumnExtractor{
		NumCols: binary.BigEndian.Uint16((*b)[0:2]),
		data:    (*b)[2:],
	}
}

// RevertToBytes dumps the message back into plain bytes
func (b *BEDataRowMsg) RevertToBytes() []byte {
	return *b
}

func (b *BEDataRowMsg) String() string {
	cols := b.Columns()
	return fmt.Sprintf("DataRow: %d column(s)", cols.NumCols)
}

func init() {
	registerBackEndMsgType('D', &BEDataRowMsg{})
}
