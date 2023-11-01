package lokipush

// This code is copied from Promtail (3478e180211c17bfe2f3f3305f668d5520f40481) with changes kept to the minimum.
// The lokipush package is used to configure and run the HTTP server that can receive loki push API requests and
// forward them to other loki components.

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/loki/client"
	"github.com/grafana/agent/component/common/loki/client/fake"
	fnet "github.com/grafana/agent/component/common/net"
	frelabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/dskit/flagext"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/grafana/river"
	"github.com/phayes/freeport"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

const localhost = "127.0.0.1"

func TestLokiPushTarget(t *testing.T) {
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	pt, port, eh := createPushServer(t, logger)

	pt.SetLabels(model.LabelSet{
		"pushserver": "pushserver1",
		"dropme":     "label",
	})
	pt.SetKeepTimestamp(true)

	relabelRule := frelabel.Config{}
	relabelStr := `
action = "labeldrop"
regex = "dropme"
`
	err := river.Unmarshal([]byte(relabelStr), &relabelRule)
	require.NoError(t, err)
	pt.SetRelabelRules(frelabel.Rules{&relabelRule})

	// Build a client to send logs
	serverURL := flagext.URLValue{}
	err = serverURL.Set("http://" + localhost + ":" + strconv.Itoa(port) + "/api/v1/push")
	require.NoError(t, err)

	ccfg := client.Config{
		URL:       serverURL,
		Timeout:   1 * time.Second,
		BatchWait: 1 * time.Second,
		BatchSize: 100 * 1024,
	}
	m := client.NewMetrics(prometheus.DefaultRegisterer)
	pc, err := client.New(m, ccfg, 0, 0, false, logger)
	require.NoError(t, err)
	defer pc.Stop()

	// Send some logs
	labels := model.LabelSet{
		"stream":             "stream1",
		"__anotherdroplabel": "dropme",
	}
	for i := 0; i < 100; i++ {
		pc.Chan() <- loki.Entry{
			Labels: labels,
			Entry: logproto.Entry{
				Timestamp: time.Unix(int64(i), 0),
				Line:      "line" + strconv.Itoa(i),
			},
		}
	}

	// Wait for them to appear in the test handler
	countdown := 10000
	for len(eh.Received()) != 100 && countdown > 0 {
		time.Sleep(1 * time.Millisecond)
		countdown--
	}

	// Make sure we didn't timeout
	require.Equal(t, 100, len(eh.Received()))

	// Verify labels
	expectedLabels := model.LabelSet{
		"pushserver": "pushserver1",
		"stream":     "stream1",
	}
	// Spot check the first value in the result to make sure relabel rules were applied properly
	require.Equal(t, expectedLabels, eh.Received()[0].Labels)

	// With keep timestamp enabled, verify timestamp
	require.Equal(t, time.Unix(99, 0).Unix(), eh.Received()[99].Timestamp.Unix())

	pt.Shutdown()
}

func TestLokiPushTargetForRedirect(t *testing.T) {
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	pt, port, eh := createPushServer(t, logger)

	pt.SetLabels(model.LabelSet{
		"pushserver": "pushserver1",
		"dropme":     "label",
	})
	pt.SetKeepTimestamp(true)

	relabelRule := frelabel.Config{}
	relabelStr := `
action = "labeldrop"
regex = "dropme"
`
	err := river.Unmarshal([]byte(relabelStr), &relabelRule)
	require.NoError(t, err)
	pt.SetRelabelRules(frelabel.Rules{&relabelRule})

	// Build a client to send logs
	serverURL := flagext.URLValue{}
	err = serverURL.Set("http://" + localhost + ":" + strconv.Itoa(port) + "/loki/api/v1/push")
	require.NoError(t, err)

	ccfg := client.Config{
		URL:       serverURL,
		Timeout:   1 * time.Second,
		BatchWait: 1 * time.Second,
		BatchSize: 100 * 1024,
	}
	m := client.NewMetrics(prometheus.DefaultRegisterer)
	pc, err := client.New(m, ccfg, 0, 0, false, logger)
	require.NoError(t, err)
	defer pc.Stop()

	// Send some logs
	labels := model.LabelSet{
		"stream":             "stream1",
		"__anotherdroplabel": "dropme",
	}
	for i := 0; i < 100; i++ {
		pc.Chan() <- loki.Entry{
			Labels: labels,
			Entry: logproto.Entry{
				Timestamp: time.Unix(int64(i), 0),
				Line:      "line" + strconv.Itoa(i),
			},
		}
	}

	// Wait for them to appear in the test handler
	countdown := 10000
	for len(eh.Received()) != 100 && countdown > 0 {
		time.Sleep(1 * time.Millisecond)
		countdown--
	}

	// Make sure we didn't timeout
	require.Equal(t, 100, len(eh.Received()))

	// Verify labels
	expectedLabels := model.LabelSet{
		"pushserver": "pushserver1",
		"stream":     "stream1",
	}
	// Spot check the first value in the result to make sure relabel rules were applied properly
	require.Equal(t, expectedLabels, eh.Received()[0].Labels)

	// With keep timestamp enabled, verify timestamp
	require.Equal(t, time.Unix(99, 0).Unix(), eh.Received()[99].Timestamp.Unix())

	pt.Shutdown()
}

