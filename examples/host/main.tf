variable "client_username" {}
variable "client_password" {}

provider "foreman" {
  server_hostname = "192.168.1.118"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = "${var.client_username}"
  client_password = "${var.client_password}"
}

data "foreman_domain" "dev" {
  name = "dev.company.com"
}

data "foreman_environment" "production" {
  name = "production"
}

data "foreman_hostgroup" "DC1VM" {
  title = "DC1/VM"
}

data "foreman_operatingsystem" "Centos74" {
  title = "CentOS 7.4"
}

data "foreman_subnet" "app1" {
  name    = "10.228.170.0 app1"
  network = "10.228.170.0"
  mask    = "255.255.255.0"
}

resource "foreman_host" "TerraformTest" {
  name  = "foremanterraformtest.dev.company.com"
  ip    = "10.228.170.38"
  mac   = "C0:FF:EE:BA:BE:00"
  build = "true"

  domain_id          = "${data.foreman_domain.dev.id}"
  environment_id     = "${data.foreman_environment.production.id}"
  hostgroup_id       = "${data.foreman_hostgroup.DC1VM.id}"
  operatingsystem_id = "${data.foreman_operatingsystem.Centos74.id}"
  subnet_id          = "${data.foreman_subnet.app1.id}"
}
