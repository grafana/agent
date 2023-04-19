package http

import (
	"context"
	"fmt"
	"github.com/efficientgo/core/errors"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/pkg/flow/logging"
	"io"
	"net/http"
)

type Arguments struct {
	HttpEndpoint string              `river:"http_endpoint,attr"`
	ForwardTo    []loki.LogsReceiver `river:"forward_to,attr"`

	//TODO: add support for additional labels, like loki push api in promtail https://grafana.com/docs/loki/next/clients/promtail/configuration/#loki_push_api
	//TODO: add support for use_incoming_timestamp like in https://grafana.com/docs/loki/next/clients/promtail/configuration/#loki_push_api
	//TODO: add support for http_server_read_timeout like in https://grafana.com/docs/loki/next/clients/promtail/configuration/#loki_push_api
	//TODO: add support for http_server_write_timeout like in https://grafana.com/docs/loki/next/clients/promtail/configuration/#loki_push_api
	//TODO: add support for http_server_idle_timeout like in https://grafana.com/docs/loki/next/clients/promtail/configuration/#loki_push_api
}

type Component struct {
	opts component.Options
	args Arguments
}

func init() {
	component.Register(component.Registration{
		Name: "loki.source.http",
		Args: Arguments{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments)), nil
		},
	})
}

func New(opts component.Options, args Arguments) component.Component {
	return &Component{
		opts: opts,
		args: args,
	}
}

func (c *Component) Run(ctx context.Context) error {
	c.opts.Logger.Log("msg", "starting component")

	server := &http.Server{Addr: c.args.HttpEndpoint, Handler: &httpSinkHandler{log: c.opts.Logger}}
	serverDone := make(chan struct{})

	go func() {
		c.opts.Logger.Log("msg", "server listening")
		err := server.ListenAndServe()
		if err != nil {
			c.opts.Logger.Log("msg", "server failed", "error", err)
		}
		serverDone <- struct{}{}
	}()

	for {
		select {
		case <-ctx.Done():
			c.opts.Logger.Log("msg", "finishing due to context done")
			return nil
		case <-serverDone:
			c.opts.Logger.Log("msg", "finishing due to server exit")
			return nil
		}
	}
}

func (c *Component) Update(args component.Arguments) error {
	if newArgs, ok := args.(Arguments); !ok {
		return errors.Newf("invalid type of arguments: %T", args)
	} else {
		c.args = newArgs
	}
	return nil
}

// ------------- debug -----------
type httpSinkHandler struct {
	log *logging.Logger
}

func (h *httpSinkHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	allBody, _ := io.ReadAll(request.Body)
	h.log.Log(
		"msg", "got request!",
		"url", request.URL.String(),
		"method", request.Method,
		"body", string(allBody),
		"headers", fmt.Sprintf("%+#v", request.Header))
	_, _ = writer.Write([]byte("thank you for your request!"))
}

// -------------------------------
