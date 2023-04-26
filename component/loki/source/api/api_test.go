package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	client2 "github.com/grafana/agent/component/common/loki/client"
	"github.com/grafana/agent/component/common/loki/client/fake"

	"github.com/grafana/dskit/flagext"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/common/model"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/regexp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLokiSourceAPI_Simple(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	receiver := fake.NewClient(func() {})
	defer receiver.Stop()

	args := defaultTestArgsWith(func(a *Arguments) {
		a.HTTPPort = 8532
		a.ForwardTo = []loki.LogsReceiver{receiver.LogsReceiver()}
		a.UseIncomingTimestamp = true
	})
	opts := defaultOptions(t)
	startTestComponent(t, opts, args, ctx)

	lokiClient := newTestLokiClient(t, args, opts)
	defer lokiClient.Stop()

	now := time.Now()
	select {
	case lokiClient.Chan() <- loki.Entry{
		Labels: map[model.LabelName]model.LabelValue{"source": "test"},
		Entry:  logproto.Entry{Timestamp: now, Line: "hello world!"},
	}:
	case <-ctx.Done():
		t.Fatalf("timed out while sending test entries via loki client")
	}

	require.Eventually(
		t,
		func() bool { return len(receiver.Received()) == 1 },
		5*time.Second,
		10*time.Millisecond,
		"did not receive the forwarded message within the timeout",
	)
	received := receiver.Received()[0]
	assert.Equal(t, received.Line, "hello world!")
	assert.Equal(t, received.Timestamp.Unix(), now.Unix())
	assert.Equal(t, received.Labels, model.LabelSet{
		"source": "test",
		"foo":    "bar",
		"fizz":   "buzz",
	})
}

func TestLokiSourceAPI_Update(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	receiver := fake.NewClient(func() {})
	defer receiver.Stop()

	args := defaultTestArgsWith(func(a *Arguments) {
		a.HTTPPort = 8583
		a.ForwardTo = []loki.LogsReceiver{receiver.LogsReceiver()}
		a.UseIncomingTimestamp = true
		a.Labels = map[string]string{"test_label": "before"}
	})
	opts := defaultOptions(t)
	c := startTestComponent(t, opts, args, ctx)

	lokiClient := newTestLokiClient(t, args, opts)
	defer lokiClient.Stop()

	now := time.Now()
	select {
	case lokiClient.Chan() <- loki.Entry{
		Labels: map[model.LabelName]model.LabelValue{"source": "test"},
		Entry:  logproto.Entry{Timestamp: now, Line: "hello world!"},
	}:
	case <-ctx.Done():
		t.Fatalf("timed out while sending test entries via loki client")
	}

	require.Eventually(
		t,
		func() bool { return len(receiver.Received()) == 1 },
		5*time.Second,
		10*time.Millisecond,
		"did not receive the forwarded message within the timeout",
	)
	received := receiver.Received()[0]
	assert.Equal(t, received.Line, "hello world!")
	assert.Equal(t, received.Timestamp.Unix(), now.Unix())
	assert.Equal(t, received.Labels, model.LabelSet{
		"test_label": "before",
		"source":     "test",
	})

	args.Labels = map[string]string{"test_label": "after"}
	err := c.Update(args)
	require.NoError(t, err)

	receiver.Clear()

	select {
	case lokiClient.Chan() <- loki.Entry{
		Labels: map[model.LabelName]model.LabelValue{"source": "test"},
		Entry:  logproto.Entry{Timestamp: now, Line: "hello brave new world!"},
	}:
	case <-ctx.Done():
		t.Fatalf("timed out while sending test entries via loki client")
	}
	require.Eventually(
		t,
		func() bool { return len(receiver.Received()) == 1 },
		5*time.Second,
		10*time.Millisecond,
		"did not receive the forwarded message within the timeout",
	)
	received = receiver.Received()[0]
	assert.Equal(t, received.Line, "hello brave new world!")
	assert.Equal(t, received.Timestamp.Unix(), now.Unix())
	assert.Equal(t, received.Labels, model.LabelSet{
		"test_label": "after",
		"source":     "test",
	})
}

func TestLokiSourceAPI_FanOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const receiversCount = 10
	var receivers = make([]*fake.Client, receiversCount)
	for i := 0; i < receiversCount; i++ {
		receivers[i] = fake.NewClient(func() {})
	}

	args := defaultTestArgsWith(func(a *Arguments) {
		a.HTTPPort = 8537
		a.ForwardTo = mapToChannels(receivers)
	})
	opts := defaultOptions(t)

	comp, err := New(opts, args)
	require.NoError(t, err)
	go func() {
		err := comp.Run(ctx)
		require.NoError(t, err)
	}()

	lokiClient := newTestLokiClient(t, args, opts)
	defer lokiClient.Stop()

	const messagesCount = 100
	for i := 0; i < messagesCount; i++ {
		entry := loki.Entry{
			Labels: map[model.LabelName]model.LabelValue{"source": "test"},
			Entry:  logproto.Entry{Line: fmt.Sprintf("test message #%d", i)},
		}
		select {
		case lokiClient.Chan() <- entry:
		case <-ctx.Done():
			t.Log("timed out while sending test entries via loki client")
		}
	}

	require.Eventually(
		t,
		func() bool {
			for i := 0; i < receiversCount; i++ {
				if len(receivers[i].Received()) != messagesCount {
					return false
				}
			}
			return true
		},
		5*time.Second,
		10*time.Millisecond,
		"did not receive all the expected messages within the timeout",
	)
}

