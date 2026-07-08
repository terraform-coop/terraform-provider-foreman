data "foreman_operatingsystem" "centos7" {
  name = "CentOS 7"
}

resource "foreman_media" "centos_mirror" {
  name = "CentOS Mirror"
  path = "http://mirror.centos.org/centos/$major.$minor/os/$arch"

  os_family           = "Redhat"
  operatingsystem_ids = [data.foreman_operatingsystem.centos7.id]
}
