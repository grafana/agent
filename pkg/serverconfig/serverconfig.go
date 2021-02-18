package serverconfig

import "github.com/weaveworks/common/server"

type ServerConfig struct {
	server.Config `yaml:",inline"`
	ClientCert    string `yaml:"client_cert_file"`
	ClientKey     string `yaml:"client_key_file"`
}
