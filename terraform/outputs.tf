output "cloud_run_service_url" {
  description = "Cloud Run service URL"
  value       = google_cloud_run_service.aiexpense_backend_prod.status[0].url
}

output "cloud_run_service_name" {
  description = "Cloud Run service name"
  value       = google_cloud_run_service.aiexpense_backend_prod.name
}


