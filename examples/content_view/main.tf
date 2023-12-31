// Simple Content View data source
data "foreman_katello_content_view" "ubuntu2204" {
  name = "Ubuntu 22.04"
}

// Query a repository to use its ID in the Content View
data "foreman_katello_repository" "ubuntu2204" {
  name = "Ubuntu 22.04"
}

// Create a new Content View with one repository and one filter
resource "foreman_katello_content_view" "test_cv_write" {
  name = "Test CV for Ubuntu Sec"
  repository_ids = [data.foreman_katello_repository.ubuntu2204.id]
  composite = false

  filter {
    name = "my filter 1"
    type = "deb"
    inclusion = true
    description = "Filters all packages except those with name 'testfilter-*'"

    rule {
      name = "testfilter-*"
    }
  }
}
