variable "client_username" {}
variable "client_password" {}

provider "foreman" {
  server_hostname = "192.168.1.118"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = "${var.client_username}"
  client_password = "${var.client_password}"
}

# Read the data resource (it's read-only at the moment)
data "foreman_setting" "append_domain" {
	name = "append_domain_name_for_hosts"
}

# Then use e.g. data.foreman_setting.append_domain:
output "setting_append_domain" {
    value = data.foreman_setting.append_domain
}

# Result:
# setting_append_domain = {
#   __meta__      = null
#   category_name = "General"
#   default       = null
#   description   = "Foreman will append domain names when new hosts are provisioned"
#   id            = "append_domain_name_for_hosts"
#   name          = "append_domain_name_for_hosts"
#   readonly      = false
#   settings_type = "boolean"
#   value         = "true"
# }