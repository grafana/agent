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
	"database/sql/driver"
	"fmt"
	"time"
)

// FEBindMsg docs
type FEBindMsg struct {
	Portal    string
	Statement string
	NamedArgs []driver.NamedValue
	OIDTypes  []int32
}

// Flatten docs
func (m *FEBindMsg) Flatten() ([]byte, byte) {

	buf := newMsgBuffer()

	buf.appendString(m.Portal)
	buf.appendString(m.Statement)

	// no parameter format codes for now
	buf.appendUint16(0)

	// number of arguments
	buf.appendUint16(uint16(len(m.NamedArgs)))

	for _, oidType := range m.OIDTypes {
		buf.appendUint32(uint32(oidType))
	}

	var strVal string

	for _, arg := range m.NamedArgs {
		switch v := arg.Value.(type) {
		case int64, float64:
			strVal = fmt.Sprintf("%v", v)
		case string:
			strVal = v
		case bool:
			if v {
				strVal = "1"
			} else {
				strVal = "0"
			}
		case nil:
			buf.appendUint32(0xffffffff)
			continue
		case time.Time:
			strVal = v.Format("2006-01-02T15:04:05.999999Z07:00")
		case []uint8:
			// Escape the byte value "\" with "\134"(octal for backslash)
			v = bytes.ReplaceAll(v, []byte("\\"), []byte("\\134"))
			buf.appendUint32(uint32(len(v)))
			buf.appendBytes(v)
			continue
		default:
			strVal = "??HELP??"
		}

		buf.appendUint32(uint32(len(strVal)))
		buf.appendBytes([]byte(strVal))
	}

	buf.appendUint16(0) // all columns in default format

	return buf.bytes(), 'B'
}

func (m *FEBindMsg) String() string {
	return fmt.Sprintf(
		"Bind: Portal='%s', Statement='%s', ArgC=%d",
		m.Portal,
		m.Statement,
		len(m.OIDTypes))
}
