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

package batchprocessor

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	"go.opentelemetry.io/collector/obsreport"
	"go.opentelemetry.io/collector/processor"
)

var (
	statBatchSizeTriggerSend = stats.Int64("batch_size_trigger_send", "Number of times the batch was sent due to a size trigger", stats.UnitDimensionless)
	statTimeoutTriggerSend   = stats.Int64("timeout_trigger_send", "Number of times the batch was sent due to a timeout trigger", stats.UnitDimensionless)
	statBatchSendSize        = stats.Int64("batch_send_size", "Number of units in the batch", stats.UnitDimensionless)
)

// MetricViews returns the metrics views related to batching
func MetricViews() []*view.View {
	processorTagKeys := []tag.Key{processor.TagProcessorNameKey}

	countBatchSizeTriggerSendView := &view.View{
		Name:        statBatchSizeTriggerSend.Name(),
		Measure:     statBatchSizeTriggerSend,
		Description: statBatchSizeTriggerSend.Description(),
		TagKeys:     processorTagKeys,
		Aggregation: view.Sum(),
	}

	countTimeoutTriggerSendView := &view.View{
		Name:        statTimeoutTriggerSend.Name(),
		Measure:     statTimeoutTriggerSend,
		Description: statTimeoutTriggerSend.Description(),
		TagKeys:     processorTagKeys,
		Aggregation: view.Sum(),
	}

	distributionBatchSendSizeView := &view.View{
		Name:        statBatchSendSize.Name(),
		Measure:     statBatchSendSize,
		Description: statBatchSendSize.Description(),
		TagKeys:     processorTagKeys,
		Aggregation: view.Distribution(10, 25, 50, 75, 100, 250, 500, 750, 1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000, 20000, 30000, 50000, 100000),
	}

	legacyViews := []*view.View{
		countBatchSizeTriggerSendView,
		countTimeoutTriggerSendView,
		distributionBatchSendSizeView,
	}

	return obsreport.ProcessorMetricViews(typeStr, legacyViews)
}
