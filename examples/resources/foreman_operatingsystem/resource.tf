data "foreman_media" "centos_mirror" {
  name = "CentOS Mirror"
}

resource "foreman_operatingsystem" "centos7" {
  name  = "CentOS"
  major = "7"
  minor = "9"

  family      = "Redhat"
  description = "CentOS 7.9 x86_64"

  medium_ids = [data.foreman_media.centos_mirror.id]

  password_hash = var.operatingsystem_password_hash

  os_parameters_attributes = {
    "custom-repo" = "http://mirror.example.com/centos/7/os/x86_64"
  }
}
