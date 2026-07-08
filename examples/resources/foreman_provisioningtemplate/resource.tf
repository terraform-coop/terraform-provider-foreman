data "foreman_templatekind" "provision" {
  name = "provision"
}

data "foreman_operatingsystem" "centos7" {
  name = "CentOS"
}

resource "foreman_provisioningtemplate" "centos_kickstart" {
  name     = "CentOS 7 Kickstart"
  template = <<-EOT
    install
    text
    reboot
  EOT

  template_kind_id    = data.foreman_templatekind.provision.id
  operatingsystem_ids = [data.foreman_operatingsystem.centos7.id]

  description = "Default kickstart template for CentOS 7 hosts"
  locked      = false
  snippet     = false
}
