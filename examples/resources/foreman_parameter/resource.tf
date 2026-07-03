resource "foreman_parameter" "ntp_server" {
  name           = "ntp_server"
  value          = "pool.ntp.org"
  parameter_type = "string"

  hidden_value = false
}
