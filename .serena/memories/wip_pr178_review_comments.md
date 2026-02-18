# PR #178 — Copilot Review Comments (M4 Scheduling Engine)

**PR**: https://github.com/simone-viozzi/bosun/pull/178
**Branch**: `012-scheduling-engine-runtime`
**Review by**: copilot-pull-request-reviewer (12 comments)
**Codecov**: 65.9% patch coverage, 251 lines missing

## Status: ALL FIXES COMPLETE ✅

All 12 review comments resolved. `go build ./...` and `go test ./internal/...` pass clean.

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
