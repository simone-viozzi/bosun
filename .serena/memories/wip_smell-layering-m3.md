# WIP: Layering / Architecture Smells — milestone3

## Scope
- Scope label: `milestone3` layering scan
- Included paths: `internal/app/`, `internal/adapters/`, `internal/domain/`, `internal/ports/`, `internal/config/`, `internal/cmd/`
- Excluded: other top-level packages (outside milestone3 scope)
- What I inspected (representative): `internal/cmd/job_run.go`, `internal/cmd/plan_list.go`, `internal/cmd/plan_show.go`, `internal/cmd/validate.go`, `internal/app/executor/*`, `internal/adapters/docker/*`, `internal/adapters/dockerlabels/*`, `internal/adapters/joblabels/*`, `internal/ports/*`, `internal/domain/*`, small set of tests and config loader files.

---

## Findings

### 1) CLI imports adapters directly (Layering violation)
- Location(s): `internal/cmd/job_run.go`, `internal/cmd/plan_list.go`, `internal/cmd/plan_show.go`, `internal/cmd/snapshot.go`, `internal/cmd/validate.go`.
- Evidence (examples):
  - `internal/cmd/job_run.go` imports:
    - `github.com/simone-viozzi/bosun/internal/adapters/docker/compose`
    - `github.com/simone-viozzi/bosun/internal/adapters/docker/worker`
    - `github.com/simone-viozzi/bosun/internal/adapters/dockerlabels`
    - `github.com/simone-viozzi/bosun/internal/adapters/joblabels`
  - `internal/cmd/plan_list.go` imports `internal/adapters/dockerlabels` and `internal/adapters/joblabels`.
