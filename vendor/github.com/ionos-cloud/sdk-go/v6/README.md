![CI](https://github.com/ionos-cloud/sdk-resources/workflows/%5B%20CI%20%5D%20CloudApi%20V6%20/%20Go/badge.svg)
[![Gitter](https://img.shields.io/gitter/room/ionos-cloud/sdk-general)](https://gitter.im/ionos-cloud/sdk-general)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=ionos-cloud_sdk-go&metric=alert_status)](https://sonarcloud.io/dashboard?id=ionos-cloud_sdk-go)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=ionos-cloud_sdk-go&metric=bugs)](https://sonarcloud.io/dashboard?id=ionos-cloud_sdk-go)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=ionos-cloud_sdk-go&metric=sqale_rating)](https://sonarcloud.io/dashboard?id=ionos-cloud_sdk-go)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=ionos-cloud_sdk-go&metric=reliability_rating)](https://sonarcloud.io/dashboard?id=ionos-cloud_sdk-go)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=ionos-cloud_sdk-go&metric=security_rating)](https://sonarcloud.io/dashboard?id=ionos-cloud_sdk-go)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=ionos-cloud_sdk-go&metric=vulnerabilities)](https://sonarcloud.io/dashboard?id=ionos-cloud_sdk-go)
[![Release](https://img.shields.io/github/v/release/ionos-cloud/sdk-go.svg)](https://github.com/ionos-cloud/sdk-go/releases/latest)
[![Release Date](https://img.shields.io/github/release-date/ionos-cloud/sdk-go.svg)](https://github.com/ionos-cloud/sdk-go/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/ionos-cloud/sdk-go.svg)](https://github.com/ionos-cloud/sdk-go)

![Alt text](.github/IONOS.CLOUD.BLU.svg?raw=true "Title")


# Go API client for ionoscloud

IONOS Enterprise-grade Infrastructure as a Service (IaaS) solutions can be managed through the Cloud API, in addition or as an alternative to the \"Data Center Designer\" (DCD) browser-based tool. 

 Both methods employ consistent concepts and features, deliver similar power and flexibility, and can be used to perform a multitude of management tasks, including adding servers, volumes, configuring networks, and so on.

## Overview
The IONOS Cloud SDK for GO provides you with access to the IONOS Cloud API. The client library supports both simple and complex requests.
It is designed for developers who are building applications in GO . The SDK for GO wraps the IONOS Cloud API. All API operations are performed over SSL and authenticated using your IONOS Cloud portal credentials.
The API can be accessed within an instance running in IONOS Cloud or directly over the Internet from any application that can send an HTTPS request and receive an HTTPS response.

## Installing

### Use go get to retrieve the SDK to add it to your GOPATH workspace, or project's Go module dependencies.
```bash
go get github.com/ionos-cloud/sdk-go/v6
```
To update the SDK use go get -u to retrieve the latest version of the SDK.
```bash
go get -u github.com/ionos-cloud/sdk-go/v6
```
### Go Modules

If you are using Go modules, your go get will default to the latest tagged release version of the SDK. To get a specific release version of the SDK use @<tag> in your go get command.
```bash
go get github.com/ionos-cloud/sdk-go/v6@v6.0.0
```
To get the latest SDK repository, use @latest.
```bash
go get github.com/ionos-cloud/sdk-go/v6@latest
```

## Environment Variables

| Environment Variable | Description                                                                                                                                                                                                                    |
|----------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `IONOS_USERNAME`     | Specify the username used to login, to authenticate against the IONOS Cloud API                                                                                                                                                |
| `IONOS_PASSWORD`     | Specify the password used to login, to authenticate against the IONOS Cloud API                                                                                                                                                |
| `IONOS_TOKEN`        | Specify the token used to login, if a token is being used instead of username and password                                                                                                                                     |
| `IONOS_API_URL`      | Specify the API URL. It will overwrite the API endpoint default value `api.ionos.com`. Note: the host URL does not contain the `/cloudapi/v6` path, so it should _not_ be included in the `IONOS_API_URL` environment variable |
| `IONOS_LOGLEVEL`     | Specify the Log Level used to log messages. Possible values: Off, Debug, Trace |
| `IONOS_PINNED_CERT`  | Specify the SHA-256 public fingerprint here, enables certificate pinning                                                                                                                                                       |

⚠️ **_Note: To overwrite the api endpoint - `api.ionos.com`, the environment variable `$IONOS_API_URL` can be set, and used with `NewConfigurationFromEnv()` function._**

## Examples

Examples for creating resources using the Go SDK can be found [here](examples/)

## Authentication

### Basic Authentication

- **Type**: HTTP basic authentication

Example

```golang
import (
	"context"
	"fmt"
	"github.com/ionos-cloud/sdk-go/v6"
	"log"
)

func basicAuthExample() error {
	cfg := ionoscloud.NewConfiguration("username_here", "pwd_here", "", "")
	cfg.Debug = true
	apiClient := ionoscloud.NewAPIClient(cfg)
	datacenters, _, err := apiClient.DataCentersApi.DatacentersGet(context.Background()).Depth(1).Execute()
	if err != nil {
		return fmt.Errorf("error retrieving datacenters %w", err)
	}
	if datacenters.HasItems() {
		for _, dc := range *datacenters.GetItems() {
			if dc.HasProperties() && dc.GetProperties().HasName() {
				fmt.Println(*dc.GetProperties().GetName())
			}
		}
	}
	return nil
}
```
### Token Authentication
There are 2 ways to generate your token:

 ### Generate token using [sdk-go-auth](https://github.com/ionos-cloud/sdk-go-auth):
```golang
    import (
        "context"
        "fmt"
        authApi "github.com/ionos-cloud/sdk-go-auth"
        "github.com/ionos-cloud/sdk-go/v6"
        "log"
    )

    func TokenAuthExample() error {
        //note: to use NewConfigurationFromEnv(), you need to previously set IONOS_USERNAME and IONOS_PASSWORD as env variables
        authClient := authApi.NewAPIClient(authApi.NewConfigurationFromEnv())
        jwt, _, err := authClient.TokensApi.TokensGenerate(context.Background()).Execute()
        if err != nil {
            return fmt.Errorf("error occurred while generating token (%w)", err)
        }
        if !jwt.HasToken() {
            return fmt.Errorf("could not generate token")
        }
        cfg := ionoscloud.NewConfiguration("", "", *jwt.GetToken(), "")
        cfg.Debug = true
        apiClient := ionoscloud.NewAPIClient(cfg)
        datacenters, _, err := apiClient.DataCentersApi.DatacenterGet(context.Background()).Depth(1).Execute()
        if err != nil {
            return fmt.Errorf("error retrieving datacenters (%w)", err)
        }
        return nil
    }
```
 ### Generate token using ionosctl:
  Install ionosctl as explained [here](https://github.com/ionos-cloud/ionosctl)
  Run commands to login and generate your token.
```golang
ionosctl login
ionosctl token generate
export IONOS_TOKEN="insert_here_token_saved_from_generate_command"
```
 Save the generated token and use it to authenticate:
```golang
    import (
        "context"
        "fmt"
        "github.com/ionos-cloud/sdk-go/v6"
        "log"
    )

    func TokenAuthExample() error {
        //note: to use NewConfigurationFromEnv(), you need to previously set IONOS_TOKEN as env variables
        authClient := authApi.NewAPIClient(authApi.NewConfigurationFromEnv())
        cfg.Debug = true
        apiClient := ionoscloud.NewAPIClient(cfg)
        datacenters, _, err := apiClient.DataCenter6Api.DatacentersGet(context.Background()).Depth(1).Execute()
        if err != nil {
            return fmt.Errorf("error retrieving datacenters (%w)", err)
        }
        return nil
    }
```

## Certificate pinning:

You can enable certificate pinning if you want to bypass the normal certificate checking procedure,
by doing the following:

Set env variable IONOS_PINNED_CERT=<insert_sha256_public_fingerprint_here>

You can get the sha256 fingerprint most easily from the browser by inspecting the certificate.

### Depth

Many of the _List_ or _Get_ operations will accept an optional _depth_ argument. Setting this to a value between 0 and 5 affects the amount of data that is returned. The details returned vary depending on the resource being queried, but it generally follows this pattern. By default, the SDK sets the _depth_ argument to the maximum value.

| Depth | Description |
| :--- | :--- |
| 0 | Only direct properties are included. Children are not included. |
| 1 | Direct properties and children's references are returned. |
| 2 | Direct properties and children's properties are returned. |
| 3 | Direct properties, children's properties, and descendants' references are returned. |
| 4 | Direct properties, children's properties, and descendants' properties are returned. |
| 5 | Returns all available properties. |


#### How to set Depth parameter:

⚠️ **_Please use this parameter with caution. We recommend using the default value and raising its value only if it is needed._**

* On the configuration level:
```go
configuration := ionoscloud.NewConfiguration("USERNAME", "PASSWORD", "TOKEN", "URL")
configuration.SetDepth(5)
```
Using this method, the depth parameter will be set **on all the API calls**.

*  When calling a method:
```go
request := apiClient.DataCenterApi.DatacentersGet(context.Background()).Depth(1)
```
Using this method, the depth parameter will be set **on the current API call**.

* Using the default value:

If the depth parameter is not set, it will have the default value from the API that can be found [here](https://api.ionos.com/cloudapi/v6/swagger.json).

> Note: The priority for setting the depth parameter is: *set on function call > set on configuration level > set using the default value from the API*

### Pretty

The operations will also accept an optional _pretty_ argument. Setting this to a value of `true` or `false` controls whether the response is pretty-printed \(with indentation and new lines\). By default, the SDK sets the _pretty_ argument to `true`.

### Changing the base URL

Base URL for the HTTP operation can be changed by using the following function:

```go
requestProperties.SetURL("https://api.ionos.com/cloudapi/v6")
```

## Debugging

You can now inject any logger that implements Printf as a logger
instead of using the default sdk logger.
There are now Loglevels that you can set: `Off`, `Debug` and `Trace`.
`Off` - does not show any logs
`Debug` - regular logs, no sensitive information
`Trace` - we recommend you only set this field for debugging purposes. Disable it in your production environments because it can log sensitive data.
          It logs the full request and response without encryption, even for an HTTPS call. Verbose request and response logging can also significantly impact your application's performance.


```golang
package main
import "github.com/ionos-cloud/sdk-go/v6"
import "github.com/sirupsen/logrus"
func main() {
    // create your configuration. replace username, password, token and url with correct values, or use NewConfigurationFromEnv()
    // if you have set your env variables as explained above
    cfg := ionoscloud.NewConfiguration("username", "password", "token", "hostUrl")
    // enable request and response logging. this is the most verbose loglevel
    cfg.LogLevel = Trace
    // inject your own logger that implements Printf
    cfg.Logger = logrus.New()
    // create you api client with the configuration
    apiClient := ionoscloud.NewAPIClient(cfg)
}
```

If you want to see the API call request and response messages, you need to set the Debug field in the Configuration struct:

⚠️ **_Note: the field `Debug` is now deprecated and will be replaced with `LogLevel` in the future.

```golang
package main
import "github.com/ionos-cloud/sdk-go/v6"
func main() {
    // create your configuration. replace username, password, token and url with correct values, or use NewConfigurationFromEnv()
    // if you have set your env variables as explained above
    cfg := ionoscloud.NewConfiguration("username", "password", "token", "hostUrl")
    // enable request and response logging
    cfg.Debug = true
    // create you api client with the configuration
    apiClient := ionoscloud.NewAPIClient(cfg)
}
```

⚠️ **_Note: We recommend you only set this field for debugging purposes.
Disable it in your production environments because it can log sensitive data.
It logs the full request and response without encryption, even for an HTTPS call.
Verbose request and response logging can also significantly impact your application's performance._**


## Documentation for API Endpoints

All URIs are relative to *https://api.ionos.com/cloudapi/v6*
<details >
<summary title="Click to toggle">API Endpoints table</summary>


Class | Method | HTTP request | Description
------------- | ------------- | ------------- | -------------
DefaultApi | [**ApiInfoGet**](docs/api/DefaultApi.md#apiinfoget) | **Get** / | Display API information
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersDelete**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersdelete) | **Delete** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId} | Delete Application Load Balancers
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersFindByApplicationLoadBalancerId**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersfindbyapplicationloadbalancerid) | **Get** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId} | Retrieve Application Load Balancers
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersFlowlogsDelete**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersflowlogsdelete) | **Delete** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId}/flowlogs/{flowLogId} | Delete ALB Flow Logs
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersFlowlogsFindByFlowLogId**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersflowlogsfindbyflowlogid) | **Get** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId}/flowlogs/{flowLogId} | Retrieve ALB Flow Logs
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersFlowlogsGet**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersflowlogsget) | **Get** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId}/flowlogs | List ALB Flow Logs
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersFlowlogsPatch**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersflowlogspatch) | **Patch** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId}/flowlogs/{flowLogId} | Partially modify ALB Flow Logs
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersFlowlogsPost**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersflowlogspost) | **Post** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId}/flowlogs | Create ALB Flow Logs
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersFlowlogsPut**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersflowlogsput) | **Put** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId}/flowlogs/{flowLogId} | Modify ALB Flow Logs
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersForwardingrulesDelete**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersforwardingrulesdelete) | **Delete** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId}/forwardingrules/{forwardingRuleId} | Delete ALB forwarding rules
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersForwardingrulesFindByForwardingRuleId**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersforwardingrulesfindbyforwardingruleid) | **Get** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId}/forwardingrules/{forwardingRuleId} | Retrieve ALB forwarding rules
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersForwardingrulesGet**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersforwardingrulesget) | **Get** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId}/forwardingrules | List ALB forwarding rules
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersForwardingrulesPatch**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersforwardingrulespatch) | **Patch** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId}/forwardingrules/{forwardingRuleId} | Partially modify ALB forwarding rules
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersForwardingrulesPost**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersforwardingrulespost) | **Post** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId}/forwardingrules | Create ALB forwarding rules
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersForwardingrulesPut**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersforwardingrulesput) | **Put** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId}/forwardingrules/{forwardingRuleId} | Modify ALB forwarding rules
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersGet**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersget) | **Get** /datacenters/{datacenterId}/applicationloadbalancers | List Application Load Balancers
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersPatch**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancerspatch) | **Patch** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId} | Partially modify Application Load Balancers
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersPost**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancerspost) | **Post** /datacenters/{datacenterId}/applicationloadbalancers | Create Application Load Balancers
ApplicationLoadBalancersApi | [**DatacentersApplicationloadbalancersPut**](docs/api/ApplicationLoadBalancersApi.md#datacentersapplicationloadbalancersput) | **Put** /datacenters/{datacenterId}/applicationloadbalancers/{applicationLoadBalancerId} | Modify Application Load Balancers
BackupUnitsApi | [**BackupunitsDelete**](docs/api/BackupUnitsApi.md#backupunitsdelete) | **Delete** /backupunits/{backupunitId} | Delete backup units
BackupUnitsApi | [**BackupunitsFindById**](docs/api/BackupUnitsApi.md#backupunitsfindbyid) | **Get** /backupunits/{backupunitId} | Retrieve backup units
BackupUnitsApi | [**BackupunitsGet**](docs/api/BackupUnitsApi.md#backupunitsget) | **Get** /backupunits | List backup units
BackupUnitsApi | [**BackupunitsPatch**](docs/api/BackupUnitsApi.md#backupunitspatch) | **Patch** /backupunits/{backupunitId} | Partially modify backup units
BackupUnitsApi | [**BackupunitsPost**](docs/api/BackupUnitsApi.md#backupunitspost) | **Post** /backupunits | Create backup units
BackupUnitsApi | [**BackupunitsPut**](docs/api/BackupUnitsApi.md#backupunitsput) | **Put** /backupunits/{backupunitId} | Modify backup units
BackupUnitsApi | [**BackupunitsSsourlGet**](docs/api/BackupUnitsApi.md#backupunitsssourlget) | **Get** /backupunits/{backupunitId}/ssourl | Retrieve BU single sign-on URLs
ContractResourcesApi | [**ContractsGet**](docs/api/ContractResourcesApi.md#contractsget) | **Get** /contracts | Retrieve contracts
DataCentersApi | [**DatacentersDelete**](docs/api/DataCentersApi.md#datacentersdelete) | **Delete** /datacenters/{datacenterId} | Delete data centers
DataCentersApi | [**DatacentersFindById**](docs/api/DataCentersApi.md#datacentersfindbyid) | **Get** /datacenters/{datacenterId} | Retrieve data centers
DataCentersApi | [**DatacentersGet**](docs/api/DataCentersApi.md#datacentersget) | **Get** /datacenters | List your data centers
DataCentersApi | [**DatacentersPatch**](docs/api/DataCentersApi.md#datacenterspatch) | **Patch** /datacenters/{datacenterId} | Partially modify data centers
DataCentersApi | [**DatacentersPost**](docs/api/DataCentersApi.md#datacenterspost) | **Post** /datacenters | Create data centers
DataCentersApi | [**DatacentersPut**](docs/api/DataCentersApi.md#datacentersput) | **Put** /datacenters/{datacenterId} | Modify data centers
FirewallRulesApi | [**DatacentersServersNicsFirewallrulesDelete**](docs/api/FirewallRulesApi.md#datacentersserversnicsfirewallrulesdelete) | **Delete** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId}/firewallrules/{firewallruleId} | Delete firewall rules
FirewallRulesApi | [**DatacentersServersNicsFirewallrulesFindById**](docs/api/FirewallRulesApi.md#datacentersserversnicsfirewallrulesfindbyid) | **Get** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId}/firewallrules/{firewallruleId} | Retrieve firewall rules
FirewallRulesApi | [**DatacentersServersNicsFirewallrulesGet**](docs/api/FirewallRulesApi.md#datacentersserversnicsfirewallrulesget) | **Get** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId}/firewallrules | List firewall rules
FirewallRulesApi | [**DatacentersServersNicsFirewallrulesPatch**](docs/api/FirewallRulesApi.md#datacentersserversnicsfirewallrulespatch) | **Patch** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId}/firewallrules/{firewallruleId} | Partially modify firewall rules
FirewallRulesApi | [**DatacentersServersNicsFirewallrulesPost**](docs/api/FirewallRulesApi.md#datacentersserversnicsfirewallrulespost) | **Post** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId}/firewallrules | Create firewall rules
FirewallRulesApi | [**DatacentersServersNicsFirewallrulesPut**](docs/api/FirewallRulesApi.md#datacentersserversnicsfirewallrulesput) | **Put** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId}/firewallrules/{firewallruleId} | Modify firewall rules
FlowLogsApi | [**DatacentersServersNicsFlowlogsDelete**](docs/api/FlowLogsApi.md#datacentersserversnicsflowlogsdelete) | **Delete** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId}/flowlogs/{flowlogId} | Delete Flow Logs
FlowLogsApi | [**DatacentersServersNicsFlowlogsFindById**](docs/api/FlowLogsApi.md#datacentersserversnicsflowlogsfindbyid) | **Get** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId}/flowlogs/{flowlogId} | Retrieve Flow Logs
FlowLogsApi | [**DatacentersServersNicsFlowlogsGet**](docs/api/FlowLogsApi.md#datacentersserversnicsflowlogsget) | **Get** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId}/flowlogs | List Flow Logs
FlowLogsApi | [**DatacentersServersNicsFlowlogsPatch**](docs/api/FlowLogsApi.md#datacentersserversnicsflowlogspatch) | **Patch** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId}/flowlogs/{flowlogId} | Partially modify Flow Logs
FlowLogsApi | [**DatacentersServersNicsFlowlogsPost**](docs/api/FlowLogsApi.md#datacentersserversnicsflowlogspost) | **Post** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId}/flowlogs | Create Flow Logs
FlowLogsApi | [**DatacentersServersNicsFlowlogsPut**](docs/api/FlowLogsApi.md#datacentersserversnicsflowlogsput) | **Put** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId}/flowlogs/{flowlogId} | Modify Flow Logs
IPBlocksApi | [**IpblocksDelete**](docs/api/IPBlocksApi.md#ipblocksdelete) | **Delete** /ipblocks/{ipblockId} | Delete IP blocks
IPBlocksApi | [**IpblocksFindById**](docs/api/IPBlocksApi.md#ipblocksfindbyid) | **Get** /ipblocks/{ipblockId} | Retrieve IP blocks
IPBlocksApi | [**IpblocksGet**](docs/api/IPBlocksApi.md#ipblocksget) | **Get** /ipblocks | List IP blocks 
IPBlocksApi | [**IpblocksPatch**](docs/api/IPBlocksApi.md#ipblockspatch) | **Patch** /ipblocks/{ipblockId} | Partially modify IP blocks
IPBlocksApi | [**IpblocksPost**](docs/api/IPBlocksApi.md#ipblockspost) | **Post** /ipblocks | Reserve IP blocks
IPBlocksApi | [**IpblocksPut**](docs/api/IPBlocksApi.md#ipblocksput) | **Put** /ipblocks/{ipblockId} | Modify IP blocks
ImagesApi | [**ImagesDelete**](docs/api/ImagesApi.md#imagesdelete) | **Delete** /images/{imageId} | Delete images
ImagesApi | [**ImagesFindById**](docs/api/ImagesApi.md#imagesfindbyid) | **Get** /images/{imageId} | Retrieve images
ImagesApi | [**ImagesGet**](docs/api/ImagesApi.md#imagesget) | **Get** /images | List images
ImagesApi | [**ImagesPatch**](docs/api/ImagesApi.md#imagespatch) | **Patch** /images/{imageId} | Partially modify images
ImagesApi | [**ImagesPut**](docs/api/ImagesApi.md#imagesput) | **Put** /images/{imageId} | Modify images
KubernetesApi | [**K8sDelete**](docs/api/KubernetesApi.md#k8sdelete) | **Delete** /k8s/{k8sClusterId} | Delete Kubernetes clusters
KubernetesApi | [**K8sFindByClusterId**](docs/api/KubernetesApi.md#k8sfindbyclusterid) | **Get** /k8s/{k8sClusterId} | Retrieve Kubernetes clusters
KubernetesApi | [**K8sGet**](docs/api/KubernetesApi.md#k8sget) | **Get** /k8s | List Kubernetes clusters
KubernetesApi | [**K8sKubeconfigGet**](docs/api/KubernetesApi.md#k8skubeconfigget) | **Get** /k8s/{k8sClusterId}/kubeconfig | Retrieve Kubernetes configuration files
KubernetesApi | [**K8sNodepoolsDelete**](docs/api/KubernetesApi.md#k8snodepoolsdelete) | **Delete** /k8s/{k8sClusterId}/nodepools/{nodepoolId} | Delete Kubernetes node pools
KubernetesApi | [**K8sNodepoolsFindById**](docs/api/KubernetesApi.md#k8snodepoolsfindbyid) | **Get** /k8s/{k8sClusterId}/nodepools/{nodepoolId} | Retrieve Kubernetes node pools
KubernetesApi | [**K8sNodepoolsGet**](docs/api/KubernetesApi.md#k8snodepoolsget) | **Get** /k8s/{k8sClusterId}/nodepools | List Kubernetes node pools
KubernetesApi | [**K8sNodepoolsNodesDelete**](docs/api/KubernetesApi.md#k8snodepoolsnodesdelete) | **Delete** /k8s/{k8sClusterId}/nodepools/{nodepoolId}/nodes/{nodeId} | Delete Kubernetes nodes
KubernetesApi | [**K8sNodepoolsNodesFindById**](docs/api/KubernetesApi.md#k8snodepoolsnodesfindbyid) | **Get** /k8s/{k8sClusterId}/nodepools/{nodepoolId}/nodes/{nodeId} | Retrieve Kubernetes nodes
KubernetesApi | [**K8sNodepoolsNodesGet**](docs/api/KubernetesApi.md#k8snodepoolsnodesget) | **Get** /k8s/{k8sClusterId}/nodepools/{nodepoolId}/nodes | List Kubernetes nodes
KubernetesApi | [**K8sNodepoolsNodesReplacePost**](docs/api/KubernetesApi.md#k8snodepoolsnodesreplacepost) | **Post** /k8s/{k8sClusterId}/nodepools/{nodepoolId}/nodes/{nodeId}/replace | Recreate Kubernetes nodes
KubernetesApi | [**K8sNodepoolsPost**](docs/api/KubernetesApi.md#k8snodepoolspost) | **Post** /k8s/{k8sClusterId}/nodepools | Create Kubernetes node pools
KubernetesApi | [**K8sNodepoolsPut**](docs/api/KubernetesApi.md#k8snodepoolsput) | **Put** /k8s/{k8sClusterId}/nodepools/{nodepoolId} | Modify Kubernetes node pools
KubernetesApi | [**K8sPost**](docs/api/KubernetesApi.md#k8spost) | **Post** /k8s | Create Kubernetes clusters
KubernetesApi | [**K8sPut**](docs/api/KubernetesApi.md#k8sput) | **Put** /k8s/{k8sClusterId} | Modify Kubernetes clusters
KubernetesApi | [**K8sVersionsDefaultGet**](docs/api/KubernetesApi.md#k8sversionsdefaultget) | **Get** /k8s/versions/default | Retrieve current default Kubernetes version
KubernetesApi | [**K8sVersionsGet**](docs/api/KubernetesApi.md#k8sversionsget) | **Get** /k8s/versions | List Kubernetes versions
LANsApi | [**DatacentersLansDelete**](docs/api/LANsApi.md#datacenterslansdelete) | **Delete** /datacenters/{datacenterId}/lans/{lanId} | Delete LANs
LANsApi | [**DatacentersLansFindById**](docs/api/LANsApi.md#datacenterslansfindbyid) | **Get** /datacenters/{datacenterId}/lans/{lanId} | Retrieve LANs
LANsApi | [**DatacentersLansGet**](docs/api/LANsApi.md#datacenterslansget) | **Get** /datacenters/{datacenterId}/lans | List LANs
LANsApi | [**DatacentersLansNicsFindById**](docs/api/LANsApi.md#datacenterslansnicsfindbyid) | **Get** /datacenters/{datacenterId}/lans/{lanId}/nics/{nicId} | Retrieve attached NICs
LANsApi | [**DatacentersLansNicsGet**](docs/api/LANsApi.md#datacenterslansnicsget) | **Get** /datacenters/{datacenterId}/lans/{lanId}/nics | List LAN members
LANsApi | [**DatacentersLansNicsPost**](docs/api/LANsApi.md#datacenterslansnicspost) | **Post** /datacenters/{datacenterId}/lans/{lanId}/nics | Attach NICs
LANsApi | [**DatacentersLansPatch**](docs/api/LANsApi.md#datacenterslanspatch) | **Patch** /datacenters/{datacenterId}/lans/{lanId} | Partially modify LANs
LANsApi | [**DatacentersLansPost**](docs/api/LANsApi.md#datacenterslanspost) | **Post** /datacenters/{datacenterId}/lans | Create LANs
LANsApi | [**DatacentersLansPut**](docs/api/LANsApi.md#datacenterslansput) | **Put** /datacenters/{datacenterId}/lans/{lanId} | Modify LANs
LabelsApi | [**DatacentersLabelsDelete**](docs/api/LabelsApi.md#datacenterslabelsdelete) | **Delete** /datacenters/{datacenterId}/labels/{key} | Delete data center labels
LabelsApi | [**DatacentersLabelsFindByKey**](docs/api/LabelsApi.md#datacenterslabelsfindbykey) | **Get** /datacenters/{datacenterId}/labels/{key} | Retrieve data center labels
LabelsApi | [**DatacentersLabelsGet**](docs/api/LabelsApi.md#datacenterslabelsget) | **Get** /datacenters/{datacenterId}/labels | List data center labels
LabelsApi | [**DatacentersLabelsPost**](docs/api/LabelsApi.md#datacenterslabelspost) | **Post** /datacenters/{datacenterId}/labels | Create data center labels
LabelsApi | [**DatacentersLabelsPut**](docs/api/LabelsApi.md#datacenterslabelsput) | **Put** /datacenters/{datacenterId}/labels/{key} | Modify data center labels
LabelsApi | [**DatacentersServersLabelsDelete**](docs/api/LabelsApi.md#datacentersserverslabelsdelete) | **Delete** /datacenters/{datacenterId}/servers/{serverId}/labels/{key} | Delete server labels
LabelsApi | [**DatacentersServersLabelsFindByKey**](docs/api/LabelsApi.md#datacentersserverslabelsfindbykey) | **Get** /datacenters/{datacenterId}/servers/{serverId}/labels/{key} | Retrieve server labels
LabelsApi | [**DatacentersServersLabelsGet**](docs/api/LabelsApi.md#datacentersserverslabelsget) | **Get** /datacenters/{datacenterId}/servers/{serverId}/labels | List server labels
LabelsApi | [**DatacentersServersLabelsPost**](docs/api/LabelsApi.md#datacentersserverslabelspost) | **Post** /datacenters/{datacenterId}/servers/{serverId}/labels | Create server labels
LabelsApi | [**DatacentersServersLabelsPut**](docs/api/LabelsApi.md#datacentersserverslabelsput) | **Put** /datacenters/{datacenterId}/servers/{serverId}/labels/{key} | Modify server labels
LabelsApi | [**DatacentersVolumesLabelsDelete**](docs/api/LabelsApi.md#datacentersvolumeslabelsdelete) | **Delete** /datacenters/{datacenterId}/volumes/{volumeId}/labels/{key} | Delete volume labels
LabelsApi | [**DatacentersVolumesLabelsFindByKey**](docs/api/LabelsApi.md#datacentersvolumeslabelsfindbykey) | **Get** /datacenters/{datacenterId}/volumes/{volumeId}/labels/{key} | Retrieve volume labels
LabelsApi | [**DatacentersVolumesLabelsGet**](docs/api/LabelsApi.md#datacentersvolumeslabelsget) | **Get** /datacenters/{datacenterId}/volumes/{volumeId}/labels | List volume labels
LabelsApi | [**DatacentersVolumesLabelsPost**](docs/api/LabelsApi.md#datacentersvolumeslabelspost) | **Post** /datacenters/{datacenterId}/volumes/{volumeId}/labels | Create volume labels
LabelsApi | [**DatacentersVolumesLabelsPut**](docs/api/LabelsApi.md#datacentersvolumeslabelsput) | **Put** /datacenters/{datacenterId}/volumes/{volumeId}/labels/{key} | Modify volume labels
LabelsApi | [**IpblocksLabelsDelete**](docs/api/LabelsApi.md#ipblockslabelsdelete) | **Delete** /ipblocks/{ipblockId}/labels/{key} | Delete IP block labels
LabelsApi | [**IpblocksLabelsFindByKey**](docs/api/LabelsApi.md#ipblockslabelsfindbykey) | **Get** /ipblocks/{ipblockId}/labels/{key} | Retrieve IP block labels
LabelsApi | [**IpblocksLabelsGet**](docs/api/LabelsApi.md#ipblockslabelsget) | **Get** /ipblocks/{ipblockId}/labels | List IP block labels
LabelsApi | [**IpblocksLabelsPost**](docs/api/LabelsApi.md#ipblockslabelspost) | **Post** /ipblocks/{ipblockId}/labels | Create IP block labels
LabelsApi | [**IpblocksLabelsPut**](docs/api/LabelsApi.md#ipblockslabelsput) | **Put** /ipblocks/{ipblockId}/labels/{key} | Modify IP block labels
LabelsApi | [**LabelsFindByUrn**](docs/api/LabelsApi.md#labelsfindbyurn) | **Get** /labels/{labelurn} | Retrieve labels by URN
LabelsApi | [**LabelsGet**](docs/api/LabelsApi.md#labelsget) | **Get** /labels | List labels 
LabelsApi | [**SnapshotsLabelsDelete**](docs/api/LabelsApi.md#snapshotslabelsdelete) | **Delete** /snapshots/{snapshotId}/labels/{key} | Delete snapshot labels
LabelsApi | [**SnapshotsLabelsFindByKey**](docs/api/LabelsApi.md#snapshotslabelsfindbykey) | **Get** /snapshots/{snapshotId}/labels/{key} | Retrieve snapshot labels
LabelsApi | [**SnapshotsLabelsGet**](docs/api/LabelsApi.md#snapshotslabelsget) | **Get** /snapshots/{snapshotId}/labels | List snapshot labels
LabelsApi | [**SnapshotsLabelsPost**](docs/api/LabelsApi.md#snapshotslabelspost) | **Post** /snapshots/{snapshotId}/labels | Create snapshot labels
LabelsApi | [**SnapshotsLabelsPut**](docs/api/LabelsApi.md#snapshotslabelsput) | **Put** /snapshots/{snapshotId}/labels/{key} | Modify snapshot labels
LoadBalancersApi | [**DatacentersLoadbalancersBalancednicsDelete**](docs/api/LoadBalancersApi.md#datacentersloadbalancersbalancednicsdelete) | **Delete** /datacenters/{datacenterId}/loadbalancers/{loadbalancerId}/balancednics/{nicId} | Detach balanced NICs
LoadBalancersApi | [**DatacentersLoadbalancersBalancednicsFindByNicId**](docs/api/LoadBalancersApi.md#datacentersloadbalancersbalancednicsfindbynicid) | **Get** /datacenters/{datacenterId}/loadbalancers/{loadbalancerId}/balancednics/{nicId} | Retrieve balanced NICs
LoadBalancersApi | [**DatacentersLoadbalancersBalancednicsGet**](docs/api/LoadBalancersApi.md#datacentersloadbalancersbalancednicsget) | **Get** /datacenters/{datacenterId}/loadbalancers/{loadbalancerId}/balancednics | List balanced NICs
LoadBalancersApi | [**DatacentersLoadbalancersBalancednicsPost**](docs/api/LoadBalancersApi.md#datacentersloadbalancersbalancednicspost) | **Post** /datacenters/{datacenterId}/loadbalancers/{loadbalancerId}/balancednics | Attach balanced NICs
LoadBalancersApi | [**DatacentersLoadbalancersDelete**](docs/api/LoadBalancersApi.md#datacentersloadbalancersdelete) | **Delete** /datacenters/{datacenterId}/loadbalancers/{loadbalancerId} | Delete Load Balancers
LoadBalancersApi | [**DatacentersLoadbalancersFindById**](docs/api/LoadBalancersApi.md#datacentersloadbalancersfindbyid) | **Get** /datacenters/{datacenterId}/loadbalancers/{loadbalancerId} | Retrieve Load Balancers
LoadBalancersApi | [**DatacentersLoadbalancersGet**](docs/api/LoadBalancersApi.md#datacentersloadbalancersget) | **Get** /datacenters/{datacenterId}/loadbalancers | List Load Balancers
LoadBalancersApi | [**DatacentersLoadbalancersPatch**](docs/api/LoadBalancersApi.md#datacentersloadbalancerspatch) | **Patch** /datacenters/{datacenterId}/loadbalancers/{loadbalancerId} | Partially modify Load Balancers
LoadBalancersApi | [**DatacentersLoadbalancersPost**](docs/api/LoadBalancersApi.md#datacentersloadbalancerspost) | **Post** /datacenters/{datacenterId}/loadbalancers | Create Load Balancers
LoadBalancersApi | [**DatacentersLoadbalancersPut**](docs/api/LoadBalancersApi.md#datacentersloadbalancersput) | **Put** /datacenters/{datacenterId}/loadbalancers/{loadbalancerId} | Modify Load Balancers
LocationsApi | [**LocationsFindByRegionId**](docs/api/LocationsApi.md#locationsfindbyregionid) | **Get** /locations/{regionId} | List locations within regions
LocationsApi | [**LocationsFindByRegionIdAndId**](docs/api/LocationsApi.md#locationsfindbyregionidandid) | **Get** /locations/{regionId}/{locationId} | Retrieve specified locations
LocationsApi | [**LocationsGet**](docs/api/LocationsApi.md#locationsget) | **Get** /locations | List locations
NATGatewaysApi | [**DatacentersNatgatewaysDelete**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysdelete) | **Delete** /datacenters/{datacenterId}/natgateways/{natGatewayId} | Delete NAT Gateways
NATGatewaysApi | [**DatacentersNatgatewaysFindByNatGatewayId**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysfindbynatgatewayid) | **Get** /datacenters/{datacenterId}/natgateways/{natGatewayId} | Retrieve NAT Gateways
NATGatewaysApi | [**DatacentersNatgatewaysFlowlogsDelete**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysflowlogsdelete) | **Delete** /datacenters/{datacenterId}/natgateways/{natGatewayId}/flowlogs/{flowLogId} | Delete NAT Gateway Flow Logs
NATGatewaysApi | [**DatacentersNatgatewaysFlowlogsFindByFlowLogId**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysflowlogsfindbyflowlogid) | **Get** /datacenters/{datacenterId}/natgateways/{natGatewayId}/flowlogs/{flowLogId} | Retrieve NAT Gateway Flow Logs
NATGatewaysApi | [**DatacentersNatgatewaysFlowlogsGet**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysflowlogsget) | **Get** /datacenters/{datacenterId}/natgateways/{natGatewayId}/flowlogs | List NAT Gateway Flow Logs
NATGatewaysApi | [**DatacentersNatgatewaysFlowlogsPatch**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysflowlogspatch) | **Patch** /datacenters/{datacenterId}/natgateways/{natGatewayId}/flowlogs/{flowLogId} | Partially modify NAT Gateway Flow Logs
NATGatewaysApi | [**DatacentersNatgatewaysFlowlogsPost**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysflowlogspost) | **Post** /datacenters/{datacenterId}/natgateways/{natGatewayId}/flowlogs | Create NAT Gateway Flow Logs
NATGatewaysApi | [**DatacentersNatgatewaysFlowlogsPut**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysflowlogsput) | **Put** /datacenters/{datacenterId}/natgateways/{natGatewayId}/flowlogs/{flowLogId} | Modify NAT Gateway Flow Logs
NATGatewaysApi | [**DatacentersNatgatewaysGet**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysget) | **Get** /datacenters/{datacenterId}/natgateways | List NAT Gateways
NATGatewaysApi | [**DatacentersNatgatewaysPatch**](docs/api/NATGatewaysApi.md#datacentersnatgatewayspatch) | **Patch** /datacenters/{datacenterId}/natgateways/{natGatewayId} | Partially modify NAT Gateways
NATGatewaysApi | [**DatacentersNatgatewaysPost**](docs/api/NATGatewaysApi.md#datacentersnatgatewayspost) | **Post** /datacenters/{datacenterId}/natgateways | Create NAT Gateways
NATGatewaysApi | [**DatacentersNatgatewaysPut**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysput) | **Put** /datacenters/{datacenterId}/natgateways/{natGatewayId} | Modify NAT Gateways
NATGatewaysApi | [**DatacentersNatgatewaysRulesDelete**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysrulesdelete) | **Delete** /datacenters/{datacenterId}/natgateways/{natGatewayId}/rules/{natGatewayRuleId} | Delete NAT Gateway rules
NATGatewaysApi | [**DatacentersNatgatewaysRulesFindByNatGatewayRuleId**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysrulesfindbynatgatewayruleid) | **Get** /datacenters/{datacenterId}/natgateways/{natGatewayId}/rules/{natGatewayRuleId} | Retrieve NAT Gateway rules
NATGatewaysApi | [**DatacentersNatgatewaysRulesGet**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysrulesget) | **Get** /datacenters/{datacenterId}/natgateways/{natGatewayId}/rules | List NAT Gateway rules
NATGatewaysApi | [**DatacentersNatgatewaysRulesPatch**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysrulespatch) | **Patch** /datacenters/{datacenterId}/natgateways/{natGatewayId}/rules/{natGatewayRuleId} | Partially modify NAT Gateway rules
NATGatewaysApi | [**DatacentersNatgatewaysRulesPost**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysrulespost) | **Post** /datacenters/{datacenterId}/natgateways/{natGatewayId}/rules | Create NAT Gateway rules
NATGatewaysApi | [**DatacentersNatgatewaysRulesPut**](docs/api/NATGatewaysApi.md#datacentersnatgatewaysrulesput) | **Put** /datacenters/{datacenterId}/natgateways/{natGatewayId}/rules/{natGatewayRuleId} | Modify NAT Gateway rules
NetworkInterfacesApi | [**DatacentersServersNicsDelete**](docs/api/NetworkInterfacesApi.md#datacentersserversnicsdelete) | **Delete** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId} | Delete NICs
NetworkInterfacesApi | [**DatacentersServersNicsFindById**](docs/api/NetworkInterfacesApi.md#datacentersserversnicsfindbyid) | **Get** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId} | Retrieve NICs
NetworkInterfacesApi | [**DatacentersServersNicsGet**](docs/api/NetworkInterfacesApi.md#datacentersserversnicsget) | **Get** /datacenters/{datacenterId}/servers/{serverId}/nics | List NICs
NetworkInterfacesApi | [**DatacentersServersNicsPatch**](docs/api/NetworkInterfacesApi.md#datacentersserversnicspatch) | **Patch** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId} | Partially modify NICs
NetworkInterfacesApi | [**DatacentersServersNicsPost**](docs/api/NetworkInterfacesApi.md#datacentersserversnicspost) | **Post** /datacenters/{datacenterId}/servers/{serverId}/nics | Create NICs
NetworkInterfacesApi | [**DatacentersServersNicsPut**](docs/api/NetworkInterfacesApi.md#datacentersserversnicsput) | **Put** /datacenters/{datacenterId}/servers/{serverId}/nics/{nicId} | Modify NICs
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersDelete**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersdelete) | **Delete** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId} | Delete Network Load Balancers
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersFindByNetworkLoadBalancerId**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersfindbynetworkloadbalancerid) | **Get** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId} | Retrieve Network Load Balancers
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersFlowlogsDelete**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersflowlogsdelete) | **Delete** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId}/flowlogs/{flowLogId} | Delete NLB Flow Logs
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersFlowlogsFindByFlowLogId**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersflowlogsfindbyflowlogid) | **Get** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId}/flowlogs/{flowLogId} | Retrieve NLB Flow Logs
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersFlowlogsGet**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersflowlogsget) | **Get** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId}/flowlogs | List NLB Flow Logs
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersFlowlogsPatch**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersflowlogspatch) | **Patch** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId}/flowlogs/{flowLogId} | Partially modify NLB Flow Logs
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersFlowlogsPost**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersflowlogspost) | **Post** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId}/flowlogs | Create NLB Flow Logs
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersFlowlogsPut**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersflowlogsput) | **Put** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId}/flowlogs/{flowLogId} | Modify NLB Flow Logs
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersForwardingrulesDelete**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersforwardingrulesdelete) | **Delete** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId}/forwardingrules/{forwardingRuleId} | Delete NLB forwarding rules
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersForwardingrulesFindByForwardingRuleId**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersforwardingrulesfindbyforwardingruleid) | **Get** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId}/forwardingrules/{forwardingRuleId} | Retrieve NLB forwarding rules
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersForwardingrulesGet**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersforwardingrulesget) | **Get** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId}/forwardingrules | List NLB forwarding rules
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersForwardingrulesPatch**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersforwardingrulespatch) | **Patch** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId}/forwardingrules/{forwardingRuleId} | Partially modify NLB forwarding rules
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersForwardingrulesPost**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersforwardingrulespost) | **Post** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId}/forwardingrules | Create NLB forwarding rules
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersForwardingrulesPut**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersforwardingrulesput) | **Put** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId}/forwardingrules/{forwardingRuleId} | Modify NLB forwarding rules
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersGet**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersget) | **Get** /datacenters/{datacenterId}/networkloadbalancers | List Network Load Balancers
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersPatch**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancerspatch) | **Patch** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId} | Partially modify Network Load Balancers
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersPost**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancerspost) | **Post** /datacenters/{datacenterId}/networkloadbalancers | Create Network Load Balancers
NetworkLoadBalancersApi | [**DatacentersNetworkloadbalancersPut**](docs/api/NetworkLoadBalancersApi.md#datacentersnetworkloadbalancersput) | **Put** /datacenters/{datacenterId}/networkloadbalancers/{networkLoadBalancerId} | Modify Network Load Balancers
PrivateCrossConnectsApi | [**PccsDelete**](docs/api/PrivateCrossConnectsApi.md#pccsdelete) | **Delete** /pccs/{pccId} | Delete private Cross-Connects
PrivateCrossConnectsApi | [**PccsFindById**](docs/api/PrivateCrossConnectsApi.md#pccsfindbyid) | **Get** /pccs/{pccId} | Retrieve private Cross-Connects
PrivateCrossConnectsApi | [**PccsGet**](docs/api/PrivateCrossConnectsApi.md#pccsget) | **Get** /pccs | List private Cross-Connects
PrivateCrossConnectsApi | [**PccsPatch**](docs/api/PrivateCrossConnectsApi.md#pccspatch) | **Patch** /pccs/{pccId} | Partially modify private Cross-Connects
PrivateCrossConnectsApi | [**PccsPost**](docs/api/PrivateCrossConnectsApi.md#pccspost) | **Post** /pccs | Create private Cross-Connects
RequestsApi | [**RequestsFindById**](docs/api/RequestsApi.md#requestsfindbyid) | **Get** /requests/{requestId} | Retrieve requests
RequestsApi | [**RequestsGet**](docs/api/RequestsApi.md#requestsget) | **Get** /requests | List requests
RequestsApi | [**RequestsStatusGet**](docs/api/RequestsApi.md#requestsstatusget) | **Get** /requests/{requestId}/status | Retrieve request status
ServersApi | [**DatacentersServersCdromsDelete**](docs/api/ServersApi.md#datacentersserverscdromsdelete) | **Delete** /datacenters/{datacenterId}/servers/{serverId}/cdroms/{cdromId} | Detach CD-ROMs
ServersApi | [**DatacentersServersCdromsFindById**](docs/api/ServersApi.md#datacentersserverscdromsfindbyid) | **Get** /datacenters/{datacenterId}/servers/{serverId}/cdroms/{cdromId} | Retrieve attached CD-ROMs
ServersApi | [**DatacentersServersCdromsGet**](docs/api/ServersApi.md#datacentersserverscdromsget) | **Get** /datacenters/{datacenterId}/servers/{serverId}/cdroms | List attached CD-ROMs 
ServersApi | [**DatacentersServersCdromsPost**](docs/api/ServersApi.md#datacentersserverscdromspost) | **Post** /datacenters/{datacenterId}/servers/{serverId}/cdroms | Attach CD-ROMs
ServersApi | [**DatacentersServersDelete**](docs/api/ServersApi.md#datacentersserversdelete) | **Delete** /datacenters/{datacenterId}/servers/{serverId} | Delete servers
ServersApi | [**DatacentersServersFindById**](docs/api/ServersApi.md#datacentersserversfindbyid) | **Get** /datacenters/{datacenterId}/servers/{serverId} | Retrieve servers by ID
ServersApi | [**DatacentersServersGet**](docs/api/ServersApi.md#datacentersserversget) | **Get** /datacenters/{datacenterId}/servers | List servers 
ServersApi | [**DatacentersServersPatch**](docs/api/ServersApi.md#datacentersserverspatch) | **Patch** /datacenters/{datacenterId}/servers/{serverId} | Partially modify servers
ServersApi | [**DatacentersServersPost**](docs/api/ServersApi.md#datacentersserverspost) | **Post** /datacenters/{datacenterId}/servers | Create servers
ServersApi | [**DatacentersServersPut**](docs/api/ServersApi.md#datacentersserversput) | **Put** /datacenters/{datacenterId}/servers/{serverId} | Modify servers
ServersApi | [**DatacentersServersRebootPost**](docs/api/ServersApi.md#datacentersserversrebootpost) | **Post** /datacenters/{datacenterId}/servers/{serverId}/reboot | Reboot servers
ServersApi | [**DatacentersServersRemoteConsoleGet**](docs/api/ServersApi.md#datacentersserversremoteconsoleget) | **Get** /datacenters/{datacenterId}/servers/{serverId}/remoteconsole | Get Remote Console link
ServersApi | [**DatacentersServersResumePost**](docs/api/ServersApi.md#datacentersserversresumepost) | **Post** /datacenters/{datacenterId}/servers/{serverId}/resume | Resume Cubes instances
ServersApi | [**DatacentersServersStartPost**](docs/api/ServersApi.md#datacentersserversstartpost) | **Post** /datacenters/{datacenterId}/servers/{serverId}/start | Start servers
ServersApi | [**DatacentersServersStopPost**](docs/api/ServersApi.md#datacentersserversstoppost) | **Post** /datacenters/{datacenterId}/servers/{serverId}/stop | Stop VMs
ServersApi | [**DatacentersServersSuspendPost**](docs/api/ServersApi.md#datacentersserverssuspendpost) | **Post** /datacenters/{datacenterId}/servers/{serverId}/suspend | Suspend Cubes instances
ServersApi | [**DatacentersServersTokenGet**](docs/api/ServersApi.md#datacentersserverstokenget) | **Get** /datacenters/{datacenterId}/servers/{serverId}/token | Get JASON Web Token
ServersApi | [**DatacentersServersUpgradePost**](docs/api/ServersApi.md#datacentersserversupgradepost) | **Post** /datacenters/{datacenterId}/servers/{serverId}/upgrade | Upgrade servers
ServersApi | [**DatacentersServersVolumesDelete**](docs/api/ServersApi.md#datacentersserversvolumesdelete) | **Delete** /datacenters/{datacenterId}/servers/{serverId}/volumes/{volumeId} | Detach volumes
ServersApi | [**DatacentersServersVolumesFindById**](docs/api/ServersApi.md#datacentersserversvolumesfindbyid) | **Get** /datacenters/{datacenterId}/servers/{serverId}/volumes/{volumeId} | Retrieve attached volumes
ServersApi | [**DatacentersServersVolumesGet**](docs/api/ServersApi.md#datacentersserversvolumesget) | **Get** /datacenters/{datacenterId}/servers/{serverId}/volumes | List attached volumes
ServersApi | [**DatacentersServersVolumesPost**](docs/api/ServersApi.md#datacentersserversvolumespost) | **Post** /datacenters/{datacenterId}/servers/{serverId}/volumes | Attach volumes
SnapshotsApi | [**SnapshotsDelete**](docs/api/SnapshotsApi.md#snapshotsdelete) | **Delete** /snapshots/{snapshotId} | Delete snapshots
SnapshotsApi | [**SnapshotsFindById**](docs/api/SnapshotsApi.md#snapshotsfindbyid) | **Get** /snapshots/{snapshotId} | Retrieve snapshots by ID
SnapshotsApi | [**SnapshotsGet**](docs/api/SnapshotsApi.md#snapshotsget) | **Get** /snapshots | List snapshots
SnapshotsApi | [**SnapshotsPatch**](docs/api/SnapshotsApi.md#snapshotspatch) | **Patch** /snapshots/{snapshotId} | Partially modify snapshots
SnapshotsApi | [**SnapshotsPut**](docs/api/SnapshotsApi.md#snapshotsput) | **Put** /snapshots/{snapshotId} | Modify snapshots
TargetGroupsApi | [**TargetGroupsDelete**](docs/api/TargetGroupsApi.md#targetgroupsdelete) | **Delete** /targetgroups/{targetGroupId} | Remove target groups
TargetGroupsApi | [**TargetgroupsFindByTargetGroupId**](docs/api/TargetGroupsApi.md#targetgroupsfindbytargetgroupid) | **Get** /targetgroups/{targetGroupId} | Retrieve target groups
TargetGroupsApi | [**TargetgroupsGet**](docs/api/TargetGroupsApi.md#targetgroupsget) | **Get** /targetgroups | List target groups
TargetGroupsApi | [**TargetgroupsPatch**](docs/api/TargetGroupsApi.md#targetgroupspatch) | **Patch** /targetgroups/{targetGroupId} | Partially modify target groups
TargetGroupsApi | [**TargetgroupsPost**](docs/api/TargetGroupsApi.md#targetgroupspost) | **Post** /targetgroups | Create target groups
TargetGroupsApi | [**TargetgroupsPut**](docs/api/TargetGroupsApi.md#targetgroupsput) | **Put** /targetgroups/{targetGroupId} | Modify target groups
TemplatesApi | [**TemplatesFindById**](docs/api/TemplatesApi.md#templatesfindbyid) | **Get** /templates/{templateId} | Retrieve Cubes Templates
TemplatesApi | [**TemplatesGet**](docs/api/TemplatesApi.md#templatesget) | **Get** /templates | List Cubes Templates
UserManagementApi | [**UmGroupsDelete**](docs/api/UserManagementApi.md#umgroupsdelete) | **Delete** /um/groups/{groupId} | Delete groups
UserManagementApi | [**UmGroupsFindById**](docs/api/UserManagementApi.md#umgroupsfindbyid) | **Get** /um/groups/{groupId} | Retrieve groups
UserManagementApi | [**UmGroupsGet**](docs/api/UserManagementApi.md#umgroupsget) | **Get** /um/groups | List all groups
UserManagementApi | [**UmGroupsPost**](docs/api/UserManagementApi.md#umgroupspost) | **Post** /um/groups | Create groups
UserManagementApi | [**UmGroupsPut**](docs/api/UserManagementApi.md#umgroupsput) | **Put** /um/groups/{groupId} | Modify groups
UserManagementApi | [**UmGroupsResourcesGet**](docs/api/UserManagementApi.md#umgroupsresourcesget) | **Get** /um/groups/{groupId}/resources | Retrieve group resources
UserManagementApi | [**UmGroupsSharesDelete**](docs/api/UserManagementApi.md#umgroupssharesdelete) | **Delete** /um/groups/{groupId}/shares/{resourceId} | Remove group shares
UserManagementApi | [**UmGroupsSharesFindByResourceId**](docs/api/UserManagementApi.md#umgroupssharesfindbyresourceid) | **Get** /um/groups/{groupId}/shares/{resourceId} | Retrieve group shares
UserManagementApi | [**UmGroupsSharesGet**](docs/api/UserManagementApi.md#umgroupssharesget) | **Get** /um/groups/{groupId}/shares | List group shares 
UserManagementApi | [**UmGroupsSharesPost**](docs/api/UserManagementApi.md#umgroupssharespost) | **Post** /um/groups/{groupId}/shares/{resourceId} | Add group shares
UserManagementApi | [**UmGroupsSharesPut**](docs/api/UserManagementApi.md#umgroupssharesput) | **Put** /um/groups/{groupId}/shares/{resourceId} | Modify group share privileges
UserManagementApi | [**UmGroupsUsersDelete**](docs/api/UserManagementApi.md#umgroupsusersdelete) | **Delete** /um/groups/{groupId}/users/{userId} | Remove users from groups
UserManagementApi | [**UmGroupsUsersGet**](docs/api/UserManagementApi.md#umgroupsusersget) | **Get** /um/groups/{groupId}/users | List group members
UserManagementApi | [**UmGroupsUsersPost**](docs/api/UserManagementApi.md#umgroupsuserspost) | **Post** /um/groups/{groupId}/users | Add group members
UserManagementApi | [**UmResourcesFindByType**](docs/api/UserManagementApi.md#umresourcesfindbytype) | **Get** /um/resources/{resourceType} | List resources by type
UserManagementApi | [**UmResourcesFindByTypeAndId**](docs/api/UserManagementApi.md#umresourcesfindbytypeandid) | **Get** /um/resources/{resourceType}/{resourceId} | Retrieve resources by type
UserManagementApi | [**UmResourcesGet**](docs/api/UserManagementApi.md#umresourcesget) | **Get** /um/resources | List all resources
UserManagementApi | [**UmUsersDelete**](docs/api/UserManagementApi.md#umusersdelete) | **Delete** /um/users/{userId} | Delete users
UserManagementApi | [**UmUsersFindById**](docs/api/UserManagementApi.md#umusersfindbyid) | **Get** /um/users/{userId} | Retrieve users
UserManagementApi | [**UmUsersGet**](docs/api/UserManagementApi.md#umusersget) | **Get** /um/users | List all users 
UserManagementApi | [**UmUsersGroupsGet**](docs/api/UserManagementApi.md#umusersgroupsget) | **Get** /um/users/{userId}/groups | Retrieve group resources by user ID
UserManagementApi | [**UmUsersOwnsGet**](docs/api/UserManagementApi.md#umusersownsget) | **Get** /um/users/{userId}/owns | Retrieve user resources by user ID
UserManagementApi | [**UmUsersPost**](docs/api/UserManagementApi.md#umuserspost) | **Post** /um/users | Create users
UserManagementApi | [**UmUsersPut**](docs/api/UserManagementApi.md#umusersput) | **Put** /um/users/{userId} | Modify users
UserS3KeysApi | [**UmUsersS3keysDelete**](docs/api/UserS3KeysApi.md#umuserss3keysdelete) | **Delete** /um/users/{userId}/s3keys/{keyId} | Delete S3 keys
UserS3KeysApi | [**UmUsersS3keysFindByKeyId**](docs/api/UserS3KeysApi.md#umuserss3keysfindbykeyid) | **Get** /um/users/{userId}/s3keys/{keyId} | Retrieve user S3 keys by key ID
UserS3KeysApi | [**UmUsersS3keysGet**](docs/api/UserS3KeysApi.md#umuserss3keysget) | **Get** /um/users/{userId}/s3keys | List user S3 keys
UserS3KeysApi | [**UmUsersS3keysPost**](docs/api/UserS3KeysApi.md#umuserss3keyspost) | **Post** /um/users/{userId}/s3keys | Create user S3 keys
UserS3KeysApi | [**UmUsersS3keysPut**](docs/api/UserS3KeysApi.md#umuserss3keysput) | **Put** /um/users/{userId}/s3keys/{keyId} | Modify S3 keys by key ID
UserS3KeysApi | [**UmUsersS3ssourlGet**](docs/api/UserS3KeysApi.md#umuserss3ssourlget) | **Get** /um/users/{userId}/s3ssourl | Retrieve S3 single sign-on URLs
VolumesApi | [**DatacentersVolumesCreateSnapshotPost**](docs/api/VolumesApi.md#datacentersvolumescreatesnapshotpost) | **Post** /datacenters/{datacenterId}/volumes/{volumeId}/create-snapshot | Create volume snapshots
VolumesApi | [**DatacentersVolumesDelete**](docs/api/VolumesApi.md#datacentersvolumesdelete) | **Delete** /datacenters/{datacenterId}/volumes/{volumeId} | Delete volumes
VolumesApi | [**DatacentersVolumesFindById**](docs/api/VolumesApi.md#datacentersvolumesfindbyid) | **Get** /datacenters/{datacenterId}/volumes/{volumeId} | Retrieve volumes
VolumesApi | [**DatacentersVolumesGet**](docs/api/VolumesApi.md#datacentersvolumesget) | **Get** /datacenters/{datacenterId}/volumes | List volumes
VolumesApi | [**DatacentersVolumesPatch**](docs/api/VolumesApi.md#datacentersvolumespatch) | **Patch** /datacenters/{datacenterId}/volumes/{volumeId} | Partially modify volumes
VolumesApi | [**DatacentersVolumesPost**](docs/api/VolumesApi.md#datacentersvolumespost) | **Post** /datacenters/{datacenterId}/volumes | Create volumes
VolumesApi | [**DatacentersVolumesPut**](docs/api/VolumesApi.md#datacentersvolumesput) | **Put** /datacenters/{datacenterId}/volumes/{volumeId} | Modify volumes
VolumesApi | [**DatacentersVolumesRestoreSnapshotPost**](docs/api/VolumesApi.md#datacentersvolumesrestoresnapshotpost) | **Post** /datacenters/{datacenterId}/volumes/{volumeId}/restore-snapshot | Restore volume snapshots

</details>

## Documentation For Models

All URIs are relative to *https://api.ionos.com/cloudapi/v6*
<details >
<summary title="Click to toggle">API models list</summary>

 - [ApplicationLoadBalancer](docs/models/ApplicationLoadBalancer)
 - [ApplicationLoadBalancerEntities](docs/models/ApplicationLoadBalancerEntities)
 - [ApplicationLoadBalancerForwardingRule](docs/models/ApplicationLoadBalancerForwardingRule)
 - [ApplicationLoadBalancerForwardingRuleProperties](docs/models/ApplicationLoadBalancerForwardingRuleProperties)
 - [ApplicationLoadBalancerForwardingRulePut](docs/models/ApplicationLoadBalancerForwardingRulePut)
 - [ApplicationLoadBalancerForwardingRules](docs/models/ApplicationLoadBalancerForwardingRules)
 - [ApplicationLoadBalancerHttpRule](docs/models/ApplicationLoadBalancerHttpRule)
 - [ApplicationLoadBalancerHttpRuleCondition](docs/models/ApplicationLoadBalancerHttpRuleCondition)
 - [ApplicationLoadBalancerProperties](docs/models/ApplicationLoadBalancerProperties)
 - [ApplicationLoadBalancerPut](docs/models/ApplicationLoadBalancerPut)
 - [ApplicationLoadBalancers](docs/models/ApplicationLoadBalancers)
 - [AttachedVolumes](docs/models/AttachedVolumes)
 - [BackupUnit](docs/models/BackupUnit)
 - [BackupUnitProperties](docs/models/BackupUnitProperties)
 - [BackupUnitSSO](docs/models/BackupUnitSSO)
 - [BackupUnits](docs/models/BackupUnits)
 - [BalancedNics](docs/models/BalancedNics)
 - [Cdroms](docs/models/Cdroms)
 - [ConnectableDatacenter](docs/models/ConnectableDatacenter)
 - [Contract](docs/models/Contract)
 - [ContractProperties](docs/models/ContractProperties)
 - [Contracts](docs/models/Contracts)
 - [CpuArchitectureProperties](docs/models/CpuArchitectureProperties)
 - [DataCenterEntities](docs/models/DataCenterEntities)
 - [Datacenter](docs/models/Datacenter)
 - [DatacenterElementMetadata](docs/models/DatacenterElementMetadata)
 - [DatacenterProperties](docs/models/DatacenterProperties)
 - [Datacenters](docs/models/Datacenters)
 - [Error](docs/models/Error)
 - [ErrorMessage](docs/models/ErrorMessage)
 - [FirewallRule](docs/models/FirewallRule)
 - [FirewallRules](docs/models/FirewallRules)
 - [FirewallruleProperties](docs/models/FirewallruleProperties)
 - [FlowLog](docs/models/FlowLog)
 - [FlowLogProperties](docs/models/FlowLogProperties)
 - [FlowLogPut](docs/models/FlowLogPut)
 - [FlowLogs](docs/models/FlowLogs)
 - [Group](docs/models/Group)
 - [GroupEntities](docs/models/GroupEntities)
 - [GroupMembers](docs/models/GroupMembers)
 - [GroupProperties](docs/models/GroupProperties)
 - [GroupShare](docs/models/GroupShare)
 - [GroupShareProperties](docs/models/GroupShareProperties)
 - [GroupShares](docs/models/GroupShares)
 - [GroupUsers](docs/models/GroupUsers)
 - [Groups](docs/models/Groups)
 - [IPFailover](docs/models/IPFailover)
 - [Image](docs/models/Image)
 - [ImageProperties](docs/models/ImageProperties)
 - [Images](docs/models/Images)
 - [Info](docs/models/Info)
 - [IpBlock](docs/models/IpBlock)
 - [IpBlockProperties](docs/models/IpBlockProperties)
 - [IpBlocks](docs/models/IpBlocks)
 - [IpConsumer](docs/models/IpConsumer)
 - [KubernetesAutoScaling](docs/models/KubernetesAutoScaling)
 - [KubernetesCluster](docs/models/KubernetesCluster)
 - [KubernetesClusterEntities](docs/models/KubernetesClusterEntities)
 - [KubernetesClusterForPost](docs/models/KubernetesClusterForPost)
 - [KubernetesClusterForPut](docs/models/KubernetesClusterForPut)
 - [KubernetesClusterProperties](docs/models/KubernetesClusterProperties)
 - [KubernetesClusterPropertiesForPost](docs/models/KubernetesClusterPropertiesForPost)
 - [KubernetesClusterPropertiesForPut](docs/models/KubernetesClusterPropertiesForPut)
 - [KubernetesClusters](docs/models/KubernetesClusters)
 - [KubernetesMaintenanceWindow](docs/models/KubernetesMaintenanceWindow)
 - [KubernetesNode](docs/models/KubernetesNode)
 - [KubernetesNodeMetadata](docs/models/KubernetesNodeMetadata)
 - [KubernetesNodePool](docs/models/KubernetesNodePool)
 - [KubernetesNodePoolForPost](docs/models/KubernetesNodePoolForPost)
 - [KubernetesNodePoolForPut](docs/models/KubernetesNodePoolForPut)
 - [KubernetesNodePoolLan](docs/models/KubernetesNodePoolLan)
 - [KubernetesNodePoolLanRoutes](docs/models/KubernetesNodePoolLanRoutes)
 - [KubernetesNodePoolProperties](docs/models/KubernetesNodePoolProperties)
 - [KubernetesNodePoolPropertiesForPost](docs/models/KubernetesNodePoolPropertiesForPost)
 - [KubernetesNodePoolPropertiesForPut](docs/models/KubernetesNodePoolPropertiesForPut)
 - [KubernetesNodePools](docs/models/KubernetesNodePools)
 - [KubernetesNodeProperties](docs/models/KubernetesNodeProperties)
 - [KubernetesNodes](docs/models/KubernetesNodes)
 - [Label](docs/models/Label)
 - [LabelProperties](docs/models/LabelProperties)
 - [LabelResource](docs/models/LabelResource)
 - [LabelResourceProperties](docs/models/LabelResourceProperties)
 - [LabelResources](docs/models/LabelResources)
 - [Labels](docs/models/Labels)
 - [Lan](docs/models/Lan)
 - [LanEntities](docs/models/LanEntities)
 - [LanNics](docs/models/LanNics)
 - [LanPost](docs/models/LanPost)
 - [LanProperties](docs/models/LanProperties)
 - [LanPropertiesPost](docs/models/LanPropertiesPost)
 - [Lans](docs/models/Lans)
 - [Loadbalancer](docs/models/Loadbalancer)
 - [LoadbalancerEntities](docs/models/LoadbalancerEntities)
 - [LoadbalancerProperties](docs/models/LoadbalancerProperties)
 - [Loadbalancers](docs/models/Loadbalancers)
 - [Location](docs/models/Location)
 - [LocationProperties](docs/models/LocationProperties)
 - [Locations](docs/models/Locations)
 - [NatGateway](docs/models/NatGateway)
 - [NatGatewayEntities](docs/models/NatGatewayEntities)
 - [NatGatewayLanProperties](docs/models/NatGatewayLanProperties)
 - [NatGatewayProperties](docs/models/NatGatewayProperties)
 - [NatGatewayPut](docs/models/NatGatewayPut)
 - [NatGatewayRule](docs/models/NatGatewayRule)
 - [NatGatewayRuleProperties](docs/models/NatGatewayRuleProperties)
 - [NatGatewayRuleProtocol](docs/models/NatGatewayRuleProtocol)
 - [NatGatewayRulePut](docs/models/NatGatewayRulePut)
 - [NatGatewayRuleType](docs/models/NatGatewayRuleType)
 - [NatGatewayRules](docs/models/NatGatewayRules)
 - [NatGateways](docs/models/NatGateways)
 - [NetworkLoadBalancer](docs/models/NetworkLoadBalancer)
 - [NetworkLoadBalancerEntities](docs/models/NetworkLoadBalancerEntities)
 - [NetworkLoadBalancerForwardingRule](docs/models/NetworkLoadBalancerForwardingRule)
 - [NetworkLoadBalancerForwardingRuleHealthCheck](docs/models/NetworkLoadBalancerForwardingRuleHealthCheck)
 - [NetworkLoadBalancerForwardingRuleProperties](docs/models/NetworkLoadBalancerForwardingRuleProperties)
 - [NetworkLoadBalancerForwardingRulePut](docs/models/NetworkLoadBalancerForwardingRulePut)
 - [NetworkLoadBalancerForwardingRuleTarget](docs/models/NetworkLoadBalancerForwardingRuleTarget)
 - [NetworkLoadBalancerForwardingRuleTargetHealthCheck](docs/models/NetworkLoadBalancerForwardingRuleTargetHealthCheck)
 - [NetworkLoadBalancerForwardingRules](docs/models/NetworkLoadBalancerForwardingRules)
 - [NetworkLoadBalancerProperties](docs/models/NetworkLoadBalancerProperties)
 - [NetworkLoadBalancerPut](docs/models/NetworkLoadBalancerPut)
 - [NetworkLoadBalancers](docs/models/NetworkLoadBalancers)
 - [Nic](docs/models/Nic)
 - [NicEntities](docs/models/NicEntities)
 - [NicProperties](docs/models/NicProperties)
 - [NicPut](docs/models/NicPut)
 - [Nics](docs/models/Nics)
 - [NoStateMetaData](docs/models/NoStateMetaData)
 - [PaginationLinks](docs/models/PaginationLinks)
 - [Peer](docs/models/Peer)
 - [PrivateCrossConnect](docs/models/PrivateCrossConnect)
 - [PrivateCrossConnectProperties](docs/models/PrivateCrossConnectProperties)
 - [PrivateCrossConnects](docs/models/PrivateCrossConnects)
 - [RemoteConsoleUrl](docs/models/RemoteConsoleUrl)
 - [Request](docs/models/Request)
 - [RequestMetadata](docs/models/RequestMetadata)
 - [RequestProperties](docs/models/RequestProperties)
 - [RequestStatus](docs/models/RequestStatus)
 - [RequestStatusMetadata](docs/models/RequestStatusMetadata)
 - [RequestTarget](docs/models/RequestTarget)
 - [Requests](docs/models/Requests)
 - [Resource](docs/models/Resource)
 - [ResourceEntities](docs/models/ResourceEntities)
 - [ResourceGroups](docs/models/ResourceGroups)
 - [ResourceLimits](docs/models/ResourceLimits)
 - [ResourceProperties](docs/models/ResourceProperties)
 - [ResourceReference](docs/models/ResourceReference)
 - [Resources](docs/models/Resources)
 - [ResourcesUsers](docs/models/ResourcesUsers)
 - [S3Bucket](docs/models/S3Bucket)
 - [S3Key](docs/models/S3Key)
 - [S3KeyMetadata](docs/models/S3KeyMetadata)
 - [S3KeyProperties](docs/models/S3KeyProperties)
 - [S3Keys](docs/models/S3Keys)
 - [S3ObjectStorageSSO](docs/models/S3ObjectStorageSSO)
 - [Server](docs/models/Server)
 - [ServerEntities](docs/models/ServerEntities)
 - [ServerProperties](docs/models/ServerProperties)
 - [Servers](docs/models/Servers)
 - [Snapshot](docs/models/Snapshot)
 - [SnapshotProperties](docs/models/SnapshotProperties)
 - [Snapshots](docs/models/Snapshots)
 - [TargetGroup](docs/models/TargetGroup)
 - [TargetGroupHealthCheck](docs/models/TargetGroupHealthCheck)
 - [TargetGroupHttpHealthCheck](docs/models/TargetGroupHttpHealthCheck)
 - [TargetGroupProperties](docs/models/TargetGroupProperties)
 - [TargetGroupPut](docs/models/TargetGroupPut)
 - [TargetGroupTarget](docs/models/TargetGroupTarget)
 - [TargetGroups](docs/models/TargetGroups)
 - [TargetPortRange](docs/models/TargetPortRange)
 - [Template](docs/models/Template)
 - [TemplateProperties](docs/models/TemplateProperties)
 - [Templates](docs/models/Templates)
 - [Token](docs/models/Token)
 - [Type](docs/models/Type)
 - [User](docs/models/User)
 - [UserMetadata](docs/models/UserMetadata)
 - [UserPost](docs/models/UserPost)
 - [UserProperties](docs/models/UserProperties)
 - [UserPropertiesPost](docs/models/UserPropertiesPost)
 - [UserPropertiesPut](docs/models/UserPropertiesPut)
 - [UserPut](docs/models/UserPut)
 - [Users](docs/models/Users)
 - [UsersEntities](docs/models/UsersEntities)
 - [Volume](docs/models/Volume)
 - [VolumeProperties](docs/models/VolumeProperties)
 - [Volumes](docs/models/Volumes)


[[Back to API list]](#documentation-for-api-endpoints) [[Back to Model list]](#documentation-for-models)

</details>



## Documentation for Utility Methods

Due to the fact that model structure members are all pointers, this package contains
a number of utility functions to easily obtain pointers to values of basic types.
Each of these functions takes a value of the given basic type and returns a pointer to it:

* `PtrBool`
* `PtrInt`
* `PtrInt32`
* `PtrInt64`
* `PtrFloat`
* `PtrFloat32`
* `PtrFloat64`
* `PtrString`
* `PtrTime`