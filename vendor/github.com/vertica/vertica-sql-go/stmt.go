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
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/vertica/vertica-sql-go/common"
	"github.com/vertica/vertica-sql-go/logger"
	"github.com/vertica/vertica-sql-go/msgs"
	"github.com/vertica/vertica-sql-go/parse"
)

var (
	stmtLogger = logger.New("stmt")
)

type parseState int

const (
	parseStateUnparsed parseState = iota
	parseStateParseError
	parseStateParsed
)

type stmt struct {
	conn         *connection
	command      string
	preparedName string
	parseState   parseState
	namedArgPos  []string
	posArgCnt    int
	paramTypes   []common.ParameterType
	lastRowDesc  *msgs.BERowDescMsg
	// set if Vertica issues an error of ROLLBACK severity
	rolledBack bool
}

func newStmt(connection *connection, command string) (*stmt, error) {

	if len(command) == 0 {
		return nil, fmt.Errorf("cannot create an empty statement")
	}

	s := &stmt{
		conn:         connection,
		preparedName: fmt.Sprintf("S%d%d%d", os.Getpid(), time.Now().Unix(), rand.Int31()),
		parseState:   parseStateUnparsed,
	}
	argCounter := func() string {
		s.posArgCnt++
		return "?"
	}
	s.command = parse.Lex(command, parse.WithNamedCallback(s.pushNamed), parse.WithPositionalSubstitution(argCounter))
	return s, nil
}

func (s *stmt) pushNamed(name string) {
	s.namedArgPos = append(s.namedArgPos, name)
}

// Close closes this statement.
func (s *stmt) Close() error {
	if s.parseState != parseStateParsed {
		return nil
	}
	if s.rolledBack {
		s.parseState = parseStateUnparsed
		s.conn.dead = true
		return nil
	}
	closeMsg := &msgs.FECloseMsg{TargetType: msgs.CmdTargetTypeStatement, TargetName: s.preparedName}

	if err := s.conn.sendMessage(closeMsg); err != nil {
		return err
	}

	if err := s.conn.sendMessage(&msgs.FEFlushMsg{}); err != nil {
		return err
	}

	for {
		bMsg, err := s.conn.recvMessage()

		if err != nil {
			return err
		}

		switch bMsg.(type) {
		case *msgs.BECloseCompleteMsg:
			s.parseState = parseStateUnparsed
			return nil
		case *msgs.BECmdDescriptionMsg:
			continue
		default:
			s.conn.defaultMessageHandler(bMsg)
		}
	}
}

// NumInput is used by database/sql to sanity check the number of arguments given
// before calling into the driver's query/exec functions. If named arguments are used
// this will return the number of unique named parameters, otherwise it is the number
// if ? placeholders.
func (s *stmt) NumInput() int {
	if len(s.namedArgPos) > 0 {
		uniqueArgs := make(map[string]bool)
		for _, arg := range s.namedArgPos {
			uniqueArgs[arg] = true
		}
		return len(uniqueArgs)
	}
	return s.posArgCnt
}

// convertToNamed takes an argument list of Value that come from the older Exec/Query functions
// and converts them to NamedValue to be forwarded to their Context equivalents.
func (s *stmt) convertToNamed(args []driver.Value) []driver.NamedValue {
	namedArgs := make([]driver.NamedValue, len(args))
	for idx, arg := range args {
		namedArgs[idx] = driver.NamedValue{
			Ordinal: idx,
			Value:   arg,
		}
	}
	return namedArgs
}

