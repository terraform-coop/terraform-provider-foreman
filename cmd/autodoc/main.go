// Package main contains the main goroutine for the autodoc command-line
// application.  This application uses the autodoc engine to create mkdocs
// style documentation for the Terraform provider.
//
// For more information on the autodoc tool, its arguments, etc see:
// pkg/github.com/HanseMerkur/terraform-provider-utils/autodoc
package main

import (
	"fmt"
	"os"

	"github.com/terraform-coop/terraform-provider-foreman/foreman"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
)

func main() {
	// Use the provider function to get information on the provider's schema,
	// resources, and data sources. The Provider() function returns a
	// *terraform.ResourceProvider (interface) which will need to be type asserted
	// to a *schema.Provider (struct)
	provider := foreman.Provider()
	// Start the autodoc engine
	errors := autodoc.Document(provider)
	if len(errors) != 0 {
		for _, err := range errors {
			fmt.Println(err)
		}
		os.Exit(autodoc.ExitError)
	}
	os.Exit(autodoc.ExitSuccess)
}
