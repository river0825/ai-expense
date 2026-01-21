variable "environment" {
  type        = string
  description = "Environment name (dev, preview, prod)"
  default     = "dev"
}

variable "gcp_region" {
  type        = string
  description = "Google Cloud region"
  default     = "us-central1"
}

variable "gcp_project_id" {
  type        = string
  description = "Google Cloud project ID"
}

variable "cloud_run_service_name" {
  type        = string
  description = "Cloud Run service name suffix"
  default     = "aiexpense-backend"
}

variable "supabase_db_password" {
  type        = string
  description = "Supabase database password"
  sensitive   = true
}

variable "supabase_project_suffix" {
  type        = string
  description = "Supabase project suffix"
  default     = "aiexpense"
}

variable "artifact_registry_image_name" {
  type        = string
  description = "Docker image name in Artifact Registry"
  default     = "aiexpense-backend"
}
