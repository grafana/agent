package vertigo

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
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/vertica/vertica-sql-go/common"
	"github.com/vertica/vertica-sql-go/logger"
	"github.com/vertica/vertica-sql-go/msgs"
	"github.com/vertica/vertica-sql-go/rowcache"
)

type rowStore interface {
	AddRow(msg *msgs.BEDataRowMsg) error
	GetRow() *msgs.BEDataRowMsg
	Peek() *msgs.BEDataRowMsg
	Close() error
	Finalize() error
}

type rows struct {
	columnDefs *msgs.BERowDescMsg
	resultData rowStore

	tzOffset      string
	inMemRowLimit int
}

var (
	paddingString        = "000000"
	defaultRowBufferSize = 256
	rowLogger            = logger.New("row")
	endsWithHalfHour     = regexp.MustCompile(".*:\\d{2}$")
)

// Columns returns the names of all of the columns
// Interface: driver.Rows
func (r *rows) Columns() []string {
	columnLabels := make([]string, len(r.columnDefs.Columns))
	for idx, cd := range r.columnDefs.Columns {
		columnLabels[idx] = cd.FieldName
	}
	return columnLabels
}

// Close closes the read cursor
// Interface: driver.Rows
func (r *rows) Close() error {
	return r.resultData.Close()
}

func (r *rows) Next(dest []driver.Value) error {
	var err error
	nextRow := r.resultData.GetRow()
	if nextRow == nil {
		return io.EOF
	}

	rowCols := nextRow.Columns()

	for idx := uint16(0); idx < rowCols.NumCols; idx++ {
		colVal := rowCols.Chunk()
		if colVal == nil {
			dest[idx] = nil
			continue
		}

		switch r.columnDefs.Columns[idx].DataTypeOID {
		case common.ColTypeBoolean: // to boolean
			dest[idx] = colVal[0] == 't'
		case common.ColTypeInt64: // to integer
			dest[idx], err = strconv.Atoi(string(colVal))
		case common.ColTypeVarChar, common.ColTypeLongVarChar, common.ColTypeChar, common.ColTypeUUID: // stays string, convert char to string
			dest[idx] = string(colVal)
		case common.ColTypeFloat64, common.ColTypeNumeric: // to float64
			dest[idx], err = strconv.ParseFloat(string(colVal), 64)
		case common.ColTypeDate: // to time.Time from YYYY-MM-DD
			dest[idx], err = parseDateColumn(string(colVal))
		case common.ColTypeTimestamp: // to time.Time from YYYY-MM-DD hh:mm:ss
			dest[idx], err = parseTimestampTZColumn(string(colVal) + r.tzOffset)
		case common.ColTypeTimestampTZ:
			dest[idx], err = parseTimestampTZColumn(string(colVal))
		case common.ColTypeTime: // to time.Time from hh:mm:ss.[fff...]
			dest[idx], err = parseTimestampTZColumn("0000-01-01 " + string(colVal) + r.tzOffset)
		case common.ColTypeTimeTZ:
			dest[idx], err = parseTimestampTZColumn("0000-01-01 " + string(colVal))
		case common.ColTypeInterval, common.ColTypeIntervalYM: // stays string
			dest[idx] = string(colVal)
		case common.ColTypeVarBinary, common.ColTypeLongVarBinary, common.ColTypeBinary:
			// to []byte; convert escaped octal (e.g. \261) into byte with \\ for \
			var out []byte
			for len(colVal) > 0 {
				c := colVal[0]
				if c == '\\' {
					if colVal[1] == '\\' { // escaped \
						colVal = colVal[2:]
					} else { // A \xxx octal string
						x, _ := strconv.ParseInt(string(colVal[1:4]), 8, 32)
						c = byte(x)
						colVal = colVal[4:]
					}
				} else {
					colVal = colVal[1:]
				}
				out = append(out, c)
			}
			dest[idx] = out
		default:
			dest[idx] = string(colVal)
		}

		if err != nil {
			rowLogger.Error("%s", err.Error())
		}
	}

	return err
}

