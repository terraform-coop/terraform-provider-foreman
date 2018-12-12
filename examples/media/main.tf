variable "client_username" {}
variable "client_password" {}

provider "foreman" {
  server_hostname = "192.168.1.118"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = "${var.client_username}"
  client_password = "${var.client_password}"
}

data "foreman_media" "centos_mirror" {
	name = "CentOS mirror"
}

resource "foreman_media" "media_terraformtest" {
	name = "CentOS Mirror Georgia Tech"
	path = "http://www.gtlib.gatech.edu/pub/centos/$major.$minor/os/$arch"
	os_family = "Redhat"
	operatingsystem_ids = [1,2,3]
}
