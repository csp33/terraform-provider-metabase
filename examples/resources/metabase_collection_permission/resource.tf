# Give the "Analysts" group write access to a collection.
resource "metabase_collection_permission" "analysts_reports" {
  group_id      = metabase_permission_group.analysts.id
  collection_id = metabase_collection.reports.id
  permission    = "write"
}

# Give a partner group read access to a collection AND its whole subtree
# (like the UI's "Also change sub-collections" toggle).
resource "metabase_collection_permission" "partner_x" {
  group_id      = metabase_permission_group.partner_x.id
  collection_id = "10"
  permission    = "read"
  propagate     = true
}

# Internal "sees everything" group: grant the virtual root — every existing
# collection gets the permission, and new ones inherit it automatically.
resource "metabase_collection_permission" "internal_viewers" {
  group_id      = metabase_permission_group.internal_viewers.id
  collection_id = "root"
  permission    = "read"
  propagate     = true
}
