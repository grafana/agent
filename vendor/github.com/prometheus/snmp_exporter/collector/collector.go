// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gosnmp/gosnmp"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/prometheus/snmp_exporter/config"
)

const (
	namespace = "snmp"
)

var (
	buckets               = prometheus.ExponentialBuckets(0.0001, 2, 15)
	snmpUnexpectedPduType = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "unexpected_pdu_type_total",
			Help:      "Unexpected Go types in a PDU.",
		},
	)
	snmpDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "packet_duration_seconds",
			Help:      "A histogram of latencies for SNMP packets.",
			Buckets:   buckets,
		},
	)
	snmpPackets = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "packets_total",
			Help:      "Number of SNMP packet sent, including retries.",
		},
	)
	snmpRetries = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "packet_retries_total",
			Help:      "Number of SNMP packet retries.",
		},
	)
	// 64-bit float mantissa: https://en.wikipedia.org/wiki/Double-precision_floating-point_format
	float64Mantissa uint64 = 9007199254740992
	wrapCounters           = kingpin.Flag("snmp.wrap-large-counters", "Wrap 64-bit counters to avoid floating point rounding.").Default("true").Bool()
)

// Types preceded by an enum with their actual type.
var combinedTypeMapping = map[string]map[int]string{
	"InetAddress": {
		1: "InetAddressIPv4",
		2: "InetAddressIPv6",
	},
	"InetAddressMissingSize": {
		1: "InetAddressIPv4",
		2: "InetAddressIPv6",
	},
	"LldpPortId": {
		1: "DisplayString",
		2: "DisplayString",
		3: "PhysAddress48",
		5: "DisplayString",
		7: "DisplayString",
	},
}

func oidToList(oid string) []int {
	result := []int{}
	for _, x := range strings.Split(oid, ".") {
		o, _ := strconv.Atoi(x)
		result = append(result, o)
	}
	return result
}

func listToOid(l []int) string {
	var result []string
	for _, o := range l {
		result = append(result, strconv.Itoa(o))
	}
	return strings.Join(result, ".")
}

type ScrapeResults struct {
	pdus    []gosnmp.SnmpPDU
	packets uint64
	retries uint64
}

