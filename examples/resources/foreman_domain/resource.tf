resource "foreman_domain" "example" {
  name     = "dev.example.com"
  fullname = "Example Development Domain"
  dns_id   = 1

  domain_parameters_attributes = {
    mtu = "1500"
  }
}
