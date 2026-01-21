# CICD Capability

## Overview
Automated CI/CD pipeline using GitHub Actions for continuous integration, Vercel for frontend deployment, Google Cloud Run for backend deployment with Supabase as cloud database, and Terraform for infrastructure as code.

## ADDED Requirements

### Requirement: GitHub Actions CI Pipeline
The system SHALL run automated checks on every pull request and push to main branch.

#### Scenario: Go backend CI on pull request
- **WHEN** a pull request is opened or updated targeting main branch
- **THEN** GitHub Actions runs `go test ./...`
- **AND** GitHub Actions runs `go build ./cmd/server`
- **AND** GitHub Actions builds Docker image for Cloud Run
- **AND** workflow fails if any step fails

#### Scenario: Dashboard CI on pull request
- **WHEN** a pull request is opened or updated targeting main branch
- **THEN** GitHub Actions runs `bun install` in dashboard directory
- **AND** GitHub Actions runs `bun run lint`
- **AND** GitHub Actions runs `bun run build`
- **AND** workflow fails if any step fails

#### Scenario: CI runs on push to main
- **WHEN** commits are pushed to main branch
- **THEN** same CI checks run as for pull requests

### Requirement: Terraform Infrastructure as Code
The system SHALL define all cloud infrastructure using Terraform configuration.

#### Scenario: Terraform configuration for GCP resources
- **WHEN** Terraform is initialized
- **THEN** Google Cloud Run services are defined in `terraform/gcp/cloud_run.tf`
- **AND** Google Artifact Registry repositories are defined in `terraform/gcp/artifact_registry.tf`
- **AND** IAM service accounts and roles are defined in `terraform/gcp/iam.tf`
- **AND** all resources use the GCP provider

#### Scenario: Terraform configuration for Supabase resources
- **WHEN** Terraform is initialized
- **THEN** Supabase project is defined in `terraform/supabase/database.tf`
- **AND** Supabase database is defined in `terraform/supabase/database.tf`
- **AND** all resources use the Supabase provider

#### Scenario: Terraform apply for preview environments
- **WHEN** a pull request is opened
- **THEN** GitHub Actions runs `terraform apply` with preview environment variables
- **AND** preview Cloud Run service is created with unique name
- **AND** preview Supabase database is created or updated
- **AND** infrastructure outputs are available for downstream steps

#### Scenario: Terraform destroy for preview cleanup
- **WHEN** a pull request is closed or merged
- **THEN** GitHub Actions runs `terraform destroy` targeting preview resources
- **AND** preview Cloud Run service is deleted
- **AND** preview Supabase database is deleted or cleaned up
- **AND** Terraform state is updated

#### Scenario: Terraform state management
- **WHEN** Terraform applies changes
- **THEN** state is stored in Terraform Cloud for development
- **AND** state is stored in GCS bucket for production
- **AND** `.terraform.lock.hcl` is committed to git for version control

### Requirement: Google Cloud Run Backend Deployment
The system SHALL deploy the Go backend as a containerized service on Google Cloud Run.

#### Scenario: Preview deployment on pull request
- **WHEN** a pull request is opened or updated
- **THEN** GitHub Actions builds Docker image tagged `preview-PR{number}`
- **AND** GitHub Actions deploys to Cloud Run service `aiexpense-backend-preview-{number}`
- **AND** the service is publicly accessible with HTTPS
- **AND** the preview URL is posted as a PR comment

#### Scenario: Production deployment on merge
- **WHEN** a pull request is merged to main branch
- **THEN** GitHub Actions builds Docker image tagged `latest`
- **AND** GitHub Actions deploys to Cloud Run service `aiexpense-backend-prod`
- **AND** the service is publicly accessible with HTTPS
- **AND** the deployment uses Supabase database connection

#### Scenario: Reuse existing business logic
- **WHEN** the Cloud Run service handles a request
- **THEN** it uses the same usecase layer as the standalone server
- **AND** it uses the same HTTP handlers from `internal/adapter/http`

#### Scenario: Environment-based configuration
- **WHEN** the Cloud Run service starts
- **THEN** it reads configuration from Cloud Run environment variables
- **AND** it uses `DATABASE_URL` to connect to Supabase PostgreSQL
- **AND** it uses the same `config.Load()` pattern as standalone server

### Requirement: Vercel Frontend Deployment
The system SHALL deploy the Next.js dashboard automatically via Vercel Git integration.

#### Scenario: Preview deployment on pull request
- **WHEN** a pull request is opened
- **THEN** Vercel creates a preview deployment
- **AND** the preview URL is posted as a PR comment
- **AND** the preview uses `NEXT_PUBLIC_API_URL` pointing to Cloud Run preview

#### Scenario: Production deployment on merge
- **WHEN** a pull request is merged to main branch
- **THEN** Vercel deploys to production automatically
- **AND** production uses `NEXT_PUBLIC_API_URL` pointing to Cloud Run production

### Requirement: Preview Environment Management
The system SHALL provide preview environments for both frontend and backend on every pull request.

#### Scenario: Synchronized preview environments
- **WHEN** a pull request is opened
- **THEN** Vercel creates a frontend preview
- **AND** GitHub Actions creates a Cloud Run backend preview via Terraform
- **AND** the frontend preview is configured to reach the backend preview
- **AND** both previews use the same PR number for identification

#### Scenario: Preview cleanup
- **WHEN** a pull request is closed or merged
- **THEN** Vercel deletes the frontend preview automatically
- **AND** GitHub Actions runs `terraform destroy` to delete the Cloud Run preview service
- **AND** Docker images for old previews are cleaned up

### Requirement: Supabase Database Integration
The system SHALL use Supabase PostgreSQL as the production database for Cloud Run deployments.

#### Scenario: Database connection
- **WHEN** the backend starts on Cloud Run
- **THEN** it connects to Supabase using `DATABASE_URL` environment variable
- **AND** it runs any pending migrations
- **AND** all repository operations use the Supabase connection

#### Scenario: Local development
- **WHEN** the backend runs locally
- **THEN** it uses SQLite via `DATABASE_PATH` environment variable
- **AND** no Supabase connection is attempted

#### Scenario: CORS configuration
- **WHEN** the dashboard makes API requests to Cloud Run
- **THEN** Supabase accepts requests from the Vercel domain
- **AND** preflight OPTIONS requests are handled correctly

#### Scenario: Supabase via Terraform
- **WHEN** Terraform applies infrastructure
- **THEN** Supabase project is created or updated via Terraform provider
- **AND** Database connection URL is output as `supabase_database_url`
- **AND** Supabase project ID is output for reference

## REMOVED Requirements

### Requirement: Vercel Serverless Go Function
**Reason**: Backend is now deployed to Google Cloud Run as a containerized service instead of Vercel serverless functions.
**Migration**: Use Docker container deployment with long-running process instead of serverless handler pattern.

### Requirement: Vercel Project Configuration (Monorepo)
**Reason**: Vercel now only deploys the frontend (Next.js dashboard). Backend is deployed separately to Cloud Run.
**Migration**: Update `vercel.json` to configure frontend-only deployment with API rewrites to Cloud Run backend URL.
