variable "client_username" {}
variable "client_password" {}

provider "foreman" {
  server_hostname = "192.168.1.118"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = "${var.client_username}"
  client_password = "${var.client_password}"
}

data "foreman_model" "poweredgem520" {
	name = "PowerEdge M520"
}

resource "foreman_model" "terraformtest" {
	"name" = "Terraform Test"
	"info" = "Testing hardware model creation with Terraform"
	"vendor_class" = "Enterprise"
}
