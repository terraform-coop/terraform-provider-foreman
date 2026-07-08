//go:build integration

package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccEndToEnd applies a complete, realistically-wired Foreman setup
// through actual Terraform - the cross-resource references (every _id/_ids
// below) are exactly what the generated per-resource basic test cannot
// exercise. The framework plans again after every apply step and fails on
// any non-empty plan, so this also proves the whole graph is
// perpetual-diff-free: every read/write asymmetry the API has must be
// absorbed somewhere below for this test to pass.
func TestAccEndToEnd(t *testing.T) {
	t.Parallel()

	fullConfig := providerConfig + `
data "foreman_templatekind" "provision" {
  name = "provision"
}

resource "foreman_architecture" "e2e" {
  name = "tf-e2e-arch"
}

resource "foreman_domain" "e2e" {
  name     = "tf-e2e.example.test"
  fullname = "TF E2E domain"
}

resource "foreman_subnet" "e2e" {
  name        = "tf-e2e-subnet"
  network     = "203.0.113.0" # TEST-NET-3
  mask        = "255.255.255.0"
  gateway     = "203.0.113.1"
  dns_primary = "203.0.113.2"
  from        = "203.0.113.10"
  to          = "203.0.113.100"
  boot_mode   = "Static"
  ipam        = "Internal DB"
  domain_ids  = [foreman_domain.e2e.id]
}

resource "foreman_partitiontable" "e2e" {
  name      = "tf-e2e-ptable"
  os_family = "Redhat"
  layout    = "zerombr\nclearpart --all --initlabel\nautopart\n"
}

resource "foreman_media" "e2e" {
  name      = "tf-e2e-medium"
  path      = "http://mirror.example.test/tf/$version/$arch"
  os_family = "Redhat"
}

resource "foreman_provisioningtemplate" "e2e" {
  name             = "tf-e2e-template"
  template         = "#!/bin/bash\n# tf e2e provision template\necho provisioned\n"
  snippet          = false
  template_kind_id = data.foreman_templatekind.provision.id
}

resource "foreman_operatingsystem" "e2e" {
  name                      = "tfE2EOS"
  major                     = "9"
  minor                     = "3"
  family                    = "Redhat"
  architecture_ids          = [foreman_architecture.e2e.id]
  medium_ids                = [foreman_media.e2e.id]
  ptable_ids                = [foreman_partitiontable.e2e.id]
  provisioning_template_ids = [foreman_provisioningtemplate.e2e.id]
}

resource "foreman_hostgroup" "e2e" {
  name               = "tf-e2e-hg"
  architecture_id    = foreman_architecture.e2e.id
  domain_id          = foreman_domain.e2e.id
  subnet_id          = foreman_subnet.e2e.id
  operatingsystem_id = foreman_operatingsystem.e2e.id
  medium_id          = foreman_media.e2e.id
  ptable_id          = foreman_partitiontable.e2e.id
  root_pass          = "tf-e2e-rootpw-secret"

  parameters = {
    e2e_tier = "integration"
  }
}

resource "foreman_parameter" "domain" {
  domain_id      = foreman_domain.e2e.id
  name           = "tf_e2e_dns_search"
  value          = "tf-e2e.example.test"
  parameter_type = "string"
}

resource "foreman_host" "e2e" {
  # Foreman stores a host's name as its FQDN once a domain is attached
  # (here inherited from the hostgroup) - config must write it that way.
  name         = "tf-e2e-host.tf-e2e.example.test"
  hostgroup_id = foreman_hostgroup.e2e.id
  managed      = false
  build        = false
  comment      = "created by TF e2e acceptance test"

  parameters = {
    e2e_role = "worker"
  }

  interfaces_attributes = [
    {
      identifier = "eth0"
      mac        = "52:54:00:ee:e0:01"
      primary    = true
      managed    = false
    },
  ]
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				// Full graph applies cleanly and the follow-up plan is empty.
				Config: fullConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("foreman_host.e2e", "id"),
					resource.TestCheckResourceAttrPair("foreman_host.e2e", "hostgroup_id", "foreman_hostgroup.e2e", "id"),
					resource.TestCheckResourceAttrPair("foreman_hostgroup.e2e", "operatingsystem_id", "foreman_operatingsystem.e2e", "id"),
					resource.TestCheckResourceAttrPair("foreman_subnet.e2e", "domain_ids.0", "foreman_domain.e2e", "id"),
					resource.TestCheckResourceAttr("foreman_host.e2e", "parameters.e2e_role", "worker"),
					resource.TestCheckResourceAttr("foreman_hostgroup.e2e", "parameters.e2e_tier", "integration"),
					resource.TestCheckResourceAttrSet("data.foreman_templatekind.provision", "id"),
				),
			},
			{
				// In-place updates across the graph stay consistent.
				Config: updateConfig(fullConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("foreman_host.e2e", "comment", "updated by TF e2e acceptance test"),
					resource.TestCheckResourceAttr("foreman_host.e2e", "parameters.e2e_role", "controller"),
					resource.TestCheckResourceAttr("foreman_subnet.e2e", "gateway", "203.0.113.254"),
				),
			},
		},
	})
}

// updateConfig derives the second step's config: host comment + parameter
// value + subnet gateway change in place.
func updateConfig(cfg string) string {
	cfg = strings.Replace(cfg, `comment      = "created by TF e2e acceptance test"`, `comment      = "updated by TF e2e acceptance test"`, 1)
	cfg = strings.Replace(cfg, `e2e_role = "worker"`, `e2e_role = "controller"`, 1)
	cfg = strings.Replace(cfg, `gateway     = "203.0.113.1"`, `gateway     = "203.0.113.254"`, 1)
	return cfg
}
