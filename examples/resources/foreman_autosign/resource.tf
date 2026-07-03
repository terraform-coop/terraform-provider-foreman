# The foreman_autosign resource currently exposes no configurable
# attributes beyond its computed id; creating it registers a new
# (empty) autosign entry.
resource "foreman_autosign" "example" {
}
