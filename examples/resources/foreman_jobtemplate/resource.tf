resource "foreman_jobtemplate" "restart_service" {
  name          = "Restart service"
  description   = "Restarts a named service via the default job template provider"
  job_category  = "Services"
  provider_type = "script"
}
