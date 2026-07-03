variable "webhook_password" {
  type      = string
  sensitive = true
}

data "foreman_webhooktemplate" "payload" {
  name = "Example Webhook Template"
}

resource "foreman_webhook" "build_notify" {
  name        = "Build entered notification"
  target_url  = "https://hooks.example.com/foreman"
  http_method = "POST"

  http_content_type   = "application/json"
  event               = "build_entered.event.foreman"
  enabled             = true
  verify_ssl          = true
  webhook_template_id = data.foreman_webhooktemplate.payload.id
  user                = "svc-foreman"
  password            = var.webhook_password
}
