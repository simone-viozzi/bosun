# WIP: General smell scan — milestone3

## Scope
- Scope label: milestone3 general discovery
- Included paths:
  - `internal/app/executor/`
  - `internal/app/planner/`
  - `internal/adapters/docker/compose/`
  - `internal/adapters/docker/worker/`
  - `internal/adapters/dockerlabels/`
  - `internal/adapters/joblabels/`
  - `internal/config/`
  - `internal/cmd/`
  - `internal/domain/jobs/`
  - `internal/ports/`
  - `integration/`
  - `internal/testutil/`
- Diff-only: No — broad scan within the above paths
- Files inspected (representative): `internal/cmd/plan_show.go`, `internal/cmd/plan_list.go`, `internal/app/executor/executor.go`, `internal/config/loader/loader.go`, `internal/config/schema/config_v1.go`, `internal/adapters/docker/worker/runner.go`, `integration/job_execution_test.go`, `internal/testutil/harness.go`

---

## Findings
(Each finding contains short evidence and remediation direction.)

1) **Confusing config property name: `BackupEnabled` (naming)**
- Location: `internal/config/schema/config_v1.go`, tests & docs (`internal/config/schema/config_v1_test.go`, `docs/config.md`).
- Evidence: field `BackupEnabled bool` and label `bosun.volume.backupEnabled` (also used in test data e.g. `internal/testutil/compose/validate-valid.yaml`).
- Why it's a smell: name is ambiguous (singular vs plural, scope implied by key); issue #134 mentions this confusion.
- Remediation: consider renaming to `BackupsEnabled` or `VolumeBackupsEnabled` (API+docs change) and add a compatibility note/alias in loader to avoid breaking labels.
- Dependencies: backward compatibility concerns; confirm preferred naming with product.

2) **Duplicate/duplicative CLI error printing (UX & duplication)**
- Location: `internal/cmd/*.go` (e.g., `plan_list.go`, `plan_show.go`, `job_run.go`) use the pattern `RunE` printing `fmt.Fprintf(os.Stderr, "Error: %v\n", err)` and lower-level code also returns formatted error messages.
- Evidence: `plan_show.go` RunE prints error and `runPlanShow` returns errors already formatted for display.
- Why it's a smell: risk of duplicated messages or inconsistent error output; harder to centralize i18n/context.
- Remediation: centralize CLI error handling (single place to format + print), standardize whether commands should return errors with user-facing text vs machine-friendly errors.
- Questions: do we want machine-friendly errors (structured) vs free-form strings? (non-blocking)

3) **Strict unknown-key rejection may be surprising (config validation UX) — possible source of #136**
- Location: `internal/config/loader/loader.go` (`FromLabels`), `internal/config/loader/errors.go`.
- Evidence: `FromLabels` filters `bosun.*` labels and calls `errs.AddUnknownKey(key, scope)` when `spec.Get(key)` is missing; tests expect unknown-key errors in `loader_test.go`.
- Why it's a smell: users reported unexpected unknown keys like `bosun.other`, `bosun.network` failing validation; rejecting any unknown `bosun.` key makes the loader brittle to auxiliary labels.
- Remediation: consider:
  - Allowlist of recognized subscopes or patterns; or
  - Provide clearer validation error messaging suggesting typos and nearby keys; or
  - Add a mode to tolerate unknown labels (warning vs error).
- Dependencies: specification of what labels are "first-class" vs allowed extensions.
- Question: Are `bosun.other` / `bosun.network` meant to be valid extension points? (blocking)

4) **Incomplete/Confusing Executor API & unused params (dead code / API mismatch)**
- Location: `internal/app/executor/executor.go`
- Evidence: `New(... discoverer ports.JobDiscoverer, ...)` takes `discoverer` but does not use it; methods `Execute` and `DryRun` return "not implemented" while `ExecuteJob` and `DryRunJob` are provided.
- Why it's a smell: mismatch between interface `ports.JobExecutor` and implementation leads to confusion; unused params suggest leftover code or incomplete refactor.
- Remediation: either implement `Execute`/`DryRun` to satisfy interface or remove those methods and adapt types; remove unused constructor params or use them.
- Confidence: high (clear unused param and unimplemented methods).

