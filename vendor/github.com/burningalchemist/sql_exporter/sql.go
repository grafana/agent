package sql_exporter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/xo/dburl"
	"k8s.io/klog/v2"
)

// OpenConnection parses a provided DSN, and opens a DB handle ensuring early termination if the context is closed
// (this is actually prevented by `database/sql` implementation), sets connection limits and returns the handle.
func OpenConnection(ctx context.Context, logContext, dsn string, maxConns, maxIdleConns int, maxConnLifetime time.Duration) (*sql.DB, error) {
	var (
		url  *dburl.URL
		conn *sql.DB
		err  error
		ch   = make(chan error)
	)

	url, err = safeParse(dsn)
	if err != nil {
		return nil, err
	}

	// Open the DB handle in a separate goroutine so we can terminate early if the context closes.
	go func() {
		conn, err = sql.Open(url.Driver, url.DSN)
		close(ch)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-ch:
		if err != nil {
			return nil, err
		}
	}

	conn.SetMaxIdleConns(maxIdleConns)
	conn.SetMaxOpenConns(maxConns)
	conn.SetConnMaxLifetime(maxConnLifetime)

	if klog.V(1).Enabled() {
		if len(logContext) > 0 {
			logContext = fmt.Sprintf("[%s] ", logContext)
		}
		klog.Infof("%sDatabase handle successfully opened with '%s' driver", logContext, url.Driver)
	}
	return conn, nil
}

// PingDB is a wrapper around sql.DB.PingContext() that terminates as soon as the context is closed.
//
// sql.DB does not actually pass along the context to the driver when opening a connection (which always happens if the
// database is down) and the driver uses an arbitrary timeout which may well be longer than ours. So we run the ping
// call in a goroutine and terminate immediately if the context is closed.
func PingDB(ctx context.Context, conn *sql.DB) error {
	ch := make(chan error, 1)

	go func() {
		ch <- conn.PingContext(ctx)
		close(ch)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-ch:
		return err
	}
}

// safeParse wraps dburl.Parse method in order to prevent leaking credentials
// if underlying url parse failed. By default it returns a raw url string in error message,
// which most likely contains a password. It's undesired here.
func safeParse(rawURL string) (*dburl.URL, error) {
	parsed, err := dburl.Parse(rawURL)
	if err != nil {
		if uerr := new(url.Error); errors.As(err, &uerr) {
			return nil, uerr.Err
		}
		return nil, errors.New("invalid URL")
	}
	return parsed, nil
}
