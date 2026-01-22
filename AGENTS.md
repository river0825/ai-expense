<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

# AIExpense Agent Guide

## Scope
- Backend: Go 1.21+ (Clean Architecture, DDD)
- Frontend: Next.js 14 dashboard in `dashboard/`
- Specs: OpenSpec in `openspec/`
- No Cursor or Copilot rules detected (.cursor/rules/, .cursorrules, .github/copilot-instructions.md)

## Build, Lint, Test Commands

### Backend (Go)
- Build: `go build ./cmd/server`
- Build named binary: `go build -o aiexpense ./cmd/server`
- Run server (manual): `go run ./cmd/server`

### Frontend (Dashboard)
- **Package Manager**: Use `bun` (NOT npm)
- Install: `cd dashboard && bun install`
- Dev server: `cd dashboard && bun run dev`
- Build: `cd dashboard && bun run build`
- Start prod: `cd dashboard && bun start`
- Lint: `cd dashboard && bun run lint`

### E2E Testing (Playwright)
- Install browsers: `cd dashboard && bunx playwright install chromium`
- Run tests: `cd dashboard && bunx playwright test`
- Run with UI: `cd dashboard && bunx playwright test --ui`

### Tests (Backend)
- All tests: `go test ./...`
- Verbose: `go test -v ./...`
- With race: `go test -race ./...`
- Coverage: `go test -cover ./...`

### Single Test / Focused Runs
- Single package: `go test ./internal/ai -v`
- Single test: `go test ./internal/ai -run TestParseExpenseRegex -v`
- Focused usecase: `go test ./internal/usecase -run AutoSignup -v`
- Usecase parse: `go test ./internal/usecase -run ParseConversation -v`

### Benchmarks and Performance
- Run all benchmarks: `go test -bench=. -benchmem ./test/bench/...`
- Single benchmark: `go test -bench=BenchmarkCreateExpense -benchmem ./test/bench/...`
- Longer runs: `go test -bench=. -benchmem -benchtime=10s ./test/bench/...`

## Architecture and Design Rules
- Clean Architecture layers: Domain -> UseCase -> Adapter -> Infrastructure
- Domain and usecase code should not depend on adapters or frameworks
- Repositories are interfaces in domain, implementations live in adapters
- Prefer in-memory repositories during early feature development

## TDD + BDD Requirements (Mandatory)
- Write Gherkin specs with checkbox status tracking:
  - `[ ]` Not started
  - `[-]` In progress
  - `[x]` Completed
- Write usecase tests before implementation
- Run `go test ./... -v` after every code change
- Test failures are blocking; fix immediately

## OpenSpec Workflow (When Required)
- New capability, breaking change, architecture change, performance/security work:
  - Create OpenSpec change proposal first
  - Do not implement until proposal is approved
- Use OpenSpec format for requirements and scenarios:
  - Requirements MUST have at least one `#### Scenario:` block

## Go Code Style
- Format with `gofmt` (standard Go formatting expected)
- Naming: exported types use PascalCase, unexported use camelCase
- Interfaces are named by capability (e.g., `UserRepository`, not `IUserRepo`)
- Prefer explicit error checks; no silent failures
- Use pointers only when optionality is required (e.g., nullable fields)
- Keep domain models in `internal/domain`
- **Always add interface compliance checks**: When implementing an interface, add `var _ InterfaceName = (*StructName)(nil)` after imports to ensure compile-time verification

## Imports
- Group imports by standard library, external, then internal
- Use fully qualified internal paths like `github.com/riverlin/aiexpense/internal/...`

## Error Handling and API Responses
- Standard Go pattern: `if err != nil { ... }`
- HTTP errors return JSON:
  - `{"status":"error","error":"message"}`
- Success responses return JSON:
  - `{"status":"success","data":...}` or `{"status":"success","message":"..."}`
- Validate required fields in handlers and return 400 on missing input

## Testing Conventions
- Tests are table-driven where appropriate
- Use in-memory mocks in unit tests
- Prefer fast, deterministic tests (no external APIs)
- Known test packages: `internal/ai`, `internal/usecase`

## Dashboard (Next.js) Conventions
- TypeScript strict mode enabled (`dashboard/tsconfig.json`)
- Tailwind CSS with design tokens via CSS variables
- Linting uses `next lint`

## Documentation Sources
- Backend overview: `README.md`
- Dashboard overview: `dashboard/README.md`
- Testing details: `TESTING.md`, `INTEGRATION_TESTS.md`
- Performance benchmarks: `PERFORMANCE.md`
- Monitoring guide: `PHASE_20_MONITORING_GUIDE.md`

## Notes
- No dedicated Go linter config found (.golangci.yml not present)
- Prefer minimal, focused changes; avoid refactors during bugfixes
