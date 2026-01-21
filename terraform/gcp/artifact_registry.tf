resource "google_artifact_registry_repository" "main" {
  location      = var.gcp_region
  repository_id = var.artifact_registry_image_name
  description   = "AIExpense Backend Docker Images"
}

resource "google_artifact_registry_docker_image" "main" {
  name     = google_artifact_registry_repository.main.name
  location = var.gcp_region
  source   = "us-docker.pkg.dev/${var.gcp_project_id}/${var.artifact_registry_image_name}"

  # Tag with Git commit SHA if available
  tags = var.environment == "prod" ? ["latest"] : ["preview-${replace(github.event.ref, "refs/heads/", "")}"]
}