func ScrapeTarget(ctx context.Context, target string, config *config.Module, logger log.Logger) (ScrapeResults, error) {
	results := ScrapeResults{}
	// Set the options.
	snmp := gosnmp.GoSNMP{}
	snmp.Context = ctx
	snmp.MaxRepetitions = config.WalkParams.MaxRepetitions
	snmp.Retries = config.WalkParams.Retries
	snmp.Timeout = config.WalkParams.Timeout
	snmp.UseUnconnectedUDPSocket = config.WalkParams.UseUnconnectedUDPSocket

	var sent time.Time
	snmp.OnSent = func(x *gosnmp.GoSNMP) {
		sent = time.Now()
		snmpPackets.Inc()
		results.packets++
	}
	snmp.OnRecv = func(x *gosnmp.GoSNMP) {
		snmpDuration.Observe(time.Since(sent).Seconds())
	}
	snmp.OnRetry = func(x *gosnmp.GoSNMP) {
		snmpRetries.Inc()
		results.retries++
	}

	snmp.Target = target
	snmp.Port = 161
	if host, port, err := net.SplitHostPort(target); err == nil {
		snmp.Target = host
		p, err := strconv.Atoi(port)
		if err != nil {
			return results, fmt.Errorf("error converting port number to int for target %s: %s", target, err)
		}
		snmp.Port = uint16(p)
	}

	// Configure auth.
	config.WalkParams.ConfigureSNMP(&snmp)

	// Do the actual walk.
	err := snmp.Connect()
	if err != nil {
		if err == context.Canceled {
			return results, fmt.Errorf("scrape canceled (possible timeout) connecting to target %s", snmp.Target)
		}
		return results, fmt.Errorf("error connecting to target %s: %s", target, err)
	}
	defer snmp.Conn.Close()

	getOids := config.Get
	maxOids := int(config.WalkParams.MaxRepetitions)
	// Max Repetition can be 0, maxOids cannot. SNMPv1 can only report one OID error per call.
	if maxOids == 0 || snmp.Version == gosnmp.Version1 {
		maxOids = 1
	}
	for len(getOids) > 0 {
		oids := len(getOids)
		if oids > maxOids {
			oids = maxOids
		}

		level.Debug(logger).Log("msg", "Getting OIDs", "oids", oids)
		getStart := time.Now()
		packet, err := snmp.Get(getOids[:oids])
		if err != nil {
			if err == context.Canceled {
				return results, fmt.Errorf("scrape canceled (possible timeout) getting target %s", snmp.Target)
			}
			return results, fmt.Errorf("error getting target %s: %s", snmp.Target, err)
		}
		level.Debug(logger).Log("msg", "Get of OIDs completed", "oids", oids, "duration_seconds", time.Since(getStart))
		// SNMPv1 will return packet error for unsupported OIDs.
		if packet.Error == gosnmp.NoSuchName && snmp.Version == gosnmp.Version1 {
			level.Debug(logger).Log("msg", "OID not supported by target", "oids", getOids[0])
			getOids = getOids[oids:]
			continue
		}
		// Response received with errors.
		// TODO: "stringify" gosnmp errors instead of showing error code.
		if packet.Error != gosnmp.NoError {
			return results, fmt.Errorf("error reported by target %s: Error Status %d", snmp.Target, packet.Error)
		}
		for _, v := range packet.Variables {
			if v.Type == gosnmp.NoSuchObject || v.Type == gosnmp.NoSuchInstance {
				level.Debug(logger).Log("msg", "OID not supported by target", "oids", v.Name)
				continue
			}
			results.pdus = append(results.pdus, v)
		}
		getOids = getOids[oids:]
	}

	for _, subtree := range config.Walk {
		var pdus []gosnmp.SnmpPDU
		level.Debug(logger).Log("msg", "Walking subtree", "oid", subtree)
		walkStart := time.Now()
		if snmp.Version == gosnmp.Version1 {
			pdus, err = snmp.WalkAll(subtree)
		} else {
			pdus, err = snmp.BulkWalkAll(subtree)
		}
		if err != nil {
			if err == context.Canceled {
				return results, fmt.Errorf("scrape canceled (possible timeout) walking target %s", snmp.Target)
			}
			return results, fmt.Errorf("error walking target %s: %s", snmp.Target, err)
		}
		level.Debug(logger).Log("msg", "Walk of subtree completed", "oid", subtree, "duration_seconds", time.Since(walkStart))

		results.pdus = append(results.pdus, pdus...)
	}
	return results, nil
}

type MetricNode struct {
	metric *config.Metric

	children map[int]*MetricNode
}

// Build a tree of metrics from the config, for fast lookup when there's lots of them.
func buildMetricTree(metrics []*config.Metric) *MetricNode {
	metricTree := &MetricNode{children: map[int]*MetricNode{}}
	for _, metric := range metrics {
		head := metricTree
		for _, o := range oidToList(metric.Oid) {
			_, ok := head.children[o]
			if !ok {
				head.children[o] = &MetricNode{children: map[int]*MetricNode{}}
			}
			head = head.children[o]
		}
		head.metric = metric
	}
	return metricTree
}

type collector struct {
	ctx    context.Context
	target string
	module *config.Module
	logger log.Logger
}

func New(ctx context.Context, target string, module *config.Module, logger log.Logger) *collector {
	return &collector{ctx: ctx, target: target, module: module, logger: logger}
}

