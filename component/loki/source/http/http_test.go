package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/regexp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLokiSourceHTTP(t *testing.T) {
	args := defaultTestArgsWith(func(a *Arguments) {
		a.HTTPAddress = "127.0.0.1"
		a.HTTPPort = 8532
	})

	comp, err := New(defaultOptions(t), args)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		err = comp.Run(ctx)
		require.NoError(t, err)

	}()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/", args.HTTPAddress, args.HTTPPort), strings.NewReader("hello"))
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	_ = res

	//c, ok := comp.(*Component)
	//require.True(t, ok)

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
				args.HTTPAddress = "192.168.0.1"
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
				args.RelabelRules = flow_relabel.Rules{}
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

			got := c.pushTargetNeedsUpdate(c.pushTargetConfigForArgs(tc.newArgs))
			assert.Equal(t, got, tc.restartRequired)
		})
	}
}

func defaultOptions(t *testing.T) component.Options {
	return component.Options{
		ID:         "loki.source.http.test",
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
		RelabelRules: flow_relabel.Rules{
			{
				SourceLabels: []string{"app"},
				Regex:        flow_relabel.Regexp{Regexp: regexp.MustCompile("backend")},
				Action:       flow_relabel.Keep,
			},
		},
		UseIncomingTimestamp: false,
	}
}
