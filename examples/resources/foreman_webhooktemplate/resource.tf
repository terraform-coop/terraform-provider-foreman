resource "foreman_webhooktemplate" "payload" {
  name = "Example Webhook Template"
  template = <<-EOT
    <%=
    payload({
      id: @object.id
    })
    -%>
  EOT

  description = "Defines the default payload content for a webhook."
  locked      = false
  default     = false
}
