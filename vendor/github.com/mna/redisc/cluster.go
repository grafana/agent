package redisc

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

// HashSlots is the number of slots supported by redis cluster.
const HashSlots = 16384

// BgErrorSrc identifies the origin of a background error as reported by calls
// to Cluster.BgError, when set.
type BgErrorSrc uint

// List of possible BgErrorSrc values.
const (
	// ClusterRefresh indicates the error comes from a background refresh of
	// cluster slots mapping, e.g. following reception of a MOVED error.
	ClusterRefresh BgErrorSrc = iota
	// RetryCloseConn indicates the error comes from the call to Close for a
	// previous connection, before retrying a command with a new one.
	RetryCloseConn
)

// A Cluster manages a redis cluster. If the CreatePool field is not nil, a
// redis.Pool is used for each node in the cluster to get connections via Get.
// If it is nil or if Dial is called, redis.Dial is used to get the connection.
//
// All fields must be set prior to using the Cluster value, and must not be
// changed afterwards, as that could be a data race.
type Cluster struct {
	// StartupNodes is the list of initial nodes that make up the cluster. The
	// values are expected as "address:port" (e.g.: "127.0.0.1:6379").
	StartupNodes []string

	// DialOptions is the list of options to set on each new connection.
	DialOptions []redis.DialOption

	// CreatePool is the function to call to create a redis.Pool for the
	// specified TCP address, using the provided options as set in DialOptions.
	// If this field is not nil, a redis.Pool is created for each node in the
	// cluster and the pool is used to manage the connections returned by Get.
	CreatePool func(address string, options ...redis.DialOption) (*redis.Pool, error)

	// PoolWaitTime is the time to wait when getting a connection from a pool
	// configured with MaxActive > 0 and Wait set to true, and MaxActive
	// connections are already in use.
	//
	// If <= 0 (or with Go < 1.7), there is no wait timeout, it will wait
	// indefinitely if Pool.Wait is true.
	PoolWaitTime time.Duration

	// BgError is an optional function to call when a background error occurs
	// that would otherwise go unnoticed. The source of the error is indicated
	// by the parameter of type BgErrorSrc, see the list of BgErrorSrc values
	// for possible error sources. The function may be called in a distinct
	// goroutine, it should not access shared values that are not meant to be
	// used concurrently.
	BgError func(BgErrorSrc, error)

	// LayoutRefresh is an optional function that is called each time a cluster
	// refresh is successfully executed, either by an explicit call to
	// Cluster.Refresh or e.g.  as required following a MOVED error. Note that
	// even though it is unlikely, the old and new mappings could be identical.
	// The function may be called in a separate goroutine, it should not access
	// shared values that are not meant to be used concurrently.
	LayoutRefresh func(old, new [HashSlots][]string)

	mu         sync.RWMutex           // protects following fields
	err        error                  // closed cluster error
	pools      map[string]*redis.Pool // created pools per node address
	masters    map[string]bool        // set of known active master nodes addresses, kept up-to-date
	replicas   map[string]bool        // set of known active replica nodes addresses, kept up-to-date
	mapping    [HashSlots][]string    // hash slot number to master and replica(s) addresses, master is always at [0]
	refreshing bool                   // indicates if there's a refresh in progress
}

// Refresh updates the cluster's internal mapping of hash slots to redis node.
// It calls CLUSTER SLOTS on each known node until one of them succeeds.
//
// It should typically be called after creating the Cluster and before using
// it. The cluster automatically keeps its mapping up-to-date afterwards, based
// on the redis commands' MOVED responses.
func (c *Cluster) Refresh() error {
	c.mu.Lock()
	err := c.err
	if err == nil {
		c.refreshing = true
	}
	c.mu.Unlock()
	if err != nil {
		return err
	}

	return c.refresh(false)
}

