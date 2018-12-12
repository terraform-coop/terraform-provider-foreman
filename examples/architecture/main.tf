variable "client_username" {}
variable "client_password" {}

provider "foreman" {
  server_hostname = "192.168.1.118"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = "${var.client_username}"
  client_password = "${var.client_password}"
}

# -----------------------------------------------------------------------------

/*
data "foreman_architecture" "arch_x64" {
	name = "x64"
}
*/

data "foreman_architecture" "arch_x86_64" {
	name = "x86_64"
}

data "foreman_architecture" "arch_i386" {
	name = "i386"
}

/*
data "foreman_architecture" "arch_amd64" {
	name = "amd64"
}
*/

# -----------------------------------------------------------------------------

#resource "foreman_architecture" "arch_TFTesting" {
#	name = "TerraformTestArch"
#}

#resource "foreman_architecture" "arch_TFTest" {
#	name = "TerraformTestArch2"
#	operatingsystem_ids = [1]
#}

#resource "foreman_architecture" "arch_TFTest2" {
#	name = "TFTest2"
#	operatingsystem_ids = [3,1,2]
#}

#resource "foreman_architecture" "arch_Test" {
#	name = "test.arch"
#}
