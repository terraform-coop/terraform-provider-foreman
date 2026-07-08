resource "foreman_architecture" "x86_64" {
  name = "x86_64"

  # IDs of operating systems that support this architecture
  operatingsystem_ids = [1, 2]
}
