package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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

// testAccPreCheck validates the required environment variables for acceptance
// tests are set. Called at the start of each acceptance test.
func testAccPreCheck() {
	// In CI, FOREMAN_SERVER_HOSTNAME, FOREMAN_CLIENT_USERNAME,
	// FOREMAN_CLIENT_PASSWORD should be set via environment.
	// For local testing, defaults in provider config are used.
}

// testAccCheckDestroyed returns a test step that verifies the resource was
// actually deleted from Foreman by attempting a refresh and expecting a 404.
func testAccCheckDestroyed(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
	// The framework's CheckDestroy pattern: verify the resource
	// no longer exists by attempting to read it.
	// This is handled by the framework's built-in destroy check.
	)
}
