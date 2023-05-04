// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pmetric // import "go.opentelemetry.io/collector/pdata/pmetric"

import (
	"go.opentelemetry.io/collector/pdata/external"
	otlpmetrics "go.opentelemetry.io/collector/pdata/external/data/protogen/metrics/v1"
)

// Deprecated: [v0.63.0] use &ProtoMarshaler{}
func NewProtoMarshaler() MarshalSizer {
	return &ProtoMarshaler{}
}

var _ MarshalSizer = (*ProtoMarshaler)(nil)

type ProtoMarshaler struct{}

func (e *ProtoMarshaler) MarshalMetrics(md Metrics) ([]byte, error) {
	pb := internal.MetricsToProto(internal.Metrics(md))
	return pb.Marshal()
}

func (e *ProtoMarshaler) MetricsSize(md Metrics) int {
	pb := internal.MetricsToProto(internal.Metrics(md))
	return pb.Size()
}

type ProtoUnmarshaler struct{}

// Deprecated: [v0.63.0] use &ProtoUnmarshaler{}
func NewProtoUnmarshaler() Unmarshaler {
	return &ProtoUnmarshaler{}
}

func (d *ProtoUnmarshaler) UnmarshalMetrics(buf []byte) (Metrics, error) {
	pb := otlpmetrics.MetricsData{}
	err := pb.Unmarshal(buf)
	return Metrics(internal.MetricsFromProto(pb)), err
}
