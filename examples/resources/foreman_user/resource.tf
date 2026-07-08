variable "jdoe_password" {
  type      = string
  sensitive = true
}

resource "foreman_user" "jdoe" {
  login          = "jdoe"
  mail           = "jdoe@example.com"
  auth_source_id = 1 # 1 is Foreman's built-in "Internal" auth source

  firstname = "Jane"
  lastname  = "Doe"
  admin     = false
  password  = var.jdoe_password
  role_ids  = [3, 4]
}
