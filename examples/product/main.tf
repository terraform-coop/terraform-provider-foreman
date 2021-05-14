resource "foreman_katello_product" "debian_10" {
  name = "Debian 10"
  description =  "Debian Buster"
  gpg_key_id = 5
  sync_plan_id = 1
  label = "Debian 10"
}