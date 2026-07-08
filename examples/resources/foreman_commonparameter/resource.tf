resource "foreman_commonparameter" "puppet_server" {
  name           = "puppet_server"
  parameter_type = "string"
  value          = "puppet.example.com"
  hidden_value   = false
}
