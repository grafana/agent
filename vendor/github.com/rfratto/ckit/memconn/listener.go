package memconn

import (
	"context"
	"fmt"
	"net"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// Listener is an in-memory net.Listener. Call DialContext to create a new
// connection.
type Listener struct {
	log log.Logger

	pending chan *conn
	closed  chan struct{}
}

var _ net.Listener = (*Listener)(nil)

// NewListener creates a new in-memory Listener.
func NewListener(l log.Logger) *Listener {
	if l == nil {
		l = log.NewNopLogger()
	}
	return &Listener{
		log:     l,
		pending: make(chan *conn),
		closed:  make(chan struct{}),
	}
}

// Accept waits for and returns the next connection to l. Connections to l are
// established by calling l.DialContext.
//
// The returned net.Conn is the server side of the connection.
func (l *Listener) Accept() (net.Conn, error) {
	select {
	case peer := <-l.pending:
		level.Debug(l.log).Log("msg", "accepted new in-memory connection")

		local := newConn()
		peer.Attach(local)
		local.Attach(peer)
		return local, nil

	case <-l.closed:
		return nil, fmt.Errorf("Listener closed")
	}
}

// Close closes l. Any blocked Accept operations will immediately be unblocked
// and return errors. Already Accepted connections are not closed.
func (l *Listener) Close() error {
	select {
	default:
		close(l.closed)
		return nil
	case <-l.closed:
		return fmt.Errorf("already closed")
	}
}

// Addr returns l's address. This will always be a fake "memory"
// address.
func (l *Listener) Addr() net.Addr {
	return Addr{}
}

// DialContext creates a new connection to l. DialContext will block until the
// connection is accepted through a blocked l.Accept call or until ctx is
// canceled.
//
// Note that unlike other Dial methods in different packages, there is no
// address to supply because the remote side of the connection is always the
// in-memory listener.
func (l *Listener) DialContext(ctx context.Context) (net.Conn, error) {
	local := newConn()

	select {
	case l.pending <- local:
		level.Debug(l.log).Log("msg", "dialed to in-memory listener")

		// Wait for our peer to be connected.
		if err := local.WaitPeer(ctx); err != nil {
			return nil, err
		}
		return local, nil
	case <-l.closed:
		return nil, fmt.Errorf("server closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
