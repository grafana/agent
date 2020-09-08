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

package exporterhelper

import (
	"errors"
)

var (
	// errNilConfig is returned when an empty name is given.
	errNilConfig = errors.New("nil config")
	// errNilPushTraceData is returned when a nil traceDataPusher is given.
	errNilPushTraceData = errors.New("nil traceDataPusher")
	// errNilPushMetricsData is returned when a nil pushMetricsData is given.
	errNilPushMetricsData = errors.New("nil pushMetricsData")
	// errNilPushLogsData is returned when a nil pushLogsData is given.
	errNilPushLogsData = errors.New("nil pushLogsData")
)
