terraform {
  required_providers {
    foreman = {
      source = "HanseMerkur/foreman"
    }
  }
}

variable "client_username" {}

provider "foreman" {
  server_hostname = "localhost"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = var.client_username
  client_password = var.client_password
}

resource "foreman_smartproxy" "main" {
  name = "local"
  url  = "https://my-foreman-server:8443"
}
