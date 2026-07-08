data "foreman_smart_class_parameter" "ntp_servers" {
  name = "ntp_servers"
}

resource "foreman_override_value" "webservers_ntp" {
  parent_id = data.foreman_smart_class_parameter.ntp_servers.id

  match = "hostgroup=Webservers"
  value = "[\"ntp1.example.com\", \"ntp2.example.com\"]"
  omit  = false
}
