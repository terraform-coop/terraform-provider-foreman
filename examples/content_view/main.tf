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


//// Content view example (repos are not defined in this example)
// Content view with repo for Ceph Pacific
resource "foreman_katello_content_view" "ubuntu_ceph_v16" {
  name = "Ubuntu Ceph Pacific"
  repository_ids = [
    data.foreman_katello_repository.debian_ceph_pacific.id
  ]
}
// Content view with repo for Ceph Quincy
resource "foreman_katello_content_view" "ubuntu_ceph_v17" {
  name = "Ubuntu Ceph Quincy"
  repository_ids = [
    data.foreman_katello_repository.debian_ceph_quincy.id
  ]
}
// Composite content view consuming both CVs above
resource "foreman_katello_content_view" "ubuntu_ceph_ccv" {
  name = "Ubuntu Ceph composite content view Pacific+Quincy"
  composite = true
  auto_publish = true

  component_ids = [
    foreman_katello_content_view.ubuntu_ceph_v16.latest_version_id,
    foreman_katello_content_view.ubuntu_ceph_v17.latest_version_id
  ]
}