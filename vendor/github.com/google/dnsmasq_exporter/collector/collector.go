// Package collector collects dnsmasq statistics as a Prometheus collector.
package collector

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

var (
	// floatMetrics contains prometheus Gauges, keyed by the stats DNS record
	// they correspond to.
	floatMetrics = map[string]*prometheus.Desc{
		"cachesize.bind.": prometheus.NewDesc(
			"dnsmasq_cachesize",
			"configured size of the DNS cache",
			nil, nil,
		),

		"insertions.bind.": prometheus.NewDesc(
			"dnsmasq_insertions",
			"DNS cache insertions",
			nil, nil,
		),

		"evictions.bind.": prometheus.NewDesc(
			"dnsmasq_evictions",
			"DNS cache exictions: numbers of entries which replaced an unexpired cache entry",
			nil, nil,
		),

		"misses.bind.": prometheus.NewDesc(
			"dnsmasq_misses",
			"DNS cache misses: queries which had to be forwarded",
			nil, nil,
		),

		"hits.bind.": prometheus.NewDesc(
			"dnsmasq_hits",
			"DNS queries answered locally (cache hits)",
			nil, nil,
		),

		"auth.bind.": prometheus.NewDesc(
			"dnsmasq_auth",
			"DNS queries for authoritative zones",
			nil, nil,
		),
	}

	serversMetrics = map[string]*prometheus.Desc{
		"queries": prometheus.NewDesc(
			"dnsmasq_servers_queries",
			"DNS queries on upstream server",
			[]string{"server"}, nil,
		),
		"queries_failed": prometheus.NewDesc(
			"dnsmasq_servers_queries_failed",
			"DNS queries failed on upstream server",
			[]string{"server"}, nil,
		),
	}

	leases = prometheus.NewDesc(
		"dnsmasq_leases",
		"Number of DHCP leases handed out",
		nil, nil,
	)
)

// From https://manpages.debian.org/stretch/dnsmasq-base/dnsmasq.8.en.html:
// The cache statistics are also available in the DNS as answers to queries of
// class CHAOS and type TXT in domain bind. The domain names are cachesize.bind,
// insertions.bind, evictions.bind, misses.bind, hits.bind, auth.bind and
// servers.bind. An example command to query this, using the dig utility would
// be:
//     dig +short chaos txt cachesize.bind

// Collector implements prometheus.Collector and exposes dnsmasq metrics.
type Collector struct {
	log         log.Logger
	dnsClient   *dns.Client
	dnsmasqAddr string
	leasesPath  string
}

// New creates a new Collector.
func New(l log.Logger, client *dns.Client, dnsmasqAddr string, leasesPath string) *Collector {
	if l == nil {
		l = log.NewNopLogger()
	}

	return &Collector{
		log:         l,
		dnsClient:   client,
		dnsmasqAddr: dnsmasqAddr,
		leasesPath:  leasesPath,
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range floatMetrics {
		ch <- d
	}
	for _, d := range serversMetrics {
		ch <- d
	}
	ch <- leases
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	var eg errgroup.Group

	eg.Go(func() error {
		msg := &dns.Msg{
			MsgHdr: dns.MsgHdr{
				Id:               dns.Id(),
				RecursionDesired: true,
			},
			Question: []dns.Question{
				question("cachesize.bind."),
				question("insertions.bind."),
				question("evictions.bind."),
				question("misses.bind."),
				question("hits.bind."),
				question("auth.bind."),
				question("servers.bind."),
			},
		}
		in, _, err := c.dnsClient.Exchange(msg, c.dnsmasqAddr)
		if err != nil {
			return err
		}
		for _, a := range in.Answer {
			txt, ok := a.(*dns.TXT)
			if !ok {
				continue
			}
			switch txt.Hdr.Name {
			case "servers.bind.":
				for _, str := range txt.Txt {
					arr := strings.Fields(str)
					if got, want := len(arr), 3; got != want {
						return fmt.Errorf("stats DNS record servers.bind.: unexpeced number of argument in record: got %d, want %d", got, want)
					}
					queries, err := strconv.ParseFloat(arr[1], 64)
					if err != nil {
						return err
					}
					failedQueries, err := strconv.ParseFloat(arr[2], 64)
					if err != nil {
						return err
					}
					ch <- prometheus.MustNewConstMetric(serversMetrics["queries"], prometheus.GaugeValue, queries, arr[0])
					ch <- prometheus.MustNewConstMetric(serversMetrics["queries_failed"], prometheus.GaugeValue, failedQueries, arr[0])
				}
			default:
				g, ok := floatMetrics[txt.Hdr.Name]
				if !ok {
					continue // ignore unexpected answer from dnsmasq
				}
				if got, want := len(txt.Txt), 1; got != want {
					return fmt.Errorf("stats DNS record %q: unexpected number of replies: got %d, want %d", txt.Hdr.Name, got, want)
				}
				f, err := strconv.ParseFloat(txt.Txt[0], 64)
				if err != nil {
					return err
				}
				ch <- prometheus.MustNewConstMetric(g, prometheus.GaugeValue, f)
			}
		}
		return nil
	})

	eg.Go(func() error {
		f, err := os.Open(c.leasesPath)
		if err != nil {
			level.Warn(c.log).Log("msg", "could not open leases file", "err", err)
			return err
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		var lines float64
		for scanner.Scan() {
			lines++
		}
		if err := scanner.Err(); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(leases, prometheus.GaugeValue, lines)
		return nil
	})

	if err := eg.Wait(); err != nil {
		level.Warn(c.log).Log("msg", "could not complete scrape", "err", err)
	}
}

func question(name string) dns.Question {
	return dns.Question{
		Name:   name,
		Qtype:  dns.TypeTXT,
		Qclass: dns.ClassCHAOS,
	}
}
