# AIExpense - Conversational Expense Tracking System

A frictionless expense tracking bot that operates through natural language conversation. Users chat with bot on LINE (with support for Telegram and other messengers in future) to log expenses and generate reports.

## 🚀 Quick Start

### Local Development
```bash
# Run server with SQLite (default)
./server

# Run server with environment variables
DATABASE_PATH=./aiexpense.db \
ENABLED_MESSENGERS=terminal \
GEMINI_API_KEY=your_gemini_key \
SERVER_PORT=8080
```

## 🏗️ Architecture

```
┌─────────────────────┐
│   GitHub Repo        │
│   (PRs + main)      │
└─────────────────────┘
         │
         ├── CI/CD Pipeline (GitHub Actions)
         │   ├── Backend: test, build
         │   ├── Next.js Dashboard
         │   └─────────────┘
         │
         ├── Backend (Google Cloud Run)
         │   │   - Go Container Service
         │   └─────────────┘
         │
         │   Database (Supabase)
         │   └─────────────────────┘
```

### Deployment Platforms

| Platform | Purpose | Free Tier |
|---------|----------|-----------|
| **Frontend**: Vercel | Next.js 14 Dashboard with automatic preview deployments |
| **Backend**: Google Cloud Run | Go containerized service with scale-to-zero (240K vCPU-sec + 450K GiB-sec free) |
| **Database**: Supabase | Managed PostgreSQL (500MB free tier) |

### Infrastructure as Code

All cloud infrastructure is defined via Terraform in `terraform/`:

```hcl
terraform/
├── main.tf                    # Root config with providers
├── variables.tf                # Input variables
├── outputs.tf                 # Output values
├── .gitignore                 # State files exclusion
├── gcp/
│   ├── provider.tf            # Google provider
│   ├── cloud_run.tf          # Cloud Run services
│   ├── artifact_registry.tf  # Docker image repository
│   ├── iam.tf                # Service accounts and IAM
│   └── supabase/
│   ├── provider.tf            # Supabase provider
│   └── database.tf            # Project and database definitions
```

### Environment Variables

#### Required Secrets (GitHub Repository Settings → Secrets)

```yaml
secrets:
  GCP_SA_KEY: [your GCP service account key JSON]
  GCP_PROJECT_ID: [your GCP project ID]
  GCP_REGION: us-central1
  SUPABASE_ACCESS_TOKEN: [your Supabase access token]
  LINE_CHANNEL_TOKEN: [your LINE channel token]
  GEMINI_API_KEY: [your Gemini API key]
  ADMIN_API_KEY: [optional admin key for metrics]
```

#### Terraform Variables

```bash
# Development (SQLite local)
terraform plan -var-file=terraform/dev.tfvars -var="environment=dev"

# Production (Supabase PostgreSQL)
terraform plan -var-file=terraform/prod.tfvars -var="environment=prod"

# Preview (PR environments)
terraform plan -var-file=terraform/preview.tfvars -var="environment=preview"
```

### CI/CD Workflow

**Location:** `.github/workflows/deploy.yml`

**Pipeline Stages:**
1. **CI Tests** - Backend and Dashboard lint, build, and tests on every PR
2. **Terraform Apply** - Create preview/prod infrastructure for PR, or apply production on main merge
3. **Docker Build & Push** - Build and push Docker image to GCP Artifact Registry
4. **Cloud Run Deploy** - Deploy Docker image from Artifact Registry to Cloud Run service
5. **Vercel Deploy** - Automatic via Vercel Git integration

### Deployment Flow

```
Pull Request Open → Preview Environment Created → Ready for Testing
  │   │
  ├─────────────┐
  │   │   GitHub Actions
  │   │   ├─────────────┘
  │   │
  │   │   Vercel (frontend preview)
  │   │   └─────────────┘
  │   │   Google Cloud Run (backend preview)
   │   │   └─────────────┘
   │   │   Supabase (preview database)
   │   │   └─────────────┘
  │   └─────────────┘
  │
  └─────────────┘
  │   └─────────────┘
  │
Pull Request Merge → Production Environment Updated
  │   │   GitHub Actions
   │   │   ├─────────────┐
  │   │   │   Vercel (frontend production)
   │   │   └─────────────┘
   │   │   Google Cloud Run (backend production)
   │   │   └─────────────┘
   │   │   Supabase (production database)
   │   │   └─────────────┘
   │   └─────────────┘
   │
   └─────────────┘
```
Pull Request Close → Preview Environment Destroyed

### Next Steps (for you)

1. **Set up GCP Project** (if not exists)
   - Create new GCP project or use existing
   - Enable Cloud Run API
   - Enable Artifact Registry API

2. **Set up Supabase Project** (if not exists)
   - Create new project via Supabase
   - Get project ID and database URL
   - Run database migrations

3. **Configure GitHub Secrets**
   - Add all required secrets listed above

4. **Test Deploy** (optional but recommended)
   - Create a test PR
   - Verify preview and production deployments work correctly

5. **Start Production Deploy** (when ready)
   - Merge to main branch
   - Monitor first production deploy
   - Run smoke tests via dashboard

## 🔧 Configuration

### Database Setup

The application uses **versioned migrations** via [golang-migrate](https://github.com/golang-migrate/migrate) for both SQLite and PostgreSQL:

#### SQLite (Local Development - Default)
```bash
# Database is created automatically at ./aiexpense.db
# Migrations run on startup
DATABASE_PATH=./aiexpense.db go run ./cmd/server/main.go
```

#### PostgreSQL with Local Supabase (Production-like Testing)
```bash
# Install Supabase CLI
brew install supabase/tap/supabase

# Start local Supabase instance
supabase start

# Get connection string
supabase status

# Run with PostgreSQL
DATABASE_URL="postgresql://postgres:postgres@localhost:54322/postgres" \
go run ./cmd/server/main.go
```

For detailed database setup instructions, see:
- [Database Migrations Guide](./docs/migrations.md) - How migrations work and how to create new ones
- [Local Supabase Setup](./docs/supabase-local-setup.md) - How to set up local PostgreSQL testing

### Repository Factory Pattern

The system uses a repository factory pattern to choose between SQLite (local) and PostgreSQL (production):

- **SQLite**: Used for local development (`DATABASE_PATH` set)
- **PostgreSQL**: Used for Cloud Run/Supabase deployments (`DATABASE_URL` set)

This selection happens at runtime via environment variables - no code changes needed.

## 📦 Testing

### Unit Tests
```bash
# Backend tests
go test -v ./...

# Dashboard tests
cd dashboard && bun install && bunx playwright test
```

### E2E Tests
```bash
# Starts backend
go build -o server ./cmd/server
./server &

# Runs Playwright tests
cd dashboard && bunx playwright test
```

## 📖 Reference

### Terraform Commands
```bash
# Development
terraform plan -var-file=terraform/dev.tfvars -var="environment=dev"

# Production
terraform plan -var-file=terraform/prod.tfvars -var="environment=prod"

# Preview (PR environments)
terraform plan -var-file=terraform/preview.tfvars -var="environment=preview"
terraform apply -auto-approve -var-file=terraform/preview.tfvars -var="environment=preview"

# Destroy preview (on PR close)
terraform destroy -auto-approve -var-file=terraform/preview.tfvars -var="environment=preview"
```
