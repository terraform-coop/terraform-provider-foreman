# foreman_smartclassparameter manages an existing Puppet smart class parameter
# override. Every attribute is Computed by Foreman (parameter, puppetclass_id,
# override, description, default_value, hidden_value) and Create is not
# supported, so there is nothing to set in configuration -- import the
# parameter by its numeric ID (see import.sh) to bring it under management.
resource "foreman_smartclassparameter" "ntp_servers" {
}
