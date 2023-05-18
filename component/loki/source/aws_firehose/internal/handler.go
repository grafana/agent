package internal

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"net/http"
	"time"
)

const (
	gzipID1     = 0x1f
	gzipID2     = 0x8b
	gzipDeflate = 8
)

type FirehoseRequest struct {
	RequestID string           `json:"requestId"`
	Timestamp int64            `json:"timestamp"`
	Records   []FirehoseRecord `json:"records"`
}

type FirehoseResponse struct {
	RequestID    string `json:"requestId"`
	Timestamp    int64  `json:"timestamp"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

type FirehoseRecord struct {
	Data string `json:"data"`
}

type CloudwatchLogsData struct {
	// Owner is the AWS Account ID of the originating log data
	Owner string `json:"owner"`

	// LogGroup is the log group name of the originating log data
	LogGroup string `json:"logGroup"`

	// LogStream is the log stream of the originating log data
	LogStream string `json:"logStream"`

	// SubscriptionFilters is the list of subscription filter names
	// that matched with the originating log data
	SubscriptionFilters []string `json:"subscriptionFilters"`

	// MessageType describes the type of LogEvents this record carries.
	// Data messages will use the "DATA_MESSAGE" type. Sometimes CloudWatch
	// Logs may emit Kinesis Data Streams records with a "CONTROL_MESSAGE" type,
	// mainly for checking if the destination is reachable.
	MessageType string `json:"messageType"`

	// LogEvents contains the actual log data.
	LogEvents []CloudwatchLogEvent `json:"logEvents"`
}

type CloudwatchLogEvent struct {
	// ID is a unique id for each log event.
	ID string `json:"id"`

	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
}

type RecordOrigin string

const (
	OriginCloudwatchLogs RecordOrigin = "cloudwatch-logs"
	OriginDirectPUT                   = "direct-put"
	OriginUnknown                     = "unknown"
)

type Sender interface {
	Send(ctx context.Context, entry loki.Entry)
}

type Handler struct {
	metrics *metrics
	logger  log.Logger
	sender  Sender
}

func NewHandler(sender Sender, logger log.Logger, reg prometheus.Registerer) *Handler {
	return &Handler{
		metrics: newMetrics(reg),
		logger:  logger,
		sender:  sender,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var err error
	defer req.Body.Close()
	level.Info(h.logger).Log("msg", "handling request")

	var bodyReader io.Reader = req.Body
	// firehose allows the user to configure gzip content-encoding, in that case
	// decompress in the reader during unmarshalling
	if req.Header.Get("Content-Encoding") == "gzip" {
		bodyReader, err = gzip.NewReader(req.Body)
		if err != nil {
			h.metrics.errors.WithLabelValues("pre_read").Inc()
			level.Error(h.logger).Log("msg", "failed to create gzip reader", "err", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// todo(pablo): use headers as labels

	firehoseReq := FirehoseRequest{}

	err = json.NewDecoder(bodyReader).Decode(&firehoseReq)
	if err != nil {
		h.metrics.errors.WithLabelValues("read_or_format").Inc()
		level.Error(h.logger).Log("msg", "failed to unmarshall request", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// todo(pablo): should parallize this?
	for _, rec := range firehoseReq.Records {
		decodedRecord, _, err := h.decodeRecord(rec.Data)

		// todo(pablo): use the decoded type for something, maybe inject as label

		if err != nil {
			h.metrics.errors.WithLabelValues("decode").Inc()
			level.Error(h.logger).Log("msg", "failed to decode request record", "err", err.Error())
			h.respondError(w, firehoseReq.RequestID, "failed to decode record", http.StatusBadRequest)

			// todo(pablo): is ok this below?
			// since all individual data record are packed in a bigger record, responding an error
			// here will mean we'll get the same individual record on the retry. Continue processing
			// the rest.
			return
		}

		h.metrics.recordsReceived.WithLabelValues().Inc()

		// todo(pablo): if cloudwatch logs we can do further decoding

		h.sender.Send(req.Context(), loki.Entry{
			Labels: nil,
			Entry: logproto.Entry{
				Timestamp: time.Now(),
				Line:      string(decodedRecord),
			},
		})
	}

	// since the response type is rather simple, use string interpolation to marshall it
	w.Write([]byte(fmt.Sprintf(successResponseTemplate, firehoseReq.RequestID, time.Now().Unix())))
}

const successResponseTemplate = `{"requestId": "%s", "timestamp": %d}`
const errorResponseTemplate = `{"requestId": "%s", "timestamp": %d, "errorMessage": "%s"}`

func (h *Handler) respondError(w http.ResponseWriter, firehoseID, errMsg string, status int) {
	timestamp := time.Now().Unix()
	w.WriteHeader(status)
	// since the response type is rather simple, use string interpolation to marshall it
	w.Write([]byte(fmt.Sprintf(errorResponseTemplate, firehoseID, timestamp, errMsg)))
	return
}

func (h *Handler) decodeRecord(rec string) ([]byte, RecordOrigin, error) {
	decodedRec, err := base64.StdEncoding.DecodeString(rec)
	if err != nil {
		return nil, OriginUnknown, fmt.Errorf("error base64-decoding record: %w", err)
	}

	// Using the same header check as the gzip library, but inlining the check to avoid unnecessary boilerplate
	// code from creating the reader.
	//
	// https://github.com/golang/go/blob/master/src/compress/gzip/gunzip.go#L185
	if !(decodedRec[0] == gzipID1 && decodedRec[1] == gzipID2 && // the first two represent the 1f8b magic bytes
		decodedRec[2] == gzipDeflate) { // the third byte represents the gzip compression method DEFLATE
		// no gzip, return decoded data
		return decodedRec, OriginDirectPUT, nil
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(decodedRec))
	if err != nil {
		return nil, OriginCloudwatchLogs, fmt.Errorf("error creating gzip reader: %w", err)
	}
	defer gzipReader.Close()

	b := bytes.Buffer{}
	if _, err := io.Copy(bufio.NewWriter(&b), gzipReader); err != nil {
		return nil, OriginCloudwatchLogs, fmt.Errorf("error reading gzipped bytes: %w", err)
	}
	return b.Bytes(), OriginCloudwatchLogs, nil

}
