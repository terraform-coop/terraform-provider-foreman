# foreman_templatekind is a read-only reference entity built into Foreman
# (e.g. "provision", "PXELinux", "finish"). Create, Update and Delete are not
# supported and its only attribute (name) is Computed, so there is nothing to
# set in configuration -- import the template kind by its numeric ID (see
# import.sh) if you need to reference it from other resources.
resource "foreman_templatekind" "provision" {
}
