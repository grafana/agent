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

package exporterhelper

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/obsreport"
)

// PushLogsData is a helper function that is similar to ConsumeLogsData but also returns
// the number of dropped logs.
type PushLogsData func(ctx context.Context, md pdata.Logs) (droppedTimeSeries int, err error)

type logsRequest struct {
	baseRequest
	ld     pdata.Logs
	pusher PushLogsData
}

func newLogsRequest(ctx context.Context, ld pdata.Logs, pusher PushLogsData) request {
	return &logsRequest{
		baseRequest: baseRequest{ctx: ctx},
		ld:          ld,
		pusher:      pusher,
	}
}

func (req *logsRequest) onPartialError(consumererror.PartialError) request {
	// TODO: Implement this
	return req
}

func (req *logsRequest) export(ctx context.Context) (int, error) {
	return req.pusher(ctx, req.ld)
}

func (req *logsRequest) count() int {
	return req.ld.LogRecordCount()
}

type logsExporter struct {
	*baseExporter
	pushLogsData PushLogsData
}

func (lexp *logsExporter) ConsumeLogs(ctx context.Context, ld pdata.Logs) error {
	exporterCtx := obsreport.ExporterContext(ctx, lexp.cfg.Name())
	_, err := lexp.sender.send(newLogsRequest(exporterCtx, ld, lexp.pushLogsData))
	return err
}

// NewLogsExporter creates an LogsExporter that records observability metrics and wraps every request with a Span.
func NewLogsExporter(cfg configmodels.Exporter, pushLogsData PushLogsData, options ...ExporterOption) (component.LogsExporter, error) {
	if cfg == nil {
		return nil, errNilConfig
	}

	if pushLogsData == nil {
		return nil, errNilPushLogsData
	}

	be := newBaseExporter(cfg, options...)
	be.wrapConsumerSender(func(nextSender requestSender) requestSender {
		return &logsExporterWithObservability{
			exporterName: cfg.Name(),
			nextSender:   nextSender,
		}
	})

	return &logsExporter{
		baseExporter: be,
		pushLogsData: pushLogsData,
	}, nil
}

type logsExporterWithObservability struct {
	exporterName string
	nextSender   requestSender
}

func (lewo *logsExporterWithObservability) send(req request) (int, error) {
	req.setContext(obsreport.StartLogsExportOp(req.context(), lewo.exporterName))
	numDroppedLogs, err := lewo.nextSender.send(req)
	obsreport.EndLogsExportOp(req.context(), req.count(), numDroppedLogs, err)
	return numDroppedLogs, err
}
