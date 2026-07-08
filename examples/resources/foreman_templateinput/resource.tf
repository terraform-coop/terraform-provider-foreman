resource "foreman_templateinput" "os_major_version" {
  name       = "os_major_version"
  input_type = "user"

  description = "The operating system major version to provision"
  required    = true
  default     = "9"
  options     = ["7", "8", "9"]
}