// Describe implements Prometheus.Collector.
func (c collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

// Collect implements Prometheus.Collector.
func (c collector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	results, err := ScrapeTarget(c.ctx, c.target, c.module, c.logger)
	if err != nil {
		level.Info(c.logger).Log("msg", "Error scraping target", "err", err)
		ch <- prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error scraping target", nil, nil), err)
		return
	}
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("snmp_scrape_walk_duration_seconds", "Time SNMP walk/bulkwalk took.", nil, nil),
		prometheus.GaugeValue,
		time.Since(start).Seconds())
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("snmp_scrape_packets_sent", "Packets sent for get, bulkget, and walk; including retries.", nil, nil),
		prometheus.GaugeValue,
		float64(results.packets))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("snmp_scrape_packets_retried", "Packets retried for get, bulkget, and walk.", nil, nil),
		prometheus.GaugeValue,
		float64(results.retries))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("snmp_scrape_pdus_returned", "PDUs returned from get, bulkget, and walk.", nil, nil),
		prometheus.GaugeValue,
		float64(len(results.pdus)))
	oidToPdu := make(map[string]gosnmp.SnmpPDU, len(results.pdus))
	for _, pdu := range results.pdus {
		oidToPdu[pdu.Name[1:]] = pdu
	}

	metricTree := buildMetricTree(c.module.Metrics)
	// Look for metrics that match each pdu.
PduLoop:
	for oid, pdu := range oidToPdu {
		head := metricTree
		oidList := oidToList(oid)
		for i, o := range oidList {
			var ok bool
			head, ok = head.children[o]
			if !ok {
				continue PduLoop
			}
			if head.metric != nil {
				// Found a match.
				samples := pduToSamples(oidList[i+1:], &pdu, head.metric, oidToPdu, c.logger)
				for _, sample := range samples {
					ch <- sample
				}
				break
			}
		}
	}
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("snmp_scrape_duration_seconds", "Total SNMP time scrape took (walk and processing).", nil, nil),
		prometheus.GaugeValue,
		time.Since(start).Seconds())
}

func getPduValue(pdu *gosnmp.SnmpPDU) float64 {
	switch pdu.Type {
	case gosnmp.Counter64:
		if *wrapCounters {
			// Wrap by 2^53.
			return float64(gosnmp.ToBigInt(pdu.Value).Uint64() % float64Mantissa)
		} else {
			return float64(gosnmp.ToBigInt(pdu.Value).Uint64())
		}
	case gosnmp.OpaqueFloat:
		return float64(pdu.Value.(float32))
	case gosnmp.OpaqueDouble:
		return pdu.Value.(float64)
	default:
		return float64(gosnmp.ToBigInt(pdu.Value).Int64())
	}
}

// parseDateAndTime extracts a UNIX timestamp from an RFC 2579 DateAndTime.
func parseDateAndTime(pdu *gosnmp.SnmpPDU) (float64, error) {
	var (
		v   []byte
		tz  *time.Location
		err error
	)
	// DateAndTime should be a slice of bytes.
	switch pduType := pdu.Value.(type) {
	case []byte:
		v = pdu.Value.([]byte)
	default:
		return 0, fmt.Errorf("invalid DateAndTime type %v", pduType)
	}
	pduLength := len(v)
	// DateAndTime can be 8 or 11 bytes depending if the time zone is included.
	switch pduLength {
	case 8:
		// No time zone included, assume UTC.
		tz = time.UTC
	case 11:
		// Extract the timezone from the last 3 bytes.
		locString := fmt.Sprintf("%s%02d%02d", string(v[8]), v[9], v[10])
		loc, err := time.Parse("-0700", locString)
		if err != nil {
			return 0, fmt.Errorf("error parsing location string: %q, error: %s", locString, err)
		}
		tz = loc.Location()
	default:
		return 0, fmt.Errorf("invalid DateAndTime length %v", pduLength)
	}
	if err != nil {
		return 0, fmt.Errorf("unable to parse DateAndTime %q, error: %s", v, err)
	}
	// Build the date from the various fields and time zone.
	t := time.Date(
		int(binary.BigEndian.Uint16(v[0:2])),
		time.Month(v[2]),
		int(v[3]),
		int(v[4]),
		int(v[5]),
		int(v[6]),
		int(v[7])*1e+8,
		tz)
	return float64(t.Unix()), nil
}

