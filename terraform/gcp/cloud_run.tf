resource "google_cloud_run_service" "aiexpense_backend_prod" {
  name     = "${var.cloud_run_service_name}-prod"
  location = var.gcp_region

  template {
    spec {
      containers {
        image = google_artifact_registry_docker_image.main.image_url

        env {
          name  = "DATABASE_URL"
          value = var.supabase_database_url
        }
        env {
          name  = "GEMINI_API_KEY"
          value = var.gemini_api_key
        }
        env {
          name  = "LINE_CHANNEL_TOKEN"
          value = var.line_channel_token
        }
        env {
          name  = "ADMIN_API_KEY"
          value = var.admin_api_key
        }
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

  depends_on = [google_artifact_registry_docker_image.main]
}

# Cloud Run Service - Preview
resource "google_cloud_run_service" "aiexpense_backend_preview" {
  count = var.environment == "preview" ? 1 : 0

  name = "${var.cloud_run_service_name}-${var.environment}"

  location = var.gcp_region

  template {
    spec {
      containers {
        image = google_artifact_registry_docker_image.main.image_url

        env {
          name  = "DATABASE_URL"
          value = var.supabase_database_url
        }
        env {
          name  = "GEMINI_API_KEY"
          value = var.gemini_api_key
        }
        env {
          name  = "LINE_CHANNEL_TOKEN"
          value = var.line_channel_token
        }
        env {
          name  = "ADMIN_API_KEY"
          value = var.admin_api_key
        }
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

  depends_on = [google_artifact_registry_docker_image.main]
}