5) **Missing/placeholder feature: restart/start step is "not yet implemented" in plan display and tests are placeholders (#135)**
- Location: `internal/cmd/plan_show.go` (Long text + `renderPlanTextOutput` prints `(Start step - not yet implemented)`), integration tests (`integration/job_execution_test.go`) have placeholders for worker-failure behavior.
- Evidence: Long help text mentions restart is future milestone; plan output shows message; `Test_Integration_JobExecution_WorkerFailure` is a placeholder that doesn't simulate a failing worker.
- Why it's a smell: user-visible feature gap + test coverage gap; may lead to incorrect expectations.
- Remediation: implement StartContainers step behavior in planner/executor and add integration test that simulates worker failure and verifies stack restart.
- Related issue: #135 (missing "restart stack" step in execution plan display).

6) **Worker runner silent error swallowing and magic values (error handling)**
- Location: `internal/adapters/docker/worker/runner.go`
- Evidence: `streamLogs` returns silently on error (no logging or returned error); `stopContainer` returns hardcoded `137` on inspect error.
- Why it's a smell: important runtime errors are dropped, magic numeric codes reduce readability and correctness.
- Remediation: return/propagate log streaming errors or at least log them; replace magic constants with named constants and surface errors where appropriate.

7) **Duplication: JSON/YAML serialization and human output duplicated across commands (duplication, DRY)**
- Location: `internal/cmd/plan_list.go`, `plan_show.go`, `snapshot.go`, `job_run.go` (`render*Output` / `render*JSONOutput` / `render*YAMLOutput` repeat pattern).
- Evidence: many files call `json.NewEncoder(os.Stdout)` / `yaml.NewEncoder(os.Stdout)` and format the same indenting and nil-slice handling.
- Why it's a smell: duplicated serialization logic increases maintenance burden and risk of inconsistencies.
- Remediation: introduce small output helpers (serializeJSON/serializeYAML) or a printer abstraction used by commands.

8) **Test housekeeping & brittle integration harness (test smell)**
- Location: `internal/testutil/harness.go` (TODO), `integration/*` tests rely on local `docker compose` CLI.
- Evidence: TODO: use testcontainers-go; tests call external `docker compose` and rely on environment (may be flaky in CI).
- Why it's a smell: brittle/integration tests may fail on CI or developer machines; there's a TODO acknowledging it.
- Remediation: migrate to deterministic harness (testcontainers or dedicated docker-in-docker fixture) and add barrier checks for environment assumptions.

9) **TODOs without issue links / unresolved decisions (process)**
- Location: multiple TODO comments (e.g., `internal/config/loader/parse.go`, `internal/app/executor/executor.go`, `internal/testutil/harness.go`)
- Evidence: spec `007-milestone-2-5-polish` requires TODOs to be converted to GitHub issues; some TODOs are present and not linked.
- Why it's a smell: technical debt tracking and discoverability suffers if TODOs are not triaged.
- Remediation: convert critical TODOs to issues or resolve them; for every TODO add TODO comments referencing `wip_smell-general-m3` or create targeted issues.

10) **Type/enum suggestions (clarity)**
- Location: `internal/domain/jobs/types.go` (`Mode` as string with values `"ro"`/`"rw"`), `internal/adapters/docker/compose/controller.go` (string state comparisons like `cnt.State != "running"`).
- Evidence: usage of magic string values in multiple places.
- Why it's a smell: increases risk for typos and inconsistent comparisons; enums/typed constants increase clarity.
- Remediation: introduce typed constants/enums for Mode and container states.

---

## Questions for user
- [blocking] Are `bosun.other` and `bosun.network` intended extension labels that should be tolerated (warnings) instead of rejected as unknown keys? (affects remediation for finding #3)
- [non-blocking] Preferred resolution for `BackupEnabled`: rename to `BackupsEnabled` or `VolumeBackupsEnabled`, or keep but add clearer docs/alias? (affects finding #1)
- [non-blocking] Should we standardize CLI error formatting (structured vs plain text)? (affects finding #2)
- [non-blocking] Do you want the executor to strictly implement `ports.JobExecutor` now (implement `Execute`/`DryRun`) or is M3 intentionally using `ExecuteJob`/`DryRunJob`? (affects finding #4)

---

## Confidence / Notes
- Confidence is high for code-grounded findings (I referenced exact files and test gaps).
- I intentionally did not propose code edits here; once triage decisions are made I can add TODO anchors in code pointing to the specific smell entry.

## Suggested next scouts
- Duplication-scout: consolidate JSON/YAML rendering and common CLI output code. (finding #7)
- Testing-scout: add an integration test simulating worker failure to validate restart behavior and replace CLI-based harness with testcontainers. (findings #5, #8)
- Design/Layering-scout: resolve Executor API mismatch and unused params, and standardize error propagation across layers. (findings #2, #4)

---

## META feedback (delete once stable)
- What happened: The repo contains many TODO markers and placeholders but not all are linked to issues.
- Why it's a problem: Hard to prioritize and track technical debt.
- Proposed change: Enforce converting TODOs into GitHub issues or include an `issue=` tag in the TODO comment.
