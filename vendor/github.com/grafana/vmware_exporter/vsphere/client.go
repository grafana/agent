package vsphere

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

// The highest number of metrics we can query for, no matter what settings
// and server say.
const absoluteMaxMetrics = 10000

type clientFactory struct {
	client     *client
	mux        sync.Mutex
	vSphereURL *url.URL
	cfg        *vSphereConfig
	logger     log.Logger
}

// client represents a connection to vSphere and is backed by a govmomi connection
type client struct {
	Client  *govmomi.Client
	Views   *view.Manager
	Root    *view.ContainerView
	Perf    *performance.Manager
	Valid   bool
	Timeout time.Duration
	logger  log.Logger
}

// newClientFactory creates a new clientFactory and prepares it for use.
func newClientFactory(l log.Logger, vSphereURL *url.URL, cfg *vSphereConfig) *clientFactory {
	return &clientFactory{
		cfg:        cfg,
		vSphereURL: vSphereURL,
		logger:     l,
	}
}

// GetClient returns a client.
func (cf *clientFactory) GetClient(ctx context.Context) (*client, error) {
	cf.mux.Lock()
	defer cf.mux.Unlock()
	retrying := false
	for {
		if cf.client == nil {
			var err error
			if cf.client, err = newClient(ctx, cf.logger, cf.vSphereURL, cf.cfg); err != nil {
				return nil, err
			}
		}

		// Execute a dummy call against the server to make sure the client is
		// still functional. If not, try to log back in. If that doesn't work,
		// we give up.
		ctx1, cancel1 := context.WithTimeout(ctx, cf.cfg.Timeout)
		defer cancel1()
		if _, err := methods.GetCurrentTime(ctx1, cf.client.Client); err != nil {
			//cf.cfg.Log.Info("client session seems to have time out. Reauthenticating!")
			ctx2, cancel2 := context.WithTimeout(ctx, cf.cfg.Timeout)
			defer cancel2()
			if err := cf.client.Client.SessionManager.Login(ctx2, url.UserPassword(cf.cfg.Username, cf.cfg.Password)); err != nil {
				if !retrying {
					// The client went stale. Probably because someone rebooted vCenter. Clear it to
					// force us to create a fresh one. We only get one chance at this. If we fail a second time
					// we will simply skip this collection round and hope things have stabilized for the next one.
					retrying = true
					cf.client = nil
					continue
				}
				return nil, fmt.Errorf("renewing authentication failed: %s", err.Error())
			}
		}

		return cf.client, nil
	}
}

// newClient creates a new vSphere client based on the url and setting passed as parameters.
// TODO: tls config
func newClient(ctx context.Context, l log.Logger, vSphereURL *url.URL, cfg *vSphereConfig) (*client, error) {
	if cfg.Username != "" {
		vSphereURL.User = url.UserPassword(cfg.Username, cfg.Password)
	}

	soapClient := soap.NewClient(vSphereURL, true)

	ctx1, cancel1 := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel1()
	vimClient, err := vim25.NewClient(ctx1, soapClient)
	if err != nil {
		return nil, err
	}
	sm := session.NewManager(vimClient)

	// Create the govmomi client.
	c := &govmomi.Client{
		Client:         vimClient,
		SessionManager: sm,
	}

	// Only login if the URL contains user information.
	if vSphereURL.User != nil {
		if err := c.Login(ctx, vSphereURL.User); err != nil {
			return nil, err
		}
	}

	c.Timeout = cfg.Timeout
	m := view.NewManager(c.Client)

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{}, true)
	if err != nil {
		return nil, err
	}

	p := performance.NewManager(c.Client)

	client := &client{
		Client:  c,
		Views:   m,
		Root:    v,
		Perf:    p,
		Valid:   true,
		Timeout: cfg.Timeout,
		logger:  l,
	}

	// Adjust max query size if needed
	ctx3, cancel3 := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel3()
	n, err := client.getMaxQueryMetrics(ctx3)
	if err != nil {
		return nil, err
	}
	if n < cfg.MaxQueryMetrics {
		cfg.MaxQueryMetrics = n
	}
	return client, nil
}

// counterInfoByKey wraps performance.CounterInfoByKey to give it proper timeouts
func (c *client) counterInfoByKey(ctx context.Context) (map[int32]*types.PerfCounterInfo, error) {
	ctx1, cancel1 := context.WithTimeout(ctx, c.Timeout)
	defer cancel1()
	return c.Perf.CounterInfoByKey(ctx1)
}

// counterInfoByName wraps performance.CounterInfoByName to give it proper timeouts
func (c *client) counterInfoByName(ctx context.Context) (map[string]*types.PerfCounterInfo, error) {
	ctx1, cancel1 := context.WithTimeout(ctx, c.Timeout)
	defer cancel1()
	return c.Perf.CounterInfoByName(ctx1)
}

// getServerTime returns the time at the vCenter server
func (c *client) getServerTime(ctx context.Context) (time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()
	t, err := methods.GetCurrentTime(ctx, c.Client)
	if err != nil {
		return time.Time{}, err
	}
	return *t, nil
}

// getMaxQueryMetrics returns the max_query_metrics setting as configured in vCenter
func (c *client) getMaxQueryMetrics(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	om := object.NewOptionManager(c.Client.Client, *c.Client.Client.ServiceContent.Setting)
	res, err := om.Query(ctx, "config.vpxd.stats.maxQueryMetrics")
	if err == nil {
		if len(res) > 0 {
			if s, ok := res[0].GetOptionValue().Value.(string); ok {
				v, err := strconv.Atoi(s)
				if err == nil {
					level.Debug(c.logger).Log("msg", "vCenter maxQueryMetrics is defined", "maxQueryMetrics", v)
					if v == -1 {
						// Whatever the server says, we never ask for more metrics than this.
						return absoluteMaxMetrics, nil
					}
					return v, nil
				}
			}
			// Fall through version-based inference if value isn't usable
		}
	} else {
		level.Debug(c.logger).Log("msg", "option query for maxMetrics failed. Using default")
	}

	// No usable maxQueryMetrics setting. Infer based on version
	ver := c.Client.Client.ServiceContent.About.Version
	parts := strings.Split(ver, ".")
	if len(parts) < 2 {
		level.Warn(c.logger).Log("msg",
			"vCenter returned an invalid version string. Using default query size=64", "version", ver)
		return 64, nil
	}
	level.Debug(c.logger).Log("vCenter version", ver)
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	if major < 6 || major == 6 && parts[1] == "0" {
		return 64, nil
	}
	return 256, nil
}
