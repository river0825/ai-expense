terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    supabase = {
      source  = "supabase/supabase"
      version = "~> 1.0"
    }
  }

  backend "gcs" {
    bucket = "aiexpense-terraform-state"
    prefix = "terraform/state"
  }
}