func (c *Cluster) refresh(bg bool) error {
	var errMsgs []string
	var oldm, newm [HashSlots][]string

	addrs, _ := c.getNodeAddrs(false)
	for _, addr := range addrs {
		m, err := c.getClusterSlots(addr)
		if err != nil {
			errMsgs = append(errMsgs, err.Error())
			continue
		}

		// succeeded, save as mapping
		c.mu.Lock()

		oldm = c.mapping
		// mark all current nodes as false
		for k := range c.masters {
			c.masters[k] = false
		}
		for k := range c.replicas {
			c.replicas[k] = false
		}

		for _, sm := range m {
			for i, node := range sm.nodes {
				if node != "" {
					target := c.masters
					if i > 0 {
						target = c.replicas
					}
					target[node] = true
				}
			}
			for ix := sm.start; ix <= sm.end; ix++ {
				c.mapping[ix] = sm.nodes
			}
		}

		// remove all nodes that are gone from the cluster
		for _, nodes := range []map[string]bool{c.masters, c.replicas} {
			for k, ok := range nodes {
				if !ok {
					delete(nodes, k)

					// close and remove all existing pools for removed nodes
					if p := c.pools[k]; p != nil {
						// Pool.Close always returns nil
						p.Close()
						delete(c.pools, k)
					}
				}
			}
		}

		// mark that no refresh is needed until another MOVED
		c.refreshing = false
		newm = c.mapping
		c.mu.Unlock()

		if c.LayoutRefresh != nil {
			c.LayoutRefresh(oldm, newm)
		}

		return nil
	}

	// reset the refreshing flag
	c.mu.Lock()
	c.refreshing = false
	c.mu.Unlock()

	msg := "redisc: all nodes failed\n"
	msg += strings.Join(errMsgs, "\n")
	err := errors.New(msg)
	if bg && c.BgError != nil {
		// in bg mode, this is already called in a distinct goroutine, so do not
		// call BgError in a distinct one.
		c.BgError(ClusterRefresh, err)
	}
	return err
}

// needsRefresh handles automatic update of the mapping, either because no node
// was found for the slot, or because a MOVED error was received.
func (c *Cluster) needsRefresh(re *RedirError) {
	c.mu.Lock()
	if re != nil {
		// update the mapping only if the address has changed, so that if a
		// READONLY replica read returns a MOVED to a master, it doesn't overwrite
		// that slot's replicas by setting just the master (i.e. this is not a
		// MOVED because the cluster is updating, it is a MOVED because the replica
		// cannot serve that key). Same goes for a request to a random connection
		// that gets a MOVED, should not overwrite the moved-to slot's
		// configuration if the master's address is the same.
		if current := c.mapping[re.NewSlot]; len(current) == 0 || current[0] != re.Addr {
			c.mapping[re.NewSlot] = []string{re.Addr}
		} else {
			// no refresh needed, the mapping already points to this address
			c.mu.Unlock()
			return
		}
	}
	if !c.refreshing {
		// refreshing is reset only once the goroutine has finished updating the
		// mapping, so a new refresh goroutine will only be started if none is
		// running.
		c.refreshing = true
		go c.refresh(true) //nolint:errcheck
	}
	c.mu.Unlock()
}

type slotMapping struct {
	start, end int
	nodes      []string // master is always at [0]
}

func (c *Cluster) getClusterSlots(addr string) ([]slotMapping, error) {
	conn, err := c.getConnForAddr(addr, false)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	vals, err := redis.Values(conn.Do("CLUSTER", "SLOTS"))
	if err != nil {
		return nil, err
	}

	m := make([]slotMapping, 0, len(vals))
	for len(vals) > 0 {
		var slotRange []interface{}
		vals, err = redis.Scan(vals, &slotRange)
		if err != nil {
			return nil, err
		}

		var start, end int
		slotRange, err = redis.Scan(slotRange, &start, &end)
		if err != nil {
			return nil, err
		}

		sm := slotMapping{start: start, end: end}
		// store the master address and all replicas
		for len(slotRange) > 0 {
			var nodes []interface{}
			slotRange, err = redis.Scan(slotRange, &nodes)
			if err != nil {
				return nil, err
			}

			var addr string
			var port int
			if _, err = redis.Scan(nodes, &addr, &port); err != nil {
				return nil, err
			}
			sm.nodes = append(sm.nodes, addr+":"+strconv.Itoa(port))
		}

		m = append(m, sm)
	}

	return m, nil
}

