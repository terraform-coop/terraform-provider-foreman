data "foreman_domain" "dc1" {
  name = "dc1.example.com"
}

data "foreman_smartproxy" "dc1_proxy" {
  name = "proxy.dc1.example.com"
}

resource "foreman_subnet" "dc1_vlan24" {
  name    = "DC1_VLAN24"
  network = "10.228.159.0"
  mask    = "255.255.255.0"

  gateway     = "10.228.159.1"
  dns_primary = "10.228.159.5"

  dhcp_id    = data.foreman_smartproxy.dc1_proxy.id
  dns_id     = data.foreman_smartproxy.dc1_proxy.id
  domain_ids = [data.foreman_domain.dc1.id]

  subnet_parameters_attributes = {
    role = "vlan24"
  }
}
