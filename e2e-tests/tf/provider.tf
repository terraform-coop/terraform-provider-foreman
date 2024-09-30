terraform {
  required_providers {
    foreman = {
      source = "terraform-coop/foreman"
      version = "0.7.0"
    }
  }
}

provider "foreman" {
  client_username = "admin"
  client_password = "admine2e"
  client_tls_insecure = "true"

  server_hostname = "HOSTNAME OF YOUR E2E TEST INSTANCE"
  server_protocol = "https"
  location_id = 2
  organization_id = 1
}


