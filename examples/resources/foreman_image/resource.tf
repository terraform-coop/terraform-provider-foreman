resource "foreman_image" "centos7" {
  name     = "CentOS 7 base image"
  username = "root"
  uuid     = "ami-0123456789abcdef0"

  compute_resource_id = "1"
  architecture_id     = "1"
  operatingsystem_id  = 1
  user_data           = true
}
