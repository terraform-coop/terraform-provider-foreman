
# foreman_provisioningtemplate


Provisioning templates are scripts used to describe how to bootstrap and install the operating system on the host.


## Example Usage

```
# Autogenerated example with required keys
resource "foreman_provisioningtemplate" "example" {
  name = "AutoYaST default"
  template = "void"
}
```


## Argument Reference

The following arguments are supported:

- `audit_comment` - (Optional) Notes and comments for auditing purposes.
- `description` - (Optional) A description of the provisioning template.
- `locked` - (Optional) Whether or not the template is locked for editing.
- `name` - (Required) Name of the provisioning template.
- `operatingsystem_ids` - (Optional) IDs of the operating systems associated with this provisioning template.
- `snippet` - (Optional) Whether or not the provisioning template is a snippet be used by other templates.
- `template` - (Required) The markup and code of the provisioning template.
- `template_combinations_attributes` - (Optional) How templates are determined:

When editing a template, you must assign a list of operating systems which this template can be used with.  Optionally, you can restrict a template to a list of host groups and/or environments.

When a host requests a template, Foreman will select the best match from the available templates of that type in the following order:

  1. host group and environment
  2. host group only
  3. environment only
  4. operating system default

Template combinations attributes contains an array of hostgroup IDs and environment ID combinations so they can be used in the provisioning template selection described above.
- `template_kind_id` - (Optional) ID of the template kind which categorizes the provisioning template. Optional for snippets, otherwise required.


## Attributes Reference

The following attributes are exported:

- `audit_comment` - Notes and comments for auditing purposes.
- `description` - A description of the provisioning template.
- `locked` - Whether or not the template is locked for editing.
- `name` - Name of the provisioning template.
- `operatingsystem_ids` - IDs of the operating systems associated with this provisioning template.
- `snippet` - Whether or not the provisioning template is a snippet be used by other templates.
- `template` - The markup and code of the provisioning template.
- `template_combinations_attributes` - How templates are determined:

When editing a template, you must assign a list of operating systems which this template can be used with.  Optionally, you can restrict a template to a list of host groups and/or environments.

When a host requests a template, Foreman will select the best match from the available templates of that type in the following order:

  1. host group and environment
  2. host group only
  3. environment only
  4. operating system default

Template combinations attributes contains an array of hostgroup IDs and environment ID combinations so they can be used in the provisioning template selection described above.
- `template_kind_id` - ID of the template kind which categorizes the provisioning template. Optional for snippets, otherwise required.

