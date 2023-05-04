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

package event

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/statsd_exporter/pkg/clock"
	"github.com/prometheus/statsd_exporter/pkg/mapper"
)

type Event interface {
	MetricName() string
	Value() float64
	Labels() map[string]string
	MetricType() mapper.MetricType
}

type CounterEvent struct {
	CMetricName string
	CValue      float64
	CLabels     map[string]string
}

func (c *CounterEvent) MetricName() string            { return c.CMetricName }
func (c *CounterEvent) Value() float64                { return c.CValue }
func (c *CounterEvent) Labels() map[string]string     { return c.CLabels }
func (c *CounterEvent) MetricType() mapper.MetricType { return mapper.MetricTypeCounter }

type GaugeEvent struct {
	GMetricName string
	GValue      float64
	GRelative   bool
	GLabels     map[string]string
}

func (g *GaugeEvent) MetricName() string            { return g.GMetricName }
func (g *GaugeEvent) Value() float64                { return g.GValue }
func (g *GaugeEvent) Labels() map[string]string     { return g.GLabels }
func (g *GaugeEvent) MetricType() mapper.MetricType { return mapper.MetricTypeGauge }

type ObserverEvent struct {
	OMetricName string
	OValue      float64
	OLabels     map[string]string
}

func (o *ObserverEvent) MetricName() string            { return o.OMetricName }
func (o *ObserverEvent) Value() float64                { return o.OValue }
func (o *ObserverEvent) Labels() map[string]string     { return o.OLabels }
func (o *ObserverEvent) MetricType() mapper.MetricType { return mapper.MetricTypeObserver }

type Events []Event

type EventQueue struct {
	C              chan Events
	q              Events
	m              sync.Mutex
	flushTicker    *time.Ticker
	flushThreshold int
	flushInterval  time.Duration
	eventsFlushed  prometheus.Counter
}

type EventHandler interface {
	Queue(event Events)
}

func NewEventQueue(c chan Events, flushThreshold int, flushInterval time.Duration, eventsFlushed prometheus.Counter) *EventQueue {
	ticker := clock.NewTicker(flushInterval)
	eq := &EventQueue{
		C:              c,
		flushThreshold: flushThreshold,
		flushInterval:  flushInterval,
		flushTicker:    ticker,
		q:              make([]Event, 0, flushThreshold),
		eventsFlushed:  eventsFlushed,
	}
	go func() {
		for {
			<-ticker.C
			eq.Flush()
		}
	}()
	return eq
}

func (eq *EventQueue) Queue(events Events) {
	eq.m.Lock()
	defer eq.m.Unlock()

	for _, e := range events {
		eq.q = append(eq.q, e)
		if len(eq.q) >= eq.flushThreshold {
			eq.FlushUnlocked()
		}
	}
}

func (eq *EventQueue) Flush() {
	eq.m.Lock()
	defer eq.m.Unlock()
	eq.FlushUnlocked()
}

func (eq *EventQueue) FlushUnlocked() {
	eq.C <- eq.q
	eq.q = make([]Event, 0, cap(eq.q))
	eq.eventsFlushed.Inc()
}

func (eq *EventQueue) Len() int {
	eq.m.Lock()
	defer eq.m.Unlock()

	return len(eq.q)
}

type UnbufferedEventHandler struct {
	C chan Events
}

func (ueh *UnbufferedEventHandler) Queue(events Events) {
	ueh.C <- events
}
