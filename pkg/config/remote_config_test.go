package config

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/prometheus/common/config"
	"gopkg.in/yaml.v2"
)

func TestRemoteConfigHTTP(t *testing.T) {
	testCfg := `
metrics:
  global:
    scrape_timeout: 33s
`
	wantCfg := &Config{}
	err := LoadBytes([]byte(testCfg), false, wantCfg)
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/agent.yml" {
			_, _ = w.Write([]byte(testCfg))
		}
	}))

	svrWithBasicAuth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		if user != "foo" && pass != "bar" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/agent.yml" {
			_, _ = w.Write([]byte(testCfg))
		}
	}))

	tempDir := t.TempDir()
	err = os.WriteFile(fmt.Sprintf("%s/password-file.txt", tempDir), []byte("bar"), 0644)
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}
	passwdFileCfg := &config.HTTPClientConfig{
		BasicAuth: &config.BasicAuth{
			Username:     "foo",
			PasswordFile: fmt.Sprintf("%s/password-file.txt", tempDir),
		},
	}
	dir, err := os.Getwd()
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}
	passwdFileCfg.SetDirectory(dir)

	type args struct {
		rawURL string
		opts   *RemoteOpts
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "HTTP config",
			args: args{
				rawURL: fmt.Sprintf("%s/agent.yml", svr.URL),
			},
			want:    wantCfg,
			wantErr: false,
		},
		{
			name: "HTTP config with basic auth",
			args: args{
				rawURL: fmt.Sprintf("%s/agent.yml", svrWithBasicAuth.URL),
				opts: &RemoteOpts{
					HTTPClientConfig: &config.HTTPClientConfig{
						BasicAuth: &config.BasicAuth{
							Username: "foo",
							Password: "bar",
						},
					},
				},
			},
			want:    wantCfg,
			wantErr: false,
		},
		{
			name: "HTTP config with basic auth password file",
			args: args{
				rawURL: fmt.Sprintf("%s/agent.yml", svrWithBasicAuth.URL),
				opts: &RemoteOpts{
					HTTPClientConfig: passwdFileCfg,
				},
			},
			want:    wantCfg,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc, err := NewRemoteConfig(tt.args.rawURL, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoteConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			cfg, err := rc.Retrieve()
			if (err != nil) != tt.wantErr {
				t.Errorf("Retrieve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			actual, _ := yaml.Marshal(cfg)
			expected, _ := yaml.Marshal(tt.want)
			if string(actual) != string(expected) {
				t.Errorf("Retrieve() cfg =\n %v\n, want\n %v", string(actual), string(expected))
			}
		})
	}
}
