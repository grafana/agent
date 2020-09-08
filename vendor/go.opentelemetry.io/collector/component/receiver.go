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

package component

import (
	"context"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer"
)

// Receiver defines functions that trace and metric receivers must implement.
type Receiver interface {
	Component
}

// A TraceReceiver is an "arbitrary data"-to-"internal format" converter.
// Its purpose is to translate data from the wild into internal trace format.
// TraceReceiver feeds a consumer.TraceConsumer with data.
//
// For example it could be Zipkin data source which translates
// Zipkin spans into consumerdata.TraceData.
type TraceReceiver interface {
	Receiver
}

// A MetricsReceiver is an "arbitrary data"-to-"internal format" converter.
// Its purpose is to translate data from the wild into internal metrics format.
// MetricsReceiver feeds a consumer.MetricsConsumer with data.
//
// For example it could be Prometheus data source which translates
// Prometheus metrics into consumerdata.MetricsData.
type MetricsReceiver interface {
	Receiver
}

// A LogReceiver is a "log data"-to-"internal format" converter.
// Its purpose is to translate data from the wild into internal data format.
// LogReceiver feeds a consumer.LogConsumer with data.
type LogReceiver interface {
	Receiver
}

// ReceiverFactoryBase defines the common functions for all receiver factories.
type ReceiverFactoryBase interface {
	Factory

	// CreateDefaultConfig creates the default configuration for the Receiver.
	// This method can be called multiple times depending on the pipeline
	// configuration and should not cause side-effects that prevent the creation
	// of multiple instances of the Receiver.
	// The object returned by this method needs to pass the checks implemented by
	// 'configcheck.ValidateConfig'. It is recommended to have such check in the
	// tests of any implementation of the Factory interface.
	CreateDefaultConfig() configmodels.Receiver

	// CustomUnmarshaler returns a custom unmarshaler for the configuration or nil if
	// there is no need for custom unmarshaling. This is typically used if viper.UnmarshalExact()
	// is not sufficient to unmarshal correctly.
	CustomUnmarshaler() CustomUnmarshaler
}

// CustomUnmarshaler is a function that un-marshals a viper data into a config struct
// in a custom way.
// componentViperSection *viper.Viper
//   The config for this specific component. May be nil or empty if no config available.
// intoCfg interface{}
//   An empty interface wrapping a pointer to the config struct to unmarshal into.
type CustomUnmarshaler func(componentViperSection *viper.Viper, intoCfg interface{}) error

// ReceiverFactoryOld can create TraceReceiver and MetricsReceiver.
type ReceiverFactoryOld interface {
	ReceiverFactoryBase

	// CreateTraceReceiver creates a trace receiver based on this config.
	// If the receiver type does not support tracing or if the config is not valid
	// error will be returned instead.
	CreateTraceReceiver(ctx context.Context, logger *zap.Logger, cfg configmodels.Receiver,
		nextConsumer consumer.TraceConsumerOld) (TraceReceiver, error)

	// CreateMetricsReceiver creates a metrics receiver based on this config.
	// If the receiver type does not support metrics or if the config is not valid
	// error will be returned instead.
	CreateMetricsReceiver(ctx context.Context, logger *zap.Logger, cfg configmodels.Receiver,
		nextConsumer consumer.MetricsConsumerOld) (MetricsReceiver, error)
}

// ReceiverCreateParams is passed to ReceiverFactory.Create* functions.
type ReceiverCreateParams struct {
	// Logger that the factory can use during creation and can pass to the created
	// component to be used later as well.
	Logger *zap.Logger
}

// ReceiverFactory can create TraceReceiver and MetricsReceiver. This is the
// new factory type that can create new style receivers.
type ReceiverFactory interface {
	ReceiverFactoryBase

	// CreateTraceReceiver creates a trace receiver based on this config.
	// If the receiver type does not support tracing or if the config is not valid
	// error will be returned instead.
	CreateTraceReceiver(ctx context.Context, params ReceiverCreateParams,
		cfg configmodels.Receiver, nextConsumer consumer.TraceConsumer) (TraceReceiver, error)

	// CreateMetricsReceiver creates a metrics receiver based on this config.
	// If the receiver type does not support metrics or if the config is not valid
	// error will be returned instead.
	CreateMetricsReceiver(ctx context.Context, params ReceiverCreateParams,
		cfg configmodels.Receiver, nextConsumer consumer.MetricsConsumer) (MetricsReceiver, error)
}

// LogReceiverFactory can create a LogReceiver.
type LogReceiverFactory interface {
	ReceiverFactoryBase

	// CreateLogReceiver creates a log receiver based on this config.
	// If the receiver type does not support the data type or if the config is not valid
	// error will be returned instead.
	CreateLogReceiver(
		ctx context.Context,
		params ReceiverCreateParams,
		cfg configmodels.Receiver,
		nextConsumer consumer.LogConsumer,
	) (LogReceiver, error)
}
