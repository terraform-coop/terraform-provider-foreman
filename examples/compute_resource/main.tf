variable "client_username" {}
variable "client_password" {}

provider "foreman" {
  server_hostname = "192.168.1.118"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = "${var.client_username}"
  client_password = "${var.client_password}"
}

data "foreman_computeresource" "kvm" {
	title = "kvm-lab"
}

/*
TODO: Add one
resource "foreman_compute_resource" "terraformtest" {
	name = "terraformtest compute_resource"
}
*/
