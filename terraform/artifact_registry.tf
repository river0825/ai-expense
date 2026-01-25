resource "google_artifact_registry_repository" "main" {
  location      = var.gcp_region
  repository_id = var.artifact_registry_image_name
  description   = "AIExpense Backend Docker Images"
  format        = "DOCKER"
}

