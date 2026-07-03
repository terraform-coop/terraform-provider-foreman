data "foreman_hostgroup" "app" {
  name = "app"
}

data "foreman_computeresource" "vcenter" {
  name = "VCenter"
}

data "foreman_computeprofile" "default" {
  name = "Default"
}

data "foreman_domain" "dev" {
  name = "dev.example.com"
}

data "foreman_operatingsystem" "centos" {
  name = "CentOS 7.4"
}

resource "foreman_host" "web" {
  name = "web01.dev.example.com"

  hostgroup_id        = data.foreman_hostgroup.app.id
  domain_id           = data.foreman_domain.dev.id
  operatingsystem_id  = data.foreman_operatingsystem.centos.id
  compute_profile_id  = data.foreman_computeprofile.default.id
  compute_resource_id = data.foreman_computeresource.vcenter.id

  root_pass = var.host_root_pass

  parameters = {
    role = "postgresql"
  }

  # Additional compute-resource-specific attributes (e.g. vSphere CPU/memory/disks)
  # are passed through as a raw JSON string.
  compute_attributes = jsonencode({
    cpus      = 4
    memory_mb = 4096
    volumes_attributes = {
      "0" = {
        size_gb   = 40
        thin      = true
        datastore = "vsanDatastore"
      }
    }
  })

  interfaces = [
    {
      type       = "interface"
      primary    = true
      identifier = "ens160"
      provision  = true
      managed    = true
      compute_attributes = jsonencode({
        model   = "VirtualVmxnet3"
        network = "AppSubnet"
      })
    }
  ]
}
