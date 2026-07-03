package main

//go:generate go run ./tools/gen/client -input apidoc/v2.json -output ./generated/ -provider ./internal/provider/ -overrides ./tools/gen/overrides.yaml

import (
	"context"
	"log"

	"github.com/terraform-coop/terraform-provider-foreman/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var version = "dev"

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/terraform-coop/foreman",
	}
	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err)
	}
}