func (c *Cluster) getConnForAddr(addr string, forceDial bool) (redis.Conn, error) {
	c.mu.Lock()

	if err := c.err; err != nil {
		c.mu.Unlock()
		return nil, err
	}
	if c.CreatePool == nil || forceDial {
		c.mu.Unlock()
		return redis.Dial("tcp", addr, c.DialOptions...)
	}

	p := c.pools[addr]
	if p == nil {
		c.mu.Unlock()
		pool, err := c.CreatePool(addr, c.DialOptions...)
		if err != nil {
			return nil, err
		}

		c.mu.Lock()
		// check again, concurrent request may have set the pool in the meantime
		if p = c.pools[addr]; p == nil {
			if c.pools == nil {
				c.pools = make(map[string]*redis.Pool, len(c.StartupNodes))
			}
			c.pools[addr] = pool
			p = pool
		} else {
			// Don't assume CreatePool just returned the pool struct, it may have
			// used a connection or something - always match CreatePool with Close.
			// Do it in a defer to keep lock time short. Pool.Close always returns
			// nil.
			defer pool.Close()
		}
	}
	c.mu.Unlock()

	return c.getFromPool(p)
}

// get connection from the pool.
// use GetContext if PoolWaitTime > 0
func (c *Cluster) getFromPool(p *redis.Pool) (redis.Conn, error) {
	if c.PoolWaitTime <= 0 {
		conn := p.Get()
		return conn, conn.Err()
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.PoolWaitTime)
	defer cancel()

	return p.GetContext(ctx)
}

var errNoNodeForSlot = errors.New("redisc: no node for slot")

func (c *Cluster) getConnForSlot(slot int, forceDial, readOnly bool) (redis.Conn, string, error) {
	c.mu.Lock()
	addrs := c.mapping[slot]
	c.mu.Unlock()
	if len(addrs) == 0 {
		return nil, "", errNoNodeForSlot
	}

	// mapping slices are never altered, they are replaced when refreshing
	// or on a MOVED response, so it's non-racy to read them outside the lock.
	addr := addrs[0]
	if readOnly && len(addrs) > 1 {
		// get the address of a replica
		if len(addrs) == 2 {
			addr = addrs[1]
		} else {
			rnd.Lock()
			ix := rnd.Intn(len(addrs) - 1)
			rnd.Unlock()
			addr = addrs[ix+1] // +1 because 0 is the master
		}
	} else {
		readOnly = false
	}
	conn, err := c.getConnForAddr(addr, forceDial)
	if err == nil && readOnly {
		_, _ = conn.Do("READONLY")
	}
	return conn, addr, err
}

// a *rand.Rand is not safe for concurrent access
var rnd = struct {
	sync.Mutex
	*rand.Rand
}{Rand: rand.New(rand.NewSource(time.Now().UnixNano()))} //nolint:gosec

func (c *Cluster) getRandomConn(forceDial, readOnly bool) (redis.Conn, string, error) {
	addrs, _ := c.getNodeAddrs(readOnly)
	rnd.Lock()
	perms := rnd.Perm(len(addrs))
	rnd.Unlock()

	var errMsgs []string //nolint:prealloc
	for _, ix := range perms {
		addr := addrs[ix]
		conn, err := c.getConnForAddr(addr, forceDial)
		if err == nil {
			if readOnly {
				_, _ = conn.Do("READONLY")
			}
			return conn, addr, nil
		}
		errMsgs = append(errMsgs, err.Error())
	}
	msg := "redisc: failed to get a connection"
	if len(errMsgs) > 0 {
		msg += "\n"
		msg += strings.Join(errMsgs, "\n")
	}
	return nil, "", errors.New(msg)
}

func (c *Cluster) getConn(preferredSlot int, forceDial, readOnly bool) (conn redis.Conn, addr string, err error) {
	if preferredSlot >= 0 {
		conn, addr, err = c.getConnForSlot(preferredSlot, forceDial, readOnly)
		if err == errNoNodeForSlot {
			c.needsRefresh(nil)
		}
	}
	if preferredSlot < 0 || err != nil {
		conn, addr, err = c.getRandomConn(forceDial, readOnly)
	}
	return conn, addr, err
}

