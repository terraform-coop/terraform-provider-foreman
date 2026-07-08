data "foreman_operatingsystem" "centos" {
  name = "CentOS 7.4"
}

data "foreman_provisioningtemplate" "kickstart" {
  name = "Kickstart default"
}

data "foreman_templatekind" "provision" {
  name = "provision"
}

# Associates a provisioning template as the default template of a given
# template kind (e.g. "provision", "PXELinux", "finish") for an operating
# system.
resource "foreman_defaulttemplate" "kickstart_default" {
  parent_id                = data.foreman_operatingsystem.centos.id
  provisioning_template_id = data.foreman_provisioningtemplate.kickstart.id
  template_kind_id         = data.foreman_templatekind.provision.id
}
