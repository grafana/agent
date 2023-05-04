package config

import (
	"io/ioutil"
	"net/url"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"

	"os"

	cfg "github.com/infinityworks/go-common/config"
)

// Config struct holds all of the runtime confgiguration for the application
type Config struct {
	*cfg.BaseConfig
	apiUrl        *url.URL
	repositories  []string
	organisations []string
	users         []string
	apiToken      string
	targetURLs    []string
}

// Init populates the Config struct based on environmental runtime configuration
func Init() Config {

	listenPort := cfg.GetEnv("LISTEN_PORT", "9171")
	os.Setenv("LISTEN_PORT", listenPort)
	ac := cfg.Init()

	appConfig := Config{
		&ac,
		nil,
		nil,
		nil,
		nil,
		"",
		nil,
	}

	err := appConfig.SetAPIURL(cfg.GetEnv("API_URL", "https://api.github.com"))
	if err != nil {
		log.Errorf("Error initialising Configuration. Unable to parse API URL. Error: %v", err)
	}
	repos := os.Getenv("REPOS")
	if repos != "" {
		appConfig.SetRepositories(strings.Split(repos, ", "))
	}
	orgs := os.Getenv("ORGS")
	if orgs != "" {
		appConfig.SetOrganisations(strings.Split(repos, ", "))
	}
	users := os.Getenv("USERS")
	if users != "" {
		appConfig.SetUsers(strings.Split(users, ", "))
	}
	tokenEnv := os.Getenv("GITHUB_TOKEN")
	tokenFile := os.Getenv("GITHUB_TOKEN_FILE")
	if tokenEnv != "" {
		appConfig.SetAPIToken(tokenEnv)
	} else if tokenFile != "" {
		err = appConfig.SetAPITokenFromFile(tokenFile)
		if err != nil {
			log.Errorf("Error initialising Configuration, Error: %v", err)
		}
	}

	return appConfig
}

// Returns the base APIURL
func (c *Config) APIURL() *url.URL {
	return c.apiUrl
}

// Returns a list of all object URLs to scrape
func (c *Config) TargetURLs() []string {
	return c.targetURLs
}

// Returns the oauth2 token for usage in http.request
func (c *Config) APIToken() string {
	return c.apiToken
}

// Sets the base API URL returning an error if the supplied string is not a valid URL
func (c *Config) SetAPIURL(u string) error {
	ur, err := url.Parse(u)
	c.apiUrl = ur
	return err
}

// Overrides the entire list of repositories
func (c *Config) SetRepositories(repos []string) {
	c.repositories = repos
	c.setScrapeURLs()
}

// Overrides the entire list of organisations
func (c *Config) SetOrganisations(orgs []string) {
	c.organisations = orgs
	c.setScrapeURLs()
}

// Overrides the entire list of users
func (c *Config) SetUsers(users []string) {
	c.users = users
	c.setScrapeURLs()
}

// SetAPIToken accepts a string oauth2 token for usage in http.request
func (c *Config) SetAPIToken(token string) {
	c.apiToken = token
}

// SetAPITokenFromFile accepts a file containing an oauth2 token for usage in http.request
func (c *Config) SetAPITokenFromFile(tokenFile string) error {
	b, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return err
	}
	c.apiToken = strings.TrimSpace(string(b))
	return nil
}

// Init populates the Config struct based on environmental runtime configuration
// All URL's are added to the TargetURL's string array
func (c *Config) setScrapeURLs() error {

	urls := []string{}

	opts := map[string]string{"per_page": "100"} // Used to set the Github API to return 100 results per page (max)

	if len(c.repositories) == 0 && len(c.organisations) == 0 && len(c.users) == 0 {
		log.Info("No targets specified. Only rate limit endpoint will be scraped")
	}

	// Append repositories to the array
	if len(c.repositories) > 0 {
		for _, x := range c.repositories {
			y := *c.apiUrl
			y.Path = path.Join(y.Path, "repos", x)
			q := y.Query()
			for k, v := range opts {
				q.Add(k, v)
			}
			y.RawQuery = q.Encode()
			urls = append(urls, y.String())
		}
	}

	// Append github orginisations to the array
	if len(c.organisations) > 0 {
		for _, x := range c.organisations {
			y := *c.apiUrl
			y.Path = path.Join(y.Path, "orgs", x, "repos")
			q := y.Query()
			for k, v := range opts {
				q.Add(k, v)
			}
			y.RawQuery = q.Encode()
			urls = append(urls, y.String())
		}
	}

	if len(c.users) > 0 {
		for _, x := range c.users {
			y := *c.apiUrl
			y.Path = path.Join(y.Path, "users", x, "repos")
			q := y.Query()
			for k, v := range opts {
				q.Add(k, v)
			}
			y.RawQuery = q.Encode()
			urls = append(urls, y.String())
		}
	}

	c.targetURLs = urls

	return nil
}
