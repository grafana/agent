package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/grafana/agent/internal/useragent"
)

const gcBaseUrl = "https://grafana.com/api/"

var userAgent = useragent.Get()

type ListStacksResponse struct {
	Items []Stack `json:"items,omitempty"`
}

type Stack struct {
	Slug             string `json:"slug"`
	MetricsUsername  int    `json:"hmInstancePromId"`
	MetricsUrl       string `json:"hmInstancePromUrl"`
	LogsUsername     int    `json:"hlInstanceId"`
	LogsUrl          string `json:"hlInstanceUrl"`
	TracesUsername   int    `json:"htInstanceId"`
	TracesUrl        string `json:"htInstanceUrl"`
	ProfilesUsername int    `json:"hpInstanceId"`
	ProfilesUrl      string `json:"hpInstanceUrl"`
	OtlpUsername     int    `json:"hoInstanceId"`
	OtlpUrl          string `json:"hoInstanceUrl"`
}

func GetListStacks(token string, org string, ctx context.Context) (ListStacksResponse, error) {
	endpoint := gcBaseUrl + "orgs/" + org + "/instances"
	req, err := http.NewRequest("GET", endpoint, bytes.NewBuffer([]byte("")))
	if err != nil {
		return ListStacksResponse{}, err
	}

	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Authorization", "Bearer "+token)
	req = req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ListStacksResponse{}, err
	}

	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil || res.StatusCode != http.StatusOK {
		return ListStacksResponse{}, err
	}

	var lsr ListStacksResponse
	err = json.Unmarshal(resBody, &lsr)
	if err != nil {
		return ListStacksResponse{}, err
	}

	return lsr, nil
}
