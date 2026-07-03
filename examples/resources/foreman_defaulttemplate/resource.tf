# Associates a provisioning template as the default template of a
# given template kind (e.g. "provision", "PXELinux", "finish").
resource "foreman_defaulttemplate" "kickstart_default" {
  provisioning_template_id = 1
  template_kind_id         = 2
}
