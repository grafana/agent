package write

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/pyroscope"
	"github.com/oklog/run"
	commonconfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"go.uber.org/multierr"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/pkg/build"
	"github.com/grafana/dskit/backoff"
	pushv1 "github.com/grafana/pyroscope/api/gen/proto/go/push/v1"
	"github.com/grafana/pyroscope/api/gen/proto/go/push/v1/pushv1connect"
	typesv1 "github.com/grafana/pyroscope/api/gen/proto/go/types/v1"
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
		Name:    "pyroscope.write",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(o component.Options, c component.Arguments) (component.Component, error) {
			return New(o, c.(Arguments))
		},
	})
}

// Arguments represents the input state of the pyroscope.write
// component.
type Arguments struct {
	ExternalLabels map[string]string  `river:"external_labels,attr,optional"`
	Endpoints      []*EndpointOptions `river:"endpoint,block,optional"`
}

// SetToDefault implements river.Defaulter.
func (rc *Arguments) SetToDefault() {
	*rc = DefaultArguments()
}

// EndpointOptions describes an individual location for where profiles
// should be delivered to using the Pyroscope push API.
type EndpointOptions struct {
	Name              string                   `river:"name,attr,optional"`
	URL               string                   `river:"url,attr"`
	RemoteTimeout     time.Duration            `river:"remote_timeout,attr,optional"`
	Headers           map[string]string        `river:"headers,attr,optional"`
	HTTPClientConfig  *config.HTTPClientConfig `river:",squash"`
	MinBackoff        time.Duration            `river:"min_backoff_period,attr,optional"`  // start backoff at this level
	MaxBackoff        time.Duration            `river:"max_backoff_period,attr,optional"`  // increase exponentially to this level
	MaxBackoffRetries int                      `river:"max_backoff_retries,attr,optional"` // give up after this many; zero means infinite retries
}

func GetDefaultEndpointOptions() EndpointOptions {
	defaultEndpointOptions := EndpointOptions{
		RemoteTimeout:     10 * time.Second,
		MinBackoff:        500 * time.Millisecond,
		MaxBackoff:        5 * time.Minute,
		MaxBackoffRetries: 10,
		HTTPClientConfig:  config.CloneDefaultHTTPClientConfig(),
	}

	return defaultEndpointOptions
}

// SetToDefault implements river.Defaulter.
func (r *EndpointOptions) SetToDefault() {
	*r = GetDefaultEndpointOptions()
}

// Validate implements river.Validator.
func (r *EndpointOptions) Validate() error {
	// We must explicitly Validate because HTTPClientConfig is squashed and it won't run otherwise
	if r.HTTPClientConfig != nil {
		return r.HTTPClientConfig.Validate()
	}

	return nil
}

// Component is the pyroscope.write component.
type Component struct {
	opts    component.Options
	cfg     Arguments
	metrics *metrics
}

// Exports are the set of fields exposed by the pyroscope.write component.
type Exports struct {
	Receiver pyroscope.Appendable `river:"receiver,attr"`
}

// New creates a new pyroscope.write component.
func New(o component.Options, c Arguments) (*Component, error) {
	metrics := newMetrics(o.Registerer)
	receiver, err := NewFanOut(o, c, metrics)
	if err != nil {
		return nil, err
	}
	// Immediately export the receiver
	o.OnStateChange(Exports{Receiver: receiver})

	return &Component{
		cfg:     c,
		opts:    o,
		metrics: metrics,
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
	level.Debug(c.opts.Logger).Log("msg", "updating pyroscope.write config", "old", c.cfg, "new", newConfig)
	receiver, err := NewFanOut(c.opts, newConfig.(Arguments), c.metrics)
	if err != nil {
		return err
	}
	c.opts.OnStateChange(Exports{Receiver: receiver})
	return nil
}

type fanOutClient struct {
	// The list of push clients to fan out to.
	clients []pushv1connect.PusherServiceClient

	config  Arguments
	opts    component.Options
	metrics *metrics
}

// NewFanOut creates a new fan out client that will fan out to all endpoints.
func NewFanOut(opts component.Options, config Arguments, metrics *metrics) (*fanOutClient, error) {
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
		metrics: metrics,
	}, nil
}

