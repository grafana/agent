package models

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/utils"
)

// SDK holds metadata about the app agent that produced the event
type SDK struct {
	Name         string           `json:"name,omitempty"`
	Version      string           `json:"version,omitempty"`
	Integrations []SDKIntegration `json:"integrations,omitempty"`
}

// KeyVal produces key->value representation of Sdk metadata
func (sdk SDK) KeyVal() *utils.KeyVal {
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

// SDKIntegration holds metadata about a plugin/integration on the app agent that collected and sent the event
type SDKIntegration struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

func (i SDKIntegration) String() string {
	return fmt.Sprintf("%s:%s", i.Name, i.Version)
}

// User holds metadata about the user related to an app event
type User struct {
	Email      string            `json:"email,omitempty"`
	ID         string            `json:"id,omitempty"`
	Username   string            `json:"username,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// KeyVal produces a key->value representation User metadata
func (u User) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()
	utils.KeyValAdd(kv, "email", u.Email)
	utils.KeyValAdd(kv, "id", u.ID)
	utils.KeyValAdd(kv, "username", u.ID)
	utils.MergeKeyValWithPrefix(kv, utils.KeyValFromMap(u.Attributes), "attr_")
	return kv
}

// Meta holds metadata about an app event
type Meta struct {
	SDK     SDK     `json:"sdk,omitempty"`
	App     App     `json:"app,omitempty"`
	User    User    `json:"user,omitempty"`
	Session Session `json:"session,omitempty"`
	Page    Page    `json:"page,omitempty"`
	Browser Browser `json:"browser,omitempty"`
}

// KeyVal produces key->value representation of the app event metadatga
func (m Meta) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()
	utils.MergeKeyValWithPrefix(kv, m.SDK.KeyVal(), "sdk_")
	utils.MergeKeyValWithPrefix(kv, m.App.KeyVal(), "app_")
	utils.MergeKeyValWithPrefix(kv, m.User.KeyVal(), "user_")
	utils.MergeKeyValWithPrefix(kv, m.Session.KeyVal(), "session_")
	utils.MergeKeyValWithPrefix(kv, m.Page.KeyVal(), "page_")
	utils.MergeKeyValWithPrefix(kv, m.Browser.KeyVal(), "browser_")
	return kv
}

// Session holds metadata about the browser session the event originates from
type Session struct {
	ID         string            `json:"id,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// KeyVal produces key->value representation of the Session metadata
func (s Session) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()
	utils.KeyValAdd(kv, "id", s.ID)
	utils.MergeKeyValWithPrefix(kv, utils.KeyValFromMap(s.Attributes), "attr_")
	return kv
}

// Page holds metadata about the web page event originates from
type Page struct {
	ID         string            `json:"id,omitempty"`
	URL        string            `json:"url,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// KeyVal produces key->val representation of Page metadata
func (p Page) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()
	utils.KeyValAdd(kv, "id", p.ID)
	utils.KeyValAdd(kv, "url", p.URL)
	utils.MergeKeyValWithPrefix(kv, utils.KeyValFromMap(p.Attributes), "attr_")
	return kv
}

// App holds metadata about the application event originates from
type App struct {
	Name        string `json:"name,omitempty"`
	Release     string `json:"release,omitempty"`
	Version     string `json:"version,omitempty"`
	Environment string `json:"environment,omitempty"`
}

// KeyVal produces key-> value representation of App metadata
func (a App) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()
	utils.KeyValAdd(kv, "name", a.Name)
	utils.KeyValAdd(kv, "release", a.Release)
	utils.KeyValAdd(kv, "version", a.Version)
	utils.KeyValAdd(kv, "environment", a.Environment)
	return kv
}

// Browser holds metadata about a client's browser
type Browser struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	OS      string `json:"os,omitempty"`
	Mobile  bool   `json:"mobile,omitempty"`
}

// KeyVal produces key->value representation of the Browser metadata
func (b Browser) KeyVal() *utils.KeyVal {
	kv := utils.NewKeyVal()
	utils.KeyValAdd(kv, "name", b.Name)
	utils.KeyValAdd(kv, "version", b.Version)
	utils.KeyValAdd(kv, "os", b.OS)
	utils.KeyValAdd(kv, "mobile", fmt.Sprintf("%v", b.Mobile))
	return kv
}
