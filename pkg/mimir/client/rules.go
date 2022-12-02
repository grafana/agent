package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/model/rulefmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// RuleGroup is a list of sequentially evaluated recording and alerting rules.
type RuleGroup struct {
	rulefmt.RuleGroup `yaml:",inline"`
	// RWConfigs is used by the remote write forwarding ruler
	RWConfigs []RemoteWriteConfig `yaml:"remote_write,omitempty"`
}

// RemoteWriteConfig is used to specify a remote write endpoint
type RemoteWriteConfig struct {
	URL string `json:"url,omitempty"`
}

// CreateRuleGroup creates a new rule group
func (r *CortexClient) CreateRuleGroup(ctx context.Context, namespace string, rg RuleGroup) error {
	payload, err := yaml.Marshal(&rg)
	if err != nil {
		return err
	}

	escapedNamespace := url.PathEscape(namespace)
	path := r.apiPath + "/" + escapedNamespace

	res, err := r.doRequest(path, "POST", payload)
	if err != nil {
		return err
	}

	res.Body.Close()

	return nil
}

// DeleteRuleGroup creates a new rule group
func (r *CortexClient) DeleteRuleGroup(ctx context.Context, namespace, groupName string) error {
	escapedNamespace := url.PathEscape(namespace)
	escapedGroupName := url.PathEscape(groupName)
	path := r.apiPath + "/" + escapedNamespace + "/" + escapedGroupName

	res, err := r.doRequest(path, "DELETE", nil)
	if err != nil {
		return err
	}

	res.Body.Close()

	return nil
}

// GetRuleGroup retrieves a rule group
func (r *CortexClient) GetRuleGroup(ctx context.Context, namespace, groupName string) (*RuleGroup, error) {
	escapedNamespace := url.PathEscape(namespace)
	escapedGroupName := url.PathEscape(groupName)
	path := r.apiPath + "/" + escapedNamespace + "/" + escapedGroupName

	fmt.Println(path)
	res, err := r.doRequest(path, "GET", nil)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	rg := RuleGroup{}
	err = yaml.Unmarshal(body, &rg)
	if err != nil {
		log.WithFields(log.Fields{
			"body": string(body),
		}).Debugln("failed to unmarshal rule group from response")

		return nil, errors.Wrap(err, "unable to unmarshal response")
	}

	return &rg, nil
}

// ListRules retrieves a rule group
func (r *CortexClient) ListRules(ctx context.Context, namespace string) (map[string][]RuleGroup, error) {
	path := r.apiPath
	if namespace != "" {
		path = path + "/" + namespace
	}

	res, err := r.doRequest(path, "GET", nil)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	ruleSet := map[string][]RuleGroup{}
	err = yaml.Unmarshal(body, &ruleSet)
	if err != nil {
		return nil, err
	}

	return ruleSet, nil
}
