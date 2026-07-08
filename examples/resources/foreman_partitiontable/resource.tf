resource "foreman_partitiontable" "centos_lvm" {
  name   = "CentOS LVM"
  layout = <<-EOT
    zerombr
    clearpart --all --initlabel
    autopart --type=lvm
  EOT

  os_family   = "Redhat"
  description = "Default LVM autopartitioning layout for CentOS hosts"
  snippet     = false
}
