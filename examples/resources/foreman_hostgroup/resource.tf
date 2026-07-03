data "foreman_hostgroup" "datacenter" {
  name = "DC1"
}

data "foreman_domain" "dev" {
  name = "dev.example.com"
}

data "foreman_operatingsystem" "centos" {
  name = "CentOS 7.4"
}

resource "foreman_hostgroup" "app" {
  name        = "app"
  description = "Application servers"
  parent_id   = data.foreman_hostgroup.datacenter.id

  domain_id          = data.foreman_domain.dev.id
  operatingsystem_id = data.foreman_operatingsystem.centos.id

  root_pass = var.hostgroup_root_pass

  parameters = {
    role = "webserver"
  }
}
