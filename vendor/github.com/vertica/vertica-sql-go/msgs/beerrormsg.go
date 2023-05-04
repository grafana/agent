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
	"fmt"
)

// BEErrorMsg docs
type BEErrorMsg struct {
	InternalQuery    string
	Severity         string
	Message          string
	SQLState         string
	Detail           string
	Hint             string
	Position         string
	Where            string
	InternalPosition string
	Routine          string
	File             string
	Line             string
	ErrorCode        string
}

// CreateFromMsgBody docs
func (b *BEErrorMsg) CreateFromMsgBody(buf *msgBuffer) (BackEndMsg, error) {

	res := &BEErrorMsg{}

	for {
		fieldType, fieldStr := buf.readTaggedString()

		if fieldType == 0 {
			buf.readByte() // There's an empty null terminator here we have to read.
			break
		}
		switch fieldType {
		case 'q':
			res.InternalQuery = fieldStr
		case 'S':
			res.Severity = fieldStr
		case 'M':
			res.Message = fieldStr
		case 'C':
			res.SQLState = fieldStr
		case 'D':
			res.Detail = fieldStr
		case 'H':
			res.Hint = fieldStr
		case 'P':
			res.Position = fieldStr
		case 'W':
			res.Where = fieldStr
		case 'p':
			res.InternalPosition = fieldStr
		case 'R':
			res.Routine = fieldStr
		case 'F':
			res.File = fieldStr
		case 'L':
			res.Line = fieldStr
		case 'V':
			res.ErrorCode = fieldStr
		}
	}

	return res, nil
}

func (b *BEErrorMsg) String() string {
	return fmt.Sprintf("ErrorResponse: %s %s: [%s] %s", b.Severity, b.ErrorCode, b.SQLState, b.Message)
}

func init() {
	registerBackEndMsgType('E', &BEErrorMsg{})
}
