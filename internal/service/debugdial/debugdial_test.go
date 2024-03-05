package debugdial

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	opamp "github.com/open-telemetry/opamp-go/protobufs"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"
)

func TestOpAmpResposne(t *testing.T) {
	// create service
	service := New(prometheus.DefaultRegisterer)
	service.Config.Store("Test", "Test123")

	// get router
	baseURL, router := service.ServiceHandler(nil)

	// create http test server
	server := httptest.NewServer(router)
	defer server.Close()

	clientPayload := &opamp.AgentToServer{
		AgentDescription: &opamp.AgentDescription{
			IdentifyingAttributes: []*opamp.KeyValue{
				&opamp.KeyValue{
					Key:   "service.name",
					Value: &opamp.AnyValue{Value: &opamp.AnyValue_StringValue{StringValue: "Test"}},
				},
			},
		},
	}

	bodyBytes, _ := proto.Marshal(clientPayload)
	req, _ := http.NewRequest("POST", server.URL+baseURL, bytes.NewBuffer(bodyBytes))
	req.Header.Set("content-type", "application/x-protobuf")
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatal("Unable to get response", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Error("Non 200 http status code on opamp response, instead got", resp.StatusCode)
	}

	bodyBytes, _ = io.ReadAll(resp.Body)

	response := opamp.ServerToAgent{}
	_ = proto.Unmarshal(bodyBytes, &response)

	if response.RemoteConfig == nil || response.RemoteConfig.Config == nil {
		t.Fatal("No remote config set in opamp response, instead got", response.String())
	}

	config, exists := response.RemoteConfig.Config.ConfigMap["Test"]
	if !exists {
		t.Fatal("No config for test service \"Test\" found")
	}

	if string(config.Body) != "Test123" {
		t.Error("Expected config returned by opamp to be \"Test123\", but got", string(config.Body))
	}
}
