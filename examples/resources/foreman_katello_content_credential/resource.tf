resource "foreman_katello_content_credential" "rpm_gpg_key_centos7" {
  name    = "RPM-GPG-KEY-CentOS-7"
  content = file("${path.module}/RPM-GPG-KEY-CentOS-7")
}
