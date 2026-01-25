resource "google_cloud_run_service" "aiexpense_backend_prod" {
  name     = "${var.cloud_run_service_name}-prod"
  location = var.gcp_region

  template {
    spec {
      containers {
        image = "${var.gcp_region}-docker.pkg.dev/${var.gcp_project_id}/${var.artifact_registry_image_name}/backend:latest"

        env {
          name  = "SERVER_PORT"
          value = "8080"
        }
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}

# Cloud Run Service - Preview
resource "google_cloud_run_service" "aiexpense_backend_preview" {
  count = var.environment == "preview" ? 1 : 0

  name = "${var.cloud_run_service_name}-${var.environment}"

  location = var.gcp_region

  template {
    spec {
      containers {
        image = "${var.gcp_region}-docker.pkg.dev/${var.gcp_project_id}/${var.artifact_registry_image_name}/backend:preview"

        env {
          name  = "SERVER_PORT"
          value = "8080"
        }
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}
