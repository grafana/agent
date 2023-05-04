package exporter

import (
	"fmt"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tomnomnom/linkheader"
)

// RateLimitExceededStatus is the status response from github when the rate limit is exceeded.
const RateLimitExceededStatus = "403 rate limit exceeded"

func asyncHTTPGets(targets []string, token string) ([]*Response, error) {
	// Expand targets by following GitHub pagination links
	targets = paginateTargets(targets, token)

	// Channels used to enable concurrent requests
	ch := make(chan *Response, len(targets))

	responses := []*Response{}

	for _, url := range targets {

		go func(url string) {
			err := getResponse(url, token, ch)
			if err != nil {
				ch <- &Response{url, nil, []byte{}, err}
			}
		}(url)

	}

	for {
		select {
		case r := <-ch:
			if r.err != nil {
				log.Errorf("Error scraping API, Error: %v", r.err)
				return nil, r.err
			}
			responses = append(responses, r)

			if len(responses) == len(targets) {
				return responses, nil
			}
		}

	}
}

// paginateTargets returns all pages for the provided targets
func paginateTargets(targets []string, token string) []string {

	paginated := targets

	for _, url := range targets {

		// make a request to the original target to get link header if it exists
		resp, err := getHTTPResponse(url, token)
		if err != nil {
			log.Errorf("Error retrieving Link headers, Error: %s", err)
			continue
		}

		if resp.Header["Link"] != nil {
			links := linkheader.Parse(resp.Header["Link"][0])

			for _, link := range links {
				if link.Rel == "last" {

					u, err := neturl.Parse(link.URL)
					if err != nil {
						log.Errorf("Unable to parse page URL, Error: %s", err)
					}

					q := u.Query()

					lastPage, err := strconv.Atoi(q.Get("page"))
					if err != nil {
						log.Errorf("Unable to convert page substring to int, Error: %s", err)
					}

					// add all pages to the slice of targets to return
					for page := 2; page <= lastPage; page++ {
						q.Set("page", strconv.Itoa(page))
						u.RawQuery = q.Encode()
						paginated = append(paginated, u.String())
					}

					break
				}
			}
		}
	}
	return paginated
}

// getResponse collects an individual http.response and returns a *Response
func getResponse(url string, token string, ch chan<- *Response) error {

	log.Infof("Fetching %s \n", url)

	resp, err := getHTTPResponse(url, token) // do this earlier
	if err != nil {
		return fmt.Errorf("Error fetching http response: %v", err)
	}
	defer resp.Body.Close()

	// Read the body to a byte array so it can be used elsewhere
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error converting body to byte array: %v", err)
	}

	// Triggers if a user specifies an invalid or not visible repository
	if resp.StatusCode == 404 {
		return fmt.Errorf("Error: Received 404 status from Github API, ensure the repsository URL is correct. If it's a privare repository, also check the oauth token is correct")
	}

	ch <- &Response{url, resp, body, err}

	return nil
}

// getHTTPResponse handles the http client creation, token setting and returns the *http.response
func getHTTPResponse(url string, token string) (*http.Response, error) {

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	// If a token is present, add it to the http.request
	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	// check rate limit exceeded.
	if resp.Status == RateLimitExceededStatus {
		resp.Body.Close()
		return nil, fmt.Errorf("%s", resp.Status)
	}

	return resp, err
}
