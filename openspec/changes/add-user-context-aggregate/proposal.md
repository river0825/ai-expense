# Change: Add user context aggregate for prompt personalization

## Why
User settings and categories are currently split across repositories, making it hard to build personalized AI prompts.

## What Changes
- Add a user context aggregate that includes core user settings (home currency, locale) and user categories.
- Expose a use case to load this aggregate by user_id for downstream prompt building.
- Update AI prompt construction to accept user context when available.

## Impact
- Affected specs: user-profile
- Affected code: internal/domain/models.go, internal/usecase/*, internal/ai/*, internal/adapter/repository/*
