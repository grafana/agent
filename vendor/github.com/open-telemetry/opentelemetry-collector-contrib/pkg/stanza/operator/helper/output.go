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
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/errors"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
)

// NewOutputConfig creates a new output config
func NewOutputConfig(operatorID, operatorType string) OutputConfig {
	return OutputConfig{
		BasicConfig: NewBasicConfig(operatorID, operatorType),
	}
}

// OutputConfig provides a basic implementation of an output operator config.
type OutputConfig struct {
	BasicConfig `mapstructure:",squash"`
}

// Build will build an output operator.
func (c OutputConfig) Build(logger *zap.SugaredLogger) (OutputOperator, error) {
	basicOperator, err := c.BasicConfig.Build(logger)
	if err != nil {
		return OutputOperator{}, err
	}

	outputOperator := OutputOperator{
		BasicOperator: basicOperator,
	}

	return outputOperator, nil
}

// OutputOperator provides a basic implementation of an output operator.
type OutputOperator struct {
	BasicOperator
}

// CanProcess will always return true for an output operator.
func (o *OutputOperator) CanProcess() bool {
	return true
}

// CanOutput will always return false for an output operator.
func (o *OutputOperator) CanOutput() bool {
	return false
}

// Outputs will always return an empty array for an output operator.
func (o *OutputOperator) Outputs() []operator.Operator {
	return []operator.Operator{}
}

// GetOutputIDs will always return an empty array for an output ID.
func (o *OutputOperator) GetOutputIDs() []string {
	return []string{}
}

// SetOutputs will return an error if called.
func (o *OutputOperator) SetOutputs(operators []operator.Operator) error {
	return errors.NewError(
		"Operator can not output, but is attempting to set an output.",
		"This is an unexpected internal error. Please submit a bug/issue.",
	)
}

// SetOutputIDs will return nothing and does nothing.
func (o *OutputOperator) SetOutputIDs(opIDs []string) {
}
