provider "foreman" {
  server_hostname = "192.168.1.118"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = "${var.client_username}"
  client_password = "${var.client_password}"
}

resource "foreman_webhooktemplate" "example_webhooktemplate_01" {
  name             = "Example Webhook Template"
  template         = "<%=\npayload({\n  id: @object.id\n})\n-%>\n"
  snippet          = false
  audit_comment    = ""
  locked           = true
  default          = true
  description      = "This template is used to define default content of payload for a webhook."
  location_ids     = [2]
  organization_ids = [1]
}
