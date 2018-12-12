variable "client_username" {}
variable "client_password" {}

provider "foreman" {
  server_hostname = "192.168.1.118"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = "${var.client_username}"
  client_password = "${var.client_password}"
}

data "foreman_smartproxy" "DC1DNSPrimary" {
	name = "dns.dc1.company.com"
}

resource "foreman_smartproxy" "terraformtest" {
	name = "terraformtestproxy.dc1.company.com"
	url  = "https://terraformtestproxy.dc1.company.com"
}
