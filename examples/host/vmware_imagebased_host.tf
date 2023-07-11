data "foreman_image" "ubuntu" {
  name ="ubuntu-22.04"
  compute_resource_id = data.foreman_computeresource.my_vmware_cluster.id
}


resource "foreman_host" "image_based_machine" {
  count      = var.hosts.count
  name       = format("my-image-based-machine-%d", count.index + 1)
  provision_method     = "image"

  # This is the Foreman-internal image ID, integer
  # TODO: Do we need two image IDs, int and UUID?
  image_id   = data.foreman_image.ubuntu.id

  # The image_id JSON attribute is required in image-based setups on VMware! It must contain the VMware-internal UUID
  compute_attributes = format(<<EOF
{
    "image_id": "%s"
}
EOF
, data.foreman_image.ubuntu.uuid)

  hostgroup_id        = foreman_hostgroup.ubuntu.id
  compute_resource_id = data.foreman_computeresource.my_vmware_cluster.id
}


resource "foreman_hostgroup" "ubuntu" {
  name          = var.group

  # Not complete, fill with your own data
}