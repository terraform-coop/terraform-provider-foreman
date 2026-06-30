//go:build integration

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// Provider name for acceptance tests
	providerConfig = `
provider "foreman" {
  server_hostname     = "localhost"
  client_username     = "admin"
  client_password     = "changeme"
  client_tls_insecure = true
}
`
)

// testAccProtoV6ProviderFactories returns a map of provider factories for
// acceptance tests using protocol v6.
func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"foreman": providerserver.NewProtocol6WithError(New("test")()),
	}
}
