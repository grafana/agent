package azidentity

import (
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

func NewAzCredential(clientOptions *azcore.ClientOptions) (azcore.TokenCredential, error) {
	// azure authorizer
	switch strings.ToLower(os.Getenv("AZURE_AUTH")) {
	case "az", "cli", "azcli":
		// azurecli authentication
		opts := azidentity.AzureCLICredentialOptions{}
		return azidentity.NewAzureCLICredential(&opts)
	case "wi", "workload", "workloadidentity", "federation":
		file := os.Getenv("AZURE_FEDERATED_TOKEN_FILE")
		tenantID := os.Getenv("AZURE_TENANT_ID")
		clientID := os.Getenv("AZURE_CLIENT_ID")

		opts := &azidentity.ClientAssertionCredentialOptions{ClientOptions: *clientOptions}

		w := &workloadIdentityCredential{file: file}
		cred, err := azidentity.NewClientAssertionCredential(tenantID, clientID, w.getAssertion, opts)
		if err != nil {
			return nil, err
		}
		w.cred = cred
		return w, nil
	default:
		// general azure authentication (env vars, service principal, msi, ...)
		opts := azidentity.DefaultAzureCredentialOptions{}
		if clientOptions != nil {
			opts.ClientOptions = *clientOptions
		}

		return azidentity.NewDefaultAzureCredential(&opts)
	}
}
