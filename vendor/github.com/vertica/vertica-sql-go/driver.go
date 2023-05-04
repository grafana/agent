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
	"database/sql"
	"database/sql/driver"
	"os"
	"strconv"

	"github.com/vertica/vertica-sql-go/logger"
)

// Driver as defined by the Go language Driver interface
type Driver struct{}

const (
	driverName      string = "vertica-sql-go"
	driverVersion   string = "1.3.0"
	protocolVersion uint32 = 0x00030009
)

var driverLogger = logger.New("driver")

// Open takes a connection string in this format:
// user:pass@host:port/database
func (d *Driver) Open(connString string) (driver.Conn, error) {
	conn, err := newConnection(connString)
	if err != nil {
		driverLogger.Error(err.Error())
	}
	return conn, err
}

// Register ourselves with the sql package.
func init() {
	logger.SetLogLevel(logger.WARN)

	if logLevel := os.Getenv("VERTICA_SQL_GO_LOG_LEVEL"); logLevel != "" {
		logVal, err := strconv.ParseUint(logLevel, 10, 32)
		if err != nil {
			driverLogger.Error(err.Error())
		} else {
			logFlag := logger.WARN
			switch logVal {
			case 0:
				logFlag = logger.TRACE
			case 1:
				logFlag = logger.DEBUG
			case 2:
				logFlag = logger.INFO
			case 3:
				logFlag = logger.WARN
			case 4:
				logFlag = logger.ERROR
			case 5:
				logFlag = logger.FATAL
			case 6:
				logFlag = logger.NONE
			default:
				driverLogger.Error("invalid VERTICA_SQL_GO_LOG_LEVEL value; should be 0-6")
			}
			logger.SetLogLevel(logFlag)
		}
	}

	if logFile := os.Getenv("VERTICA_SQL_GO_LOG_FILE"); logFile != "" {
		if loggerBackend, err := logger.NewFileLogger(logFile); err == nil {
			logger.SetLogger(loggerBackend)
		} else {
			driverLogger.Error("unable to create file logger: %v", err)
		}
	}

	sql.Register("vertica", &Driver{})
}
