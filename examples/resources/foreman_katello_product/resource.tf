# Reference a sync plan to attach to the product
data "foreman_katello_sync_plan" "daily" {
  name = "daily"
}

resource "foreman_katello_product" "debian_12" {
  name        = "Debian 12"
  description = "Debian Bookworm"
  label       = "debian_12"

  sync_plan_id = tonumber(data.foreman_katello_sync_plan.daily.id)
  gpg_key_id   = 5
}
