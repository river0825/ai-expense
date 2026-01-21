# Service Account for Cloud Run
resource "google_service_account" "cloud_run_sa" {
  account_id   = "${var.cloud_run_service_name}-sa"
  display_name = "AIExpense Cloud Run Service Account"

  # Grant roles needed for Cloud Run and Artifact Registry
  roles = [
    "roles/run.admin",
    "roles/artifactregistry.reader",
    "roles/storage.objectAdmin", # For state backend
  ]
}

# IAM policy to allow Cloud Run to pull from Artifact Registry
resource "google_artifact_registry_repository_iam_member" "registry_iam" {
  project    = var.gcp_project_id
  location   = var.gcp_region
  repository = google_artifact_registry_repository.main.name
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${google_service_account.cloud_run_sa.email}"
}
