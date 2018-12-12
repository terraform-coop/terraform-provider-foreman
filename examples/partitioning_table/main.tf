variable "client_username" {}
variable "client_password" {}

provider "foreman" {
  server_hostname = "192.168.1.118"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = "${var.client_username}"
  client_password = "${var.client_password}"
}

data "foreman_partitiontable" "Default_Centos" {
	name = "Default CentOS"
}

//resource "foreman_partitiontable" "TerraformTestSnippet" {
//	name = "Terraform Test Snippet Partition Table"
//	layout = "void"
//	snippet = "true"
//}

resource "foreman_partitiontable" "TerraformTest" {
	name = "Terraform Test Partition Table"
	layout = "void"
	snippet = "false"
}
