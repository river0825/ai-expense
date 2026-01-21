# Tasks: Add CI/CD Pipeline (Vercel + Google Cloud Run + Supabase + Terraform IaC)

## 1. Docker Container for Cloud Run
- [x] 1.1 Update `Dockerfile` with multi-stage build for Cloud Run optimization
- [x] 1.2 Add health check endpoint `/health` to Go backend (if not present)
- [x] 1.3 Ensure Go binary is statically linked (CGO_ENABLED=1 for SQLite)
- [x] 1.4 Test Docker image locally: `docker build -t aiexpense . && docker run -p 8080:8080 aiexpense`

## 2. Terraform Infrastructure Setup
- [x] 2.1 Create `terraform/` directory structure
- [x] 2.2 Create `terraform/main.tf` with provider declarations
- [x] 2.3 Create `terraform/variables.tf` with input variables (environment, region, project_id)
- [x] 2.4 Create `terraform/outputs.tf` with output values (service URLs, database URLs)
- [x] 2.5 Create `terraform/.gitignore` for `.tfstate`, `.tfstate.backup`, and `.tfvars` files

## 3. Terraform GCP Resources
- [x] 3.1 Create `terraform/gcp/provider.tf` with Google provider configuration
- [x] 3.2 Create `terraform/gcp/cloud_run.tf` with Cloud Run service definitions
  - Preview service (environment-specific naming)
  - Production service
- [x] 3.3 Create `terraform/gcp/artifact_registry.tf` for Docker image repositories
- [x] 3.4 Create `terraform/gcp/iam.tf` for service accounts and IAM roles
- [x] 3.5 Create `terraform/gcp/firewall.tf` for VPC connectors (if needed)

## 4. Terraform Supabase Resources
- [x] 4.1 Create `terraform/supabase/provider.tf` with Supabase provider configuration
- [x] 4.2 Create `terraform/supabase/database.tf` with project and database definitions
  - Single project with environment-specific databases (dev, preview, prod)

## 5. Supabase Database Integration
- [ ] 5.1 Run existing migrations on Supabase database (via psql or migration tool)
- [ ] 5.2 Configure Supabase project settings to allow CORS from Vercel domains
- [ ] 5.3 Add Supabase connection credentials to GitHub Secrets (`SUPABASE_DATABASE_URL`, `SUPABASE_ANON_KEY`)
- [ ] 5.4 Update repository implementation to support PostgreSQL (adapter pattern or conditional logic)
- [ ] 5.5 Test Supabase connection via Terraform outputs

## 6. Vercel Configuration (Frontend Only)
- [x] 6.1 Update `vercel.json` to remove Go function configuration
- [x] 6.2 Configure Next.js build settings
- [x] 6.3 Add environment variables in Vercel dashboard:
    - `NEXT_PUBLIC_API_URL` (production Cloud Run URL)
- [ ] 6.4 Add API rewrites to route `/api/*` to Cloud Run backend (optional, for unified domain)

## 7. GitHub Actions CI/CD Workflow
- [x] 7.1 Update existing `.github/workflows/ci.yml` (keep CI jobs)
- [x] 7.2 Create `.github/workflows/deploy.yml` with deployment and Terraform jobs
- [x] 7.3 Add Terraform setup job:
  - Install Terraform CLI
  - Authenticate with GCP (Service Account key from Secrets)
- [x] 7.4 Add Terraform apply-preview job:
  - Trigger on pull_request events
  - Run `terraform apply -auto-approve -var-file=terraform/preview.tfvars`
  - Extract Supabase database URL from Terraform outputs
  - Extract Cloud Run service URL from Terraform outputs
- [x] 7.5 Add Terraform apply-prod job:
  - Trigger on push to main
  - Run `terraform apply -auto-approve -var-file=terraform/prod.tfvars`
- [x] 7.6 Add Terraform destroy-preview job:
  - Trigger on pull_request close (not merged)
  - Run `terraform destroy -auto-approve -var-file=terraform/preview.tfvars`
- [x] 7.7 Add backend preview deployment job (builds and pushes Docker image after Terraform apply)
- [x] 7.8 Add backend production deployment job (builds and pushes Docker image after Terraform apply)
- [x] 7.9 Add Google Cloud authentication in GitHub Secrets (`GCP_SA_KEY`, `GCP_PROJECT_ID`, `GCP_REGION`, `SUPABASE_ACCESS_TOKEN`)
- [x] 7.10 Configure Terraform state backend (Terraform Cloud for dev, GCS bucket for prod)

## 8. Repository Updates for PostgreSQL Support
- [x] 8.1 Add PostgreSQL driver dependency (`lib/pq`) to `go.mod`
- [x] 8.2 Update repository interface to support multiple database backends
- [x] 8.3 Implement PostgreSQL repository adapter (or update existing adapter with conditional logic)
- [x] 8.4 Add environment-based repository selection in `repository` initialization
- [x] 8.5 Test locally with SQLite and verify with Supabase connection

## 9. Environment Configuration
- [x] 9.1 Update `internal/config/` to support `DATABASE_URL` (PostgreSQL) in addition to `DATABASE_PATH` (SQLite)
- [x] 9.2 Add validation for required environment variables
- [x] 9.3 Document all environment variables for both platforms:
  - Cloud Run: `DATABASE_URL`, `GEMINI_API_KEY`, `LINE_CHANNEL_TOKEN`, `ADMIN_API_KEY`
  - Vercel: `NEXT_PUBLIC_API_URL`

## 10. Documentation
- [x] 10.1 Update README.md with split deployment architecture and Terraform setup
- [x] 10.2 Document Terraform structure and variables
- [x] 10.3 Document Supabase setup and migration process via Terraform
- [x] 10.4 Document GitHub Secrets required for CI/CD

## 11. Validation
- [ ] 11.1 Test CI workflow: Pull request should trigger all CI jobs
- [ ] 11.2 Test Terraform apply: Create PR, verify GCP and Supabase resources are created
- [ ] 11.3 Test backend preview deployment: Verify Cloud Run preview service is created via Terraform
- [ ] 11.4 Test backend production deployment: Merge PR, verify production service is updated via Terraform
- [ ] 11.5 Test Vercel frontend preview: Verify preview deployment uses correct `NEXT_PUBLIC_API_URL` from Terraform outputs
- [ ] 11.6 Test end-to-end flow: Create expense via dashboard, verify it saves to Supabase
- [ ] 11.7 Test CORS: Verify dashboard can call Cloud Run backend without errors
- [ ] 11.8 Verify cleanup: Close PR without merging, confirm Terraform destroy removes preview resources
- [ ] 11.9 Verify Terraform state: Check state is properly backed up and locked

## 12. Cleanup (Remove Vercel Serverless Approach)
- [x] 12.1 Delete `cmd/vercel/main.go` (no longer needed)
- [x] 12.2 Remove any `api/` directory if created for serverless
- [x] 12.3 Update existing CI workflow to remove `go build ./api` step (only `go build ./cmd/server`)

