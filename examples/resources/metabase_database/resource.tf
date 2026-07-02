# A Postgres database connection managed by Terraform.
resource "metabase_database" "sales" {
  name   = "Analytics Postgres"
  engine = "postgres"
  details = jsonencode({
    host     = "db.example.com"
    port     = 5432
    dbname   = "analytics"
    user     = "metabase"
    password = "..."
    ssl      = false
  })

  # Metabase hard-deletes a database and ALL content built on it. deletion_protection
  # (default true) makes the provider refuse to delete it; set false + apply first.
  deletion_protection = true

  # Belt-and-suspenders: Terraform itself also blocks `destroy` for this resource.
  lifecycle {
    prevent_destroy = true
  }
}
