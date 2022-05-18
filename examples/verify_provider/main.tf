terraform {
  required_providers {
    foreman = {
      source = "HanseMerkur/foreman"
    }
  }
}


provider "foreman" {
  server_hostname = "localhost"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = "admin"
  client_password = "null"
}

resource "foreman_smartproxy" "main" {
  name = "local"
  url  = "https://my-foreman-server:8443"
}
