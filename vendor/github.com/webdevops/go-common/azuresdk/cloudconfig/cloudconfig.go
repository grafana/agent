package cloudconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
)

type (
	CloudEnvironment struct {
		cloud.Configuration
		Name CloudName
	}
)

// NewCloudConfig creates a new cloud configuration object based on cloud name (eg. AzurePublicCloud)
func NewCloudConfig(cloudName string) (config CloudEnvironment, err error) {
	switch strings.ToLower(cloudName) {
	// ----------------------------------------------------
	// Azure Public cloud (default)
	case "azurepublic", "azurepubliccloud", "azurecloud":
		config, err = CloudEnvironment{
			Name:          AzurePublicCloud,
			Configuration: cloud.AzurePublic,
		}, nil
		injectServiceConfig(&config.Configuration, ServiceNameMicrosoftGraph, cloud.ServiceConfiguration{
			Audience: "https://graph.microsoft.com/",
			Endpoint: "https://graph.microsoft.com",
		})
		injectServiceConfig(&config.Configuration, ServiceNameLogAnalyticsWorkspace, cloud.ServiceConfiguration{
			Audience: "https://api.loganalytics.io/",
			Endpoint: "https://api.loganalytics.io",
		})

	// ----------------------------------------------------
	// Azure China cloud
	case "azurechina", "azurechinacloud":
		config, err = CloudEnvironment{
			Name:          AzureChinaCloud,
			Configuration: cloud.AzureChina,
		}, nil
		injectServiceConfig(&config.Configuration, ServiceNameMicrosoftGraph, cloud.ServiceConfiguration{
			Audience: "https://microsoftgraph.chinacloudapi.cn/",
			Endpoint: "https://microsoftgraph.chinacloudapi.cn",
		})
		injectServiceConfig(&config.Configuration, ServiceNameLogAnalyticsWorkspace, cloud.ServiceConfiguration{
			Audience: "https://api.loganalytics.azure.cn/",
			Endpoint: "https://api.loganalytics.azure.cn",
		})

	// ----------------------------------------------------
	// Azure Government cloud
	case "azuregovernment", "azuregovernmentcloud", "azureusgovernmentcloud":
		config, err = CloudEnvironment{
			Name:          AzureGovernmentCloud,
			Configuration: cloud.AzureGovernment,
		}, nil
		injectServiceConfig(&config.Configuration, ServiceNameMicrosoftGraph, cloud.ServiceConfiguration{
			Audience: "https://login.microsoftonline.us/",
			Endpoint: "https://login.microsoftonline.us",
		})
		injectServiceConfig(&config.Configuration, ServiceNameLogAnalyticsWorkspace, cloud.ServiceConfiguration{
			Audience: "https://api.loganalytics.us/",
			Endpoint: "https://api.loganalytics.us",
		})

	// ----------------------------------------------------
	// Azure Private Cloud (onpremise, custom configuration via json)
	case "azureprivate", "azurepprivatecloud":
		config, err = CloudEnvironment{
			Name: AzurePrivateCloud,
		}, nil

		if cloudConfig, privateCloudConfigErr := createAzurePrivateCloudConfig(); privateCloudConfigErr == nil {
			config.Configuration = cloudConfig
		} else {
			err = privateCloudConfigErr
		}

	default:
		err = fmt.Errorf(`unable to set Azure Cloud "%v", not valid`, cloudName)
	}

	return
}

// injectServiceConfig injects a serviceconfiguration into cloud config
func injectServiceConfig(config *cloud.Configuration, serviceName cloud.ServiceName, serviceConfig cloud.ServiceConfiguration) {
	if config.Services == nil {
		config.Services = map[cloud.ServiceName]cloud.ServiceConfiguration{}
	}

	config.Services[serviceName] = serviceConfig
}

// createAzurePrivateCloudConfig creates azureprivate (onpremise) cloudconfig from either AZURE_CLOUD_CONFIG (string) or AZURE_CLOUD_CONFIG_FILE (file)
func createAzurePrivateCloudConfig() (cloud.Configuration, error) {
	var cloudConfigJson []byte
	cloudConfig := cloud.Configuration{}

	if val := os.Getenv("AZURE_CLOUD_CONFIG"); len(val) > 0 {
		// cloud config via JSON string
		cloudConfigJson = []byte(val)
	} else if val := os.Getenv("AZURE_CLOUD_CONFIG_FILE"); len(val) > 0 {
		// cloud config via JSON file
		data, err := os.ReadFile(val) // #nosec G304
		if err != nil {
			return cloudConfig, fmt.Errorf(`unable to parse json for AzurePrivateCloud from env var AZURE_CLOUD_CONFIG_FILE, see https://github.com/webdevops/go-common/tree/main/azuresdk: %w`, err)
		}
		cloudConfigJson = data
	}

	if len(cloudConfigJson) == 0 {
		return cloudConfig, fmt.Errorf(`AzurePrivateCloud needs cloudconfig json passed via env var AZURE_CLOUD_CONFIG or AZURE_CLOUD_CONFIG_FILE, see https://github.com/webdevops/go-common/tree/main/azuresdk`)
	}

	if err := json.Unmarshal([]byte(cloudConfigJson), &cloudConfig); err != nil {
		return cloudConfig, fmt.Errorf(`unable to parse json for AzurePrivateCloud from env var AZURE_CLOUD_CONFIG or AZURE_CLOUD_CONFIG_FILE, see https://github.com/webdevops/go-common/tree/main/azuresdk: %w`, err)
	}

	return cloudConfig, nil
}
