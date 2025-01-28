provider "foreman" {
  server_hostname = "192.168.1.118"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = "${var.client_username}"
  client_password = "${var.client_password}"
}

resource "foreman_webhook" "example_webhook_01" {
  name                = "Example Webhook"
  target_url          = "https://example-webhook.local:8080"
  http_method         = "GET"
  http_content_type   = "application/json"
  http_headers        = "{\n\"X-Shellhook-Arg-1\":\"value\"\n}"
  event               = "build_entered.event.foreman"
  enabled             = true
  verify_ssl          = false
  ssl_ca_certs        = ""
  proxy_authorization = false
  user                = "foo"
  password            = "bar"
  webhook_template_id = 1
}
