variable "client_username" {}
variable "client_password" {}

provider "foreman" {
  server_hostname = "192.168.1.118"
  server_protocol = "https"

  client_tls_insecure = true

  client_username = "${var.client_username}"
  client_password = "${var.client_password}"
}

data "foreman_hostgroup" "app" {
  title = "APP"
}

data "foreman_computeresource" "vcenter" {
  name = "VCenter"
}

data "foreman_computeprofile" "default" {
  name = "Default"
}

data "foreman_domain" "dev" {
  name = "dev.company.com"
}

data "foreman_environment" "production" {
  name = "production"
}

data "foreman_hostgroup" "DC1VM" {
  title = "DC1/VM"
}

data "foreman_operatingsystem" "Centos74" {
  title = "CentOS 7.4"
}

data "foreman_subnet" "app1" {
  name    = "10.228.170.0 app1"
  network = "10.228.170.0"
  mask    = "255.255.255.0"
}

resource "foreman_host" "TerraformTest" {
  name = "terraformed"

  hostgroup_id        = data.foreman_hostgroup.app.id
  environment_id      = data.foreman_hostgroup.app.environment_id
  operatingsystem_id  = data.foreman_operatingsystem.rhel7.id
  compute_profile_id  = data.foreman_computeprofile.default.id
  compute_resource_id = data.foreman_computeresource.vcenter.id

  owner_id   = data.foreman_usergroup.root.id
  owner_type = "Usergroup"

  parameters = {
    role = "postgresql"
  }

// Example uses vSphere compute attributes
  compute_attributes = <<EOF
{
    "cpus": 4,
    "memory_mb": 4096,
    "volumes_attributes": {
      "0": {
        "size_gb": 40,
        "thin": true,
        "datastore": "vsanDatastore"
      },
      "1": {
        "size_gb": 30,
        "thin": true,
        "datastore": "vsanDatastore"
      },
      "2": {
        "size_gb": 35,
        "thin": true,
        "datastore": "vsanDatastore"
      }
    }
}
EOF

  interfaces_attributes {
    type       = "interface"
    primary    = true
    identifier = "ens160"
    provision  = true
    managed    = true
    compute_attributes = {
      model   = "VirtualVmxnet3"
      network = "AppSubnet"
    }
  }
}