// injectNamedArgs takes a list of arguments, builds a symbol table of name => arg and then
// fills a list of positional arguments based on the names from the args parameter
// This will return an error if any of the given args lack a name
func (s *stmt) injectNamedArgs(args []driver.NamedValue) ([]driver.NamedValue, error) {
	if len(s.namedArgPos) == 0 {
		return args, nil
	}
	symbols := make(map[string]driver.NamedValue, len(args))
	for _, arg := range args {
		if len(arg.Name) > 0 {
			symbols[strings.ToUpper(arg.Name)] = arg
			continue
		}
		namedVal, ok := arg.Value.(driver.NamedValue)
		if !ok || len(namedVal.Name) == 0 {
			return nil, errors.New("all parameters must have names when using named parameters")
		}
		symbols[strings.ToUpper(namedVal.Name)] = namedVal
	}
	realArgs := make([]driver.NamedValue, len(s.namedArgPos))
	for pos, name := range s.namedArgPos {
		realArgs[pos] = symbols[name]
		realArgs[pos].Ordinal = pos
	}
	return realArgs, nil
}

// Exec docs
func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.ExecContext(context.Background(), s.convertToNamed(args))
}

// Query docs
func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	stmtLogger.Debug("stmt.Query(): %s\n", s.command)
	return s.QueryContext(context.Background(), s.convertToNamed(args))
}

// ExecContext docs
func (s *stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	stmtLogger.Trace("stmt.ExecContext()")

	rows, err := s.QueryContext(ctx, args)

	if err != nil {
		return driver.ResultNoRows, err
	}

	numCols := len(rows.Columns())
	vals := make([]driver.Value, numCols)

	if rows.Next(vals) == io.EOF {
		return driver.ResultNoRows, nil
	}

	rv := reflect.ValueOf(vals[0])

	return &result{lastInsertID: 0, rowsAffected: rv.Int()}, nil
}

func (s *stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return s.QueryContextRaw(ctx, args)
}

// QueryContext docs
func (s *stmt) QueryContextRaw(ctx context.Context, baseArgs []driver.NamedValue) (*rows, error) {
	stmtLogger.Debug("stmt.QueryContextRaw(): %s", s.command)

	var cmd string
	var err error
	var portalName string

	args, err := s.injectNamedArgs(baseArgs)
	if err != nil {
		return newEmptyRows(), err
	}

	doneChan := make(chan bool, 1)
	go func(pid, key uint32) {
		select {
		case <-doneChan:
			return
		case <-ctx.Done():
			stmtLogger.Info("Context cancelled, cancelling %s", s.preparedName)
			cancelMsg := msgs.FECancelMsg{PID: pid, Key: key}
			conn, err := s.conn.establishSocketConnection()
			if err != nil {
				stmtLogger.Warn("unable to establish connection for cancellation")
				return
			}
			conn.SetDeadline(time.Now().Add(10 * time.Second))
			if err := s.conn.sendMessageTo(&cancelMsg, conn); err != nil {
				stmtLogger.Warn("unable to send cancel message: %v", err)
			}
			if err := conn.Close(); err != nil {
				stmtLogger.Warn("error closing cancel connection: %v", err)
			}
			stmtLogger.Info("Cancelled %s", s.preparedName)
		}
	}(s.conn.backendPID, s.conn.cancelKey)

	s.conn.lockSessionMutex()
	defer s.conn.unlockSessionMutex()
	defer func() {
		doneChan <- true
	}()

	// If we have a prepared statement, go through bind/execute() phases instead.
	if s.parseState == parseStateParsed {
		if err = s.bindAndExecute(portalName, args); err != nil {
			return newEmptyRows(), err
		}

		return s.collectResults(ctx)
	}

	rows := newEmptyRows()

	// We aren't a prepared statement, manually interpolate and do as a simple query.
	cmd, err = s.interpolate(args)

	if err != nil {
		return rows, err
	}

	if err = s.conn.sendMessage(&msgs.FEQueryMsg{Query: cmd}); err != nil {
		return rows, err
	}

	for {
		bMsg, err := s.conn.recvMessage()

		if err != nil {
			return newEmptyRows(), err
		}

		switch msg := bMsg.(type) {
		case *msgs.BEDataRowMsg:
			err = rows.addRow(msg)
			if err != nil {
				return rows, err
			}
		case *msgs.BERowDescMsg:
			rows = newRows(ctx, msg, s.conn.serverTZOffset)
		case *msgs.BECmdCompleteMsg:
			break
		case *msgs.BEErrorMsg:
			return newEmptyRows(), s.evaluateErrorMsg(msg)
		case *msgs.BEEmptyQueryResponseMsg:
			return newEmptyRows(), nil
		case *msgs.BEReadyForQueryMsg, *msgs.BEPortalSuspendedMsg:
			err = rows.finalize()
			if err != nil {
				return rows, err
			}
			return rows, ctx.Err()
		case *msgs.BEInitSTDINLoadMsg:
			s.copySTDIN(ctx)
		default:
			s.conn.defaultMessageHandler(bMsg)
		}
	}
}

