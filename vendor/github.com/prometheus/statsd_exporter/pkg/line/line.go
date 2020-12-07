// Copyright 2013 The Prometheus Authors
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

package line

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/statsd_exporter/pkg/event"
	"github.com/prometheus/statsd_exporter/pkg/mapper"
)

// Parser is a struct to hold configuration for parsing behavior
type Parser struct {
	DogstatsdTagsEnabled bool
	InfluxdbTagsEnabled  bool
	LibratoTagsEnabled   bool
	SignalFXTagsEnabled  bool
}

// NewParser returns a new line parser
func NewParser() *Parser {
	p := Parser{}
	return &p
}

// EnableDogstatsdParsing option to enable dogstatsd tag parsing
func (p *Parser) EnableDogstatsdParsing() {
	p.DogstatsdTagsEnabled = true
}

// EnableInfluxdbParsing option to enable influxdb tag parsing
func (p *Parser) EnableInfluxdbParsing() {
	p.InfluxdbTagsEnabled = true
}

// EnableLibratoParsing option to enable librato tag parsing
func (p *Parser) EnableLibratoParsing() {
	p.LibratoTagsEnabled = true
}

// EnableSignalFXParsing option to enable signalfx tag parsing
func (p *Parser) EnableSignalFXParsing() {
	p.SignalFXTagsEnabled = true
}

func buildEvent(statType, metric string, value float64, relative bool, labels map[string]string) (event.Event, error) {
	switch statType {
	case "c":
		return &event.CounterEvent{
			CMetricName: metric,
			CValue:      float64(value),
			CLabels:     labels,
		}, nil
	case "g":
		return &event.GaugeEvent{
			GMetricName: metric,
			GValue:      float64(value),
			GRelative:   relative,
			GLabels:     labels,
		}, nil
	case "ms":
		return &event.ObserverEvent{
			OMetricName: metric,
			OValue:      float64(value) / 1000, // prometheus presumes seconds, statsd millisecond
			OLabels:     labels,
		}, nil
	case "h", "d":
		return &event.ObserverEvent{
			OMetricName: metric,
			OValue:      float64(value),
			OLabels:     labels,
		}, nil
	case "s":
		return nil, fmt.Errorf("no support for StatsD sets")
	default:
		return nil, fmt.Errorf("bad stat type %s", statType)
	}
}

func parseTag(component, tag string, separator rune, labels map[string]string, tagErrors prometheus.Counter, logger log.Logger) {
	// Entirely empty tag is an error
	if len(tag) == 0 {
		tagErrors.Inc()
		level.Debug(logger).Log("msg", "Empty name tag", "component", component)
		return
	}

	for i, c := range tag {
		if c == separator {
			k := tag[:i]
			v := tag[i+1:]

			if len(k) == 0 || len(v) == 0 {
				// Empty key or value is an error
				tagErrors.Inc()
				level.Debug(logger).Log("msg", "Malformed name tag", "k", k, "v", v, "component", component)
			} else {
				labels[mapper.EscapeMetricName(k)] = v
			}
			return
		}
	}

	// Missing separator (no value) is an error
	tagErrors.Inc()
	level.Debug(logger).Log("msg", "Malformed name tag", "tag", tag, "component", component)
}

func parseNameTags(component string, labels map[string]string, tagErrors prometheus.Counter, logger log.Logger) {
	lastTagEndIndex := 0
	for i, c := range component {
		if c == ',' {
			tag := component[lastTagEndIndex:i]
			lastTagEndIndex = i + 1
			parseTag(component, tag, '=', labels, tagErrors, logger)
		}
	}

	// If we're not off the end of the string, add the last tag
	if lastTagEndIndex < len(component) {
		tag := component[lastTagEndIndex:]
		parseTag(component, tag, '=', labels, tagErrors, logger)
	}
}

func trimLeftHash(s string) string {
	if s != "" && s[0] == '#' {
		return s[1:]
	}
	return s
}

func (p *Parser) ParseDogStatsDTags(component string, labels map[string]string, tagErrors prometheus.Counter, logger log.Logger) {
	if p.DogstatsdTagsEnabled {
		lastTagEndIndex := 0
		for i, c := range component {
			if c == ',' {
				tag := component[lastTagEndIndex:i]
				lastTagEndIndex = i + 1
				parseTag(component, trimLeftHash(tag), ':', labels, tagErrors, logger)
			}
		}

		// If we're not off the end of the string, add the last tag
		if lastTagEndIndex < len(component) {
			tag := component[lastTagEndIndex:]
			parseTag(component, trimLeftHash(tag), ':', labels, tagErrors, logger)
		}
	}
}

