# foreman_webhook

Webhooks provide integration to 3rd parties via web services with configurable payload (via foreman_webhooktemplate).

## Example Usage

```terraform
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
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the webhook.
- `target_url` - (Required) The URL for the webhook.
- `http_method` - (Optional) Must be one of: POST, GET, PUT, DELETE, PATCH.
- `http_content_type` - (Optional) Content Type Header.
- `http_headers` - (Optional) Additional Headers to send. Must be a json object.
- `event` - (Required) An string specifying the event type for which the webhook is triggered.
- `enabled` - (Optional) A boolean value indicating whether the webhook is enabled. When set to `true`, the rule is active and will be evaluated.
- `verify_ssl` - (Optional) A boolean value indicating if SSL certs should be verified.
- `ssl_ca_certs` - (Optional) X509 Certification Authorities for verification concatenated in PEM format.
- `proxy_authorization` - (Optional) Indicating whether to authorize with Foreman client certificate and validate smart-proxy CA from Settings
- `user` - (Optional) User name for basic authentication.
- `password` - (Optional) Password for basic authentication.
- `webhook_template_id` - (Optional) ID of the webhook template containing the payload.

## Attributes Reference

The following attributes are exported:

- `name` - The name of the webhook.
- `target_url` - The URL for the webhook.
- `http_method` - Must be one of: POST, GET, PUT, DELETE, PATCH.
- `http_content_type` - Content Type Header.
- `http_headers` - Additional Headers to send. Must be a json object.
- `event` - An string specifying the event type for which the webhook is triggered.
- `enabled` - A boolean value indicating whether the webhook is enabled. When set to `true`, the rule is active and will be evaluated.
- `verify_ssl` - A boolean value indicating if SSL certs should be verified.
- `ssl_ca_certs` - X509 Certification Authorities for verification concatenated in PEM format.
- `proxy_authorization` - Indicating whether to authorize with Foreman client certificate and validate smart-proxy CA from Settings.
- `user` - User name for basic authentication.
- `password_set` - If a password is set for basic authentication.
- `webhook_template` - Webhook template containing the payload.