func pduToSamples(indexOids []int, pdu *gosnmp.SnmpPDU, metric *config.Metric, oidToPdu map[string]gosnmp.SnmpPDU, logger log.Logger) []prometheus.Metric {
	var err error
	// The part of the OID that is the indexes.
	labels := indexesToLabels(indexOids, metric, oidToPdu)

	value := getPduValue(pdu)
	t := prometheus.UntypedValue

	labelnames := make([]string, 0, len(labels)+1)
	labelvalues := make([]string, 0, len(labels)+1)
	for k, v := range labels {
		labelnames = append(labelnames, k)
		labelvalues = append(labelvalues, v)
	}

	switch metric.Type {
	case "counter":
		t = prometheus.CounterValue
	case "gauge":
		t = prometheus.GaugeValue
	case "Float", "Double":
		t = prometheus.GaugeValue
	case "DateAndTime":
		t = prometheus.GaugeValue
		value, err = parseDateAndTime(pdu)
		if err != nil {
			level.Debug(logger).Log("msg", "Error parsing DateAndTime", "err", err)
			return []prometheus.Metric{}
		}
	case "EnumAsInfo":
		return enumAsInfo(metric, int(value), labelnames, labelvalues)
	case "EnumAsStateSet":
		return enumAsStateSet(metric, int(value), labelnames, labelvalues)
	case "Bits":
		return bits(metric, pdu.Value, labelnames, labelvalues)
	default:
		// It's some form of string.
		t = prometheus.GaugeValue
		value = 1.0
		metricType := metric.Type

		if typeMapping, ok := combinedTypeMapping[metricType]; ok {
			// Lookup associated sub type in previous object.
			oids := strings.Split(metric.Oid, ".")
			i, _ := strconv.Atoi(oids[len(oids)-1])
			oids[len(oids)-1] = strconv.Itoa(i - 1)
			prevOid := fmt.Sprintf("%s.%s", strings.Join(oids, "."), listToOid(indexOids))
			if prevPdu, ok := oidToPdu[prevOid]; ok {
				val := int(getPduValue(&prevPdu))
				if t, ok := typeMapping[val]; ok {
					metricType = t
				} else {
					metricType = "OctetString"
					level.Debug(logger).Log("msg", "Unable to handle type value", "value", val, "oid", prevOid, "metric", metric.Name)
				}
			} else {
				metricType = "OctetString"
				level.Debug(logger).Log("msg", "Unable to find type at oid for metric", "oid", prevOid, "metric", metric.Name)
			}
		}

		if len(metric.RegexpExtracts) > 0 {
			return applyRegexExtracts(metric, pduValueAsString(pdu, metricType), labelnames, labelvalues, logger)
		}
		// For strings we put the value as a label with the same name as the metric.
		// If the name is already an index, we do not need to set it again.
		if _, ok := labels[metric.Name]; !ok {
			labelnames = append(labelnames, metric.Name)
			labelvalues = append(labelvalues, pduValueAsString(pdu, metricType))
		}
	}

	sample, err := prometheus.NewConstMetric(prometheus.NewDesc(metric.Name, metric.Help, labelnames, nil),
		t, value, labelvalues...)
	if err != nil {
		sample = prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error calling NewConstMetric", nil, nil),
			fmt.Errorf("error for metric %s with labels %v from indexOids %v: %v", metric.Name, labelvalues, indexOids, err))
	}

	return []prometheus.Metric{sample}
}

