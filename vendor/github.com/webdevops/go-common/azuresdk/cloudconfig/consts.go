package cloudconfig

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
)

type (
	CloudName string
)

const (
	// Environment names
	AzurePublicCloud     = CloudName("AzurePublicCloud")
	AzureChinaCloud      = CloudName("AzureChinaCloud")
	AzureGovernmentCloud = CloudName("AzureGovernmentCloud")
	AzurePrivateCloud    = CloudName("AzurePrivateCloud")

	// Service name
	ServiceNameMicrosoftGraph        cloud.ServiceName = "microsoftGraph"
	ServiceNameLogAnalyticsWorkspace cloud.ServiceName = "logAnalytics"
)