func parseDateColumn(fullString string) (driver.Value, error) {
	var result driver.Value
	var err error

	// Dates Before Christ (YYYY-MM-DD BC) are special
	if strings.HasSuffix(fullString, " BC") {
		var t time.Time
		t, err = time.Parse("2006-01-02 BC", fullString)
		if err != nil {
			return time.Time{}, err
		}
		result = t.AddDate(-2*t.Year(), 0, 0)
	} else {
		result, err = time.Parse("2006-01-02", fullString)
	}
	return result, err
}

func parseTimestampTZColumn(fullString string) (driver.Value, error) {
	var result driver.Value
	var err error
	var isBC bool

	// +infinity or -infinity value
	if strings.Contains(fullString, "infinity") {
		return time.Time{}, fmt.Errorf("cannot parse an infinity timestamp to time.Time")
	}

	if isBC = strings.Contains(fullString, " BC"); isBC {
		fullString = strings.Replace(fullString, " BC", "", -1)
	}

	if !endsWithHalfHour.MatchString(fullString) {
		fullString = fullString + ":00"
	}

	// ensures ms are included with the desired length
	if strings.IndexByte(fullString, '.') == 19 {
		neededPadding := 32 - len(fullString)
		if neededPadding > 0 {
			fullString = fullString[0:26-neededPadding] + paddingString[0:neededPadding] + fullString[26-neededPadding:]
		}
	} else {
		fullString = fullString[0:19] + "." + paddingString[0:6] + fullString[19:]
	}

	// Note: The date/time output format for the current session (sql=> SHOW DATESTYLE) should be 'ISO'
	result, err = time.Parse("2006-01-02 15:04:05.000000-07:00", fullString)
	if isBC {
		result = result.(time.Time).AddDate(-2*result.(time.Time).Year(), 0, 0)
	}

	return result, err
}

func (r *rows) finalize() error {
	return r.resultData.Finalize()
}

func (r *rows) addRow(rowData *msgs.BEDataRowMsg) error {
	return r.resultData.AddRow(rowData)
}

func newRows(ctx context.Context, columnsDefsMsg *msgs.BERowDescMsg, tzOffset string) *rows {

	rowBufferSize := defaultRowBufferSize
	inMemRowLimit := 0
	var resultData rowStore
	var err error

	if vCtx, ok := ctx.(VerticaContext); ok {
		rowBufferSize = vCtx.GetInMemoryResultRowLimit()
		inMemRowLimit = rowBufferSize
	}
	if inMemRowLimit != 0 {
		resultData, err = rowcache.NewFileCache(inMemRowLimit)
		if err != nil {
			resultData = rowcache.NewMemoryCache(rowBufferSize)
		}
	} else {
		resultData = rowcache.NewMemoryCache(rowBufferSize)
	}

	res := &rows{
		columnDefs:    columnsDefsMsg,
		resultData:    resultData,
		tzOffset:      tzOffset,
		inMemRowLimit: inMemRowLimit,
	}

	return res
}

// Returns the database system type name without the length. Type names should be uppercase.
// Interface: driver.RowsColumnTypeDatabaseTypeName
func (r *rows) ColumnTypeDatabaseTypeName(index int) string {
	return r.columnDefs.Columns[index].DataTypeName
}

// The nullable value should be true if it is known the column may be null, or false if the column
// is known to be not nullable. The ok value should always be true as column nullability is known.
// Interface: driver.RowsColumnTypeNullable
func (r *rows) ColumnTypeNullable(index int) (nullable, ok bool) {
	return r.columnDefs.Columns[index].Nullable, true
}

