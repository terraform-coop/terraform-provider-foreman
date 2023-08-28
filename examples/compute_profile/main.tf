data "foreman_computeresource" "vmware" {
  name = "VMware Cluster ABC"
}

resource "foreman_computeprofile" "Small VM" {
  name = "Small VM"

  compute_attributes {
    compute_resource_id = data.foreman_computeresource.vmware.id
    vm_attrs = {
        cpus = 2
        memory_mb = 4096
    }
  }
}