// Get the root 'Library' environment as data source
data "foreman_katello_lifecycle_environment" "library" {
  name = "Library"
}

// Then create a new lifecycle environment which uses the Library as the prior environment
resource "foreman_katello_lifecycle_environment" "newenv" {
  name = "My new lifecycle env"
  prior_id = data.foreman_katello_lifecycle_environment.library.id
  organization_id = 1
}
