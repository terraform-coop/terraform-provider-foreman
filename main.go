package main

import (
	"github.com/terraform-coop/terraform-provider-foreman/foreman"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	// opts contains the configurations to serve the Foreman plugin.
	opts := plugin.ServeOpts{
		ProviderFunc: foreman.Provider,
	}
	// Serves the foreman plugin in the defined configurations.
	plugin.Serve(&opts)
}
