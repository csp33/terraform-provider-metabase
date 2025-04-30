resource "metabase_permission_group" "analysts" {
  name = "👨🏻‍💻 Analysts"
}

resource "metabase_user" "john_doe" {
  email      = "john@doe.com"
  first_name = "John"
  last_name  = "Doe"
}

resource "metabase_user_permission_group_membership" "john_doe_analyst" {
  user_id  = metabase_user.john_doe.id
  group_id = metabase_permission_group.analysts.id
}
