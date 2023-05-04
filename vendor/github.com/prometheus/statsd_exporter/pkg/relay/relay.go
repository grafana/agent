// Copyright 2021 The Prometheus Authors
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

package relay

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/prometheus/statsd_exporter/pkg/clock"

	"github.com/go-kit/log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/statsd_exporter/pkg/level"
)

type Relay struct {
	addr          *net.UDPAddr
	bufferChannel chan []byte
	conn          *net.UDPConn
	logger        log.Logger
	packetLength  uint

	packetsTotal      prometheus.Counter
	longLinesTotal    prometheus.Counter
	relayedLinesTotal prometheus.Counter
}

var (
	relayPacketsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "statsd_exporter_relay_packets_total",
			Help: "The number of StatsD packets relayed.",
		},
		[]string{"target"},
	)
	relayLongLinesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "statsd_exporter_relay_long_lines_total",
			Help: "The number lines that were too long to relay.",
		},
		[]string{"target"},
	)
	relayLinesRelayedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "statsd_exporter_relay_lines_relayed_total",
			Help: "The number of lines that were buffered to be relayed.",
		},
		[]string{"target"},
	)
)

// NewRelay creates a statsd UDP relay. It can be used to send copies of statsd raw
// lines to a separate service.
func NewRelay(l log.Logger, target string, packetLength uint) (*Relay, error) {
	addr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve target %s, err: %w", target, err)
	}
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, fmt.Errorf("unable to listen on UDP, err: %w", err)
	}

	c := make(chan []byte, 100)

	r := Relay{
		addr:          addr,
		bufferChannel: c,
		conn:          conn,
		logger:        l,
		packetLength:  packetLength,

		packetsTotal:      relayPacketsTotal.WithLabelValues(target),
		longLinesTotal:    relayLongLinesTotal.WithLabelValues(target),
		relayedLinesTotal: relayLinesRelayedTotal.WithLabelValues(target),
	}

	// Startup the UDP sender.
	go r.relayOutput()

	return &r, nil
}

// relayOutput buffers statsd lines and sends them to the relay target.
func (r *Relay) relayOutput() {
	var buffer bytes.Buffer
	var err error

	relayInterval := clock.NewTicker(1 * time.Second)
	defer relayInterval.Stop()

	for {
		select {
		case <-relayInterval.C:
			err = r.sendPacket(buffer.Bytes())
			if err != nil {
				level.Error(r.logger).Log("msg", "Error sending UDP packet", "error", err)
				return
			}
			// Clear out the buffer.
			buffer.Reset()
		case b := <-r.bufferChannel:
			if uint(len(b)+buffer.Len()) > r.packetLength {
				level.Debug(r.logger).Log("msg", "Buffer full, sending packet", "length", buffer.Len())
				err = r.sendPacket(buffer.Bytes())
				if err != nil {
					level.Error(r.logger).Log("msg", "Error sending UDP packet", "error", err)
					return
				}
				// Seed the new buffer with the new line.
				buffer.Reset()
				buffer.Write(b)
			} else {
				level.Debug(r.logger).Log("msg", "Adding line to buffer", "line", b)
				buffer.Write(b)
			}
		}
	}
}

// sendPacket sends a single relay line to the destination target.
func (r *Relay) sendPacket(buf []byte) error {
	if len(buf) == 0 {
		level.Debug(r.logger).Log("msg", "Empty buffer, nothing to send")
		return nil
	}
	level.Debug(r.logger).Log("msg", "Sending packet", "length", len(buf), "data", buf)
	_, err := r.conn.WriteToUDP(buf, r.addr)
	r.packetsTotal.Inc()
	return err
}

// RelayLine processes a single statsd line and forwards it to the relay target.
func (r *Relay) RelayLine(l string) {
	lineLength := uint(len(l))
	if lineLength == 0 {
		level.Debug(r.logger).Log("msg", "Empty line, not relaying")
		return
	}
	if lineLength > r.packetLength-1 {
		level.Warn(r.logger).Log("msg", "line too long, not relaying", "length", lineLength, "max", r.packetLength)
		r.longLinesTotal.Inc()
		return
	}
	level.Debug(r.logger).Log("msg", "Relaying line", "line", l)
	if !strings.HasSuffix(l, "\n") {
		l = l + "\n"
	}
	r.relayedLinesTotal.Inc()
	r.bufferChannel <- []byte(l)
}
