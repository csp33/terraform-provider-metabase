# Give the "Analysts" group query-builder access to a database.
# On OSS create_queries is the effective data-access control.
resource "metabase_database_permission" "analysts_sales" {
  group_id       = metabase_permission_group.analysts.id
  database_id    = metabase_database.sales.id
  create_queries = "query-builder" # "no" | "query-builder" | "query-builder-and-native"
}
