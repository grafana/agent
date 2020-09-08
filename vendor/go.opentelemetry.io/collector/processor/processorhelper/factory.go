// Copyright  OpenTelemetry Authors
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

package processorhelper

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configerror"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer"
)

// FactoryOption apply changes to ProcessorOptions.
type FactoryOption func(o *factory)

// CreateDefaultConfig is the equivalent of component.ProcessorFactory.CreateDefaultConfig()
type CreateDefaultConfig func() configmodels.Processor

// CreateTraceProcessor is the equivalent of component.ProcessorFactory.CreateTraceProcessor()
type CreateTraceProcessor func(context.Context, component.ProcessorCreateParams, configmodels.Processor, consumer.TraceConsumer) (component.TraceProcessor, error)

// CreateMetricsProcessor is the equivalent of component.ProcessorFactory.CreateMetricsProcessor()
type CreateMetricsProcessor func(context.Context, component.ProcessorCreateParams, configmodels.Processor, consumer.MetricsConsumer) (component.MetricsProcessor, error)

// CreateMetricsProcessor is the equivalent of component.ProcessorFactory.CreateLogProcessor()
type CreateLogProcessor func(context.Context, component.ProcessorCreateParams, configmodels.Processor, consumer.LogConsumer) (component.LogProcessor, error)

// factory is the factory for Jaeger gRPC exporter.
type factory struct {
	cfgType                configmodels.Type
	createDefaultConfig    CreateDefaultConfig
	createTraceProcessor   CreateTraceProcessor
	createMetricsProcessor CreateMetricsProcessor
	createLogProcessor     CreateLogProcessor
}

var _ component.LogProcessorFactory = new(factory)

// WithTraces overrides the default "error not supported" implementation for CreateTraceReceiver.
func WithTraces(createTraceProcessor CreateTraceProcessor) FactoryOption {
	return func(o *factory) {
		o.createTraceProcessor = createTraceProcessor
	}
}

// WithMetrics overrides the default "error not supported" implementation for CreateMetricsReceiver.
func WithMetrics(createMetricsProcessor CreateMetricsProcessor) FactoryOption {
	return func(o *factory) {
		o.createMetricsProcessor = createMetricsProcessor
	}
}

// WithLogs overrides the default "error not supported" implementation for CreateLogReceiver.
func WithLogs(createLogProcessor CreateLogProcessor) FactoryOption {
	return func(o *factory) {
		o.createLogProcessor = createLogProcessor
	}
}

// NewFactory returns a component.ProcessorFactory that only supports all types.
func NewFactory(
	cfgType configmodels.Type,
	createDefaultConfig CreateDefaultConfig,
	options ...FactoryOption) component.ProcessorFactory {
	f := &factory{
		cfgType:             cfgType,
		createDefaultConfig: createDefaultConfig,
	}
	for _, opt := range options {
		opt(f)
	}
	return f
}

// Type gets the type of the Processor config created by this factory.
func (f *factory) Type() configmodels.Type {
	return f.cfgType
}

// CreateDefaultConfig creates the default configuration for processor.
func (f *factory) CreateDefaultConfig() configmodels.Processor {
	return f.createDefaultConfig()
}

// CreateTraceProcessor creates a component.TraceProcessor based on this config.
func (f *factory) CreateTraceProcessor(
	ctx context.Context,
	params component.ProcessorCreateParams,
	nextConsumer consumer.TraceConsumer,
	cfg configmodels.Processor) (component.TraceProcessor, error) {
	if f.createTraceProcessor != nil {
		return f.createTraceProcessor(ctx, params, cfg, nextConsumer)
	}
	return nil, configerror.ErrDataTypeIsNotSupported
}

// CreateMetricsProcessor creates a consumer.MetricsConsumer based on this config.
func (f *factory) CreateMetricsProcessor(
	ctx context.Context,
	params component.ProcessorCreateParams,
	nextConsumer consumer.MetricsConsumer,
	cfg configmodels.Processor) (component.MetricsProcessor, error) {
	if f.createMetricsProcessor != nil {
		return f.createMetricsProcessor(ctx, params, cfg, nextConsumer)
	}
	return nil, configerror.ErrDataTypeIsNotSupported
}

// CreateLogProcessor creates a metrics processor based on this config.
func (f *factory) CreateLogProcessor(
	ctx context.Context,
	params component.ProcessorCreateParams,
	cfg configmodels.Processor,
	nextConsumer consumer.LogConsumer,
) (component.LogProcessor, error) {
	if f.createLogProcessor != nil {
		return f.createLogProcessor(ctx, params, cfg, nextConsumer)
	}
	return nil, configerror.ErrDataTypeIsNotSupported
}
