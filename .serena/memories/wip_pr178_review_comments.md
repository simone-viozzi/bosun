# PR #178 — Copilot Review Comments (M4 Scheduling Engine)

**PR**: https://github.com/simone-viozzi/bosun/pull/178
**Branch**: `012-scheduling-engine-runtime`
**Review by**: copilot-pull-request-reviewer (12 comments)
**Codecov**: 65.9% patch coverage, 251 lines missing

## Status: ALL 22 COMMENTS ADDRESSED ✅

- Round 1 (C01-C12): ✅ All 12 resolved and committed
- Round 2 (C13-C14): ✅ Fixed
- Round 3 (C15-C22): ✅ Fixed (2026-02-18)

## Review Comments

### CRITICAL — Empty Selector bugs (0 jobs discovered)

- [x] **C03** [daemon.go] Fixed: `ports.Selector{Prefixes: []string{dlabels.DefaultLabelPrefix}}` in `makeDiscoverFunc`
- [x] **C11** [daemon.go] Fixed: same fix in `discoverAndRegisterJobs`
- [x] **C12** [job_list.go] Fixed: same fix in `runJobList`

**Root cause**: `dockerlabels.FilterByPrefixes` returns empty map when Prefixes slice is empty.

### CRITICAL — context.Background() ignores shutdown

- [x] **C10** [scheduler.go] Fixed: added `ctx context.Context` + `cancel context.CancelFunc` fields to `Scheduler` struct; `New()` initialises with Background; `Start()` replaces with `context.WithCancel(ctx)`; `Stop()` calls `s.cancel()`; `makeJobFunc` and `skipIfRunning` use `s.ctx`.

### HIGH — Resource leak

- [x] **C01** [app.go / dockerlabels/source.go] Fixed: added `NewFromClient(*client.Client)` constructor; `Bootstrap` now passes the existing `dockerClient` instead of creating a second one via `NewFromEnv()`.

### HIGH — Validation gaps

- [x] **C04** [discoverer.go] Fixed: `jobs.ValidateOverlapPolicy()` called during label merge; invalid values emit `ValidationError`.
- [x] **C05** [scheduler.go] Fixed: `AddJob` calls `ValidateOverlapPolicy` and returns an error; silent "treat as queue" fallback removed.
- [x] **C06** [scheduler.go] Fixed: removed `cron.WithSeconds()` — scheduler now uses standard 5-field parser; all test cron expressions updated (`"0 0 * * * *"` → `"0 0 * * *"`, `"* * * * * *"` → `"@every 500ms"`).

### MEDIUM — Documentation & description mismatches

- [x] **C02** [job_list.go] Fixed: `Short` updated to "List jobs discovered from Docker labels".
- [x] **C07** [docs/config.md] Fixed: `bosun.job.enabled` → `Default: false`, `Required: No`.

### MEDIUM — Incomplete #141 resolution

- [x] **C08** [job_run.go] Fixed: added NOTE to existing TODO acknowledging migration to `app.Bootstrap()` is deferred beyond M4 scope.

### MEDIUM — Unlock panic risk

- [x] **C09** [stack_lock.go] Fixed: `UnlockAll` now deduplicates input via `dedupSort()` before unlocking.

## Round 2 — New Comments (2026-02-18)

### HIGH — Nil EventEmitter in executor

- [x] **C13** [job_run.go:248] Fixed: added `noopEventEmitter` in `internal/app/executor/noop_emitter.go`; `executor.New()` nil-guards `events` param so all existing `nil` callers are safe without call-site changes.

### MEDIUM — Goroutine leak in stack lock

- [x] **C14** [stack_lock.go:55] Fixed (doc): expanded comment on cleanup goroutine explaining the bounded risk — in practice `executeJob` always defers `UnlockAll`, so holders always unlock.

## Round 3 — New Comments (2026-02-18, second batch)

### HIGH — Scheduler start error ignored

- [x] **C15** [daemon.go:145] Fixed: `runWithSignalHandling` now selects on both `schedErr` and `sigCtx.Done()` — scheduler start failure returns the error immediately.

### HIGH — Graceful shutdown defeated

- [x] **C22** [scheduler.go:403] Fixed: `Start()` now returns `nil` on `ctx.Done()` without calling `Stop()`. Daemon's `runWithSignalHandling` owns the `Stop(shutdownCtx)` call with a fresh timeout context for proper graceful drain.

### MEDIUM — Duplicate event emission

- [x] **C16** [refresh.go:104] Fixed: removed `s.events.EmitJobRemoved(ctx, name)` from `ApplyRefresh` Removed loop — `RemoveJob` already emits it. Added clarifying comment.

### MEDIUM — Cron parser missing Descriptor support

- [x] **C17** [discoverer.go:33] Fixed: added `cron.Descriptor` to `cronParser` flags. `@every 10s`, `@daily`, `@hourly` etc. now pass label validation, matching `cron.New()` behaviour.

### MEDIUM — Enabled field default comment wrong

- [x] **C19** [types.go:37] Fixed: `Enabled` comment now says "Default: false. Containers without `bosun.job.enabled=true` not discovered."

### MEDIUM — Unlock still panic-prone

- [x] **C21** [stack_lock.go:99] Fixed (doc): removed false "idempotent" claim; comment now clearly states Unlock must be called exactly once per successful Lock and will panic on double-unlock. Channel-based refactor deferred.

### LOW — job_list Long description stale

- [x] **C18** [job_list.go:22] Fixed: added `- Target stacks` bullet; appended note that command does label discovery only, not daemon runtime state.

### LOW — #141 still claimed as resolved

- [x] **C20** [job_run.go:9] Fixed (PR metadata): PR body updated from "resolves #141" to "partially addresses #141; full migration of job_run.go deferred beyond M4 scope". No `Closes #141` in the Closes list.

## Coverage Gaps (Codecov)

| File | Coverage | Missing |
|------|----------|---------|
| internal/cmd/daemon.go | 38.7% | 85 + 2 partial |
| internal/app/scheduler/scheduler.go | 77.1% | 44 + 4 partial |
| internal/cmd/job_list.go | 57.6% | 38 + 1 partial |
| internal/app/app.go | 2.9% | 33 |
| internal/app/scheduler/refresh.go | 64.2% | 24 + 5 partial |
| internal/adapters/joblabels/discoverer.go | 15.4% | 10 + 1 partial |
| internal/app/concurrency/stack_lock.go | 93.4% | 2 + 2 partial |