func (p *Parser) parseNameAndTags(name string, labels map[string]string, tagErrors prometheus.Counter, logger log.Logger) string {
	if p.SignalFXTagsEnabled {
		// check for SignalFx tags first
		// `[` delimits start of tags by SignalFx
		// `]` delimits end of tags by SignalFx
		// https://docs.signalfx.com/en/latest/integrations/agent/monitors/collectd-statsd.html
		startIdx := strings.IndexRune(name, '[')
		endIdx := strings.IndexRune(name, ']')

		switch {
		case startIdx != -1 && endIdx != -1:
			// good signalfx tags
			parseNameTags(name[startIdx+1:endIdx], labels, tagErrors, logger)
			return name[:startIdx] + name[endIdx+1:]
		case (startIdx != -1) != (endIdx != -1):
			// only one bracket, return unparsed
			level.Debug(logger).Log("msg", "invalid SignalFx tags, not parsing", "metric", name)
			tagErrors.Inc()
			return name
		}
	}

	for i, c := range name {
		// `#` delimits start of tags by Librato
		// https://www.librato.com/docs/kb/collect/collection_agents/stastd/#stat-level-tags
		// `,` delimits start of tags by InfluxDB
		// https://www.influxdata.com/blog/getting-started-with-sending-statsd-metrics-to-telegraf-influxdb/#introducing-influx-statsd
		if (c == '#' && p.LibratoTagsEnabled) || (c == ',' && p.InfluxdbTagsEnabled) {
			parseNameTags(name[i+1:], labels, tagErrors, logger)
			return name[:i]
		}
	}
	return name
}

func (p *Parser) LineToEvents(line string, sampleErrors prometheus.CounterVec, samplesReceived prometheus.Counter, tagErrors prometheus.Counter, tagsReceived prometheus.Counter, logger log.Logger) event.Events {
	events := event.Events{}
	if line == "" {
		return events
	}

	elements := strings.SplitN(line, ":", 2)
	if len(elements) < 2 || len(elements[0]) == 0 || !utf8.ValidString(line) {
		sampleErrors.WithLabelValues("malformed_line").Inc()
		level.Debug(logger).Log("msg", "Bad line from StatsD", "line", line)
		return events
	}

	labels := map[string]string{}
	metric := p.parseNameAndTags(elements[0], labels, tagErrors, logger)

	var samples []string
	if strings.Contains(elements[1], "|#") {
		// using DogStatsD tags

		// don't allow mixed tagging styles
		if len(labels) > 0 {
			sampleErrors.WithLabelValues("mixed_tagging_styles").Inc()
			level.Debug(logger).Log("msg", "Bad line (multiple tagging styles) from StatsD", "line", line)
			return events
		}

		// disable multi-metrics
		samples = elements[1:]
	} else {
		samples = strings.Split(elements[1], ":")
	}

samples:
	for _, sample := range samples {
		samplesReceived.Inc()
		components := strings.Split(sample, "|")
		samplingFactor := 1.0
		if len(components) < 2 || len(components) > 4 {
			sampleErrors.WithLabelValues("malformed_component").Inc()
			level.Debug(logger).Log("msg", "Bad component", "line", line)
			continue
		}
		valueStr, statType := components[0], components[1]

		var relative = false
		if strings.Index(valueStr, "+") == 0 || strings.Index(valueStr, "-") == 0 {
			relative = true
		}

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			level.Debug(logger).Log("msg", "Bad value", "value", valueStr, "line", line)
			sampleErrors.WithLabelValues("malformed_value").Inc()
			continue
		}

		multiplyEvents := 1
		if len(components) >= 3 {
			for _, component := range components[2:] {
				if len(component) == 0 {
					level.Debug(logger).Log("msg", "Empty component", "line", line)
					sampleErrors.WithLabelValues("malformed_component").Inc()
					continue samples
				}
			}

			for _, component := range components[2:] {
				switch component[0] {
				case '@':

					samplingFactor, err = strconv.ParseFloat(component[1:], 64)
					if err != nil {
						level.Debug(logger).Log("msg", "Invalid sampling factor", "component", component[1:], "line", line)
						sampleErrors.WithLabelValues("invalid_sample_factor").Inc()
					}
					if samplingFactor == 0 {
						samplingFactor = 1
					}

					if statType == "g" {
						continue
					} else if statType == "c" {
						value /= samplingFactor
					} else if statType == "ms" || statType == "h" || statType == "d" {
						multiplyEvents = int(1 / samplingFactor)
					}
				case '#':
					p.ParseDogStatsDTags(component[1:], labels, tagErrors, logger)
				default:
					level.Debug(logger).Log("msg", "Invalid sampling factor or tag section", "component", components[2], "line", line)
					sampleErrors.WithLabelValues("invalid_sample_factor").Inc()
					continue
				}
			}
		}

		if len(labels) > 0 {
			tagsReceived.Inc()
		}

		for i := 0; i < multiplyEvents; i++ {
			event, err := buildEvent(statType, metric, value, relative, labels)
			if err != nil {
				level.Debug(logger).Log("msg", "Error building event", "line", line, "error", err)
				sampleErrors.WithLabelValues("illegal_event").Inc()
				continue
			}
			events = append(events, event)
		}
	}
	return events
}
