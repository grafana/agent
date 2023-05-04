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

	"github.com/vertica/vertica-sql-go/common"
)

type customType struct {
	TypeOID  uint32
	TypeName string
}

//
type BEParameterDescMsg struct {
	ParameterTypes []common.ParameterType
}

// CreateFromMsgBody docs
func (m *BEParameterDescMsg) CreateFromMsgBody(buf *msgBuffer) (BackEndMsg, error) {

	res := &BEParameterDescMsg{}

	// Read in the number of parameters
	numParams := buf.readInt16()
	res.ParameterTypes = make([]common.ParameterType, numParams)

	// Get number of custom types
	numTypes := buf.readInt32()
	customTypes := make([]customType, numTypes)

	// Read in those types
	for i := int32(0); i < numTypes; i++ {
		customTypes[i].TypeOID = buf.readUint32()
		customTypes[i].TypeName = buf.readString()
	}

	// Now decipher the list of parameters
	for i := int16(0); i < numParams; i++ {
		isUserType := buf.readBool()
		typeOIDOrIndex := buf.readUint32()
		typeModifier := buf.readInt32()
		nullOK := buf.readInt16()

		if isUserType {
			res.ParameterTypes[i].TypeOID = customTypes[typeOIDOrIndex].TypeOID
			res.ParameterTypes[i].TypeName = customTypes[typeOIDOrIndex].TypeName
		} else {
			res.ParameterTypes[i].TypeOID = typeOIDOrIndex
			res.ParameterTypes[i].TypeName = common.ColumnTypeString(typeOIDOrIndex, typeModifier)
		}

		res.ParameterTypes[i].TypeModifier = typeModifier
		res.ParameterTypes[i].Nullable = nullOK == 1

	}

	return res, nil
}

func (m *BEParameterDescMsg) String() string {
	return fmt.Sprintf("ParameterDesc: %d parameter(s) described: %v", len(m.ParameterTypes), m.ParameterTypes)
}

func init() {
	registerBackEndMsgType('t', &BEParameterDescMsg{})
}