// Returns the precision and scale for column types. If not applicable, ok should be false.
// Interface: driver.RowsColumnTypePrecisionScale
func (r *rows) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	// The type modifier of -1 is used when the size of a type is unknown.
	// In those cases we assume the maximum possible size.
	var typeMod = int64(r.columnDefs.Columns[index].DataTypeMod)
	switch r.columnDefs.Columns[index].DataTypeOID {
	case common.ColTypeNumeric:
		// For numerics, precision is the total number of digits (in base 10) that can fit in the type
		if typeMod == -1 {
			return 1024, 15, true
		} else {
			precision := ((typeMod - 4) >> 16) & 0xFFFF
			scale := (typeMod - 4) & 0xFF
			return precision, scale, true
		}
	case common.ColTypeTime, common.ColTypeTimeTZ, common.ColTypeTimestamp,
		common.ColTypeTimestampTZ, common.ColTypeInterval, common.ColTypeIntervalYM:
		// For intervals, time and timestamps, precision is the number of digits to the
		// right of the decimal point in the seconds portion of the time.
		if typeMod == -1 {
			return 6, 0, true
		} else {
			return typeMod & 0xF, 0, true
		}
	default:
		return 0, 0, false
	}
}

// Returns the length of the column type. If the column is not a variable length type ok should
// return false.
// Interface: driver.RowsColumnTypeLength
func (r *rows) ColumnTypeLength(index int) (length int64, ok bool) {
	var typeMod = int64(r.columnDefs.Columns[index].DataTypeMod)
	switch r.columnDefs.Columns[index].DataTypeOID {
	case common.ColTypeBoolean, common.ColTypeInt64, common.ColTypeFloat64,
		common.ColTypeDate, common.ColTypeTimestamp, common.ColTypeTimestampTZ,
		common.ColTypeTime, common.ColTypeTimeTZ, common.ColTypeInterval,
		common.ColTypeIntervalYM, common.ColTypeUUID:
		return int64(r.columnDefs.Columns[index].Length), false
	case common.ColTypeChar, common.ColTypeVarChar,
		common.ColTypeBinary, common.ColTypeVarBinary:
		if typeMod == -1 {
			return 65000, true
		} else {
			return typeMod - 4, true
		}
	case common.ColTypeLongVarChar, common.ColTypeLongVarBinary:
		if typeMod == -1 {
			return 32000000, true
		} else {
			return typeMod - 4, true
		}
	case common.ColTypeNumeric:
		precision, _, _ := r.ColumnTypePrecisionScale(index)
		return (precision/19 + 1) * 8, true
	default:
		return 0, false
	}
}

// Returns the value type that can be used to scan types into.
// Interface: driver.RowsColumnTypeScanType
func (r *rows) ColumnTypeScanType(index int) reflect.Type {
	switch r.columnDefs.Columns[index].DataTypeOID {
	case common.ColTypeBoolean:
		return reflect.TypeOf(sql.NullBool{})
	case common.ColTypeInt64:
		return reflect.TypeOf(sql.NullInt64{})
	case common.ColTypeFloat64, common.ColTypeNumeric:
		return reflect.TypeOf(sql.NullFloat64{})
	case common.ColTypeVarChar, common.ColTypeLongVarChar, common.ColTypeChar,
		common.ColTypeVarBinary, common.ColTypeLongVarBinary, common.ColTypeBinary,
		common.ColTypeUUID, common.ColTypeInterval, common.ColTypeIntervalYM:
		return reflect.TypeOf(sql.NullString{})
	case common.ColTypeDate, common.ColTypeTimestamp, common.ColTypeTimestampTZ,
		common.ColTypeTime, common.ColTypeTimeTZ:
		return reflect.TypeOf(sql.NullTime{})
	default:
		return reflect.TypeOf(new(interface{}))
	}
}

func newEmptyRows() *rows {
	cdf := make([]*msgs.BERowDescColumnDef, 0)
	be := &msgs.BERowDescMsg{Columns: cdf}
	return newRows(context.Background(), be, "")
}
