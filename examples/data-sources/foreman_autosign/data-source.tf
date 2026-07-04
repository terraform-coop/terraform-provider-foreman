data "foreman_smartproxy" "puppet" {
  name = "puppet.example.com"
}

data "foreman_autosign" "app_wildcard" {
  smart_proxy_id = data.foreman_smartproxy.puppet.id
  id             = "*.app.example.com"
}
