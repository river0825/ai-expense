## 1. Domain & Repository
- [x] 1.1 Define `Policy` struct in `internal/domain/models.go` (fields: ID, Key, Title, Content, Version, CreatedAt, UpdatedAt)
- [x] 1.2 Define `PolicyRepository` interface in `internal/domain/repositories.go` (GetByKey)

## 2. Infrastructure (Database)
- [x] 2.1 Create SQL migration for `policies` table
- [x] 2.2 Implement `PolicyRepository` in `internal/adapter/repository/sqlite`
- [x] 2.3 Add `PolicyRepo` to initialization logic

## 3. Use Case & Application Logic
- [x] 3.1 Create `GetPolicyUseCase` in `internal/usecase`
- [x] 3.2 Implement logic to retrieve latest version of policy

## 4. API & Interface
- [x] 4.1 Create `PolicyHandler` in `internal/adapter/http`
- [x] 4.2 Register route `GET /api/policies/{key}`
- [x] 4.3 Add integration tests for the endpoint

## 5. Data Seeding
- [x] 5.1 Create seed script or migration to insert initial "Privacy Policy" and "Terms of Use" content
