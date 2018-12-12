variable "client_username" {}
variable "client_password" {}

provider "foreman" {
  server_hostname = "192.168.1.118"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = "${var.client_username}"
  client_password = "${var.client_password}"
}

data "foreman_templatekind" "PXELinux" {
  name = "PXELinux"
}

/*
data "foreman_provisioningtemplate" "PXELinux" {
	name = "Kickstart default PXELinux"
}

resource "foreman_provisioningtemplate" "TerraformTest" {
	name = "Terraform Test Template"
	template = <<EOF
line 1
line 2

line 4
EOF
	template_kind_id = "${data.foreman_templatekind.PXELinux.id}"
}

resource "foreman_provisioningtemplate" "TerraformTestSnippet" {
	name     = "Terraform Test Snippet"
	template = "void"
	snippet  = "true"
}
*/

resource "foreman_provisioningtemplate" "TerraformTestCombo" {
  name             = "Terraform Test Template with Combination Attributes"
  template         = "test template - do not use"
  snippet          = "false"
  template_kind_id = "${data.foreman_templatekind.PXELinux.id}"

  template_combinations_attributes = [
    {
      # DC1/terraformtest
      hostgroup_id = 163

      # staging_web
      environment_id = 6196
    },
    {
      # DC1/storefront
      hostgroup_id = 151

      # staging_web
      environment_id = 6196
    },
  ]
}

resource "foreman_provisioningtemplate" "TerraformTestNoCombo" {
  name = "Terraform Test Template without Template Combinations"

  template = <<EOF
just checking handling
multiple
	lines with whitespace

EOF

  snippet          = "false"
  template_kind_id = "${data.foreman_templatekind.PXELinux.id}"

  //operatingsystem_ids = [1,2,3]
}