func applyRegexExtracts(metric *config.Metric, pduValue string, labelnames, labelvalues []string, logger log.Logger) []prometheus.Metric {
	results := []prometheus.Metric{}
	for name, strMetricSlice := range metric.RegexpExtracts {
		for _, strMetric := range strMetricSlice {
			indexes := strMetric.Regex.FindStringSubmatchIndex(pduValue)
			if indexes == nil {
				level.Debug(logger).Log("msg", "No match found for regexp", "metric", metric.Name, "value", pduValue, "regex", strMetric.Regex.String())
				continue
			}
			res := strMetric.Regex.ExpandString([]byte{}, strMetric.Value, pduValue, indexes)
			v, err := strconv.ParseFloat(string(res), 64)
			if err != nil {
				level.Debug(logger).Log("msg", "Error parsing float64 from value", "metric", metric.Name, "value", pduValue, "regex", strMetric.Regex.String(), "extracted_value", res)
				continue
			}
			newMetric, err := prometheus.NewConstMetric(prometheus.NewDesc(metric.Name+name, metric.Help+" (regex extracted)", labelnames, nil),
				prometheus.GaugeValue, v, labelvalues...)
			if err != nil {
				newMetric = prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error calling NewConstMetric for regex_extract", nil, nil),
					fmt.Errorf("error for metric %s with labels %v: %v", metric.Name+name, labelvalues, err))
			}
			results = append(results, newMetric)
			break
		}
	}
	return results
}

func enumAsInfo(metric *config.Metric, value int, labelnames, labelvalues []string) []prometheus.Metric {
	// Lookup enum, default to the value.
	state, ok := metric.EnumValues[int(value)]
	if !ok {
		state = strconv.Itoa(int(value))
	}
	labelnames = append(labelnames, metric.Name)
	labelvalues = append(labelvalues, state)

	newMetric, err := prometheus.NewConstMetric(prometheus.NewDesc(metric.Name+"_info", metric.Help+" (EnumAsInfo)", labelnames, nil),
		prometheus.GaugeValue, 1.0, labelvalues...)
	if err != nil {
		newMetric = prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error calling NewConstMetric for EnumAsInfo", nil, nil),
			fmt.Errorf("error for metric %s with labels %v: %v", metric.Name, labelvalues, err))
	}
	return []prometheus.Metric{newMetric}
}

func enumAsStateSet(metric *config.Metric, value int, labelnames, labelvalues []string) []prometheus.Metric {
	labelnames = append(labelnames, metric.Name)
	results := []prometheus.Metric{}

	state, ok := metric.EnumValues[value]
	if !ok {
		// Fallback to using the value.
		state = strconv.Itoa(value)
	}
	newMetric, err := prometheus.NewConstMetric(prometheus.NewDesc(metric.Name, metric.Help+" (EnumAsStateSet)", labelnames, nil),
		prometheus.GaugeValue, 1.0, append(labelvalues, state)...)
	if err != nil {
		newMetric = prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error calling NewConstMetric for EnumAsStateSet", nil, nil),
			fmt.Errorf("error for metric %s with labels %v: %v", metric.Name, labelvalues, err))
	}
	results = append(results, newMetric)

	for k, v := range metric.EnumValues {
		if k == value {
			continue
		}
		newMetric, err := prometheus.NewConstMetric(prometheus.NewDesc(metric.Name, metric.Help+" (EnumAsStateSet)", labelnames, nil),
			prometheus.GaugeValue, 0.0, append(labelvalues, v)...)
		if err != nil {
			newMetric = prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error calling NewConstMetric for EnumAsStateSet", nil, nil),
				fmt.Errorf("error for metric %s with labels %v: %v", metric.Name, labelvalues, err))
		}
		results = append(results, newMetric)
	}
	return results
}

