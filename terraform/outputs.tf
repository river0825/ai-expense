output "cloud_run_service_url" {
  description = "Cloud Run service URL"
  value       = google_cloud_run_service.aiexpense_backend.status[0].url
}

output "cloud_run_service_name" {
  description = "Cloud Run service name"
  value       = google_cloud_run_service.aiexpense_backend.name
}

output "supabase_database_url" {
  description = "Supabase database connection URL"
  value       = supabase_database.main.database_uri
}

output "supabase_project_id" {
  description = "Supabase project ID"
  value       = supabase_project.aiexpense.id
}

output "supabase_anon_key" {
  description = "Supabase anonymous key for client access"
  value       = supabase_project.aiexpense.anon_key
}
