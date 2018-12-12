variable "client_username" {}
variable "client_password" {}

provider "foreman" {
  server_hostname = "192.168.1.118"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = "${var.client_username}"
  client_password = "${var.client_password}"
}

data "foreman_hostgroup" "DC1" {
	title = "DC1"
}

/*
resource "foreman_hostgroup" "terraformtest" {
	name = "terraformtest hostgroup"
	parent_id = "${data.foreman_hostgroup.DC1.id}"

	# TFTest2
	#architecture_id = 13

	# aaanderson_test
	#environment_id = 6196
}
*/
