resource "foreman_discovery_rule" "example" {
  name         = "example-rule-01"
  search       = "facts.bios_vendor = HPE"
  hostgroup_id = 5
  hostname     = "host-<%= @host.mac.delete(':') %>"
  priority     = 100
  hosts_limit  = 0
  enabled      = true
}
