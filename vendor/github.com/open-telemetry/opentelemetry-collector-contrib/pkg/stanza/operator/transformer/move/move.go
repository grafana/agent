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

package move // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/transformer/move"

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
)

const operatorType = "move"

func init() {
	operator.Register(operatorType, func() operator.Builder { return NewConfig() })
}

// NewConfig creates a new move operator config with default values
func NewConfig() *Config {
	return NewConfigWithID(operatorType)
}

// NewConfigWithID creates a new move operator config with default values
func NewConfigWithID(operatorID string) *Config {
	return &Config{
		TransformerConfig: helper.NewTransformerConfig(operatorID, operatorType),
	}
}

// Config is the configuration of a move operator
type Config struct {
	helper.TransformerConfig `mapstructure:",squash"`
	From                     entry.Field `mapstructure:"from"`
	To                       entry.Field `mapstructure:"to"`
}

// Build will build a Move operator from the supplied configuration
func (c Config) Build(logger *zap.SugaredLogger) (operator.Operator, error) {
	transformerOperator, err := c.TransformerConfig.Build(logger)
	if err != nil {
		return nil, err
	}

	if c.To == entry.NewNilField() || c.From == entry.NewNilField() {
		return nil, fmt.Errorf("move: missing to or from field")
	}

	return &Transformer{
		TransformerOperator: transformerOperator,
		From:                c.From,
		To:                  c.To,
	}, nil
}

// Transformer is an operator that moves a field's value to a new field
type Transformer struct {
	helper.TransformerOperator
	From entry.Field
	To   entry.Field
}

// Process will process an entry with a move transformation.
func (p *Transformer) Process(ctx context.Context, entry *entry.Entry) error {
	return p.ProcessWith(ctx, entry, p.Transform)
}

// Transform will apply the move operation to an entry
func (p *Transformer) Transform(e *entry.Entry) error {
	val, exist := p.From.Delete(e)
	if !exist {
		return fmt.Errorf("move: field does not exist")
	}
	return p.To.Set(e, val)
}
