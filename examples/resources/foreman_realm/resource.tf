data "foreman_smartproxy" "idm" {
  name = "idm.example.com"
}

resource "foreman_realm" "example_com" {
  name           = "EXAMPLE.COM"
  realm_type     = "FreeIPA"
  realm_proxy_id = data.foreman_smartproxy.idm.id
}
