# Change: Add CI/CD Pipeline with Vercel (Frontend) and Google Cloud Run (Backend) with IaC

## Why
The project lacks automated CI/CD. Manual deployment is error-prone and slows down iteration. Separating frontend (Next.js) and backend (Go) deployment leverages each platform's strengths: Vercel for optimal Next.js experience, Google Cloud Run for scalable containerized Go services with Supabase for cloud database. IaC ensures infrastructure is reproducible, version-controlled, and consistent across environments.

## What Changes
- **BREAKING**: Remove Vercel serverless approach (delete `cmd/vercel/main.go`)
- Add Docker container for Go backend (Cloud Run compatible)
- Add Terraform IaC for GCP resources (Cloud Run, Artifact Registry, IAM)
- Add Terraform IaC for Supabase project configuration
- Add GitHub Actions workflows for CI and CD (deploy to GCR and Vercel)
- Configure Supabase as cloud database (replacing SQLite for production)
- Set up preview deployments on every PR for both frontend (Vercel) and backend (GCR)

## Impact
- Affected specs: None (new capability)
- Affected code:
  - NEW: `Dockerfile` - Container for Cloud Run
  - NEW: `terraform/` - IaC definitions for GCP and Supabase
  - NEW: `terraform/main.tf` - Root Terraform configuration
  - NEW: `terraform/gcp/` - Google Cloud resources
  - NEW: `terraform/supabase/` - Supabase resources
  - NEW: `.github/workflows/deploy.yml` - Deployment with Terraform apply
  - MODIFIED: `.github/workflows/ci.yml` - Add GCR deployment jobs
  - MODIFIED: `vercel.json` - Update for frontend-only deployment
  - REMOVED: `cmd/vercel/main.go` - No longer needed
  - NEW: Supabase database configuration
