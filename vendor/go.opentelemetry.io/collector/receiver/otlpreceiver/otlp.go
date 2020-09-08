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

package otlpreceiver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"

	gatewayruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	collectormetrics "go.opentelemetry.io/collector/internal/data/opentelemetry-proto-gen/collector/metrics/v1"
	collectortrace "go.opentelemetry.io/collector/internal/data/opentelemetry-proto-gen/collector/trace/v1"
	"go.opentelemetry.io/collector/receiver/otlpreceiver/metrics"
	"go.opentelemetry.io/collector/receiver/otlpreceiver/trace"
)

// Receiver is the type that exposes Trace and Metrics reception.
type Receiver struct {
	cfg        *Config
	serverGRPC *grpc.Server
	gatewayMux *gatewayruntime.ServeMux
	serverHTTP *http.Server

	traceReceiver   *trace.Receiver
	metricsReceiver *metrics.Receiver

	stopOnce        sync.Once
	startServerOnce sync.Once
}

// New just creates the OpenTelemetry receiver services. It is the caller's
// responsibility to invoke the respective Start*Reception methods as well
// as the various Stop*Reception methods to end it.
func New(cfg *Config) (*Receiver, error) {
	r := &Receiver{
		cfg: cfg,
	}
	if cfg.GRPC != nil {
		opts, err := cfg.GRPC.ToServerOption()
		if err != nil {
			return nil, err
		}
		r.serverGRPC = grpc.NewServer(opts...)
	}
	if cfg.HTTP != nil {
		r.gatewayMux = gatewayruntime.NewServeMux(
			gatewayruntime.WithMarshalerOption("application/x-protobuf", &xProtobufMarshaler{}),
		)
	}

	return r, nil
}

// Start runs the trace receiver on the gRPC server. Currently
// it also enables the metrics receiver too.
func (r *Receiver) Start(ctx context.Context, host component.Host) error {
	if r.traceReceiver == nil && r.metricsReceiver == nil {
		return errors.New("cannot start receiver: no consumers were specified")
	}

	var err error
	r.startServerOnce.Do(func() {
		if r.cfg.GRPC != nil {
			var gln net.Listener
			gln, err = r.cfg.GRPC.ToListener()
			if err != nil {
				return
			}
			go func() {
				if errGrpc := r.serverGRPC.Serve(gln); errGrpc != nil {
					host.ReportFatalError(errGrpc)
				}
			}()
		}
		if r.cfg.HTTP != nil {
			r.serverHTTP = r.cfg.HTTP.ToServer(r.gatewayMux)
			var hln net.Listener
			hln, err = r.cfg.HTTP.ToListener()
			if err != nil {
				return
			}
			go func() {
				if errHTTP := r.serverHTTP.Serve(hln); errHTTP != nil {
					host.ReportFatalError(errHTTP)
				}
			}()
		}
	})
	return err
}

// Shutdown is a method to turn off receiving.
func (r *Receiver) Shutdown(context.Context) error {
	var err error
	r.stopOnce.Do(func() {
		err = nil

		if r.serverHTTP != nil {
			err = r.serverHTTP.Close()
		}

		if r.serverGRPC != nil {
			r.serverGRPC.Stop()
		}
	})
	return err
}

func (r *Receiver) registerTraceConsumer(ctx context.Context, tc consumer.TraceConsumer) error {
	if tc == nil {
		return componenterror.ErrNilNextConsumer
	}
	r.traceReceiver = trace.New(r.cfg.Name(), tc)
	if r.serverGRPC != nil {
		collectortrace.RegisterTraceServiceServer(r.serverGRPC, r.traceReceiver)
	}
	if r.gatewayMux != nil {
		return collectortrace.RegisterTraceServiceHandlerServer(ctx, r.gatewayMux, r.traceReceiver)
	}
	return nil
}

func (r *Receiver) registerMetricsConsumer(ctx context.Context, mc consumer.MetricsConsumer) error {
	if mc == nil {
		return componenterror.ErrNilNextConsumer
	}
	r.metricsReceiver = metrics.New(r.cfg.Name(), mc)
	if r.serverGRPC != nil {
		collectormetrics.RegisterMetricsServiceServer(r.serverGRPC, r.metricsReceiver)
	}
	if r.gatewayMux != nil {
		return collectormetrics.RegisterMetricsServiceHandlerServer(ctx, r.gatewayMux, r.metricsReceiver)
	}
	return nil
}
