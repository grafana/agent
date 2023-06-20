package auth

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/grafana/agent/pkg/river/rivertypes"
)

type GitAuthConfig struct {
	BasicAuth *BasicAuth `river:"basic_auth,block,optional"`
	SSHKey    *SSHKey    `river:"ssh_key,block,optional"`
}

// Convert converts HTTPClientConfig to the native Prometheus type. If h is
// nil, the default client config is returned.
func (h *GitAuthConfig) Convert() transport.AuthMethod {
	if h == nil {
		return nil
	}

	if h.BasicAuth != nil {
		return h.BasicAuth.Convert()
	}

	if h.SSHKey != nil {
		return h.SSHKey.Convert()
	}
	return nil
}

type BasicAuth struct {
	Username string            `river:"username,attr,optional"`
	Password rivertypes.Secret `river:"password,attr,optional"`
}

// Convert converts our type to the native prometheus type
func (b *BasicAuth) Convert() (t transport.AuthMethod) {
	if b == nil {
		return nil
	}
	return &http.BasicAuth{
		Username: b.Username,
		Password: string(b.Password),
	}
}

type SSHKey struct {
	Username   string            `river:"username,attr,optional"`
	Keyfile    string            `river:"keyfile,attr,optional"`
	Passphrase rivertypes.Secret `river:"passphrase,attr,optional"`
}

// Convert converts our type to the native prometheus type
func (s *SSHKey) Convert() (t transport.AuthMethod) {
	if s == nil {
		return nil
	}
	publickeys, err := ssh.NewPublicKeysFromFile(s.Username, s.Keyfile, string(s.Passphrase))
	if err != nil {
		fmt.Sprintf("generate publickeys failed: %s\n", err.Error())
		return
	}
	return publickeys
}
