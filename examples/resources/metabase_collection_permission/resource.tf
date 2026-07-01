# Give the "Analysts" group write access to a collection.
resource "metabase_collection_permission" "analysts_reports" {
  group_id      = metabase_permission_group.analysts.id
  collection_id = metabase_collection.reports.id
  permission    = "write"
}
