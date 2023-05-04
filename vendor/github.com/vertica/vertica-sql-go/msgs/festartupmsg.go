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
	"os/user"

	"github.com/elastic/go-sysinfo"
)

// FEStartupMsg docs
type FEStartupMsg struct {
	ProtocolVersion uint32
	DriverName      string
	DriverVersion   string
	Username        string
	Database        string
	SessionID       string
	ClientPID       int
	ClientOS        string
	OSUsername      string
}

// Flatten docs
func (m *FEStartupMsg) Flatten() ([]byte, byte) {

	m.ClientOS = ""
	host, err := sysinfo.Host()
	if err == nil {
		info := host.Info()
		m.ClientOS = fmt.Sprintf("%s %s %s", info.OS.Name, info.KernelVersion, info.Architecture)
	}

	m.OSUsername = ""
	currentUser, err := user.Current()
	if err == nil {
		m.OSUsername = currentUser.Username
	}

	buf := newMsgBuffer()
	const fixedProtocolVersion uint32 = 0x00030005
	buf.appendUint32(fixedProtocolVersion)

	// Requested protocol version
	buf.appendString("protocol_version")
	buf.appendUint32(m.ProtocolVersion)
	buf.appendBytes([]byte{0})

	if len(m.Username) > 0 {
		buf.appendLabeledString("user", m.Username)
	}

	if len(m.Database) > 0 {
		buf.appendLabeledString("database", m.Database)
	}

	buf.appendLabeledString("client_type", m.DriverName)
	buf.appendLabeledString("client_version", m.DriverVersion)
	buf.appendLabeledString("client_label", m.SessionID)
	buf.appendLabeledString("client_pid", fmt.Sprintf("%d", m.ClientPID))
	buf.appendLabeledString("client_os", m.ClientOS)
	buf.appendLabeledString("client_os_user_name", m.OSUsername)
	buf.appendBytes([]byte{0})

	return buf.bytes(), 0
}

func (m *FEStartupMsg) String() string {
	return fmt.Sprintf(
		"Startup (packet): ProtocolVersion:%08X, DriverName='%s', DriverVersion='%s', UserName='%s', Database='%s', SessionID='%s', ClientPID=%d, ClientOS='%s', ClientOSUserName='%s'",
		m.ProtocolVersion,
		m.DriverName,
		m.DriverVersion,
		m.Username,
		m.Database,
		m.SessionID,
		m.ClientPID,
		m.ClientOS,
		m.OSUsername)
}