func (s *stmt) copySTDIN(ctx context.Context) {

	var streamToUse io.Reader
	streamToUse = os.Stdin

	var copyBlockSize = stdInDefaultCopyBlockSize

	if vCtx, ok := ctx.(VerticaContext); ok {
		streamToUse = vCtx.GetCopyInputStream()
		copyBlockSize = vCtx.GetCopyBlockSizeBytes()
	}

	block := make([]byte, copyBlockSize)
	for {
		bytesRead, err := streamToUse.Read(block)
		if err == io.EOF {
			s.conn.sendMessage(&msgs.FELoadDoneMsg{})
			break
		}
		if err != nil {
			s.conn.sendMessage(&msgs.FELoadFailMsg{Message: err.Error()})
			break
		}
		s.conn.sendMessage(&msgs.FELoadDataMsg{Data: block, UsedBytes: bytesRead})
	}
	s.conn.sendMessage(&msgs.FEFlushMsg{})
}

func (s *stmt) cleanQuotes(val string) string {
	re := regexp.MustCompile(`'+`)
	pairs := re.FindAllStringIndex(val, -1)
	if pairs == nil {
		return val
	}
	cleaned := strings.Builder{}
	cleaned.Grow(len(val))
	cleanedTo := 0
	for _, matchPair := range pairs {
		if (matchPair[1]-matchPair[0])%2 != 0 {
			cleaned.WriteString(val[cleanedTo:matchPair[1]])
			cleaned.WriteRune('\'')
			cleanedTo = matchPair[1]
		}
	}
	cleaned.WriteString(val[cleanedTo:])
	return cleaned.String()
}

func (s *stmt) formatArg(arg driver.NamedValue) string {
	var replaceStr string
	switch v := arg.Value.(type) {
	case nil:
		replaceStr = "NULL"
	case int64, float64:
		replaceStr = fmt.Sprintf("%v", v)
	case string:
		replaceStr = fmt.Sprintf("'%s'", s.cleanQuotes(v))
	case bool:
		if v {
			replaceStr = "true"
		} else {
			replaceStr = "false"
		}
	case time.Time:
		replaceStr = fmt.Sprintf("'%02d-%02d-%02d %02d:%02d:%02d.%09d'",
			v.Year(),
			v.Month(),
			v.Day(),
			v.Hour(),
			v.Minute(),
			v.Second(),
			v.Nanosecond())
	default:
		replaceStr = "?unknown_type?"
	}
	return replaceStr
}

func (s *stmt) interpolate(args []driver.NamedValue) (string, error) {

	numArgs := s.NumInput()

	if numArgs == 0 {
		return s.command, nil
	}

	curArg := 0
	argSwapper := func() string {
		arg := s.formatArg(args[curArg])
		curArg++
		return arg
	}

	result := parse.Lex(s.command, parse.WithPositionalSubstitution(argSwapper))
	return result, nil
}

func (s *stmt) evaluateErrorMsg(msg *msgs.BEErrorMsg) error {
	if msg.Severity == "ROLLBACK" {
		s.rolledBack = true
	}
	return errorMsgToVError(msg)
}

