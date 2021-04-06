resource "foreman_katello_content_credential" "RPM-GPG-KEY-centos7" {
   name = "RPM-GPG-KEY-azure"
   content = file("RPM-GPG-KEY-centos7")
}