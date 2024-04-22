package syslogtarget

// This code is copied from Promtail. The syslogtarget package is used to
// configure and run the targets that can read syslog entries and forward them
// to other loki components.

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/grafana/agent/internal/component/common/loki/client/fake"

	"github.com/go-kit/log"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	"github.com/grafana/loki/clients/pkg/promtail/targets/syslog/syslogparser"
	"github.com/influxdata/go-syslog/v3"
	promconfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var (
	caCert = []byte(`
-----BEGIN CERTIFICATE-----
Test Cert
-----END CERTIFICATE-----
`)

	// Unused, but can be useful to (re)generate some certificates
	// nolint:deadcode,unused,varcheck
	caKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
Test Cert
-----END RSA PRIVATE KEY-----
`)

	serverCert = []byte(`
-----BEGIN CERTIFICATE-----
Test Cert
-----END CERTIFICATE-----
`)

	serverKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
Test Cert
-----END RSA PRIVATE KEY-----
`)

	clientCert = []byte(`
-----BEGIN CERTIFICATE-----
Test Cert
-----END CERTIFICATE-----
`)
	clientKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
Test Cert
-----END RSA PRIVATE KEY-----
`)
)

type formatFunc func(string) string

var (
	fmtOctetCounting = func(s string) string { return fmt.Sprintf("%d %s", len(s), s) }
	fmtNewline       = func(s string) string { return s + "\n" }
)

func Benchmark_SyslogTarget(b *testing.B) {
	for _, tt := range []struct {
		name       string
		protocol   string
		formatFunc formatFunc
	}{
		{"tcp", protocolTCP, fmtOctetCounting},
		{"udp", protocolUDP, fmtOctetCounting},
	} {
		tt := tt
		b.Run(tt.name, func(b *testing.B) {
			client := fake.NewClient(func() {})

			metrics := NewMetrics(nil)
			tgt, _ := NewSyslogTarget(metrics, log.NewNopLogger(), client, []*relabel.Config{}, &scrapeconfig.SyslogTargetConfig{
				ListenAddress:       "127.0.0.1:0",
				ListenProtocol:      tt.protocol,
				LabelStructuredData: true,
				Labels: model.LabelSet{
					"test": "syslog_target",
				},
			})
			b.Cleanup(func() {
				require.NoError(b, tgt.Stop())
			})
			require.Eventually(b, tgt.Ready, time.Second, 10*time.Millisecond)

			addr := tgt.ListenAddress().String()

			messages := []string{
				`<165>1 2022-04-08T22:14:10.001Z host1 app - id1 [custom@32473 exkey="1"] An application event log entry...`,
				`<165>1 2022-04-08T22:14:11.002Z host2 app - id2 [custom@32473 exkey="1"] An application event log entry...`,
				`<165>1 2022-04-08T22:14:12.003Z host1 app - id3 [custom@32473 exkey="1"] An application event log entry...`,
				`<165>1 2022-04-08T22:14:13.004Z host2 app - id4 [custom@32473 exkey="1"] An application event log entry...`,
				`<165>1 2022-04-08T22:14:14.005Z host1 app - id5 [custom@32473 exkey="1"] An application event log entry...`,
				`<165>1 2022-04-08T22:14:15.002Z host2 app - id6 [custom@32473 exkey="1"] An application event log entry...`,
				`<165>1 2022-04-08T22:14:16.003Z host1 app - id7 [custom@32473 exkey="1"] An application event log entry...`,
				`<165>1 2022-04-08T22:14:17.004Z host2 app - id8 [custom@32473 exkey="1"] An application event log entry...`,
				`<165>1 2022-04-08T22:14:18.005Z host1 app - id9 [custom@32473 exkey="1"] An application event log entry...`,
				`<165>1 2022-04-08T22:14:19.001Z host2 app - id10 [custom@32473 exkey="1"] An application event log entry...`,
			}

			b.ReportAllocs()
			b.ResetTimer()

			c, _ := net.Dial(tt.protocol, addr)
			for n := 0; n < b.N; n++ {
				_ = writeMessagesToStream(c, messages, tt.formatFunc)
			}
			c.Close()

			require.Eventuallyf(b, func() bool {
				return len(client.Received()) == len(messages)*b.N
			}, 15*time.Second, time.Second, "expected: %d got:%d", len(messages)*b.N, len(client.Received()))
		})
	}
}

func TestSyslogTarget(t *testing.T) {
	for _, tt := range []struct {
		name     string
		protocol string
		fmtFunc  formatFunc
	}{
		{"tcp newline separated", protocolTCP, fmtNewline},
		{"tcp octetcounting", protocolTCP, fmtOctetCounting},
		{"udp newline separated", protocolUDP, fmtNewline},
		{"udp octetcounting", protocolUDP, fmtOctetCounting},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			w := log.NewSyncWriter(os.Stderr)
			logger := log.NewLogfmtLogger(w)
			client := fake.NewClient(func() {})

			metrics := NewMetrics(nil)
			tgt, err := NewSyslogTarget(metrics, logger, client, relabelConfig(t), &scrapeconfig.SyslogTargetConfig{
				MaxMessageLength:    1 << 12, // explicitly not use default value
				ListenAddress:       "127.0.0.1:0",
				ListenProtocol:      tt.protocol,
				LabelStructuredData: true,
				Labels: model.LabelSet{
					"test": "syslog_target",
				},
			})
			require.NoError(t, err)

			require.Eventually(t, tgt.Ready, time.Second, 10*time.Millisecond)

			addr := tgt.ListenAddress().String()
			c, err := net.Dial(tt.protocol, addr)
			require.NoError(t, err)

			messages := []string{
				`<165>1 2018-10-11T22:14:15.003Z host5 e - id1 [custom@32473 exkey="1"] An application event log entry...`,
				`<165>1 2018-10-11T22:14:15.005Z host5 e - id2 [custom@32473 exkey="2"] An application event log entry...`,
				`<165>1 2018-10-11T22:14:15.007Z host5 e - id3 [custom@32473 exkey="3"] An application event log entry...`,
			}

			err = writeMessagesToStream(c, messages, tt.fmtFunc)
			require.NoError(t, err)
			require.NoError(t, c.Close())

			if tt.protocol == protocolUDP {
				time.Sleep(time.Second)
				require.NoError(t, tgt.Stop())
			} else {
				defer func() {
					require.NoError(t, tgt.Stop())
				}()
			}

			require.Eventuallyf(t, func() bool {
				return len(client.Received()) == len(messages)
			}, time.Second, 10*time.Millisecond, "Expected to receive %d messages.", len(messages))

			labels := make([]model.LabelSet, 0, len(messages))
			for _, entry := range client.Received() {
				labels = append(labels, entry.Labels)
			}
			// we only check if one of the received entries contain the wanted label set
			// because UDP does not guarantee the order of the messages
			require.Contains(t, labels, model.LabelSet{
				"test": "syslog_target",

				"severity": "notice",
				"facility": "local4",
				"hostname": "host5",
				"app_name": "e",
				"msg_id":   "id1",

				"sd_custom_exkey": "1",
			})
			require.Equal(t, "An application event log entry...", client.Received()[0].Line)

			require.NotZero(t, client.Received()[0].Timestamp)
		})
	}
}

func relabelConfig(t *testing.T) []*relabel.Config {
	relabelCfg := `
