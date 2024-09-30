## Data sources

data "foreman_architecture" "x86_64" {
  name = "x86_64"
}

data "foreman_media" "almalinux9" {
  name = "AlmaLinux"
}

data "foreman_partitiontable" "almalinux9" {
  name = "Kickstart default"
}





## Resources

resource "foreman_domain" "dot-invalid" {
  name = "e2e.invalid"
}

resource "foreman_operatingsystem" "e2eos" {
  name = "E2E-OS"
  major = "1"
  minor = "0"

  architectures = [data.foreman_architecture.x86_64.id]
  media = [data.foreman_media.almalinux9.id]
  partitiontables = [data.foreman_partitiontable.almalinux9.id]
}

resource "foreman_host" "e2ehost" {
  name = "my e2e host"
  root_password = "sepgnapngapwn"

  domain_id = foreman_domain.dot-invalid.id
  operatingsystem_id = foreman_operatingsystem.e2eos.id
  architecture_id = data.foreman_architecture.x86_64.id
  medium_id = data.foreman_media.almalinux9.id
  ptable_id = data.foreman_partitiontable.almalinux9.id

  enable_bmc = false
  manage_power_operations = false
  interfaces_attributes {
    type = "interface"
    primary = true
    mac = "16:02:a7:1f:8e:f2"
    identifier = "e2eif"
    managed = true
    provision = false
  }
}
