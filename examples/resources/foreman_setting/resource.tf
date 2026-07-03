# foreman_setting is a singleton-style resource: Foreman ships a fixed set of
# global settings and this resource can only manage the "value" of an
# existing one. Create is not supported -- import the setting by its
# numeric ID first (see import.sh), then manage its value going forward.
resource "foreman_setting" "append_domain_name_for_hosts" {
  value = "true"
}
