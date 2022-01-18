package models

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/utils"
)

type Sdk struct {
	Name         string           `json:"name,omitempty"`
	Version      string           `json:"version,omitempty"`
	Integrations []SdkIntegration `json:"integrations,omitempty"`
}

func (sdk Sdk) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()
	utils.KeyValAdd(kv, "name", sdk.Name)
	utils.KeyValAdd(kv, "version", sdk.Version)

	if len(sdk.Integrations) > 0 {
		integrations := make([]string, len(sdk.Integrations))

		for i, integration := range sdk.Integrations {
			integrations[i] = integration.String()
		}

		utils.KeyValAdd(kv, "integrations", strings.Join(integrations, ","))
	}

	return kv
}

type SdkIntegration struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

func (i SdkIntegration) String() string {
	return fmt.Sprintf("%s:%s", i.Name, i.Version)
}

type User struct {
	Email      string            `json:"email,omitempty"`
	ID         string            `json:"id,omitempty"`
	Username   string            `json:"username,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

func (u User) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()
	utils.KeyValAdd(kv, "email", u.Email)
	utils.KeyValAdd(kv, "id", u.ID)
	utils.KeyValAdd(kv, "username", u.ID)
	utils.MergeKeyValWithPrefix(kv, utils.KeyValFromMap(u.Attributes), "attr_")
	return kv
}

type Meta struct {
	Sdk     Sdk     `json:"sdk,omitempty"`
	App     App     `json:"app,omitempty"`
	User    User    `json:"user,omitempty"`
	Session Session `json:"session,omitempty"`
	Page    Page    `json:"page,omitempty"`
	Browser Browser `json:"browser,omitempty"`
}

func (m Meta) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()
	utils.MergeKeyValWithPrefix(kv, m.Sdk.KeyVal(), "sdk_")
	utils.MergeKeyValWithPrefix(kv, m.App.KeyVal(), "app_")
	utils.MergeKeyValWithPrefix(kv, m.User.KeyVal(), "user_")
	utils.MergeKeyValWithPrefix(kv, m.Session.KeyVal(), "session_")
	utils.MergeKeyValWithPrefix(kv, m.Page.KeyVal(), "page_")
	utils.MergeKeyValWithPrefix(kv, m.Browser.KeyVal(), "browser_")
	return kv
}

type Session struct {
	ID         string            `json:"id,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

func (s Session) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()
	utils.KeyValAdd(kv, "id", s.ID)
	utils.MergeKeyValWithPrefix(kv, utils.KeyValFromMap(s.Attributes), "attr_")
	return kv
}

type Page struct {
	ID         string            `json:"id,omitempty"`
	URL        string            `json:"url,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

func (p Page) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()
	utils.KeyValAdd(kv, "id", p.ID)
	utils.KeyValAdd(kv, "url", p.URL)
	utils.MergeKeyValWithPrefix(kv, utils.KeyValFromMap(p.Attributes), "attr_")
	return kv
}

type App struct {
	Name        string `json:"name,omitempty"`
	Release     string `json:"release,omitempty"`
	Version     string `json:"version,omitempty"`
	Environment string `json:"environment,omitempty"`
}

func (a App) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()
	utils.KeyValAdd(kv, "name", a.Name)
	utils.KeyValAdd(kv, "release", a.Release)
	utils.KeyValAdd(kv, "version", a.Version)
	utils.KeyValAdd(kv, "environment", a.Environment)
	return kv
}

type Browser struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	OS      string `json:"os,omitempty"`
	Mobile  bool   `json:"mobile,omitempty"`
}

func (b Browser) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()
	utils.KeyValAdd(kv, "name", b.Name)
	utils.KeyValAdd(kv, "version", b.Version)
	utils.KeyValAdd(kv, "os", b.OS)
	utils.KeyValAdd(kv, "mobile", fmt.Sprintf("%v", b.Mobile))
	return kv
}
