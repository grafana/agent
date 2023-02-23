package write

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	commonconfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"go.uber.org/multierr"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/phlare"
	"github.com/grafana/agent/pkg/build"
	pushv1 "github.com/grafana/phlare/api/gen/proto/go/push/v1"
	pushv1connect "github.com/grafana/phlare/api/gen/proto/go/push/v1/pushv1connect"
	typesv1 "github.com/grafana/phlare/api/gen/proto/go/types/v1"
)

var (
	userAgent        = fmt.Sprintf("GrafanaAgent/%s", build.Version)
	DefaultArguments = func() Arguments {
		return Arguments{}
	}
	_ component.Component = (*Component)(nil)
)

func init() {
	component.Register(component.Registration{
		Name:    "phlare.write",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(o component.Options, c component.Arguments) (component.Component, error) {
			return NewComponent(o, c.(Arguments))
		},
	})
}

// Arguments represents the input state of the phlare.write
// component.
type Arguments struct {
	ExternalLabels map[string]string  `river:"external_labels,attr,optional"`
	Endpoints      []*EndpointOptions `river:"endpoint,block,optional"`
}

// UnmarshalRiver implements river.Unmarshaler.
func (rc *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*rc = DefaultArguments()

	type config Arguments
	return f((*config)(rc))
}

// EndpointOptions describes an individual location for where profiles
// should be delivered to using the Phlare push API.
type EndpointOptions struct {
	Name             string                   `river:"name,attr,optional"`
	URL              string                   `river:"url,attr"`
	RemoteTimeout    time.Duration            `river:"remote_timeout,attr,optional"`
	Headers          map[string]string        `river:"headers,attr,optional"`
	HTTPClientConfig *config.HTTPClientConfig `river:",squash"`
}

func GetDefaultEndpointOptions() EndpointOptions {
	var defaultEndpointOptions = EndpointOptions{
		RemoteTimeout:    30 * time.Second,
		HTTPClientConfig: config.CloneDefaultHTTPClientConfig(),
	}

	return defaultEndpointOptions
}

// UnmarshalRiver implements river.Unmarshaler.
func (r *EndpointOptions) UnmarshalRiver(f func(v interface{}) error) error {
	*r = GetDefaultEndpointOptions()

	type arguments EndpointOptions
	err := f((*arguments)(r))
	if err != nil {
		return err
	}

	// We must explicitly Validate because HTTPClientConfig is squashed and it won't run otherwise
	if r.HTTPClientConfig != nil {
		return r.HTTPClientConfig.Validate()
	}

	return nil
}

// Component is the phlare.write component.
type Component struct {
	opts component.Options
	cfg  Arguments
}

// Exports are the set of fields exposed by the phlare.write component.
type Exports struct {
	Receiver phlare.Appendable `river:"receiver,attr"`
}

// NewComponent creates a new phlare.write component.
func NewComponent(o component.Options, c Arguments) (*Component, error) {
	receiver, err := NewFanOut(o, c)
	if err != nil {
		return nil, err
	}
	// Immediately export the receiver
	o.OnStateChange(Exports{Receiver: receiver})

	return &Component{
		cfg:  c,
		opts: o,
	}, nil
}

var _ component.Component = (*Component)(nil)

// Run implements Component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

// Update implements Component.
func (c *Component) Update(newConfig component.Arguments) error {
	c.cfg = newConfig.(Arguments)
	level.Debug(c.opts.Logger).Log("msg", "updating phlare.write config", "old", c.cfg, "new", newConfig)
	receiver, err := NewFanOut(c.opts, newConfig.(Arguments))
	if err != nil {
		return err
	}
	c.opts.OnStateChange(Exports{Receiver: receiver})
	return nil
}

type fanOutClient struct {
	// The list of push clients to fan out to.
	clients []pushv1connect.PusherServiceClient

	config Arguments
	opts   component.Options
}

