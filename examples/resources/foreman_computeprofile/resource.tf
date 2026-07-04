data "foreman_computeresource" "vcenter" {
  name = "VCenter"
}

resource "foreman_computeprofile" "small_vm" {
  name = "1-CPU 2GB"

  compute_attributes = [
    {
      compute_resource_id = data.foreman_computeresource.vcenter.id
      vm_attrs = jsonencode({
        cpus      = 1
        memory_mb = 2048
      })
    }
  ]
}