func TestComponent_detectsWhenUpdateRequiresARestart(t *testing.T) {
	tests := []struct {
		name            string
		args            Arguments
		newArgs         Arguments
		restartRequired bool
	}{
		{
			name:            "identical args don't require server restart",
			args:            defaultTestArgs(),
			newArgs:         defaultTestArgs(),
			restartRequired: false,
		},
		{
			name: "change in address requires server restart",
			args: defaultTestArgs(),
			newArgs: defaultTestArgsWith(func(args *Arguments) {
				args.HTTPAddress = "localhost"
			}),
			restartRequired: true,
		},
		{
			name: "change in port requires server restart",
			args: defaultTestArgs(),
			newArgs: defaultTestArgsWith(func(args *Arguments) {
				args.HTTPPort = 7777
			}),
			restartRequired: true,
		},
		{
			name: "change in forwardTo does NOT requires server restart",
			args: defaultTestArgs(),
			newArgs: defaultTestArgsWith(func(args *Arguments) {
				args.ForwardTo = []loki.LogsReceiver{}
			}),
			restartRequired: false,
		},
		{
			name: "change in labels requires server restart",
			args: defaultTestArgs(),
			newArgs: defaultTestArgsWith(func(args *Arguments) {
				args.Labels = map[string]string{"some": "label"}
			}),
			restartRequired: true,
		},
		{
			name: "change in labels requires server restart",
			args: defaultTestArgs(),
			newArgs: defaultTestArgsWith(func(args *Arguments) {
				args.Labels = map[string]string{"some": "label"}
			}),
			restartRequired: true,
		},
		{
			name: "change in relabel rules requires server restart",
			args: defaultTestArgs(),
			newArgs: defaultTestArgsWith(func(args *Arguments) {
				args.RelabelRules = relabel.Rules{}
			}),
			restartRequired: true,
		},
		{
			name: "change in use incoming timestamp requires server restart",
			args: defaultTestArgs(),
			newArgs: defaultTestArgsWith(func(args *Arguments) {
				args.UseIncomingTimestamp = !args.UseIncomingTimestamp
			}),
			restartRequired: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			comp, err := New(
				defaultOptions(t),
				tc.args,
			)
			require.NoError(t, err)

			c, ok := comp.(*Component)
			require.True(t, ok)

			pushTargetBefore := c.pushTarget

			err = c.Update(tc.newArgs)
			require.NoError(t, err)

			restarted := pushTargetBefore != c.pushTarget
			assert.Equal(t, restarted, tc.restartRequired)
		})
	}
}

func startTestComponent(t *testing.T, opts component.Options, args Arguments, ctx context.Context) component.Component {
	comp, err := New(opts, args)
	require.NoError(t, err)
	go func() {
		err := comp.Run(ctx)
		require.NoError(t, err)
	}()
	return comp
}

func mapToChannels(clients []*fake.Client) []loki.LogsReceiver {
	channels := make([]loki.LogsReceiver, len(clients))
	for i := 0; i < len(clients); i++ {
		channels[i] = clients[i].LogsReceiver()
	}
	return channels
}

func newTestLokiClient(t *testing.T, args Arguments, opts component.Options) client2.Client {
	url := flagext.URLValue{}
	err := url.Set(fmt.Sprintf("http://%s:%d/api/v1/push", args.HTTPAddress, args.HTTPPort))
	require.NoError(t, err)

	lokiClient, err := client2.New(
		client2.NewMetrics(nil, nil),
		client2.Config{
			URL:     url,
			Timeout: 5 * time.Second,
		},
		[]string{},
		0,
		opts.Logger,
	)
	require.NoError(t, err)
	return lokiClient
}

func defaultOptions(t *testing.T) component.Options {
	return component.Options{
		ID:         "loki.source.api.test",
		Logger:     util.TestFlowLogger(t),
		Registerer: prometheus.NewRegistry(),
	}
}

func defaultTestArgsWith(mutator func(arguments *Arguments)) Arguments {
	a := defaultTestArgs()
	mutator(&a)
	return a
}

func defaultTestArgs() Arguments {
	return Arguments{
		HTTPAddress: "127.0.0.1",
		HTTPPort:    0,
		ForwardTo:   []loki.LogsReceiver{make(chan loki.Entry), make(chan loki.Entry)},
		Labels:      map[string]string{"foo": "bar", "fizz": "buzz"},
		RelabelRules: relabel.Rules{
			{
				SourceLabels: []string{"tag"},
				Regex:        relabel.Regexp{Regexp: regexp.MustCompile("ignore")},
				Action:       relabel.Drop,
			},
		},
		UseIncomingTimestamp: false,
	}
}
