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
	"bytes"

	"go.opentelemetry.io/collector/pdata/external"
	otlpmetrics "go.opentelemetry.io/collector/pdata/external/data/protogen/metrics/v1"
	"go.opentelemetry.io/collector/pdata/pmetric/external/pmetricjson"
)

var delegate = pmetricjson.JSONMarshaler

// Deprecated: [v0.63.0] use &JSONMarshaler{}
func NewJSONMarshaler() Marshaler {
	return &JSONMarshaler{}
}

var _ Marshaler = (*JSONMarshaler)(nil)

type JSONMarshaler struct{}

func (*JSONMarshaler) MarshalMetrics(md Metrics) ([]byte, error) {
	buf := bytes.Buffer{}
	pb := internal.MetricsToProto(internal.Metrics(md))
	err := delegate.Marshal(&buf, &pb)
	return buf.Bytes(), err
}

type JSONUnmarshaler struct{}

// Deprecated: [v0.63.0] use &JSONUnmarshaler{}
func NewJSONUnmarshaler() Unmarshaler {
	return &JSONUnmarshaler{}
}

func (*JSONUnmarshaler) UnmarshalMetrics(buf []byte) (Metrics, error) {
	var md otlpmetrics.MetricsData
	if err := pmetricjson.UnmarshalMetricsData(buf, &md); err != nil {
		return Metrics{}, err
	}
	return Metrics(internal.MetricsFromProto(md)), nil
}
