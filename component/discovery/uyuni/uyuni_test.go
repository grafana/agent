package uyuni

import (
	"testing"
)

// type Arguments struct {
// 	Server           string                  `river:"server,attr"`
// 	Username         string                  `river:"username,attr"`
// 	Password         rivertypes.Secret       `river:"password,attr"`
// 	HTTPClientConfig config.HTTPClientConfig `river:",squash"`
// 	Entitlement      string                  `river:"entitlement,attr,optional"`
// 	Separator        string                  `river:"separator,attr,optional"`
// 	RefreshInterval  time.Duration           `river:"refresh_interval,attr,optional"`
// }

func TestUnmarshal(t *testing.T) {
	cfg := `
	server = "https://uyuni.example.com"
	username = "exampleuser"
	password = "examplepassword"
	refresh_interval = "1m"
	http_client_config {
		tls_config {
			ca_file = "/etc/ssl/certs/ca-certificates.crt"
			
	}
	`
}