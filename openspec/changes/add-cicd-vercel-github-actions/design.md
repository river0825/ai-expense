# Design: CI/CD with Vercel (Frontend) and Google Cloud Run (Backend) with IaC

## Context
AIExpense is a monorepo with:
- **Go Backend**: Clean Architecture, SQLite (local), HTTP server (`cmd/server/`)
- **Next.js Dashboard**: React dashboard (`dashboard/`)

This change separates deployment concerns with IaC:
- **Frontend (Vercel)**: Optimal for Next.js with automatic preview deployments and global CDN
- **Backend (Google Cloud Run)**: Containerized Go service, scale-to-zero, generous free tier
- **Database (Supabase)**: Managed PostgreSQL replacing SQLite for production persistence
- **Infrastructure as Code (Terraform)**: Reproducible, versioned infrastructure deployment

## Goals
- Automated CI on every PR (lint, test, build)
- Automated deployment with preview environments for both frontend and backend
- Zero-downtime deployments
- Reuse existing Clean Architecture layers
- Cost-effective free-tier deployment
- Infrastructure defined as code for reproducibility

## Non-Goals
- Database migration automation (handled separately)
- Multi-region deployment (single region for MVP)
- Custom domain setup (manual configuration)
- Complex Kubernetes or managed instance setup
- Terraform state management (use Terraform Cloud for production, local backend for dev)

## Architecture

### Deployment Architecture
```
┌─────────────────┐
│   GitHub Repo   │
│   (PRs + main) │
└────────┬────────┘
         │
         ├── CI Pipeline (GitHub Actions)
         │   ├── Backend: test → build → terraform apply → deploy to GCR (preview + prod)
         │   └── Frontend: lint → test → deploy to Vercel (preview + prod)
         │
         ├── IaC Pipeline (Terraform)
         │   ├── GCP resources: Cloud Run, Artifact Registry, IAM roles
         │   └── Supabase resources: Project, Database, Tables
         │
         ├──────────────────────┬──────────────────────┐
         │                      │                      │
    ┌────▼────┐          ┌──────▼──────┐    ┌──────▼──────┐
    │  Vercel  │          │ Cloud Run    │    │  Supabase   │
    │ Frontend  │          │ Backend      │    │  Database   │
    │ (Next.js) │          │ (Go Docker)  │    │ (PostgreSQL) │
    └───────────┘          └──────┬──────┘    └─────────────┘
                                  │
                                  │ (DATABASE_URL)
                           Preview on every PR:
                           - gcr-preview-PR123.run.app
                           - vercel preview URL
```

### Project Structure
```
aiexpense/
├── Dockerfile               # Cloud Run container (removes serverless approach)
├── cmd/
│   ├── server/              # Main Go server (unchanged)
│   └── vercel/            # REMOVED (no longer needed)
├── dashboard/              # Next.js app (deployed to Vercel)
├── internal/               # Shared Go business logic
├── terraform/              # NEW: Infrastructure as Code
│   ├── main.tf             # Root Terraform config
│   ├── variables.tf         # Input variables
│   ├── outputs.tf          # Output values (service URLs)
│   ├── gcp/
│   │   ├── cloud_run.tf    # Cloud Run services
│   │   ├── artifact_registry.tf  # Docker registry
│   │   ├── iam.tf         # Service accounts and roles
│   │   └── provider.tf    # GCP provider config
│   ├── supabase/
│   │   ├── provider.tf    # Supabase provider config
│   │   └── database.tf    # Database project and schema
│   └── dev.tfvars         # Development variables (gitignored)
│   └── prod.tfvars        # Production variables (gitignored)
├── vercel.json             # Updated: frontend-only config
├── .github/workflows/
│   ├── ci.yml              # CI tests (unchanged)
│   └── deploy.yml          # NEW: Terraform apply + GCR + Vercel deployment
└── go.mod                 # Root go.mod
```

### Cloud Run Deployment Pattern
```yaml
# .github/workflows/deploy.yml
deploy-backend:
  # Build and push to Artifact Registry
  # Deploy to Cloud Run
  # Unique service name per PR: aiexpense-backend-preview-123
  # Production: aiexpense-backend-prod
```

### Database Configuration
- **Local Development**: SQLite (`aiexpense.db` via DATABASE_PATH)
- **Production (Cloud Run)**: Supabase PostgreSQL via `DATABASE_URL`
- **Environment-based selection**: Update repository implementation to support both SQLite and PostgreSQL
- **Migration Strategy**: Supabase has schema migration support; existing `migrations/` compatible

### GitHub Actions Workflow
```yaml
on:
  push: [main]
  pull_request: [main]

jobs:
  # CI (existing, unchanged)
  ci-backend:
    - go test ./...
    - go build ./cmd/server

  ci-dashboard:
    - bun install
    - bun run lint
    - bun run build

  # NEW: Terraform Infrastructure
  terraform-apply-preview:
    if: github.event_name == 'pull_request'
    - Setup Terraform
    - terraform init
    - terraform apply -auto-approve
    - Use: preview.tfvars (environment: preview)

  terraform-apply-prod:
    if: github.ref == 'refs/heads/main'
    - Setup Terraform
    - terraform init
    - terraform apply -auto-approve
    - Use: prod.tfvars (environment: production)

  # NEW: Deployments
  deploy-backend-preview:
    if: github.event_name == 'pull_request'
    - Build Docker image
    - Tag: preview-PR${{ github.event.pull_request.number }}
    - Push to Artifact Registry (defined in Terraform)
    - Cloud Run service created by Terraform references image

  deploy-backend-prod:
    if: github.ref == 'refs/heads/main'
    - Build Docker image
    - Tag: latest
    - Push to Artifact Registry (defined in Terraform)
    - Cloud Run service created by Terraform references image

  deploy-frontend:
    # Vercel handles automatically via Git integration
    # Preview on PR, production on main merge
```

