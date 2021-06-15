resource "foreman_katello_repository" "centos7base" {
  name = "centos7base"
  content_type = "yum"
  product_id = 2
  gpg_key_id = 2
  url = "http://mirror.centos.org/centos/7/os/x86_64/"
  label = "centos7base"
  http_proxy_policy = "global_default_http_proxy"
  checksum_type = "sha256"
  http_proxy_id = 1
  download_policy = "immediate"
}