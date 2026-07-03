resource "foreman_smartproxy" "dc1_proxy" {
  name = "proxy.dc1.example.com"
  url  = "https://proxy.dc1.example.com:8443"
}
