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

// CmdTargetType describes the target of a command
type CmdTargetType byte

// Possible command targets
const (
	CmdTargetTypePortal    CmdTargetType = 'P'
	CmdTargetTypeStatement CmdTargetType = 'S'
)

// FrontEndMsg is sent from the adapter to the database
type FrontEndMsg interface {
	Flatten() ([]byte, byte)
	String() string
}

// BackEndMsg is received from the database
type BackEndMsg interface {
	CreateFromMsgBody(*msgBuffer) (BackEndMsg, error)
	String() string
}

// backEndMsgTypeMap is a global map of message descriptor bytes to instances
// of that message. The instances are not used directly, but instead are used to
// construct new values of that message type. This is populated on init.
var backEndMsgTypeMap = make(map[byte]BackEndMsg)

func registerBackEndMsgType(msgType byte, bem BackEndMsg) {
	backEndMsgTypeMap[msgType] = bem
}

// CreateBackEndMsg docs
func CreateBackEndMsg(msgType byte, body []byte) (BackEndMsg, error) {
	if bem, ok := backEndMsgTypeMap[msgType]; ok {
		buffer := NewMsgBufferFromBytes(body)
		newMsg, err := bem.CreateFromMsgBody(buffer)

		if err != nil {
			return nil, err
		}

		if bytesLeft := buffer.remainingBytes(); bytesLeft > 0 {
			return nil, fmt.Errorf("error creating message of type '%c': %d byte(s) remaining", msgType, bytesLeft)
		}

		return newMsg, nil
	}

	return nil, fmt.Errorf("unsupported backend msg type: %c", msgType)
}
