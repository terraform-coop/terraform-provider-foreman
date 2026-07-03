resource "foreman_usergroup" "operators" {
  name = "Operators"

  admin         = false
  role_ids      = [3]
  user_ids      = [12, 15]
  usergroup_ids = [4]
}