func bits(metric *config.Metric, value interface{}, labelnames, labelvalues []string) []prometheus.Metric {
	bytes, ok := value.([]byte)
	if !ok {
		return []prometheus.Metric{prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "BITS type was not a BISTRING on the wire.", nil, nil),
			fmt.Errorf("error for metric %s with labels %v: %T", metric.Name, labelvalues, value))}
	}
	labelnames = append(labelnames, metric.Name)
	results := []prometheus.Metric{}

	for k, v := range metric.EnumValues {
		bit := 0.0
		// Most significant byte most significant bit, then most significant byte 2nd most significant bit etc.
		if k < len(bytes)*8 {
			if (bytes[k/8] & (128 >> (k % 8))) != 0 {
				bit = 1.0
			}
		}
		newMetric, err := prometheus.NewConstMetric(prometheus.NewDesc(metric.Name, metric.Help+" (Bits)", labelnames, nil),
			prometheus.GaugeValue, bit, append(labelvalues, v)...)
		if err != nil {
			newMetric = prometheus.NewInvalidMetric(prometheus.NewDesc("snmp_error", "Error calling NewConstMetric for Bits", nil, nil),
				fmt.Errorf("error for metric %s with labels %v: %v", metric.Name, labelvalues, err))
		}
		results = append(results, newMetric)
	}
	return results
}

// Right pad oid with zeros, and split at the given point.
// Some routers exclude trailing 0s in responses.
func splitOid(oid []int, count int) ([]int, []int) {
	head := make([]int, count)
	tail := []int{}
	for i, v := range oid {
		if i < count {
			head[i] = v
		} else {
			tail = append(tail, v)
		}
	}
	return head, tail
}

