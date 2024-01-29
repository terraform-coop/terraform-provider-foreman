resource "foreman_katello_sync_plan" "sync_plan_daily" {
   name = "daily"
   interval = "daily"
   enabled = true
   description = "My sync plan description"

   // Run the first sync plan on Jan 1st 2024 at 5:10 in the morning in UTC time.
   // If parameter 'interval' is defined, time defined in 'sync_date' will determine
   // when the sync runs in general.
   // Example: If sync_date is Jan 1st 2024 and interval is daily, next run will be at Jan 2nd at 5:10.
   sync_date = "2024-01-01 05:10:00 +0000"

   //cron_expression = "*/5 * * * *" // every 5min
}
