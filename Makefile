PROJECT ?= $(shell gcloud config get-value project 2>/dev/null)
REGION ?= $(strip $(shell gcloud config get-value run/region 2>/dev/null))
ifeq ($(REGION),)
REGION := us-central1
endif
SERVICE ?= aiexpense
REPO ?= aiexpense-backend
IMAGE_NAME ?= aiexpense
TAG ?= latest
IMAGE ?= $(REGION)-docker.pkg.dev/$(PROJECT)/$(REPO)/$(IMAGE_NAME)
PLATFORM ?= linux/amd64
DOCKER ?= docker
GCLOUD ?= gcloud
ENV_FILE ?= .env.prod
BUILD_CONTEXT ?= .
LOG_LIMIT ?= 100
LOG_FRESHNESS ?= 1h
LOG_ORDER ?= desc
PRICING_PROVIDER ?= gemini

.PHONY: image push deploy tail update-price
image:
	$(DOCKER) buildx build --platform $(PLATFORM) -t $(IMAGE):$(TAG) --load $(BUILD_CONTEXT)

push:
	$(DOCKER) buildx build --platform $(PLATFORM) -t $(IMAGE):$(TAG) --push $(BUILD_CONTEXT)

deploy: push
	@set -e; tmpfile=$$(mktemp); trap 'rm -f $$tmpfile' EXIT; \
	grep -v '^[[:space:]]*#' $(ENV_FILE) | sed '/^[[:space:]]*$$/d' | \
	while IFS= read -r line; do \
		key=$${line%%=*}; val=$${line#*=}; \
		val=$${val//\\/\\\\}; val=$${val//\"/\\\"}; \
		printf '%s: "%s"\n' "$$key" "$$val"; \
	done > $$tmpfile; \
	$(GCLOUD) run deploy $(SERVICE) --image $(IMAGE):$(TAG) --region $(REGION) --platform managed --allow-unauthenticated --env-vars-file $$tmpfile

tail:
	$(GCLOUD) logging read 'resource.type="cloud_run_revision" AND resource.labels.service_name="$(SERVICE)" AND resource.labels.location="$(REGION)"' --project $(PROJECT) --limit $(LOG_LIMIT) --freshness $(LOG_FRESHNESS) --order $(LOG_ORDER) --format='table(timestamp, severity, textPayload, jsonPayload.message)'

update-price:
	@set -e; \
	admin_key=$${ADMIN_API_KEY:-$$(grep -v '^[[:space:]]*#' $(ENV_FILE) | sed '/^[[:space:]]*$$/d' | grep '^ADMIN_API_KEY=' | head -n 1 | cut -d= -f2-)}; \
	if [ -z "$$admin_key" ]; then echo "ADMIN_API_KEY not set (or missing in $(ENV_FILE))" >&2; exit 1; fi; \
	service_url=$${SERVICE_URL:-$$( $(GCLOUD) run services describe $(SERVICE) --region $(REGION) --format='value(status.url)' )}; \
	if [ -z "$$service_url" ]; then echo "SERVICE_URL not found; set SERVICE_URL or check Cloud Run service." >&2; exit 1; fi; \
	curl -sS -X POST "$$service_url/api/pricing/sync?provider=$(PRICING_PROVIDER)" -H "X-API-Key: $$admin_key" -H "Content-Type: application/json"

deploy-fe:
	cd dashboard && vercel --prod --build-env NEXT_PUBLIC_API_URL=https://aiexpense-996531141309.us-central1.run.app --env NEXT_PUBLIC_API_URL=https://aiexpense-996531141309.us-central1.run.app
