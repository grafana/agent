package auth

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/grafana/agent/pkg/agentflow/config"
	"github.com/grafana/agent/pkg/agentflow/types/actorstate"
	"github.com/grafana/agent/pkg/agentflow/types/exchange"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type CredentialsManager struct {
	cfg  *config.Credentials
	self *actor.PID
	name string
	out  []*actor.PID
}

func (m *CredentialsManager) AllowableInputs() []actorstate.InOutType {
	return []actorstate.InOutType{actorstate.None}
}

func (m *CredentialsManager) Output() actorstate.InOutType {
	return actorstate.Credentials
}

func NewCredentialsManager(name string, cfg *config.CredentialsManager) (actorstate.FlowActor, error) {
	fileCfg := &config.Credentials{}
	if cfg.File != "" {
		out, err := ioutil.ReadFile(cfg.File)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(out, fileCfg)
		if err != nil {
			return nil, err
		}
	}
	// Union the credentials
	if cfg.Credentials != nil {
		if cfg.Credentials.Github != nil {
			fileCfg.Github = append(fileCfg.Github, cfg.Credentials.Github...)
		}
		if cfg.Credentials.BasicAuth != nil {
			fileCfg.BasicAuth = append(fileCfg.BasicAuth, cfg.Credentials.BasicAuth...)
		}
		if cfg.Credentials.MySQL != nil {
			fileCfg.MySQL = append(fileCfg.MySQL, cfg.Credentials.MySQL...)
		}
		if cfg.Credentials.Redis != nil {
			fileCfg.Redis = append(fileCfg.Redis, cfg.Credentials.Redis...)
		}
	}
	// TODO ensure the names are unique
	return &CredentialsManager{
		cfg:  fileCfg,
		name: name,
	}, nil
}

func (m *CredentialsManager) Receive(c actor.Context) {
	switch msg := c.Message().(type) {
	case actorstate.Init:
		m.self = c.Self()
		m.out = msg.Children
	case actorstate.Start:
		for _, o := range m.out {
			ex := convert(m.cfg)
			c.Send(o, ex)
		}
	}
}

func (m *CredentialsManager) Name() string {
	return m.name
}

func (m *CredentialsManager) PID() *actor.PID {
	return m.self
}

func convert(in *config.Credentials) exchange.Credentials {
	ex := exchange.Credentials{
		BasicAuth: make([]exchange.BasicAuthCredential, 0),
		Redis:     make([]exchange.RedisCredential, 0),
		Github:    make([]exchange.GithubCredential, 0),
		MySQL:     make([]exchange.MySQLCredential, 0),
	}
	if in.BasicAuth != nil {
		for _, a := range in.BasicAuth {
			ex.BasicAuth = append(ex.BasicAuth, exchange.BasicAuthCredential{
				Name:     a.Name,
				URL:      a.URL,
				Username: a.Username,
				Password: a.Password,
			})
		}
	}
	if in.Redis != nil {
		for _, a := range in.Redis {
			ex.Redis = append(ex.Redis, exchange.RedisCredential{
				Name: a.Name,
				Auth: a.Auth,
			})
		}
	}
	if in.Github != nil {
		for _, a := range in.Github {
			ex.Github = append(ex.Github, exchange.GithubCredential{
				Name:     a.Name,
				APIToken: a.APIToken,
			})
		}
	}
	if in.MySQL != nil {
		for _, a := range in.MySQL {
			ex.MySQL = append(ex.MySQL, exchange.MySQLCredential{
				Name:     a.Name,
				Password: a.Password,
				Username: a.Username,
			})
		}
	}
	return ex
}
