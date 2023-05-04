package memconn

import (
	"context"
	"net"
	"time"
)

// Pipe creates a single pair of in-memory net.Conns. Use NewListener if you
// want to create more than one connection.
func Pipe() (local, remote net.Conn) {
	// NOTE(rfratto): Things are horribly broken if they fail here, so we panic
	// on errors instead of making the user deal with it.

	lis := NewListener(nil)
	defer func() {
		_ = lis.Close()
	}()

	connCh := make(chan net.Conn, 1)

	go func() {
		conn, err := lis.Accept()
		if err != nil {
			panic(err)
		}
		connCh <- conn
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	local, err := lis.DialContext(ctx)
	if err != nil {
		panic(err)
	}

	select {
	case <-ctx.Done():
		panic(ctx.Err())
	case remote = <-connCh:
		return local, remote
	}
}
