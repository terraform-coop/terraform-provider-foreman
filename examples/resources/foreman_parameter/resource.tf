resource "foreman_parameter" "ntp_server" {
  # exactly one of host_id, hostgroup_id, domain_id, operatingsystem_id,
  # subnet_id, location_id, organization_id must be set
  host_id        = foreman_host.example.id
  name           = "ntp_server"
  value          = "pool.ntp.org"
  parameter_type = "string"

  hidden_value = false
}