// NewFanOut creates a new fan out client that will fan out to all endpoints.
func NewFanOut(opts component.Options, config Arguments) (*fanOutClient, error) {
	clients := make([]pushv1connect.PusherServiceClient, 0, len(config.Endpoints))
	for _, endpoint := range config.Endpoints {
		httpClient, err := commonconfig.NewClientFromConfig(*endpoint.HTTPClientConfig.Convert(), endpoint.Name)
		if err != nil {
			return nil, err
		}
		clients = append(clients, pushv1connect.NewPusherServiceClient(httpClient, endpoint.URL, WithUserAgent(userAgent)))
	}
	return &fanOutClient{
		clients: clients,
		config:  config,
		opts:    opts,
	}, nil
}

// Push implements the PusherServiceClient interface.
func (f *fanOutClient) Push(ctx context.Context, req *connect.Request[pushv1.PushRequest]) (*connect.Response[pushv1.PushResponse], error) {
	// Don't flow the context down to the `run.Group`.
	// We want to fan out to all even in case of failures to one.
	var (
		g    run.Group
		errs error
	)

	for i, client := range f.clients {
		client := client
		i := i
		g.Add(func() error {
			ctx, cancel := context.WithTimeout(ctx, f.config.Endpoints[i].RemoteTimeout)
			defer cancel()

			req := connect.NewRequest(req.Msg)
			for k, v := range f.config.Endpoints[i].Headers {
				req.Header().Set(k, v)
			}
			_, err := client.Push(ctx, req)
			if err != nil {
				f.opts.Logger.Log("msg", "failed to push to endpoint", "endpoint", f.config.Endpoints[i].Name, "err", err)
				errs = multierr.Append(errs, err)
			}
			return err
		}, func(err error) {})
	}
	if err := g.Run(); err != nil {
		return nil, err
	}
	if errs != nil {
		return nil, errs
	}
	return connect.NewResponse(&pushv1.PushResponse{}), nil
}

// Append implements the phlare.Appendable interface.
func (f *fanOutClient) Appender() phlare.Appender {
	return f
}

// Append implements the Appender interface.
func (f *fanOutClient) Append(ctx context.Context, lbs labels.Labels, samples []*phlare.RawSample) error {
	// todo(ctovena): we should probably pool the label pair arrays and label builder to avoid allocs.
	var (
		protoLabels  = make([]*typesv1.LabelPair, 0, len(lbs)+len(f.config.ExternalLabels))
		protoSamples = make([]*pushv1.RawSample, 0, len(samples))
		lbsBuilder   = labels.NewBuilder(nil)
	)

	for _, label := range lbs {
		// only __name__ is required as a private label.
		if strings.HasPrefix(label.Name, model.ReservedLabelPrefix) && label.Name != labels.MetricName {
			continue
		}
		lbsBuilder.Set(label.Name, label.Value)
	}
	for name, value := range f.config.ExternalLabels {
		lbsBuilder.Set(name, value)
	}
	for _, l := range lbsBuilder.Labels(lbs) {
		protoLabels = append(protoLabels, &typesv1.LabelPair{
			Name:  l.Name,
			Value: l.Value,
		})
	}
	for _, sample := range samples {
		protoSamples = append(protoSamples, &pushv1.RawSample{
			RawProfile: sample.RawProfile,
		})
	}
	// push to all clients
	_, err := f.Push(ctx, connect.NewRequest(&pushv1.PushRequest{
		Series: []*pushv1.RawProfileSeries{
			{Labels: protoLabels, Samples: protoSamples},
		},
	}))
	return err
}

// WithUserAgent returns a `connect.ClientOption` that sets the User-Agent header on.
func WithUserAgent(agent string) connect.ClientOption {
	return connect.WithInterceptors(&agentInterceptor{agent})
}

type agentInterceptor struct {
	agent string
}

func (i *agentInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		req.Header().Set("User-Agent", i.agent)
		return next(ctx, req)
	}
}

func (i *agentInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		conn.RequestHeader().Set("User-Agent", i.agent)
		return conn
	}
}

func (i *agentInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}
