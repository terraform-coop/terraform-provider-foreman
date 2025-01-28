
# foreman_discovery_rules

Discovery rules in Foreman are used to automatically provision hosts based on predefined criteria.
These rules help streamline the process of adding new hosts to your infrastructure by automating the provisioning process based on specific conditions.

## Example Usage

```terraform
resource "foreman_discovery_rule" "example_rule_01" {
  name              = "Example Rule HPE servers"
  search            = "facts.bios_vendor = HPE"
  hostgroup_ids     = 3
  hostname          = "<%= @host.facts['nmprimary_dhcp4_option_host_name'] %>"
  max_count         = 0
  priority          = 100
  enabled           = true
  location_ids      = [2]
  organization_ids  = [1]
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the discovery rule.
- `search` - (Required) The search criteria used to match hosts.
- `priority` - (Required) The priority of the rule.
- `hostgroup_id` - (Optional) The ID of the host group to which the discovered host will be assigned.
- `enabled` - (Optional) A boolean value indicating whether the discovery rule is enabled. When set to `true`, the rule is active and will be evaluated.
- `order` - (Optional) An integer specifying the order in which the rule is evaluated. Lower numbers are evaluated first.
- `parameters` - (Optional) A map of key-value pairs that will be saved as discovery rule parameters. These parameters can be used to pass additional information to the rule.
- `max_count` - (Optional) The maximum number of hosts that can be discovered by this rule. A value of `0` means unlimited.
- `hostname` - (Optional) The hostname pattern to be used for the discovered hosts.
- `location_ids` - (Optional) A list of location IDs where the discovered hosts will be assigned.
- `organization_ids` - (Optional) A list of organization IDs where the discovered hosts will be assigned.

## Attributes Reference

The following attributes are exported:

- `name` - The name of the discovery rule.
- `search` - The search criteria used to match hosts.
- `priority` - The priority of the rule.
- `hostgroup_id` - The ID of the host group to which the discovered host will be assigned.
- `enabled` - Whether the discovery rule is enabled.
- `order` - The order in which the rule is evaluated.
- `parameters` - A map of parameters that will be saved as discovery rule parameters.
- `max_count` - (Optional) The maximum number of hosts that can be discovered by this rule. A value of `0` means unlimited.
- `hostname` - (Optional) The hostname pattern to be used for the discovered hosts.
- `location_ids` - (Optional) A list of location IDs where the discovered hosts will be assigned.
- `organization_ids` - (Optional) A list of organization IDs where the discovered hosts will be assigned.
