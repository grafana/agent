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

package add // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/transformer/add"

import (
	"context"
	"fmt"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
)

const operatorType = "add"

func init() {
	operator.Register(operatorType, func() operator.Builder { return NewConfig() })
}

// NewConfig creates a new add operator config with default values
func NewConfig() *Config {
	return NewConfigWithID(operatorType)
}

// NewConfigWithID creates a new add operator config with default values
func NewConfigWithID(operatorID string) *Config {
	return &Config{
		TransformerConfig: helper.NewTransformerConfig(operatorID, operatorType),
	}
}

// Config is the configuration of an add operator
type Config struct {
	helper.TransformerConfig `mapstructure:",squash"`
	Field                    entry.Field `mapstructure:"field"`
	Value                    interface{} `mapstructure:"value,omitempty"`
}

// Build will build an add operator from the supplied configuration
func (c Config) Build(logger *zap.SugaredLogger) (operator.Operator, error) {
	transformerOperator, err := c.TransformerConfig.Build(logger)
	if err != nil {
		return nil, err
	}

	addOperator := &Transformer{
		TransformerOperator: transformerOperator,
		Field:               c.Field,
	}
	strVal, ok := c.Value.(string)
	if !ok || !isExpr(strVal) {
		addOperator.Value = c.Value
		return addOperator, nil
	}
	exprStr := strings.TrimPrefix(strVal, "EXPR(")
	exprStr = strings.TrimSuffix(exprStr, ")")

	compiled, err := expr.Compile(exprStr, expr.AllowUndefinedVariables())
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression '%s': %w", c.IfExpr, err)
	}

	addOperator.program = compiled
	return addOperator, nil
}

// Transformer is an operator that adds a string value or an expression value
type Transformer struct {
	helper.TransformerOperator

	Field   entry.Field
	Value   interface{}
	program *vm.Program
}

// Process will process an entry with a add transformation.
func (p *Transformer) Process(ctx context.Context, entry *entry.Entry) error {
	return p.ProcessWith(ctx, entry, p.Transform)
}

// Transform will apply the add operations to an entry
func (p *Transformer) Transform(e *entry.Entry) error {
	if p.Value != nil {
		return e.Set(p.Field, p.Value)
	}
	if p.program != nil {
		env := helper.GetExprEnv(e)
		defer helper.PutExprEnv(env)

		result, err := vm.Run(p.program, env)
		if err != nil {
			return fmt.Errorf("evaluate value_expr: %w", err)
		}
		return e.Set(p.Field, result)
	}
	return fmt.Errorf("add: missing required field 'value'")
}

func isExpr(str string) bool {
	return strings.HasPrefix(str, "EXPR(") && strings.HasSuffix(str, ")")
}