func (s *stmt) prepareAndDescribe() error {

	parseMsg := &msgs.FEParseMsg{
		PreparedName: s.preparedName,
		Command:      s.command,
		NumArgs:      0,
	}

	// If we've already been parsed, no reason to do it again.
	if s.parseState == parseStateParsed {
		return nil
	}

	s.parseState = parseStateParseError

	s.conn.lockSessionMutex()
	defer s.conn.unlockSessionMutex()

	if err := s.conn.sendMessage(parseMsg); err != nil {
		return err
	}

	describeMsg := &msgs.FEDescribeMsg{TargetType: msgs.CmdTargetTypeStatement, TargetName: s.preparedName}

	if err := s.conn.sendMessage(describeMsg); err != nil {
		return err
	}

	if err := s.conn.sendMessage(&msgs.FEFlushMsg{}); err != nil {
		return err
	}

	for {
		bMsg, err := s.conn.recvMessage()

		if err != nil {
			return err
		}

		switch msg := bMsg.(type) {
		case *msgs.BEErrorMsg:
			s.conn.sync()
			return errorMsgToVError(msg)
		case *msgs.BEParseCompleteMsg:
			s.parseState = parseStateParsed
		case *msgs.BERowDescMsg:
			s.lastRowDesc = msg
			return nil
		case *msgs.BENoDataMsg:
			s.lastRowDesc = nil
			return nil
		case *msgs.BEParameterDescMsg:
			s.paramTypes = msg.ParameterTypes
		case *msgs.BECmdDescriptionMsg:
			continue
		default:
			s.conn.defaultMessageHandler(msg)
		}
	}
}

func (s *stmt) bindAndExecute(portalName string, args []driver.NamedValue) error {

	// We only need to send the OID types
	paramOIDs := make([]int32, len(s.paramTypes))
	for i, p := range s.paramTypes {
		paramOIDs[i] = int32(p.TypeOID)
	}

	if err := s.conn.sendMessage(&msgs.FEBindMsg{Portal: portalName, Statement: s.preparedName, NamedArgs: args, OIDTypes: paramOIDs}); err != nil {
		return err
	}

	if err := s.conn.sendMessage(&msgs.FEExecuteMsg{Portal: portalName}); err != nil {
		return err
	}

	if err := s.conn.sendMessage(&msgs.FEFlushMsg{}); err != nil {
		return err
	}

	return nil
}

func (s *stmt) collectResults(ctx context.Context) (*rows, error) {
	rows := newEmptyRows()

	if s.lastRowDesc != nil {
		rows = newRows(ctx, s.lastRowDesc, s.conn.serverTZOffset)
	}

	for {
		bMsg, err := s.conn.recvMessage()

		if err != nil {
			return newEmptyRows(), err
		}

		switch msg := bMsg.(type) {
		case *msgs.BEDataRowMsg:
			err = rows.addRow(msg)
			if err != nil {
				return rows, err
			}
		case *msgs.BERowDescMsg:
			s.lastRowDesc = msg
			rows = newRows(ctx, s.lastRowDesc, s.conn.serverTZOffset)
		case *msgs.BEErrorMsg:
			s.conn.sync()
			return newEmptyRows(), s.evaluateErrorMsg(msg)
		case *msgs.BEEmptyQueryResponseMsg:
			return newEmptyRows(), nil
		case *msgs.BEBindCompleteMsg, *msgs.BECmdDescriptionMsg:
			continue
		case *msgs.BEReadyForQueryMsg, *msgs.BEPortalSuspendedMsg, *msgs.BECmdCompleteMsg:
			err = rows.finalize()
			if err != nil {
				return rows, err
			}
			return rows, ctx.Err()
		case *msgs.BEInitSTDINLoadMsg:
			s.copySTDIN(ctx)
		default:
			_, _ = s.conn.defaultMessageHandler(msg)
		}
	}
}
