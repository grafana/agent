package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const query = "http://localhost:3100/loki/api/v1/series"
const filename = "logs.txt"

type data struct {
	Filename string `json:"filename"`
}

type jsonResponse struct {
	Status string `json:"status"`
	Data   []data `json:"data"`
}

func TestAgentIntegration(t *testing.T) {
	const maxRetries = 20
	const retryInterval = time.Second * 5

	for i := 0; i < maxRetries; i++ {
		fmt.Println("retry", i)
		resp, err := http.Get(query)
		if err != nil {
			t.Fatalf("Failed to get logs from Loki: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status OK but got %v", resp.Status)
		}

		var jsonResponse jsonResponse
		if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
			t.Fatalf("Failed to decode JSON: %v", err)
		}

		if len(jsonResponse.Data) > 0 {
			if len(jsonResponse.Data) > 1 {
				t.Fatalf("Retrieved more files than expected(%d): %s", len(jsonResponse.Data), jsonResponse.Data)
			}
			if jsonResponse.Data[0].Filename == filename {
				return
			}
			t.Fatalf("Wrong file retrieved, expected %s got %s", filename, jsonResponse.Data[0].Filename)
		}

		time.Sleep(retryInterval)
	}

	t.Fatal("The result array is empty after all retries")
}
