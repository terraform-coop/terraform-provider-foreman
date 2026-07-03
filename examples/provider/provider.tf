terraform {
  required_providers {
    foreman = {
      source = "terraform-coop/foreman"
    }
  }
}

provider "foreman" {
  server_hostname = "foreman.example.com"
  server_protocol = "https"

  # Credentials can also be supplied via FOREMAN_CLIENT_USERNAME /
  # FOREMAN_CLIENT_PASSWORD environment variables instead of hardcoding them here.
  client_username = "admin"
  client_password = var.foreman_password

  # Only needed when the Foreman server uses a self-signed certificate.
  client_tls_insecure = false

  # Scope all API calls to a specific organization/location. Leave unset (or 0)
  # to disable Foreman's organizations/locations feature.
  organization_id = 1
  location_id     = 1
}
