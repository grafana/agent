// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helper // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/errors"
)

// ExprStringConfig is a string that represents an expression
type ExprStringConfig string

const (
	exprStartToken = "EXPR("
	exprEndToken   = ")"
)

// Build creates an ExprStr string from the specified config
func (e ExprStringConfig) Build() (*ExprString, error) {
	s := string(e)
	rangeStart := 0

	subStrings := make([]string, 0, 4)
	subExprStrings := make([]string, 0, 4)

	for {
		rangeEnd := len(s)

		// Find the first instance of the start token
		indexStart := strings.Index(s[rangeStart:rangeEnd], exprStartToken)
		if indexStart == -1 {
			// Start token does not exist in the remainder of the string,
			// so treat the rest as a string literal
			subStrings = append(subStrings, s[rangeStart:])
			break
		} else {
			indexStart = rangeStart + indexStart
		}

		// Restrict our end token search range to the next instance of the start token
		nextIndexStart := strings.Index(s[indexStart+len(exprStartToken):], exprStartToken)
		if nextIndexStart == -1 {
			rangeEnd = len(s)
		} else {
			rangeEnd = indexStart + len(exprStartToken) + nextIndexStart
		}

		// Greedily search for the last end token in the search range
		indexEnd := strings.LastIndex(s[indexStart:rangeEnd], exprEndToken)
		if indexEnd == -1 {
			// End token does not exist before the next start token
			// or end of expression string, so treat the remainder of the string
			// as a string literal
			subStrings = append(subStrings, s[rangeStart:])
			break
		} else {
			indexEnd = indexStart + indexEnd
		}

		// Unscope the indexes and add the partitioned strings
		subStrings = append(subStrings, s[rangeStart:indexStart])
		subExprStrings = append(subExprStrings, s[indexStart+len(exprStartToken):indexEnd])

		// Reset the starting range and finish if it reaches the end of the string
		rangeStart = indexEnd + len(exprEndToken)
		if rangeStart > len(s) {
			break
		}
	}

	subExprs := make([]*vm.Program, 0, len(subExprStrings))
	for _, subExprString := range subExprStrings {
		program, err := expr.Compile(subExprString, expr.AllowUndefinedVariables())
		if err != nil {
			return nil, errors.Wrap(err, "compile embedded expression")
		}
		subExprs = append(subExprs, program)
	}

	return &ExprString{
		SubStrings: subStrings,
		SubExprs:   subExprs,
	}, nil
}

// An ExprString is made up of a list of string literals
// interleaved with expressions. len(SubStrings) == len(SubExprs) + 1
type ExprString struct {
	SubStrings []string
	SubExprs   []*vm.Program
}

// Render will render an ExprString as a string
func (e *ExprString) Render(env map[string]interface{}) (string, error) {
	var b strings.Builder
	for i := 0; i < len(e.SubExprs); i++ {
		b.WriteString(e.SubStrings[i])
		out, err := vm.Run(e.SubExprs[i], env)
		if err != nil {
			return "", errors.Wrap(err, "render embedded expression")
		}
		outString, ok := out.(string)
		if !ok {
			return "", fmt.Errorf("embedded expression returned non-string %v", out)
		}
		b.WriteString(outString)
	}
	b.WriteString(e.SubStrings[len(e.SubStrings)-1])

	return b.String(), nil
}

var envPool = sync.Pool{
	New: func() interface{} {
		return map[string]interface{}{
			"env": os.Getenv,
		}
	},
}

// GetExprEnv returns a map of key/value pairs that can be be used to evaluate an expression
func GetExprEnv(e *entry.Entry) map[string]interface{} {
	env := envPool.Get().(map[string]interface{})
	env["$"] = e.Body
	env["body"] = e.Body
	env["attributes"] = e.Attributes
	env["resource"] = e.Resource
	env["timestamp"] = e.Timestamp

	return env
}

// PutExprEnv adds a key/value pair that will can be used to evaluate an expression
func PutExprEnv(e map[string]interface{}) {
	envPool.Put(e)
}
