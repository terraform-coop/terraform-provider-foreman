provider "foreman" {
  server_hostname = "192.168.1.118"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = "${var.client_username}"
  client_password = "${var.client_password}"
}

resource "foreman_discovery_rule" "example_rule_01" {
  name              = "example-rule-01"
  search            = "facts.bios_vendor = HPE"
  hostgroup_ids     = 5
  hostname          = "<%= @host.facts['nmprimary_dhcp4_option_host_name'] %>"
  max_count         = 0
  priority          = 100
  enabled           = true
  location_ids      = [1]
  organization_ids  = [1]
}
