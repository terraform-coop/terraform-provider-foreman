resource "foreman_katello_sync_plan" "sync_plan_daily" {
   name = "daily"
   interval = "daily"
   enabled = true
   sync_date = "1970-01-01 00:00:00 UTC"
}