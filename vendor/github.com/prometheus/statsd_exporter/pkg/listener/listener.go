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

package listener

import (
	"bufio"
	"io"
	"net"
	"os"
	"strings"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/statsd_exporter/pkg/event"
	"github.com/prometheus/statsd_exporter/pkg/level"
	"github.com/prometheus/statsd_exporter/pkg/relay"
)

type Parser interface {
	LineToEvents(line string, sampleErrors prometheus.CounterVec, samplesReceived prometheus.Counter, tagErrors prometheus.Counter, tagsReceived prometheus.Counter, logger log.Logger) event.Events
}

type StatsDUDPListener struct {
	Conn            *net.UDPConn
	EventHandler    event.EventHandler
	Logger          log.Logger
	LineParser      Parser
	UDPPackets      prometheus.Counter
	LinesReceived   prometheus.Counter
	EventsFlushed   prometheus.Counter
	Relay           *relay.Relay
	SampleErrors    prometheus.CounterVec
	SamplesReceived prometheus.Counter
	TagErrors       prometheus.Counter
	TagsReceived    prometheus.Counter
}

func (l *StatsDUDPListener) SetEventHandler(eh event.EventHandler) {
	l.EventHandler = eh
}

func (l *StatsDUDPListener) Listen() {
	buf := make([]byte, 65535)
	for {
		n, _, err := l.Conn.ReadFromUDP(buf)
		if err != nil {
			// https://github.com/golang/go/issues/4373
			// ignore net: errClosing error as it will occur during shutdown
			if strings.HasSuffix(err.Error(), "use of closed network connection") {
				return
			}
			level.Error(l.Logger).Log("error", err)
			return
		}
		l.HandlePacket(buf[0:n])
	}
}

func (l *StatsDUDPListener) HandlePacket(packet []byte) {
	l.UDPPackets.Inc()
	lines := strings.Split(string(packet), "\n")
	for _, line := range lines {
		level.Debug(l.Logger).Log("msg", "Incoming line", "proto", "udp", "line", line)
		l.LinesReceived.Inc()
		if l.Relay != nil && len(line) > 0 {
			l.Relay.RelayLine(line)
		}
		l.EventHandler.Queue(l.LineParser.LineToEvents(line, l.SampleErrors, l.SamplesReceived, l.TagErrors, l.TagsReceived, l.Logger))
	}
}

type StatsDTCPListener struct {
	Conn            *net.TCPListener
	EventHandler    event.EventHandler
	Logger          log.Logger
	LineParser      Parser
	LinesReceived   prometheus.Counter
	EventsFlushed   prometheus.Counter
	Relay           *relay.Relay
	SampleErrors    prometheus.CounterVec
	SamplesReceived prometheus.Counter
	TagErrors       prometheus.Counter
	TagsReceived    prometheus.Counter
	TCPConnections  prometheus.Counter
	TCPErrors       prometheus.Counter
	TCPLineTooLong  prometheus.Counter
}

func (l *StatsDTCPListener) SetEventHandler(eh event.EventHandler) {
	l.EventHandler = eh
}

func (l *StatsDTCPListener) Listen() {
	for {
		c, err := l.Conn.AcceptTCP()
		if err != nil {
			// https://github.com/golang/go/issues/4373
			// ignore net: errClosing error as it will occur during shutdown
			if strings.HasSuffix(err.Error(), "use of closed network connection") {
				return
			}
			level.Error(l.Logger).Log("msg", "AcceptTCP failed", "error", err)
			os.Exit(1)
		}
		go l.HandleConn(c)
	}
}

func (l *StatsDTCPListener) HandleConn(c *net.TCPConn) {
	defer c.Close()

	l.TCPConnections.Inc()

	r := bufio.NewReader(c)
	for {
		line, isPrefix, err := r.ReadLine()
		if err != nil {
			if err != io.EOF {
				l.TCPErrors.Inc()
				level.Debug(l.Logger).Log("msg", "Read failed", "addr", c.RemoteAddr(), "error", err)
			}
			break
		}
		level.Debug(l.Logger).Log("msg", "Incoming line", "proto", "tcp", "line", line)
		if isPrefix {
			l.TCPLineTooLong.Inc()
			level.Debug(l.Logger).Log("msg", "Read failed: line too long", "addr", c.RemoteAddr())
			break
		}
		l.LinesReceived.Inc()
		if l.Relay != nil && len(line) > 0 {
			l.Relay.RelayLine(string(line))
		}
		l.EventHandler.Queue(l.LineParser.LineToEvents(string(line), l.SampleErrors, l.SamplesReceived, l.TagErrors, l.TagsReceived, l.Logger))
	}
}

type StatsDUnixgramListener struct {
	Conn            *net.UnixConn
	EventHandler    event.EventHandler
	Logger          log.Logger
	LineParser      Parser
	UnixgramPackets prometheus.Counter
	LinesReceived   prometheus.Counter
	EventsFlushed   prometheus.Counter
	Relay           *relay.Relay
	SampleErrors    prometheus.CounterVec
	SamplesReceived prometheus.Counter
	TagErrors       prometheus.Counter
	TagsReceived    prometheus.Counter
}

func (l *StatsDUnixgramListener) SetEventHandler(eh event.EventHandler) {
	l.EventHandler = eh
}

func (l *StatsDUnixgramListener) Listen() {
	buf := make([]byte, 65535)
	for {
		n, _, err := l.Conn.ReadFromUnix(buf)
		if err != nil {
			// https://github.com/golang/go/issues/4373
			// ignore net: errClosing error as it will occur during shutdown
			if strings.HasSuffix(err.Error(), "use of closed network connection") {
				return
			}
			level.Error(l.Logger).Log(err)
			os.Exit(1)
		}
		l.HandlePacket(buf[:n])
	}
}

func (l *StatsDUnixgramListener) HandlePacket(packet []byte) {
	l.UnixgramPackets.Inc()
	lines := strings.Split(string(packet), "\n")
	for _, line := range lines {
		level.Debug(l.Logger).Log("msg", "Incoming line", "proto", "unixgram", "line", line)
		l.LinesReceived.Inc()
		if l.Relay != nil && len(line) > 0 {
			l.Relay.RelayLine(line)
		}
		l.EventHandler.Queue(l.LineParser.LineToEvents(line, l.SampleErrors, l.SamplesReceived, l.TagErrors, l.TagsReceived, l.Logger))
	}
}
