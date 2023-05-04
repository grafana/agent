// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package keyvalue // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/parser/keyvalue"

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
)

const operatorType = "key_value_parser"

func init() {
	operator.Register(operatorType, func() operator.Builder { return NewConfig() })
}

// NewConfig creates a new key value parser config with default values
func NewConfig() *Config {
	return NewConfigWithID(operatorType)
}

// NewConfigWithID creates a new key value parser config with default values
func NewConfigWithID(operatorID string) *Config {
	return &Config{
		ParserConfig: helper.NewParserConfig(operatorID, operatorType),
		Delimiter:    "=",
	}
}

// Config is the configuration of a key value parser operator.
type Config struct {
	helper.ParserConfig `mapstructure:",squash"`

	Delimiter     string `mapstructure:"delimiter"`
	PairDelimiter string `mapstructure:"pair_delimiter"`
}

// Build will build a key value parser operator.
func (c Config) Build(logger *zap.SugaredLogger) (operator.Operator, error) {
	parserOperator, err := c.ParserConfig.Build(logger)
	if err != nil {
		return nil, err
	}

	if c.Delimiter == c.PairDelimiter {
		return nil, errors.New("delimiter and pair_delimiter cannot be the same value")
	}

	if c.Delimiter == "" {
		return nil, errors.New("delimiter is a required parameter")
	}

	// split on whitespace by default, if pair delimiter is set, use
	// strings.Split()
	pairSplitFunc := splitStringByWhitespace
	if c.PairDelimiter != "" {
		pairSplitFunc = func(input string) []string {
			return strings.Split(input, c.PairDelimiter)
		}
	}

	return &Parser{
		ParserOperator: parserOperator,
		delimiter:      c.Delimiter,
		pairSplitFunc:  pairSplitFunc,
	}, nil
}

// Parser is an operator that parses key value pairs.
type Parser struct {
	helper.ParserOperator
	delimiter     string
	pairSplitFunc func(input string) []string
}

// Process will parse an entry for key value pairs.
func (kv *Parser) Process(ctx context.Context, entry *entry.Entry) error {
	return kv.ParserOperator.ProcessWith(ctx, entry, kv.parse)
}

// parse will parse a value as key values.
func (kv *Parser) parse(value interface{}) (interface{}, error) {
	switch m := value.(type) {
	case string:
		return kv.parser(m, kv.delimiter)
	default:
		return nil, fmt.Errorf("type %T cannot be parsed as key value pairs", value)
	}
}

func (kv *Parser) parser(input string, delimiter string) (map[string]interface{}, error) {
	if input == "" {
		return nil, fmt.Errorf("parse from field %s is empty", kv.ParseFrom.String())
	}

	parsed := make(map[string]interface{})

	var err error
	for _, raw := range kv.pairSplitFunc(input) {
		m := strings.Split(raw, delimiter)
		if len(m) != 2 {
			e := fmt.Errorf("expected '%s' to split by '%s' into two items, got %d", raw, delimiter, len(m))
			err = multierr.Append(err, e)
			continue
		}

		key := strings.TrimSpace(strings.Trim(m[0], "\"'"))
		value := strings.TrimSpace(strings.Trim(m[1], "\"'"))

		parsed[key] = value
	}

	return parsed, err
}

// split on whitespace and preserve quoted text
func splitStringByWhitespace(input string) []string {
	quoted := false
	raw := strings.FieldsFunc(input, func(r rune) bool {
		if r == '"' || r == '\'' {
			quoted = !quoted
		}
		return !quoted && r == ' '
	})
	return raw
}