- source_labels: ['__syslog_message_severity']
  target_label: 'severity'
- source_labels: ['__syslog_message_facility']
  target_label: 'facility'
- source_labels: ['__syslog_message_hostname']
  target_label: 'hostname'
- source_labels: ['__syslog_message_app_name']
  target_label: 'app_name'
- source_labels: ['__syslog_message_proc_id']
  target_label: 'proc_id'
- source_labels: ['__syslog_message_msg_id']
  target_label: 'msg_id'
- source_labels: ['__syslog_message_sd_custom_32473_exkey']
  target_label: 'sd_custom_exkey'
`

	var relabels []*relabel.Config
	err := yaml.Unmarshal([]byte(relabelCfg), &relabels)
	require.NoError(t, err)

	return relabels
}

func writeMessagesToStream(w io.Writer, messages []string, formatter formatFunc) error {
	for _, msg := range messages {
		_, err := fmt.Fprint(w, formatter(msg))
		if err != nil {
			return err
		}
	}
	return nil
}

func TestSyslogTarget_RFC5424Messages(t *testing.T) {
	for _, tt := range []struct {
		name     string
		protocol string
		fmtFunc  formatFunc
	}{
		{"tcp newline separated", protocolTCP, fmtNewline},
		{"tcp octetcounting", protocolTCP, fmtOctetCounting},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			w := log.NewSyncWriter(os.Stderr)
			logger := log.NewLogfmtLogger(w)
			client := fake.NewClient(func() {})

			metrics := NewMetrics(nil)
			tgt, err := NewSyslogTarget(metrics, logger, client, []*relabel.Config{}, &scrapeconfig.SyslogTargetConfig{
				ListenAddress:       "127.0.0.1:0",
				ListenProtocol:      tt.protocol,
				LabelStructuredData: true,
				Labels: model.LabelSet{
					"test": "syslog_target",
				},
				UseRFC5424Message: true,
			})
			require.NoError(t, err)
			require.Eventually(t, tgt.Ready, time.Second, 10*time.Millisecond)
			defer func() {
				require.NoError(t, tgt.Stop())
			}()

			addr := tgt.ListenAddress().String()
			c, err := net.Dial(tt.protocol, addr)
			require.NoError(t, err)

			messages := []string{
				`<165>1 2018-10-11T22:14:15.003Z host5 e - id1 [custom@32473 exkey="1"] An application event log entry...`,
				`<165>1 2018-10-11T22:14:15.005Z host5 e - id2 [custom@32473 exkey="2"] An application event log entry...`,
				`<165>1 2018-10-11T22:14:15.007Z host5 e - id3 [custom@32473 exkey="3"] An application event log entry...`,
			}

			err = writeMessagesToStream(c, messages, tt.fmtFunc)
			require.NoError(t, err)
			require.NoError(t, c.Close())

			require.Eventuallyf(t, func() bool {
				return len(client.Received()) == len(messages)
			}, time.Second, time.Millisecond, "Expected to receive %d messages, got %d.", len(messages), len(client.Received()))

			for i := range messages {
				require.Equal(t, model.LabelSet{
					"test": "syslog_target",
				}, client.Received()[i].Labels)
				require.Contains(t, messages, client.Received()[i].Line)
				require.NotZero(t, client.Received()[i].Timestamp)
			}
		})
	}
}

func TestSyslogTarget_TLSConfigWithoutServerCertificate(t *testing.T) {
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	client := fake.NewClient(func() {})

	metrics := NewMetrics(nil)
	_, err := NewSyslogTarget(metrics, logger, client, relabelConfig(t), &scrapeconfig.SyslogTargetConfig{
		ListenAddress: "127.0.0.1:0",
		TLSConfig: promconfig.TLSConfig{
			KeyFile: "foo",
		},
	})
	require.Error(t, err, "error setting up syslog target: certificate and key files are required")
}

func TestSyslogTarget_TLSConfigWithoutServerKey(t *testing.T) {
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	client := fake.NewClient(func() {})

	metrics := NewMetrics(nil)
	_, err := NewSyslogTarget(metrics, logger, client, relabelConfig(t), &scrapeconfig.SyslogTargetConfig{
		ListenAddress: "127.0.0.1:0",
		TLSConfig: promconfig.TLSConfig{
			CertFile: "foo",
		},
	})
	require.Error(t, err, "error setting up syslog target: certificate and key files are required")
}

func TestSyslogTarget_TLSConfig(t *testing.T) {
	t.Run("NewlineSeparatedMessages", func(t *testing.T) {
		testSyslogTargetWithTLS(t, fmtNewline)
	})
	t.Run("OctetCounting", func(t *testing.T) {
		testSyslogTargetWithTLS(t, fmtOctetCounting)
	})
}

func testSyslogTargetWithTLS(t *testing.T, fmtFunc formatFunc) {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	serverCertFile, err := createTempFile(serverCert)
	if err != nil {
		t.Fatalf("Unable to create server certificate temporary file: %s", err)
	}
	defer os.Remove(serverCertFile.Name())

	serverKeyFile, err := createTempFile(serverKey)
	if err != nil {
		t.Fatalf("Unable to create server key temporary file: %s", err)
	}
	defer os.Remove(serverKeyFile.Name())

	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	client := fake.NewClient(func() {})

	metrics := NewMetrics(nil)
	tgt, err := NewSyslogTarget(metrics, logger, client, relabelConfig(t), &scrapeconfig.SyslogTargetConfig{
		ListenAddress:       "127.0.0.1:0",
		LabelStructuredData: true,
		Labels: model.LabelSet{
			"test": "syslog_target",
		},
		TLSConfig: promconfig.TLSConfig{
			CertFile: serverCertFile.Name(),
			KeyFile:  serverKeyFile.Name(),
		},
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, tgt.Stop())
	}()

	tlsConfig := tls.Config{
		RootCAs:    caCertPool,
		ServerName: "promtail.example.com",
	}

	addr := tgt.ListenAddress().String()
	c, err := tls.Dial("tcp", addr, &tlsConfig)
	require.NoError(t, err)

	validMessages := []string{
		`<165>1 2018-10-11T22:14:15.003Z host5 e - id1 [custom@32473 exkey="1"] An application event log entry...`,
		`<165>1 2018-10-11T22:14:15.005Z host5 e - id2 [custom@32473 exkey="2"] An application event log entry...`,
		`<165>1 2018-10-11T22:14:15.007Z host5 e - id3 [custom@32473 exkey="3"] An application event log entry...`,
	}
	// Messages that are malformed but still valid.
	// This causes error messages being written, but the parser does not stop and close the connection.
	malformeddMessages := []string{
		`<165>1    -   An application event log entry...`,
		`<165>1 2018-10-11T22:14:15.007Z host5 e -   An application event log entry...`,
	}
	messages := append(malformeddMessages, validMessages...)

	err = writeMessagesToStream(c, messages, fmtFunc)
	require.NoError(t, err)
	require.NoError(t, c.Close())

	require.Eventuallyf(t, func() bool {
		return len(client.Received()) == len(validMessages)
	}, time.Second, time.Millisecond, "Expected to receive %d messages, got %d.", len(validMessages), len(client.Received()))

	require.Equal(t, model.LabelSet{
		"test": "syslog_target",

		"severity": "notice",
		"facility": "local4",
		"hostname": "host5",
		"app_name": "e",
		"msg_id":   "id1",

		"sd_custom_exkey": "1",
	}, client.Received()[0].Labels)
	require.Equal(t, "An application event log entry...", client.Received()[0].Line)

	require.NotZero(t, client.Received()[0].Timestamp)
}

func createTempFile(data []byte) (*os.File, error) {
	tmpFile, err := os.CreateTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %s", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write data to temporary file: %s", err)
	}

	if err := tmpFile.Close(); err != nil {
		return nil, err
	}

	return tmpFile, nil
}

func TestSyslogTarget_TLSConfigVerifyClientCertificate(t *testing.T) {
	t.Run("NewlineSeparatedMessages", func(t *testing.T) {
		testSyslogTargetWithTLSVerifyClientCertificate(t, fmtNewline)
	})
	t.Run("OctetCounting", func(t *testing.T) {
		testSyslogTargetWithTLSVerifyClientCertificate(t, fmtOctetCounting)
	})
}

func testSyslogTargetWithTLSVerifyClientCertificate(t *testing.T, fmtFunc formatFunc) {
	caCertFile, err := createTempFile(caCert)
	if err != nil {
		t.Fatalf("Unable to create CA certificate temporary file: %s", err)
	}
	defer os.Remove(caCertFile.Name())

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	serverCertFile, err := createTempFile(serverCert)
	if err != nil {
		t.Fatalf("Unable to create server certificate temporary file: %s", err)
	}
	defer os.Remove(serverCertFile.Name())

	serverKeyFile, err := createTempFile(serverKey)
	if err != nil {
		t.Fatalf("Unable to create server key temporary file: %s", err)
	}
	defer os.Remove(serverKeyFile.Name())

	clientCertFile, err := createTempFile(clientCert)
	if err != nil {
		t.Fatalf("Unable to create client certificate temporary file: %s", err)
	}
	defer os.Remove(clientCertFile.Name())

	clientKeyFile, err := createTempFile(clientKey)
	if err != nil {
		t.Fatalf("Unable to create client key temporary file: %s", err)
	}
	defer os.Remove(clientKeyFile.Name())

	clientCerts, err := tls.LoadX509KeyPair(clientCertFile.Name(), clientKeyFile.Name())
	if err != nil {
		t.Fatalf("Unable to load client certificate or key: %s", err)
	}

	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	client := fake.NewClient(func() {})

	metrics := NewMetrics(nil)
	tgt, err := NewSyslogTarget(metrics, logger, client, relabelConfig(t), &scrapeconfig.SyslogTargetConfig{
		ListenAddress:       "127.0.0.1:0",
		LabelStructuredData: true,
		Labels: model.LabelSet{
			"test": "syslog_target",
		},
		TLSConfig: promconfig.TLSConfig{
			CAFile:   caCertFile.Name(),
			CertFile: serverCertFile.Name(),
			KeyFile:  serverKeyFile.Name(),
		},
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, tgt.Stop())
	}()

	tlsConfig := tls.Config{
		RootCAs:    caCertPool,
		ServerName: "promtail.example.com",
	}

	addr := tgt.ListenAddress().String()

	t.Run("WithoutClientCertificate", func(t *testing.T) {
		c, err := tls.Dial("tcp", addr, &tlsConfig)
		require.NoError(t, err)

		err = c.SetDeadline(time.Now().Add(time.Second))
		require.NoError(t, err)

		buf := make([]byte, 1)
		_, err = c.Read(buf)
		require.ErrorContains(t, err, "remote error: tls:")
	})

	t.Run("WithClientCertificate", func(t *testing.T) {
		tlsConfig.Certificates = []tls.Certificate{clientCerts}
		c, err := tls.Dial("tcp", addr, &tlsConfig)
		require.NoError(t, err)

		messages := []string{
			`<165>1 2018-10-11T22:14:15.003Z host5 e - id1 [custom@32473 exkey="1"] An application event log entry...`,
			`<165>1 2018-10-11T22:14:15.005Z host5 e - id2 [custom@32473 exkey="2"] An application event log entry...`,
			`<165>1 2018-10-11T22:14:15.007Z host5 e - id3 [custom@32473 exkey="3"] An application event log entry...`,
		}

		err = writeMessagesToStream(c, messages, fmtFunc)
		require.NoError(t, err)
		require.NoError(t, c.Close())

		require.Eventuallyf(t, func() bool {
			return len(client.Received()) == len(messages)
		}, time.Second, time.Millisecond, "Expected to receive %d messages, got %d.", len(messages), len(client.Received()))

		require.Equal(t, model.LabelSet{
			"test": "syslog_target",

			"severity": "notice",
			"facility": "local4",
			"hostname": "host5",
			"app_name": "e",
			"msg_id":   "id1",

			"sd_custom_exkey": "1",
		}, client.Received()[0].Labels)
		require.Equal(t, "An application event log entry...", client.Received()[0].Line)

		require.NotZero(t, client.Received()[0].Timestamp)
	})
}

func TestSyslogTarget_InvalidData(t *testing.T) {
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	client := fake.NewClient(func() {})
	metrics := NewMetrics(nil)

	tgt, err := NewSyslogTarget(metrics, logger, client, relabelConfig(t), &scrapeconfig.SyslogTargetConfig{
		ListenAddress: "127.0.0.1:0",
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, tgt.Stop())
	}()

	addr := tgt.ListenAddress().String()
	c, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer c.Close()

	_, err = fmt.Fprint(c, "xxx")
	require.NoError(t, err)

	// syslog target should immediately close the connection if sent invalid data
	err = c.SetDeadline(time.Now().Add(time.Second))
	require.NoError(t, err)

	buf := make([]byte, 1)
	_, err = c.Read(buf)
	require.EqualError(t, err, "EOF")
}

func TestSyslogTarget_NonUTF8Message(t *testing.T) {
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	client := fake.NewClient(func() {})
	metrics := NewMetrics(nil)

	tgt, err := NewSyslogTarget(metrics, logger, client, relabelConfig(t), &scrapeconfig.SyslogTargetConfig{
		ListenAddress: "127.0.0.1:0",
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, tgt.Stop())
	}()

	addr := tgt.ListenAddress().String()
	c, err := net.Dial("tcp", addr)
	require.NoError(t, err)

	msg1 := "Some non utf8 \xF8\xF7\xE3\xE4 characters"
	require.False(t, utf8.ValidString(msg1), "msg must no be valid utf8")
	msg2 := "\xF8 other \xF7\xE3\xE4 characters \xE3"
	require.False(t, utf8.ValidString(msg2), "msg must no be valid utf8")

	err = writeMessagesToStream(c, []string{
		"<165>1 - - - - - - " + msg1,
		"<123>1 - - - - - - " + msg2,
	}, fmtOctetCounting)
	require.NoError(t, err)
	require.NoError(t, c.Close())

	require.Eventuallyf(t, func() bool {
		return len(client.Received()) == 2
	}, time.Second, time.Millisecond, "Expected to receive 2 messages, got %d.", len(client.Received()))

	require.Equal(t, msg1, client.Received()[0].Line)
	require.Equal(t, msg2, client.Received()[1].Line)
}

func TestSyslogTarget_IdleTimeout(t *testing.T) {
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	client := fake.NewClient(func() {})
	metrics := NewMetrics(nil)

	tgt, err := NewSyslogTarget(metrics, logger, client, relabelConfig(t), &scrapeconfig.SyslogTargetConfig{
		ListenAddress: "127.0.0.1:0",
		IdleTimeout:   time.Millisecond,
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, tgt.Stop())
	}()

	addr := tgt.ListenAddress().String()
	c, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer c.Close()

	// connection should be closed before the higher timeout
	// from SetDeadline fires
	err = c.SetDeadline(time.Now().Add(time.Second))
	require.NoError(t, err)

	buf := make([]byte, 1)
	_, err = c.Read(buf)
	require.EqualError(t, err, "EOF")
}

func TestParseStream_WithAsyncPipe(t *testing.T) {
	lines := [3]string{
		"<165>1 2018-10-11T22:14:15.003Z host5 e - id1 [custom@32473 exkey=\"1\"] An application event log entry...\n",
		"<165>1 2018-10-11T22:14:15.005Z host5 e - id2 [custom@32473 exkey=\"2\"] An application event log entry...\n",
		"<165>1 2018-10-11T22:14:15.007Z host5 e - id3 [custom@32473 exkey=\"3\"] An application event log entry...\n",
	}

	addr := &net.UDPAddr{IP: net.IP{127, 0, 0, 1}, Port: 1514}
	pipe := NewConnPipe(addr)
	go func() {
		for _, line := range lines {
			_, _ = pipe.Write([]byte(line))
		}
		pipe.Close()
	}()

	results := make([]*syslog.Result, 0)
	cb := func(res *syslog.Result) {
		results = append(results, res)
	}

	err := syslogparser.ParseStream(pipe, cb, DefaultMaxMessageLength)
	require.NoError(t, err)
	require.Equal(t, 3, len(results))
}
