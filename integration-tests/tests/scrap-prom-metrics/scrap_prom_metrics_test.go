package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const query = "http://localhost:9009/prometheus/api/v1/query?query=avalanche_metric_mmmmm_0_0"

type jsonResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func TestAgentIntegration(t *testing.T) {
	const maxRetries = 20
	const retryInterval = time.Second * 5

	for i := 0; i < maxRetries; i++ {
		fmt.Println("retry", i)
		resp, err := http.Get(query)
		if err != nil {
			t.Fatalf("Failed to get metrics from Mimir: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status OK but got %v", resp.Status)
		}

		var jsonResponse jsonResponse
		if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
			t.Fatalf("Failed to decode JSON: %v", err)
		}

		if len(jsonResponse.Data.Result) > 0 {
			return
		}

		time.Sleep(retryInterval)
	}

	t.Fatal("The result array is empty after all retries")
}