func TestPlaintextPushTarget(t *testing.T) {
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	//Create PushAPIServerOld
	eh := fake.NewClient(func() {})
	defer eh.Stop()

	// Get a randomly available port by open and closing a TCP socket
	addr, err := net.ResolveTCPAddr("tcp", localhost+":0")
	require.NoError(t, err)
	l, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	err = l.Close()
	require.NoError(t, err)

	serverConfig := &fnet.ServerConfig{
		HTTP: &fnet.HTTPConfig{
			ListenAddress: localhost,
			ListenPort:    port,
		},
		GRPC: &fnet.GRPCConfig{ListenPort: getFreePort(t)},
	}

	pt, err := NewPushAPIServer(logger, serverConfig, eh, prometheus.NewRegistry())
	require.NoError(t, err)

	err = pt.Run()
	require.NoError(t, err)

	pt.SetLabels(model.LabelSet{
		"pushserver": "pushserver2",
		"keepme":     "label",
	})
	pt.SetKeepTimestamp(true)

	// Send some logs
	ts := time.Now()
	body := new(bytes.Buffer)
	for i := 0; i < 100; i++ {
		body.WriteString("line" + strconv.Itoa(i))
		_, err := http.Post(fmt.Sprintf("http://%s:%d/api/v1/raw", localhost, port), "text/json", body)
		require.NoError(t, err)
		body.Reset()
	}

	// Wait for them to appear in the test handler
	countdown := 10000
	for len(eh.Received()) != 100 && countdown > 0 {
		time.Sleep(1 * time.Millisecond)
		countdown--
	}

	// Make sure we didn't timeout
	require.Equal(t, 100, len(eh.Received()))

	// Verify labels
	expectedLabels := model.LabelSet{
		"pushserver": "pushserver2",
		"keepme":     "label",
	}
	// Spot check the first value in the result to make sure relabel rules were applied properly
	require.Equal(t, expectedLabels, eh.Received()[0].Labels)

	// Timestamp is always set in the handler, we expect received timestamps to be slightly higher than the timestamp when we started sending logs.
	require.GreaterOrEqual(t, eh.Received()[99].Timestamp.Unix(), ts.Unix())

	pt.Shutdown()
}

func TestReady(t *testing.T) {
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)

	//Create PushAPIServerOld
	eh := fake.NewClient(func() {})
	defer eh.Stop()

	// Get a randomly available port by open and closing a TCP socket
	addr, err := net.ResolveTCPAddr("tcp", localhost+":0")
	require.NoError(t, err)
	l, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	err = l.Close()
	require.NoError(t, err)

	serverConfig := &fnet.ServerConfig{
		HTTP: &fnet.HTTPConfig{
			ListenAddress: localhost,
			ListenPort:    port,
		},
		GRPC: &fnet.GRPCConfig{ListenPort: getFreePort(t)},
	}

	pt, err := NewPushAPIServer(logger, serverConfig, eh, prometheus.NewRegistry())
	require.NoError(t, err)

	err = pt.Run()
	require.NoError(t, err)

	pt.SetLabels(model.LabelSet{
		"pushserver": "pushserver2",
		"keepme":     "label",
	})
	pt.SetKeepTimestamp(true)

	url := fmt.Sprintf("http://%s:%d/ready", localhost, port)
	response, err := http.Get(url)
	if err != nil {
		require.NoError(t, err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		require.NoError(t, err)
	}
	responseCode := fmt.Sprint(response.StatusCode)
	responseBody := string(body)

	fmt.Println(responseBody)
	wantedResponse := "ready"
	if responseBody != wantedResponse {
		t.Errorf("got the response %q, want %q", responseBody, wantedResponse)
	}
	wantedCode := "200"
	if responseCode != wantedCode {
		t.Errorf("Got the response code %q, want %q", responseCode, wantedCode)
	}

	t.Cleanup(pt.Shutdown)
}

func getFreePort(t *testing.T) int {
	port, err := freeport.GetFreePort()
	require.NoError(t, err)
	return port
}

func createPushServer(t *testing.T, logger log.Logger) (*PushAPIServer, int, *fake.Client) {
	//Create PushAPIServerOld
	eh := fake.NewClient(func() {})
	t.Cleanup(func() {
		eh.Stop()
	})

	// Get a randomly available port by open and closing a TCP socket
	port := getFreePort(t)

	serverConfig := &fnet.ServerConfig{
		HTTP: &fnet.HTTPConfig{
			ListenAddress: localhost,
			ListenPort:    port,
		},
		GRPC: &fnet.GRPCConfig{ListenPort: getFreePort(t)},
	}

	pt, err := NewPushAPIServer(logger, serverConfig, eh, prometheus.NewRegistry())
	require.NoError(t, err)

	err = pt.Run()
	require.NoError(t, err)
	return pt, port, eh
}
