package vertigo

import (
	"context"
	"fmt"
	"io"
	"os"
)

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

const (
	minCopyBlockSize          = 16384
	stdInDefaultCopyBlockSize = 65536
)

type VerticaContext interface {
	context.Context

	SetCopyInputStream(inputStream io.Reader) error
	GetCopyInputStream() io.Reader

	SetCopyBlockSizeBytes(blockSize int) error
	GetCopyBlockSizeBytes() int

	SetInMemoryResultRowLimit(rowLimit int) error
	GetInMemoryResultRowLimit() int
}

type verticaContext struct {
	context.Context

	inputStream io.Reader
	blockSize   int
	rowLimit    int
}

// NewVerticaContext creates a new context that inherits the values and behavior of the provided parent context.
func NewVerticaContext(parentCtx context.Context) VerticaContext {
	return &verticaContext{
		Context:     parentCtx,
		inputStream: os.Stdin,
		blockSize:   stdInDefaultCopyBlockSize,
		rowLimit:    0,
	}
}

// SetCopyInputStream sets the input stream to be used when copying from stdin. If not set, copying from stdin will
// read from os.stdin.
func (c *verticaContext) SetCopyInputStream(inputStream io.Reader) error {
	if inputStream == nil {
		return fmt.Errorf("cannot SetInputStream to a nil value")
	}

	c.inputStream = inputStream

	return nil
}

// GetCopyInputStream returns the currently active input stream to be used when copying from stdin.
func (c *verticaContext) GetCopyInputStream() io.Reader {
	return c.inputStream
}

// SetCopyBlockSizeBytes sets the size of the buffer used to transfer from the input stream to Vertica. By
// default, it's 65536 (64k). It must be at least 16384 (16k) bytes.
func (c *verticaContext) SetCopyBlockSizeBytes(blockSize int) error {
	if blockSize < minCopyBlockSize {
		return fmt.Errorf("cannot set copy block size to less than %d", minCopyBlockSize)
	}

	c.blockSize = blockSize

	return nil
}

// GetCopyBlockSizeBytes gets the size of the buffer used to transfer from the input stream to Vertica.
func (c *verticaContext) GetCopyBlockSizeBytes() int {
	return c.blockSize
}

func (c *verticaContext) SetInMemoryResultRowLimit(rowLimit int) error {
	if rowLimit < 0 {
		return fmt.Errorf("cannot set result limit to a negative number")
	}

	c.rowLimit = rowLimit

	return nil
}

func (c *verticaContext) GetInMemoryResultRowLimit() int {
	return c.rowLimit
}
