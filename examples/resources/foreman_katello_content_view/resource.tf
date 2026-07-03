# Look up an existing repository to include in the content view
data "foreman_katello_repository" "epel8" {
  name = "EPEL 8"
}

resource "foreman_katello_content_view" "epel8_cv" {
  name            = "EPEL 8 Content View"
  description     = "Content view containing the EPEL 8 repository"
  organization_id = 1

  repository_ids = [tonumber(data.foreman_katello_repository.epel8.id)]

  composite    = false
  auto_publish = false

  filters = [
    {
      name        = "exclude-testfilter"
      type        = "rpm"
      inclusion   = false
      description = "Excludes all packages named testfilter-*"

      rule = [
        {
          name = "testfilter-*"
        }
      ]
    }
  ]
}
