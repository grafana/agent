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
	"github.com/vertica/vertica-sql-go/msgs"
)

// MemoryCache is a simple in memory row store
type MemoryCache struct {
	resultData []*msgs.BEDataRowMsg
	readIdx    int
}

// NewMemoryCache initializes the memory store with a given size, but it can continue
// to grow
func NewMemoryCache(size int) *MemoryCache {
	return &MemoryCache{
		resultData: make([]*msgs.BEDataRowMsg, 0, size),
	}
}

// AddRow adds a row to the store
func (m *MemoryCache) AddRow(msg *msgs.BEDataRowMsg) error {
	m.resultData = append(m.resultData, msg)
	return nil
}

// Finalize signals the end of new rows, a noop for the memory cache
func (m *MemoryCache) Finalize() error {
	return nil
}

// GetRow pulls a row from the cache, returning nil if none remain
func (m *MemoryCache) GetRow() *msgs.BEDataRowMsg {
	if m.readIdx >= len(m.resultData) {
		return nil
	}
	result := m.resultData[m.readIdx]
	m.readIdx++
	return result
}

// Peek returns the next row without changing the state
func (m *MemoryCache) Peek() *msgs.BEDataRowMsg {
	if len(m.resultData) == 0 {
		return nil
	}
	return m.resultData[0]
}

// Close provides an opportunity to free resources, a noop for the memory cache
func (m *MemoryCache) Close() error {
	return nil
}
