resource "foreman_httpproxy" "example" {
  name     = "proxy.example.com"
  url      = "https://proxy.example.com:8443"
  username = "proxyuser"
  password = var.httpproxy_password
}
