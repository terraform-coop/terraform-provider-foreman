resource "foreman_computeprofile" "vmware_webserver" {
  name = "VMware Webserver"

  compute_attributes {
    name = "Webserver middle (2 CPUs and 16GB memory)"
    compute_resource_id = data.foreman_computeresource.vmware.id

    vm_attrs = {
        cpus = 2
        corespersocket = 1
        memory_mb = 16384
        firmware = "bios"
        resource_pool = "pool1"
        guest_id = "ubuntu64Guest"

        boot_order = jsonencode([ "disk", "network" ])

        interfaces_attributes = jsonencode({
            0: { type: "VirtualE1000", network: "dmz-net" },
            1: { type: "VirtualE1000", network: "webserver-net" } }
        )
    }
  }
}
