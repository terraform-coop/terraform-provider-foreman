# The "Library" environment is the root of every organization's lifecycle path
data "foreman_katello_lifecycle_environment" "library" {
  name = "Library"
}

resource "foreman_katello_lifecycle_environment" "dev" {
  name            = "Development"
  description     = "Development lifecycle environment"
  label           = "development"
  organization_id = 1

  prior_id = tonumber(data.foreman_katello_lifecycle_environment.library.id)
}