func (c *Cluster) getNodeAddrs(preferReplicas bool) (addrs []string, replicas bool) {
	c.mu.Lock()

	// populate nodes lazily, only once
	if c.masters == nil {
		c.masters = make(map[string]bool, len(c.StartupNodes))
		c.replicas = make(map[string]bool)

		// StartupNodes should be masters
		for _, n := range c.StartupNodes {
			c.masters[n] = true
		}
	}

	from := c.masters
	if preferReplicas && len(c.replicas) > 0 {
		from = c.replicas
		replicas = true
	}

	// grab a slice of addresses
	addrs = make([]string, 0, len(from))
	for addr := range from {
		addrs = append(addrs, addr)
	}
	c.mu.Unlock()

	return addrs, replicas
}

// Dial returns a connection the same way as Get, but it guarantees that the
// connection will not be managed by the pool, even if CreatePool is set. The
// actual returned type is *Conn, see its documentation for details.
func (c *Cluster) Dial() (redis.Conn, error) {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()

	if err != nil {
		return nil, err
	}

	return &Conn{
		cluster:   c,
		forceDial: true,
	}, nil
}

// Get returns a redis.Conn interface that can be used to call redis commands
// on the cluster. The application must close the returned connection. The
// actual returned type is *Conn, see its documentation for details.
func (c *Cluster) Get() redis.Conn {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()

	return &Conn{
		cluster: c,
		err:     err,
	}
}

// EachNode calls fn for each node in the cluster, with a connection bound to
// that node. The connection is automatically closed (and potentially returned
// to the pool if Cluster.CreatePool is set) after the function executes. Note
// that conn is not a RetryConn and using one is inappropriate, as the goal of
// EachNode is to connect to specific nodes, not to target specific keys. The
// visited nodes are those that are known at the time of the call - it does not
// force a refresh of the cluster layout. If no nodes are known, it returns an
// error.
//
// If fn returns an error, no more nodes are visited and that error is returned
// by EachNode. If replicas is true, it will visit each replica node instead,
// otherwise the primary nodes are visited. Keep in mind that if replicas is
// true, it will visit all known replicas - which is great e.g. to run
// diagnostics on each node, but can be surprising if the goal is e.g. to
// collect all keys, as it is possible that more than one node is acting as
// replica for the same primary, meaning that the same keys could be seen
// multiple times - you should be prepared to handle this scenario. The
// connection provided to fn is not a ReadOnly connection (conn.ReadOnly hasn't
// been called on it), it is up to fn to execute the READONLY redis command if
// required.
func (c *Cluster) EachNode(replicas bool, fn func(addr string, conn redis.Conn) error) error {
	addrs, ok := c.getNodeAddrs(replicas)
	if len(addrs) == 0 || replicas && !ok {
		return errors.New("redisc: no known node address")
	}

	for _, addr := range addrs {
		conn, err := c.getConnForAddr(addr, false)
		cconn := &Conn{
			cluster:   c,
			boundAddr: addr,
			rc:        conn,
			// in case of error, create a failed connection and still call fn, so
			// that it can decide whether or not to keep visiting nodes.
			err: err,
		}
		err = func() error {
			defer cconn.Close()
			return fn(addr, cconn)
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

// Close releases the resources used by the cluster. It closes all the pools
// that were created, if any.
func (c *Cluster) Close() error {
	c.mu.Lock()
	err := c.err
	if err == nil {
		c.err = errors.New("redisc: closed")
		for _, p := range c.pools {
			if e := p.Close(); e != nil && err == nil {
				// note that Pool.Close always returns nil.
				err = e
			}
		}
		// keep c.pools around so that Stats can still be called after Close
	}
	c.mu.Unlock()

	return err
}

// Stats returns the current statistics for all pools. Keys are node's
// addresses.
func (c *Cluster) Stats() map[string]redis.PoolStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]redis.PoolStats, len(c.pools))
	for address, pool := range c.pools {
		stats[address] = pool.Stats()
	}
	return stats
}
