package rowcache

// Copyright (c) 2020-2022 Micro Focus or one of its affiliates.
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
	"bufio"
	"encoding/binary"
	"io"
	"io/ioutil"
	"os"

	"github.com/vertica/vertica-sql-go/msgs"
)

// FileCache stores rows from the wire and puts excess into a temporary file
type FileCache struct {
	maxInMemory int
	rowCount    int
	readIdx     int
	resultData  []*msgs.BEDataRowMsg
	file        *os.File
	rwbuffer    *bufio.ReadWriter
	scratch     [512]byte
}

// NewFileCache returns a file cache with a set row limit
func NewFileCache(rowLimit int) (*FileCache, error) {
	file, err := ioutil.TempFile("", ".vertica-sql-go.*.dat")
	if err != nil {
		return nil, err
	}
	return &FileCache{
		maxInMemory: rowLimit,
		resultData:  make([]*msgs.BEDataRowMsg, 0, rowLimit),
		file:        file,
		rwbuffer:    bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriterSize(file, 1<<16)),
	}, nil
}

func (f *FileCache) writeCached(msg *msgs.BEDataRowMsg) error {
	sizeBuf := f.scratch[:4]
	binary.LittleEndian.PutUint32(sizeBuf, uint32(len(*msg)))
	if _, err := f.rwbuffer.Write(sizeBuf); err != nil {
		return err
	}
	if _, err := f.rwbuffer.Write(*msg); err != nil {
		return err
	}
	return nil
}

// AddRow adds a row to the cache
func (f *FileCache) AddRow(msg *msgs.BEDataRowMsg) error {
	f.rowCount++
	if len(f.resultData) >= f.maxInMemory {
		if err := f.writeCached(msg); err != nil {
			return err
		}
		return nil
	}
	f.resultData = append(f.resultData, msg)
	return nil
}

// Finalize signals the end of rows from the wire and readies the cache for reading
func (f *FileCache) Finalize() error {
	var err error
	name := f.file.Name()
	if err = f.rwbuffer.Flush(); err != nil {
		return err
	}
	if err = f.file.Close(); err != nil {
		return err
	}
	f.file, err = os.OpenFile(name, os.O_RDONLY|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	f.rwbuffer = bufio.NewReadWriter(bufio.NewReader(f.file), bufio.NewWriter(f.file))
	return err
}

// GetRow pulls a row message out of the cache, returning nil of none remain
func (f *FileCache) GetRow() *msgs.BEDataRowMsg {
	if f.readIdx >= len(f.resultData) {
		if !f.reloadFromCache() {
			return nil
		}
	}
	result := f.resultData[f.readIdx]
	f.readIdx++
	return result
}

// Peek returns the next row without changing the state
func (f *FileCache) Peek() *msgs.BEDataRowMsg {
	if len(f.resultData) == 0 {
		return nil
	}
	return f.resultData[0]
}

// Close clears resources associated with the cache, deleting the temp file
func (f *FileCache) Close() error {
	name := f.file.Name()
	f.rwbuffer.Flush()
	f.file.Close()
	return os.Remove(name)
}

func (f *FileCache) reloadFromCache() bool {
	hadData := false

	f.readIdx = 0
	indexCount := 0

	for {
		sizeBuf := f.scratch[:4]

		if _, err := io.ReadFull(f.rwbuffer, sizeBuf); err != nil {
			if err == io.EOF {
				if indexCount == 0 {
					return false
				}
				f.resultData = f.resultData[0:indexCount]
				return true
			}
			return false
		}

		rowDataSize := binary.LittleEndian.Uint32(sizeBuf)

		var rowBuf []byte
		rowBytes := f.scratch[4:]
		if rowDataSize <= uint32(len(rowBytes)) {
			rowBuf = rowBytes[:rowDataSize]
		} else {
			rowBuf = make([]byte, rowDataSize)
		}
		if _, err := io.ReadFull(f.rwbuffer, rowBuf); err != nil {
			return false
		}

		msgBuf := msgs.NewMsgBufferFromBytes(rowBuf)

		drm := &msgs.BEDataRowMsg{}

		msg, _ := drm.CreateFromMsgBody(msgBuf)

		f.resultData[indexCount] = msg.(*msgs.BEDataRowMsg)
		indexCount++

		hadData = true

		// If we've reached the original capacity of the slice, we're done.
		if indexCount == len(f.resultData) {
			break
		}
	}

	return hadData
}
