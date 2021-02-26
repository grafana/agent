package ha

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/agentproto"
	"github.com/grafana/agent/pkg/prom/ha/configapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// APIHandler is a function that returns a configapi Response type
// and optionally an error.
type APIHandler func(r *http.Request) (interface{}, error)

// API can wire HTTP and gRPC routes.
type API struct {
	mut         sync.Mutex
	server      *Server
	logger      log.Logger
	scrapingApi *deferredScrapingServiceServer
}

// NewAPI creates a new API for a given server s. s may be nil and updated later.
func NewAPI(log log.Logger, s *Server) *API {
	return &API{
		server:      s,
		logger:      log,
		scrapingApi: &deferredScrapingServiceServer{server: s},
	}
}

// SetServer updates the server used by the API.
func (api *API) SetServer(s *Server) {
	api.mut.Lock()
	defer api.mut.Unlock()

	api.server = s
	api.scrapingApi.SetServer(s)
}

// WireAPI injects routes into the provided mux router for the config
// management API.
func (api *API) WireAPI(r *mux.Router) {
	var (
		listConfig   = api.wrapHandler(func(s *Server) APIHandler { return s.ListConfigurations })
		getConfig    = api.wrapHandler(func(s *Server) APIHandler { return s.GetConfiguration })
		deleteConfig = api.wrapHandler(func(s *Server) APIHandler { return s.DeleteConfiguration })

		// putConfig doesn't support multiple concurrent requests since it acts as
		// a non-atomic CAS.
		putConfig = nonConcurrentHTTPHandler(
			api.wrapHandler(func(s *Server) APIHandler { return s.PutConfiguration }),
		)
	)

	// Support URL-encoded config names. The handlers will need to decode the
	// name when reading the path variable.
	r = r.UseEncodedPath()

	r.HandleFunc("/agent/api/v1/configs", listConfig).Methods("GET")
	r.HandleFunc("/agent/api/v1/configs/{name}", getConfig).Methods("GET")
	r.HandleFunc("/agent/api/v1/config/{name}", putConfig).Methods("PUT", "POST")
	r.HandleFunc("/agent/api/v1/config/{name}", deleteConfig).Methods("DELETE")

	r.HandleFunc("/debug/ring", func(rw http.ResponseWriter, r *http.Request) {
		api.mut.Lock()
		defer api.mut.Unlock()

		if api.server == nil {
			err := configapi.WriteError(rw, http.StatusNotFound, fmt.Errorf("scraping service not enabled"))
			if err != nil {
				level.Error(api.logger).Log("msg", "failed writing error response to client", "err", err)
			}
			return
		}

		api.server.ring.ServeHTTP(rw, r)
	})
}

// WireGRPC injects gRPC server handlers into the provided gRPC server.
func (api *API) WireGRPC(srv *grpc.Server) {
	agentproto.RegisterScrapingServiceServer(srv, api.scrapingApi)
}

// wrapHandler is responsible for turning an APIHandler into an HTTP
// handler by wrapping responses and writing them as JSON.
func (api *API) wrapHandler(getHandler func(s *Server) APIHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.mut.Lock()
		defer api.mut.Unlock()

		if api.server == nil {
			err := configapi.WriteError(w, http.StatusNotFound, fmt.Errorf("scraping service not enabled"))
			if err != nil {
				level.Error(api.logger).Log("msg", "failed writing error response to client", "err", err)
			}
			return
		}

		next := getHandler(api.server)
		resp, err := next(r)
		if err != nil {
			httpErr, ok := err.(*httpError)

			if ok {
				err = configapi.WriteError(w, httpErr.StatusCode, httpErr.Err)
			} else {
				err = configapi.WriteError(w, http.StatusInternalServerError, err)
			}

			if err != nil {
				level.Error(api.logger).Log("msg", "failed writing error response to client", "err", err)
			}
			return
		}

		// Prepare data and status code to send back to the writer: if the handler
		// returned an *httpResponse, use the status code defined there and send the
		// internal data. Otherwise, assume HTTP 200 OK and marshal the raw response.
		var (
			data       = resp
			statusCode = http.StatusOK
		)
		if httpResp, ok := data.(*httpResponse); ok {
			data = httpResp.Data
			statusCode = httpResp.StatusCode
		}

		if err := configapi.WriteResponse(w, statusCode, data); err != nil {
			level.Error(api.logger).Log("msg", "failed to write valid response", "err", err)
		}
	})
}

// nonConcurrentHTTPHandler wraps an http.HandlerFunc in a mutex.
func nonConcurrentHTTPHandler(next http.HandlerFunc) http.HandlerFunc {
	var mut sync.Mutex
	return func(rw http.ResponseWriter, r *http.Request) {
		mut.Lock()
		defer mut.Unlock()
		next(rw, r)
	}
}

type deferredScrapingServiceServer struct {
	mut    sync.Mutex
	server *Server
}

func (s *deferredScrapingServiceServer) SetServer(server *Server) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.server = server
}

func (s *deferredScrapingServiceServer) Reshard(ctx context.Context, req *agentproto.ReshardRequest) (*empty.Empty, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	if s.server == nil {
		return nil, status.Errorf(codes.Unimplemented, "scraping service not running")
	}

	return s.server.Reshard(ctx, req)
}

type httpError struct {
	StatusCode int
	Err        error
}

func (e httpError) Error() string { return e.Err.Error() }

type httpResponse struct {
	StatusCode int
	Data       interface{}
}
