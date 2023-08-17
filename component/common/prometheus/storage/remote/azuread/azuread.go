package azuread

import (
	internal "github.com/prometheus/prometheus/storage/remote/azuread"
)

// AzureADConfig is used to store the config values.
type AzureADConfig struct { // nolint:revive
	// ManagedIdentity is the managed identity that is being used to authenticate.
	ManagedIdentity *ManagedIdentityConfig `river:"managed_identity,block,optional"`

	// Cloud is the Azure cloud in which the service is running. Example: AzurePublic/AzureGovernment/AzureChina.
	Cloud string `river:"cloud,string,optional"`
}

func (c *AzureADConfig) ToInternal() internal.AzureADConfig {
	managedIdentityConfig := c.ManagedIdentity.ToInternal()
	return internal.AzureADConfig{
		ManagedIdentity: &managedIdentityConfig,
		Cloud:           c.Cloud,
	}
}

// ManagedIdentityConfig is used to store managed identity config values
type ManagedIdentityConfig struct {
	// ClientID is the clientId of the managed identity that is being used to authenticate.
	ClientID string `river:"client_id,string,optional"`
}

func (c *ManagedIdentityConfig) ToInternal() internal.ManagedIdentityConfig {
	return internal.ManagedIdentityConfig{
		ClientID: c.ClientID,
	}
}
