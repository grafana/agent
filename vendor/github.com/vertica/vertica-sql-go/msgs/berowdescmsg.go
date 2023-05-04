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

// BERowDescColumnDef docs
type BERowDescColumnDef struct {
	FieldName    string
	SchemaName   string
	TableName    string
	TableOID     int64
	Nullable     bool
	IsIdentity   bool
	AttribNum    int16
	DataTypeOID  uint32
	DataTypeName string
	Length       int16
	DataTypeMod  int32
	FormatCode   uint16
}

type userType struct {
	BaseTypeOID  uint32
	DataTypeName string
}

// BERowDescMsg docs
type BERowDescMsg struct {
	Columns []*BERowDescColumnDef
}

// CreateFromMsgBody docs
func (m *BERowDescMsg) CreateFromMsgBody(buf *msgBuffer) (BackEndMsg, error) {

	res := &BERowDescMsg{}

	numCols := buf.readInt16()
	res.Columns = make([]*BERowDescColumnDef, numCols)

	numTypes := buf.readInt32()
	userTypes := make([]userType, numTypes)

	// Read User Types
	for i := int32(0); i < numTypes; i++ {
		userTypes[i].BaseTypeOID = buf.readUint32()
		userTypes[i].DataTypeName = buf.readString()
	}

	// Read Column Descriptions
	for i := int16(0); i < numCols; i++ {

		col := &BERowDescColumnDef{}

		col.FieldName = buf.readString()
		col.TableOID = buf.readInt64()

		if col.TableOID != 0 {
			col.SchemaName = buf.readString()
			col.TableName = buf.readString()
		}

		col.AttribNum = buf.readInt16()

		isUserType := buf.readBool()
		dataTypeID := buf.readUint32()

		col.Length = buf.readInt16()
		col.Nullable = buf.readInt16() == 1
		col.IsIdentity = buf.readInt16() == 1
		col.DataTypeMod = buf.readInt32()
		col.FormatCode = buf.readUint16()

		if isUserType {
			col.DataTypeOID = userTypes[dataTypeID].BaseTypeOID
			col.DataTypeName = userTypes[dataTypeID].DataTypeName
		} else {
			col.DataTypeOID = dataTypeID
			col.DataTypeName = common.ColumnTypeString(dataTypeID, col.DataTypeMod)
		}

		res.Columns[i] = col
	}

	return res, nil
}

func (m *BERowDescMsg) String() string {
	return fmt.Sprintf("RowDesc: %d column(s)", len(m.Columns))
}

func init() {
	registerBackEndMsgType('T', &BERowDescMsg{})
}
