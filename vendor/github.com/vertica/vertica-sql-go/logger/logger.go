package logger

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
	"os"
	"runtime"
)

// LogLevel is the enum type for the log levels.
type LogLevel int32

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	NONE
)

// Backend defines the public interface that must be implemented by all logger backends
type Backend interface {
	Write(prefix string, name string, msg string)
	Close()
}

var (
	prefixes         = []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	level            = WARN
	logger   Backend = &STDIOLogger{}
)

type Logger struct {
	name string
}

func (l *Logger) print(level LogLevel, format string, args ...interface{}) {
	logger.Write(prefixes[level], l.name, fmt.Sprintf(format, args...))
}

func New(name string) *Logger {
	return &Logger{name: name}
}

func (l *Logger) LineTrace() {
	if level == TRACE {
		_, file, line, ok := runtime.Caller(1)

		if !ok {
			l.Warn("unable to determine stack frame in LineTrace()")
		} else {
			l.Trace("%s(%d)", file, line)
		}
	}
}

func (l *Logger) Trace(format string, args ...interface{}) {
	if level == TRACE {
		l.print(TRACE, format, args...)
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if level <= DEBUG {
		l.print(DEBUG, format, args...)
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if level <= INFO {
		l.print(INFO, format, args...)
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	if level <= ERROR {
		l.print(ERROR, format, args...)
	}
}

func (l *Logger) Warn(format string, args ...interface{}) {
	if level <= WARN {
		l.print(WARN, format, args...)
	}
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	if level <= FATAL {
		l.print(FATAL, format, args...)
		os.Exit(1)
	}
}

func SetLogLevel(newLevel LogLevel) {
	level = newLevel
}

func SetLogger(newLogger Backend) {
	if logger != newLogger {
		logger.Close()
		logger = newLogger
	}
}
