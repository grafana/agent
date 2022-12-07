package client

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	log "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/dskit/crypto/tls"
	"github.com/pkg/errors"
)

const (
	rulerAPIPath  = "/prometheus/config/v1/rules"
	legacyAPIPath = "/api/v1/rules"
)

var (
	ErrNoConfig         = errors.New("No config exists for this user")
	ErrResourceNotFound = errors.New("requested resource not found")
)

// Config is used to configure a MimirClient.
type Config struct {
	User            string `yaml:"user"`
	Key             string `yaml:"key"`
	Address         string `yaml:"address"`
	ID              string `yaml:"id"`
	TLS             tls.ClientConfig
	UseLegacyRoutes bool   `yaml:"use_legacy_routes"`
	AuthToken       string `yaml:"auth_token"`
}

type Interface interface {
	CreateRuleGroup(ctx context.Context, namespace string, rg RuleGroup) error
	DeleteRuleGroup(ctx context.Context, namespace, groupName string) error
	ListRules(ctx context.Context, namespace string) (map[string][]RuleGroup, error)
}

// MimirClient is a client to the Mimir API.
type MimirClient struct {
	user      string
	key       string
	id        string
	endpoint  *url.URL
	Client    http.Client
	apiPath   string
	authToken string
	logger    log.Logger
}

// New returns a new MimirClient.
func New(logger log.Logger, cfg Config) (*MimirClient, error) {
	endpoint, err := url.Parse(cfg.Address)
	if err != nil {
		return nil, err
	}

	level.Debug(logger).Log("msg", "New Mimir client created", "address", cfg.Address, "id", cfg.ID)

	client := http.Client{}

	// Setup TLS client
	tlsConfig, err := cfg.TLS.GetTLSConfig()
	if err != nil {
		level.Error(logger).Log(
			"msg", "error loading TLS files",
			"tls-ca", cfg.TLS.CAPath,
			"tls-cert", cfg.TLS.CertPath,
			"tls-key", cfg.TLS.KeyPath,
			"err", err,
		)
		return nil, fmt.Errorf("Mimir client initialization unsuccessful")
	}

	if tlsConfig != nil {
		transport := &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: tlsConfig,
		}
		client = http.Client{Transport: transport}
	}

	path := rulerAPIPath
	if cfg.UseLegacyRoutes {
		path = legacyAPIPath
	}

	return &MimirClient{
		user:      cfg.User,
		key:       cfg.Key,
		id:        cfg.ID,
		endpoint:  endpoint,
		Client:    client,
		apiPath:   path,
		authToken: cfg.AuthToken,
		logger:    logger,
	}, nil
}

func (r *MimirClient) doRequest(path, method string, payload []byte) (*http.Response, error) {
	req, err := buildRequest(path, method, *r.endpoint, payload)
	if err != nil {
		return nil, err
	}

	if (r.user != "" || r.key != "") && r.authToken != "" {
		err := errors.New("atmost one of basic auth or auth token should be configured")
		level.Error(r.logger).Log(
			"msg", "error during setting up request to mimir api",
			"url", req.URL.String(),
			"method", req.Method,
			"error", err,
		)
		return nil, err
	}

	if r.user != "" {
		req.SetBasicAuth(r.user, r.key)
	} else if r.key != "" {
		req.SetBasicAuth(r.id, r.key)
	}

	if r.authToken != "" {
		req.Header.Add("Authorization", "Bearer "+r.authToken)
	}

	req.Header.Add("X-Scope-OrgID", r.id)

	level.Debug(r.logger).Log(
		"msg", "sending request to Grafana Mimir API",
		"url", req.URL.String(),
		"method", req.Method,
	)

	resp, err := r.Client.Do(req)
	if err != nil {
		level.Error(r.logger).Log(
			"msg", "error during request to Grafana Mimir API",
			"url", req.URL.String(),
			"method", req.Method,
			"error", err,
		)
		return nil, err
	}

	if err := checkResponse(r.logger, resp); err != nil {
		_ = resp.Body.Close()
		return nil, errors.Wrapf(err, "%s request to %s failed", req.Method, req.URL.String())
	}

	return resp, nil
}

// checkResponse checks an API response for errors.
func checkResponse(logger log.Logger, r *http.Response) error {
	level.Debug(logger).Log("msg", "checking response", "status", r.Status)
	if 200 <= r.StatusCode && r.StatusCode <= 299 {
		return nil
	}

	var msg, errMsg string
	scanner := bufio.NewScanner(io.LimitReader(r.Body, 512))
	if scanner.Scan() {
		msg = scanner.Text()
	}

	if msg == "" {
		errMsg = fmt.Sprintf("server returned HTTP status %s", r.Status)
	} else {
		errMsg = fmt.Sprintf("server returned HTTP status %s: %s", r.Status, msg)
	}

	if r.StatusCode == http.StatusNotFound {
		level.Debug(logger).Log("msg", msg, "status", r.Status)
		return ErrResourceNotFound
	}

	level.Error(logger).Log("msg", msg, "status", r.Status)

	return errors.New(errMsg)
}

func joinPath(baseURLPath, targetPath string) string {
	// trim exactly one slash at the end of the base URL, this expects target
	// path to always start with a slash
	return strings.TrimSuffix(baseURLPath, "/") + targetPath
}

func buildRequest(p, m string, endpoint url.URL, payload []byte) (*http.Request, error) {
	// parse path parameter again (as it already contains escaped path information
	pURL, err := url.Parse(p)
	if err != nil {
		return nil, err
	}

	// if path or endpoint contains escaping that requires RawPath to be populated, also join rawPath
	if pURL.RawPath != "" || endpoint.RawPath != "" {
		endpoint.RawPath = joinPath(endpoint.EscapedPath(), pURL.EscapedPath())
	}
	endpoint.Path = joinPath(endpoint.Path, pURL.Path)
	return http.NewRequest(m, endpoint.String(), bytes.NewBuffer(payload))
}
