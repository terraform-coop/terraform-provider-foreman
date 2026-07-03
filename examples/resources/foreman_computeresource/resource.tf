variable "vmware_password" {
  type      = string
  sensitive = true
}

resource "foreman_computeresource" "vmware" {
  name        = "VMware Cluster ABC"
  description = "Production VMware compute resource"
  compute_resource_provider = "Vmware"
  server      = "vcenter.example.com"
  datacenter  = "DC1"
  user        = "svc-foreman"
  password    = var.vmware_password

  caching_enabled = true
}