// This mirrors decodeValue in gosnmp's helper.go.
func pduValueAsString(pdu *gosnmp.SnmpPDU, typ string) string {
	switch pdu.Value.(type) {
	case int:
		return strconv.Itoa(pdu.Value.(int))
	case uint:
		return strconv.FormatUint(uint64(pdu.Value.(uint)), 10)
	case uint64:
		return strconv.FormatUint(pdu.Value.(uint64), 10)
	case float32:
		return strconv.FormatFloat(float64(pdu.Value.(float32)), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(pdu.Value.(float64), 'f', -1, 64)
	case string:
		if pdu.Type == gosnmp.ObjectIdentifier {
			// Trim leading period.
			return pdu.Value.(string)[1:]
		}
		// DisplayString.
		return pdu.Value.(string)
	case []byte:
		if typ == "" {
			typ = "OctetString"
		}
		// Reuse the OID index parsing code.
		parts := make([]int, len(pdu.Value.([]byte)))
		for i, o := range pdu.Value.([]byte) {
			parts[i] = int(o)
		}
		if typ == "OctetString" || typ == "DisplayString" {
			// Prepend the length, as it is explicit in an index.
			parts = append([]int{len(pdu.Value.([]byte))}, parts...)
		}
		str, _, _ := indexOidsAsString(parts, typ, 0, false, nil)
		return str
	case nil:
		return ""
	default:
		// This shouldn't happen.
		snmpUnexpectedPduType.Inc()
		return fmt.Sprintf("%s", pdu.Value)
	}
}

// Convert oids to a string index value.
//
// Returns the string, the oids that were used and the oids left over.
func indexOidsAsString(indexOids []int, typ string, fixedSize int, implied bool, enumValues map[int]string) (string, []int, []int) {
	if typeMapping, ok := combinedTypeMapping[typ]; ok {
		subOid, valueOids := splitOid(indexOids, 2)
		if typ == "InetAddressMissingSize" {
			// The size of the main index value is missing.
			subOid, valueOids = splitOid(indexOids, 1)
		}
		var str string
		var used, remaining []int
		if t, ok := typeMapping[subOid[0]]; ok {
			str, used, remaining = indexOidsAsString(valueOids, t, 0, false, enumValues)
			return str, append(subOid, used...), remaining
		}
		if typ == "InetAddressMissingSize" {
			// We don't know the size, so pass everything remaining.
			return indexOidsAsString(indexOids, "OctetString", 0, true, enumValues)
		}
		// The 2nd oid is the length.
		return indexOidsAsString(indexOids, "OctetString", subOid[1]+2, false, enumValues)
	}

	switch typ {
	case "Integer32", "Integer", "gauge", "counter":
		// Extract the oid for this index, and keep the remainder for the next index.
		subOid, indexOids := splitOid(indexOids, 1)
		return fmt.Sprintf("%d", subOid[0]), subOid, indexOids
	case "PhysAddress48":
		subOid, indexOids := splitOid(indexOids, 6)
		parts := make([]string, 6)
		for i, o := range subOid {
			parts[i] = fmt.Sprintf("%02X", o)
		}
		return strings.Join(parts, ":"), subOid, indexOids
	case "OctetString":
		var subOid []int
		// The length of fixed size indexes come from the MIB.
		// For varying size, we read it from the first oid.
		length := fixedSize
		if implied {
			length = len(indexOids)
		}
		if length == 0 {
			subOid, indexOids = splitOid(indexOids, 1)
			length = subOid[0]
		}
		content, indexOids := splitOid(indexOids, length)
		subOid = append(subOid, content...)
		parts := make([]byte, length)
		for i, o := range content {
			parts[i] = byte(o)
		}
		if len(parts) == 0 {
			return "", subOid, indexOids
		} else {
			return fmt.Sprintf("0x%X", string(parts)), subOid, indexOids
		}
	case "DisplayString":
		var subOid []int
		length := fixedSize
		if implied {
			length = len(indexOids)
		}
		if length == 0 {
			subOid, indexOids = splitOid(indexOids, 1)
			length = subOid[0]
		}
		content, indexOids := splitOid(indexOids, length)
		subOid = append(subOid, content...)
		parts := make([]byte, length)
		for i, o := range content {
			parts[i] = byte(o)
		}
		// ASCII, so can convert staight to utf-8.
		return string(parts), subOid, indexOids
	case "InetAddressIPv4":
		subOid, indexOids := splitOid(indexOids, 4)
		parts := make([]string, 4)
		for i, o := range subOid {
			parts[i] = strconv.Itoa(o)
		}
		return strings.Join(parts, "."), subOid, indexOids
	case "InetAddressIPv6":
		subOid, indexOids := splitOid(indexOids, 16)
		parts := make([]interface{}, 16)
		for i, o := range subOid {
			parts[i] = o
		}
		return fmt.Sprintf("%02X%02X:%02X%02X:%02X%02X:%02X%02X:%02X%02X:%02X%02X:%02X%02X:%02X%02X", parts...), subOid, indexOids
	case "EnumAsInfo":
		subOid, indexOids := splitOid(indexOids, 1)
		value, ok := enumValues[subOid[0]]
		if ok {
			return value, subOid, indexOids
		} else {
			return fmt.Sprintf("%d", subOid[0]), subOid, indexOids
		}
	default:
		panic(fmt.Sprintf("Unknown index type %s", typ))
		return "", nil, nil
	}
}

func indexesToLabels(indexOids []int, metric *config.Metric, oidToPdu map[string]gosnmp.SnmpPDU) map[string]string {
	labels := map[string]string{}
	labelOids := map[string][]int{}

	// Covert indexes to useful strings.
	for _, index := range metric.Indexes {
		str, subOid, remainingOids := indexOidsAsString(indexOids, index.Type, index.FixedSize, index.Implied, index.EnumValues)
		// The labelvalue is the text form of the index oids.
		labels[index.Labelname] = str
		// Save its oid in case we need it for lookups.
		labelOids[index.Labelname] = subOid
		// For the next iteration.
		indexOids = remainingOids
	}

	// Perform lookups.
	for _, lookup := range metric.Lookups {
		if len(lookup.Labels) == 0 {
			delete(labels, lookup.Labelname)
			continue
		}
		oid := lookup.Oid
		for _, label := range lookup.Labels {
			oid = fmt.Sprintf("%s.%s", oid, listToOid(labelOids[label]))
		}
		if pdu, ok := oidToPdu[oid]; ok {
			labels[lookup.Labelname] = pduValueAsString(&pdu, lookup.Type)
			labelOids[lookup.Labelname] = []int{int(gosnmp.ToBigInt(pdu.Value).Int64())}
		} else {
			labels[lookup.Labelname] = ""
		}
	}

	return labels
}
