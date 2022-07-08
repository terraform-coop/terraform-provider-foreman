package foreman

import (
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// Config struct defines the necessary information needed to configure the
// terraform provider for communication with the Foreman API.
type Config struct {
	// Server definition.  The server's URL will be the 'base URL' the REST
	// client uses for issuing API requests to the API gateway.
	Server api.Server
	// Whether or not to verify the server's certificate/hostname.  This flag
	// is passed to the TLS config when initializing the REST client for API
	// communication.
	//
	// See 'pkg/crypto/tls/#Config.InsecureSkipVerify' for more information.
	ClientTLSInsecure bool
	// Whether or not the client should try to authenticate to foreman
	// through the HTTP negotiate mechanism.
	NegotiateAuthEnabled bool
	// Set of credentials needed to authenticate against Foreman
	ClientCredentials api.ClientCredentials
	// Location for all API Calls
	LocationID int
	// Organization for all API Calls
	OrganizationID int
}

// Client creates a client reference for the Foreman REST API given the
// provider configuration options.  After creating a client reference, the
// client is then authenticated with the credentials supplied to the provider
// configuration.
func (c *Config) Client() (*api.Client, diag.Diagnostics) {
	log.Tracef("config.go#Client")

	client := api.NewClient(
		c.Server,
		c.ClientCredentials,
		api.ClientConfig{
			TLSInsecureEnabled:   c.ClientTLSInsecure,
			LocationID:           c.LocationID,
			OrganizationID:       c.OrganizationID,
			NegotiateAuthEnabled: c.NegotiateAuthEnabled,
		},
	)

	log.Debugf("Rest Client configured")

	return client, diag.Diagnostics{}
}