- Why it’s a smell: CLI directly depending on concrete adapters increases coupling, duplicates wiring across commands, and leaks adapter composition details into the presentation layer. It makes testing/DI harder and hinders future changes to adapter boundaries.
- Remediation direction: Move adapter wiring into the application layer (or a dedicated bootstrap package). Provide higher-level app APIs (e.g., `app.NewExecutorWithDefaults()` or `app.NewCLIService()`) so CLI only calls app-level constructors or service functions and doesn't import adapter packages directly.
- Dependencies: If project convention is to keep wiring in `internal/cmd` (small scope), this may be an intentional trade-off — confirm with team.
- Relationship to other findings: This is the root cause of duplicated discovery code and CLI business logic (see Finding #2).

---

### 2) Business logic & discovery work in CLI handlers (Misplaced responsibilities)
- Location(s): `internal/cmd/job_run.go`, `internal/cmd/plan_list.go`, `internal/cmd/plan_show.go`, `internal/cmd/validate.go`.
- Evidence (examples):
  - `job_run.go` creates Docker client, label source (`dockerlabels.NewFromEnv()`), runs `Snapshot()`, `joblabels.NewDiscoverer().DiscoverJobs()`, then selects the job by name.
  - `plan_list.go` performs snapshot -> discover -> sort -> render logic.
  - `validate.go` performs snapshot loading, loader.FromLabels() per entity, merging and result formatting.
- Why it’s a smell: These are application/domain-level behaviors (discovery, selection, config merging) that should be encapsulated in app/service packages. Having them in CLI duplicates code across commands, reduces testability, and means the CLI contains a lot of business logic rather than just presentation/wiring.
- Remediation direction: Introduce app-level service functions (e.g., `app.DiscoverJobs(ctx, selector)`, `app.FindJob(ctx, selector, name)`, `app.ValidateSnapshot(ctx, opts)`), and move the logic there. Make CLI call these services and only handle UI concerns (flags, output formatting, exit codes).
- Dependencies: Requires defining a small app interface to expose discovery/validation operations and deciding where default adapter factories live.

---

### 3) CLI performs low-level dependency creation (Docker client, adapters) and wiring
- Location(s): `internal/cmd/job_run.go` (creates `client.NewClientWithOpts` and constructs `compose.NewController(...)`, `worker.NewRunner(...)`, `executor.New(...)`).
- Why it’s a smell: Wiring in multiple commands means duplication and inconsistent lifecycle/DI. Centralizing wiring makes it easier to swap implementations and to introduce test doubles.
- Remediation direction: Provide a single place (app constructor or bootstrap) to create default dependencies and return an app object suitable for CLI use.

---

### 4) Positive: Adapters do NOT depend on other adapters (OK)
- Location(s): inspected adapter packages `internal/adapters/docker/...`, `internal/adapters/dockerlabels`, `internal/adapters/joblabels`.
- Evidence: Adapter files import `internal/domain/*`, `internal/ports` and `internal/config/schema` where needed; I did not find imports from one adapter package to another.
- Why it’s good: Adapters are implemented as proper outward-facing modules against port interfaces, preserving hexagonal boundaries.
- Remediation: None.

---

### 5) Positive: Domain does NOT import adapters (OK)
- Evidence: No matches for imports of `internal/adapters` in `internal/domain/**` (search found only test import of `internal/config/schema`).
- Why it’s good: Domain layer remains independent of infrastructure.

---

### 6) Config knowledge leaking into domain — limited / test-only
- Evidence: `internal/domain/jobs/types_test.go` imports `internal/config/schema` (test only).
- Why it’s low-risk: The import is in tests; production domain code doesn't import config packages. Keep tests as-is or refactor test helpers if desired.

---

### 7) Circular dependencies
- Evidence: Quick import scans did not reveal direct circular imports across the scoped packages. No packages in `internal/domain` import `internal/cmd`, adapters do not import CLI or app in a way that looks cyclic.
- Confidence: Medium — I looked at package imports and representative files; a full `go list` style cycle check would raise confidence to high.

---

## Questions for user
- [blocking] Do you want the project rule to be: "CLI must not import adapter packages; CLI only calls app-level services and a single bootstrap"? (Yes/No) — this determines how I recommend changes.
- [non-blocking] If yes, where should the default wiring live: (a) `internal/app` (e.g., a factory like `app.NewWithDefaults()`), (b) `internal/bootstrap` package, or (c) `cmd/bosun` main only? Pick one preference.
- [non-blocking] Are you willing to accept a short-term trade-off where CLI still wires adapters for operational simplicity in M3, with a plan to refactor later? (Yes/No)
- [non-blocking] Do you want me to add linter / import-check tests to prevent future CLI→Adapter imports automatically? (Yes/No)

---

## Confidence / Notes
- Confidence: High for the CLI→Adapter import smell and the presence of business logic in CLI commands. Medium for absence of circular dependencies (did not run `go list -json -deps` but scanned imports and files in-scope).
- What I could not fully verify within scope: a programmatic check for package import cycles across the entire module (requires `go list` outside of this scan). If you want, I can run that next.

---

## Suggested next scout(s)
1. Refactor Scout (implementation): Implement/apply the chosen boundary rule — create `app` factory methods to centralize wiring, update CLI to use those, and move discovery/selection/validation logic into app packages.
2. Linter Scout (prevent regressions): Add an import lint rule (e.g., `revive`, `golangci-lint` config or a small `go vet` checker) to forbid `internal/cmd` → `internal/adapters` imports.
3. Cycle-check Scout (verification): Run `go list`/`go mod` based cycle checks across the module to assert there are no hidden circular deps.

---

## META feedback (optional)
- What happened: The codebase currently lets CLI directly import adapters, leading to duplicated wiring and logic in multiple commands.
- Why it’s a problem: Harder to maintain and test; makes future adapter replacement or dependency inversion harder.
- Proposed change: Add a short guideline in repo docs: "CLI should use app-level services; wiring lives in `internal/app` or `cmd/main.go` (pick one)" and add a linter rule to enforce it.


*WIP — file will be updated if I find more instances or after clarifications.*