// Push implements the PusherServiceClient interface.
func (f *fanOutClient) Push(ctx context.Context, req *connect.Request[pushv1.PushRequest]) (*connect.Response[pushv1.PushResponse], error) {
	// Don't flow the context down to the `run.Group`.
	// We want to fan out to all even in case of failures to one.
	var (
		g                     run.Group
		errs                  error
		reqSize, profileCount = requestSize(req)
	)

	for i, client := range f.clients {
		var (
			client  = client
			i       = i
			backoff = backoff.New(ctx, backoff.Config{
				MinBackoff: f.config.Endpoints[i].MinBackoff,
				MaxBackoff: f.config.Endpoints[i].MaxBackoff,
				MaxRetries: f.config.Endpoints[i].MaxBackoffRetries,
			})
			err error
		)
		g.Add(func() error {
			req := connect.NewRequest(req.Msg)
			for k, v := range f.config.Endpoints[i].Headers {
				req.Header().Set(k, v)
			}
			for {
				err = func() error {
					ctx, cancel := context.WithTimeout(ctx, f.config.Endpoints[i].RemoteTimeout)
					defer cancel()

					_, err := client.Push(ctx, req)
					return err
				}()
				if err == nil {
					f.metrics.sentBytes.WithLabelValues(f.config.Endpoints[i].URL).Add(float64(reqSize))
					f.metrics.sentProfiles.WithLabelValues(f.config.Endpoints[i].URL).Add(float64(profileCount))
					break
				}
				level.Warn(f.opts.Logger).Log("msg", "failed to push to endpoint", "endpoint", f.config.Endpoints[i].URL, "err", err)
				if !shouldRetry(err) {
					break
				}
				backoff.Wait()
				if !backoff.Ongoing() {
					break
				}
				f.metrics.retries.WithLabelValues(f.config.Endpoints[i].URL).Inc()
			}
			if err != nil {
				f.metrics.droppedBytes.WithLabelValues(f.config.Endpoints[i].URL).Add(float64(reqSize))
				f.metrics.droppedProfiles.WithLabelValues(f.config.Endpoints[i].URL).Add(float64(profileCount))
				level.Warn(f.opts.Logger).Log("msg", "final error sending to profiles to endpoint", "endpoint", f.config.Endpoints[i].URL, "err", err)
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

func shouldRetry(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	switch connect.CodeOf(err) {
	case connect.CodeDeadlineExceeded, connect.CodeUnknown,
		connect.CodeResourceExhausted, connect.CodeInternal,
		connect.CodeUnavailable, connect.CodeDataLoss, connect.CodeAborted:
		return true
	}
	return false
}

func requestSize(req *connect.Request[pushv1.PushRequest]) (int64, int64) {
	var size, profiles int64
	for _, raw := range req.Msg.Series {
		for _, sample := range raw.Samples {
			size += int64(len(sample.RawProfile))
			profiles++
		}
	}
	return size, profiles
}

// Append implements the pyroscope.Appendable interface.
func (f *fanOutClient) Appender() pyroscope.Appender {
	return f
}

// Append implements the Appender interface.
func (f *fanOutClient) Append(ctx context.Context, lbs labels.Labels, samples []*pyroscope.RawSample) error {
	// todo(ctovena): we should probably pool the label pair arrays and label builder to avoid allocs.
	var (
		protoLabels  = make([]*typesv1.LabelPair, 0, len(lbs)+len(f.config.ExternalLabels))
		protoSamples = make([]*pushv1.RawSample, 0, len(samples))
		lbsBuilder   = labels.NewBuilder(nil)
	)

	for _, label := range lbs {
		// filter reserved labels, with exceptions for __name__ and __delta__.
		if strings.HasPrefix(label.Name, model.ReservedLabelPrefix) &&
			label.Name != labels.MetricName &&
			label.Name != pyroscope.LabelNameDelta {

			continue
		}
		lbsBuilder.Set(label.Name, label.Value)
	}
	for name, value := range f.config.ExternalLabels {
		lbsBuilder.Set(name, value)
	}
	for _, l := range lbsBuilder.Labels() {
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
