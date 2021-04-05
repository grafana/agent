package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/common/server"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
)

// Config is a server config.
type Config = server.Config

// Server is a Weaveworks server with support for reloading.
type Server struct {
	reg *util.Unregisterer
	log log.Logger

	// Last received config, used for seeing if any changes need to be made.
	cfg Config

	// The current server is stored for shutting down when closing, but otherwise
	// the most recent server to run is kept in srvCh. srv should always be shut
	// down before replacing it, reventing the run loop from running an old
	// server.
	srvMut sync.Mutex
	srv    *server.Server
	srvCh  chan *server.Server

	// reloading determine if a Server ApplyConfig is currently running.
	// This is required by the Run loop to know if a new server will
	// be created after the old one shuts down.
	reloading *atomic.Bool

	// doneCh is used by the Run loop to detected a closed Server.
	// closeOnce is used to close it.
	closeOnce sync.Once
	doneCh    chan bool
}

// New creates a new Server. ApplyConfig must be called after creating a server.
func New(r prometheus.Registerer, l log.Logger) *Server {
	return &Server{
		reg: util.WrapWithUnregisterer(r),
		log: l,

		srvCh:     make(chan *server.Server, 1),
		reloading: atomic.NewBool(false),
		doneCh:    make(chan bool),
	}
}

// ApplyConfig applies changes to the Server block. wire will be called when
// the server is recreated, and should be used to hook up endpoints.
//
// If the SignalHandler is not set, it will default to not watching for
// signals. This contrasts the upstream default of watching for termination
// signals.
//
// ApplyConfig will override the registerer of the Config to the registerer
// passed to New.
func (s *Server) ApplyConfig(cfg Config, wire func(mux *mux.Router, grpc *grpc.Server)) error {
	s.srvMut.Lock()
	defer s.srvMut.Unlock()

	if util.CompareYAML(&s.cfg, &cfg) {
		return nil
	}

	level.Info(s.log).Log("msg", "server configuration changed, restarting server")

	// We're going to create a new server, so we need to unregister existing
	// metrics.
	s.reg.UnregisterAll()
	cfg.Registerer = s.reg

	if cfg.SignalHandler == nil {
		cfg.SignalHandler = newNoopSignalHandler()
	}

	if s.srv != nil {
		s.reloading.Store(true)
		s.srv.Shutdown()
	}

	var err error
	s.srv, err = server.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to recreate server: %w", err)
	}

	wire(s.srv.HTTP, s.srv.GRPC)

	s.srvCh <- s.srv

	s.cfg = cfg
	return nil
}

// Run starts the Server. Run will block until an error occurs or until Close
// is called.
func (s *Server) Run() error {
	// Read in servers as they get created and run them.
	//
	// During a reload, the current server will be shut down, and a new
	// one will take its place. This scenario is detected through an atomic
	// bool that indicates a reload, and is always set before the shutdown.
	//
	// If the server shuts down independently of a reload, this is treated as
	// a fatal error, and the loop exits.
NextServer:
	for {
		select {
		case <-s.doneCh:
			return fmt.Errorf("server stopping")
		case nextSrv := <-s.srvCh:
			// If the reload failed, s will be nil. Skip this loop and wait for the
			// next recv.
			if s == nil {
				continue NextServer
			}

			err := nextSrv.Run()

			// If we're reloading, wait for the next server. Note this causes an edge
			// case where the server shuts down from a problem in the middle of a
			// reload. Since a new server is going to replace it, it's safe to
			// ignore.
			if s.reloading.CAS(true, false) {
				continue NextServer
			}

			return err
		}
	}
}

// Close closes the Server.
func (s *Server) Close() {
	s.srvMut.Lock()
	defer s.srvMut.Unlock()

	// Prevent the run loop from running again.
	s.closeOnce.Do(func() {
		close(s.doneCh)
	})

	// Shut down the current server. This will cause the run loop to stop, if
	// it's currently running.
	if s.srv != nil {
		s.srv.Shutdown()
	}
}

// noopSignalHandler implements the SignalHandler interface used by
// weaveworks/common/server.
type noopSignalHandler struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func newNoopSignalHandler() *noopSignalHandler {
	var sh noopSignalHandler
	sh.ctx, sh.cancel = context.WithCancel(context.Background())
	return &sh
}

// Equal implements the equality checking interface used by cmp.
func (sh *noopSignalHandler) Equal(*noopSignalHandler) bool {
	return true
}

func (sh *noopSignalHandler) Loop() {
	<-sh.ctx.Done()
}

func (sh *noopSignalHandler) Stop() {
	sh.cancel()
}
