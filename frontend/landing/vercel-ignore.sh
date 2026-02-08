#!/usr/bin/env sh
set -eu

# Build on first deployment when previous SHA is unavailable.
if [ -z "${VERCEL_GIT_PREVIOUS_SHA:-}" ] || [ "${VERCEL_GIT_PREVIOUS_SHA:-}" = "0000000000000000000000000000000000000000" ]; then
  exit 1
fi

# Only production branch should deploy.
if [ "${VERCEL_GIT_COMMIT_REF:-}" != "main" ]; then
  exit 0
fi

# Only deploy when frontend/landing changed.
if git diff --quiet "${VERCEL_GIT_PREVIOUS_SHA}" "${VERCEL_GIT_COMMIT_SHA}" -- frontend/landing; then
  exit 0
fi

exit 1
