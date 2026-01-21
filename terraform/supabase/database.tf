resource "supabase_project" "aiexpense" {
  name        = "${var.supabase_project_suffix}-${var.environment}"
  db_password = var.supabase_db_password
}

resource "supabase_database" "main" {
  project_id = supabase_project.aiexpense.id
  name       = "aiexpense"
}

# Note: The Supabase provider does not support CORS configuration via Terraform
# CORS will be configured manually in Supabase dashboard
