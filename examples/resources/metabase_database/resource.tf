resource "metabase_database" "postgres" {
  name   = "Analytics Postgres"
  engine = "postgres"
  details = jsonencode({
    host   = "db.example.com"
    port   = 5432
    dbname = "analytics"
    user   = "metabase"
    ssl    = false
  })
}
