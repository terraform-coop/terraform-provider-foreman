# Reference the product this repository belongs to
data "foreman_katello_product" "debian_12" {
  name = "Debian 12"
}

resource "foreman_katello_repository" "debian_12_base" {
  name  = "debian12base"
  label = "debian12base"

  product_id   = tonumber(data.foreman_katello_product.debian_12.id)
  content_type = "deb"
  url          = "http://deb.debian.org/debian/"

  checksum_type   = "sha256"
  download_policy = "immediate"

  deb_releases      = "bookworm"
  deb_components    = "main"
  deb_architectures = "amd64"
}