### Terraform IaC Structure

```hcl
# terraform/main.tf
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

# terraform/gcp/cloud_run.tf
resource "google_cloud_run_service" "aiexpense_backend" {
  name     = "aiexpense-backend-${var.environment}"
  location = var.gcp_region

  template {
    spec {
      containers {
        image = "${google_artifact_registry_docker_image.aiexpense.image_url}"

        env {
          name  = "DATABASE_URL"
          value = var.supabase_database_url
        }
        env {
          name  = "GEMINI_API_KEY"
          value = var.gemini_api_key
        }
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}

# terraform/supabase/database.tf
resource "supabase_project" "aiexpense" {
  name        = "aiexpense-${var.environment}"
  db_password = var.supabase_db_password
  db_name     = "aiexpense"
}

resource "supabase_database" "main" {
  project_id = supabase_project.aiexpense.id
  name       = "aiexpense"
}
```

## Decisions

### Decision 1: Terraform for IaC
- **Choice**: Use Terraform with Google and Supabase providers
- **Alternatives**:
  - Google Cloud Deployment Manager: GCP-specific, limited to Google resources
  - Pulumi: Less ecosystem support, more complex
  - Manual `gcloud` CLI: Not reproducible, not version-controlled
- **Rationale**: Industry standard, excellent GCP and Supabase provider support, state management built-in

### Decision 2: Split Deployment (Vercel + Cloud Run)
- **Choice**: Frontend on Vercel, Backend on Cloud Run, Database on Supabase
- **Alternatives**:
  - All on Vercel (serverless Go): Limiting, no persistent DB
  - All on Cloud Run: Complex, loses Vercel's Next.js optimizations
  - All on Render: Limited free tier, spin-down delays
- **Rationale**: Best-in-class platforms for each component, generous free tiers combined

### Decision 2: Preview Environments for Both
- **Choice**: Create Cloud Run preview service AND Vercel preview on every PR
- **Alternatives**:
  - Preview only frontend: Cannot test API changes before merge
  - Preview only backend: Cannot test integration before merge
  - Manual preview on demand: Slows down development
- **Rationale**: Full-stack testing before merging, catches integration issues early

### Decision 3: Supabase for Cloud Database
- **Choice**: Supabase PostgreSQL (500MB free, auth, storage) managed via Terraform
- **Alternatives**:
  - Neon (0.5GB serverless): Good, but Supabase has more features
  - PlanetScale (MySQL 5GB): Not PostgreSQL native
  - Cloud SQL (paid): Expensive for MVP
- **Rationale**: PostgreSQL-native (standard), generous free tier, additional features (auth, real-time), Terraform provider available

### Decision 4: Docker Container for Cloud Run
- **Choice**: Single Dockerfile with multi-stage build
- **Alternatives**:
  - Buildpacks: Less control, larger images
  - Separate build step: More complex CI/CD
- **Rationale**: Optimized image size, Go binary-only in final stage

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Terraform state management | Use Terraform Cloud for production, local backend for dev; add `.terraform.lock.hcl` to git |
| Preview environment cost | Cloud Run free tier (2M requests) covers most testing; Terraform auto-deletes preview resources after PR close |
| CORS complexity | Configure Supabase CORS for Vercel domains; Terraform manages CORS policies |
| Database migration dual-support | Add repository interface; SQLite for local, PostgreSQL adapter for Supabase |
| Cold start latency on Cloud Run | Scale-to-zero is acceptable for MVP; can set min instances later |
| Preview service management | Terraform destroy workflow step to delete preview infrastructure |
| Supabase project limits | Use environment-specific projects (dev, preview, prod) to avoid quota conflicts |

## Migration Plan

1. **Phase 1**: Update Dockerfile for Cloud Run compatibility
2. **Phase 2**: Create Terraform configuration structure (providers, variables, main)
3. **Phase 3**: Define GCP resources in Terraform (Cloud Run, Artifact Registry, IAM)
4. **Phase 4**: Define Supabase resources in Terraform (project, database)
5. **Phase 5**: Update `vercel.json` for frontend-only deployment
6. **Phase 6**: Add `.github/workflows/deploy.yml` with Terraform + GCR deployment
7. **Phase 7**: Update repository layer for PostgreSQL support
8. **Phase 8**: Test Terraform infrastructure creation (dev environment)
9. **Phase 9**: Test preview deployments on PR (Terraform apply with preview variables)
10. **Phase 10**: Delete `cmd/vercel/main.go` (no longer needed)

## Open Questions

1. **Terraform state backend**: Should Terraform state be stored in Terraform Cloud or GCS bucket?
    - **Recommendation**: Terraform Cloud for MVP (free), migrate to GCS for production
2. **Preview service naming**: Should preview services auto-delete after PR merge or keep for N days?
    - **Recommendation**: Delete immediately after PR close via `terraform destroy`
3. **Environment variables**: How to manage LINE_CHANNEL_TOKEN, SUPABASE_URL across two platforms?
    - **Answer**: GitHub Secrets for GitHub Actions; Terraform passes to Cloud Run environment variables; Vercel UI for frontend
4. **Supabase project structure**: Single project with multiple environments or separate projects per environment?
    - **Recommendation**: Single Supabase project with environment-specific databases (dev, preview, prod) to stay within free tier
5. **Custom domain**: Should frontend and backend share a custom domain?
    - **Recommendation**: Optional for MVP; can configure `vercel.json` rewrites and Cloud Run domain mapping later
