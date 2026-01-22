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

## Agent Development Setup

**Development Methodology**: TDD + BDD

All development must follow:
- **Test-Driven Development (TDD)**: Write tests first, implementation second
- **Behavior-Driven Development (BDD)**: Use Gherkin format for specifications
- **Checkbox Status**: Mark each scenario with state:
  - `[ ]` Not started
  - `[-]` In progress
  - `[x]` Completed

**Architecture**: Clean Architecture + Domain-Driven Design (DDD)

- Four-layer architecture: Domain → UseCase → Adapter → Infrastructure
- Repository pattern for data access abstraction
- Initial implementation: **In-Memory Repositories** (before database integration)
- **Interface Compliance**: Always add `var _ InterfaceName = (*StructName)(nil)` after imports when implementing interfaces for compile-time verification

**Development Workflow**:
1. Write Gherkin specs with checkbox status tracking
2. Design DDD domain models and boundaries
3. Write UseCase tests (test the business logic)
4. Implement UseCase with In-Memory repositories
5. **Run tests after every code change** - No changes proceed without test validation
6. Update checkbox states as progress advances
7. Graduate to database/external integration once logic is proven

**Test Execution Rules**:
- ✅ Run tests after every modification: `go test ./... -v`
- ✅ All tests must pass before proceeding to next task
- ✅ Test failures are blocking - fix immediately
- ✅ Show test results transparently to user
- ✅ Domain layer tests MUST pass
- ✅ UseCase layer tests MUST pass (using In-Memory repositories)