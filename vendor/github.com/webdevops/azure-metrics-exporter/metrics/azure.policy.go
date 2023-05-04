package metrics

import (
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

type noCachePolicy struct{}

func (p noCachePolicy) Do(req *policy.Request) (*http.Response, error) {
	// Mutate/process request.
	req.Raw().Header.Set("cache-control", "no-cache")

	// Forward the request to the next policy in the pipeline.
	return req.Next()
}
