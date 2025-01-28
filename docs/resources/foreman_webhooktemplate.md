# foreman_webhooktemplate

Webhook templates allow to configure a payload to send via webhook.

## Example Usage

```terraform
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
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the webhook template.
- `template` - (Required) The content of the webhook template.
- `snippet` - (Optional) Specifies if webhook template is a snippet.
- `audit_comment` - (Optional) Comment for audits.
- `locked` - (Optional) Whether the template is locked for editing.
- `default` - (Optional) Whether the template is automatically added to new organizations and locations.
- `description` - (Optional) Webhook Template description.
- `location_ids` - (Optional) A list of location IDs where the discovered hosts will be assigned.
- `organization_ids` - (Optional) A list of organization IDs where the discovered hosts will be assigned.

## Attributes Reference

The following attributes are exported:

- `name` - The name of the webhook template.
- `template` - The content of the webhook template.
- `snippet` - Specifies if webhook template is a snippet.
- `audit_comment` - Comment for audits.
- `locked` - Whether the template is locked for editing.
- `default` - Whether the template is automatically added to new organizations and locations.
- `description` - Webhook Template description.
- `locations` - A list of locations where the discovered hosts will be assigned.
- `organizations` - A list of organizations where the discovered hosts will be assigned.
