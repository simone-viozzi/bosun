# WIP: Duplication Smell Scan — milestone3

## Scope
- Scope label: milestone3 duplication scan
- Included paths (per request):
  - `internal/app/`
  - `internal/adapters/`
  - `internal/config/`
  - `internal/cmd/`
  - `internal/domain/`
  - `internal/ports/`
  - `integration/`
  - `internal/testutil/`
- Diff-only: No (full-scan across the listed paths)
- What I inspected: CLI commands under `internal/cmd/` (notably `plan_list.go`, `plan_show.go`, `job_run.go`, `validate.go`, `snapshot.go`), adapters creating Docker clients, `internal/tools` JSON output helpers, and integration tests using `internal/testutil`.

---

## Findings

### 1) Repeated JSON/YAML output handling across CLI commands
- **Location(s):**
  - `internal/cmd/plan_list.go` (`renderJSONOutput`, `renderYAMLOutput`) — uses `json.NewEncoder(os.Stdout)` with `SetIndent` and `yaml.NewEncoder(os.Stdout)` + `SetIndent`.
  - `internal/cmd/plan_show.go` (`renderPlanJSONOutput`, `renderPlanYAMLOutput`) — same encoder & indent usage.
  - `internal/cmd/snapshot.go` (`runSnapshot`) — `json.NewEncoder(os.Stdout)` with `SetIndent`.
  - `internal/cmd/job_run.go` (`printDryRunJSON`) — `json.NewEncoder(os.Stdout)` with `SetIndent`.
  - `internal/cmd/validate.go` (`outputResults`) — prints merged config using a `json.NewEncoder(os.Stdout)` + `SetIndent` when `--print`.
  - Tools: `internal/tools/configdoc/jsonschema.go` uses `json.MarshalIndent`.
- **Evidence:** multiple calls to `json.NewEncoder(os.Stdout); enc.SetIndent("", "  "); enc.Encode(...)` and `yaml.NewEncoder(os.Stdout); enc.SetIndent(2); enc.Encode(...)`. Also repeated normalization to ensure empty slices (e.g., `if output.Jobs == nil { output.Jobs = []jobs.Job{} }`).
- **Why it's a smell:** duplicated encoder setup and formatting logic leads to inconsistent formatting behavior and duplicated test coverage for the same behavior. When tweaks are needed (indent, nil-slice normalization, YAML tag options), changes must be replicated across many files.
- **Remediation direction:** add a shared package (e.g., `internal/ui/format` or `internal/cmd/output`) exposing small helpers like `PrintJSON(w io.Writer, v interface{}) error`, `PrintYAML(w io.Writer, v interface{}) error`, plus a helper `EnsureNilSlices` or a serialization wrapper that normalizes types for consistent `[]` vs `null` behavior. Update CLI commands to call these helpers.
- **Dependencies:** Ensure chosen helper signatures are generic enough (accept io.Writer + interface{}), and consider common options (indent width) as package-level defaults or config.

### 2) Duplicated CLI flag setup & format validation
- **Location(s):**
  - `internal/cmd/plan_list.go`, `plan_show.go`, `job_run.go`, `snapshot.go` — repeated flags: `--format`, `--stopped`, `--project`, `--stack` (some with short `-f`).
  - Format validation logic duplicated: checks like `format = strings.ToLower(format); if format != "text" && format != "json" && format != "yaml" { ... }` appear in multiple commands.
- **Evidence:** repeated `cmd.Flags().StringVarP(&format, "format", "f", "text", ...)` and `format` validation code in `run*` functions.
- **Why it's a smell:** flag definitions and allowed-value validation are copy/paste-prone; adding or changing formats or allowed values requires edits in each place.
- **Remediation direction:** provide a small CLI utility to attach common flags (e.g., `AttachCommonDiscoveryFlags(cmd *cobra.Command, &options)`) and a shared `ValidateFormat(format string, allowed []string) error` helper; or define a `type Format string` with validation methods.

### 3) Docker client wiring duplicated in CLI & tests (#18)
- **Location(s):**
  - `internal/cmd/job_run.go` — `client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())` (used twice: normal run + dry-run).
  - `internal/adapters/dockerlabels/source.go` — adapter also calls `client.NewClientWithOpts(...)`.
  - `internal/testutil/docker.go`, many tests and integration helpers call `client.NewClientWithOpts(...)`.
- **Evidence:** multiple direct calls to `client.NewClientWithOpts` across CLI commands, adapters, and tests.
- **Why it's a smell:** low-level Docker client creation and related configuration (APIVersionNegotiation, host overrides) is scattered; changes for environment variables, TLS, or fallback behavior must be replicated.
- **Remediation direction:** add a single factory helper, e.g., `internal/adapters/dockerclient.NewFromEnv()` returning `*client.Client` and consistent errors/messages, or a small `internal/docker` package that encapsulates client creation and provides helpers for `ComposeController` / `WorkerRunner` creation.
- **Dependencies:** careful not to hide required configuration for testing (allow dependency injection or pass options); avoid introducing global mutable state.

### 4) Repeated error handling & messages
- **Location(s):**
  - Many commands use the same error text/behaviour: `failed to connect to Docker`, `failed to get Docker snapshot`, `invalid format` etc.
  - Some commands print to `os.Stderr` and exit; others return errors with wrapped messages and let the caller handle exit codes.
- **Evidence:** repeated strings and the same branching behavior (print extra "Is Docker running?" hint in some places).
- **Why it's a smell:** inconsistent error reporting patterns increase the risk of divergent user-facing messages and duplicate test assertions.
- **Remediation direction:** centralize common error text constructors and define CLI-facing helpers for rendering errors and mapping to exit codes; unify the style (either return errors with metadata or always let main() handle exit codes consistently).

### 5) Near-identical test setup calls
- **Location(s):**
  - Integration tests under `integration/` repeatedly call `testutil.StartCompose(t, ctx, "<compose file>")` followed by similar HostPort/Wait steps and log dumps.
- **Evidence:** many tests call `stack := testutil.StartCompose(t, ctx, "...")` repeatedly with identical follow-up operations.
- **Why it's a smell:** repeated boilerplate reduces readability and makes it harder to add standardized test preconditions (like logging, health checks, or env setup) across tests.
- **Remediation direction:** expand `testutil` with higher-level helpers (e.g., `StartJobExecutionCompose(t, ctx)` or `RunComposeScenario(t, ctx, scenarioName, func(stack *testutil.Stack) {...})`) to encapsulate common patterns.

### 6) Duplicate discovery/wiring sequences
- **Location(s):**
  - Several commands perform the sequence: create Docker labels source -> build `ports.Selector` (prefixes + filters) -> `source.Snapshot(ctx, selector)` -> `joblabels.NewDiscoverer().DiscoverJobs(...)`. This sequence is replicated in `plan_list`, `plan_show`, `job_run` (and dry-run), `validate`.
- **Evidence:** identical selector construction and snapshot/discover calls in multiple `run*` functions.
- **Why it's a smell:** The orchestration is repeated; extracting it reduces duplication and centralizes filter behavior.
- **Remediation direction:** introduce a small `app.DiscoverJobs(ctx, options)` or `cli.Discover(ctx, opts) (jobs []Job, errors []ValidationError, err error)` that returns found jobs and validation errors and is used by commands.

---

## Questions for user
- [blocking] Should I prioritize **CLI-facing behavior** (i.e., preserve exact error text and output format semantics) over internal refactors that may change wording? (Yes/No)
- [non-blocking] Do you prefer **one** shared output serializer (generic: `PrintJSON/PrintYAML`) or **per-domain** serializers (e.g., `plan.PrintJSON`, `execution.PrintJSON`) to keep domain types explicit?
- [non-blocking] Is it acceptable to introduce a small `internal/dockerclient` factory package for client creation and tests to use it (with injection hooks), or should existing tests continue to manage their own clients directly?

---

## Confidence / Notes
- Confidence: **High** for the CLI duplication findings (evidence is concrete and code-localized).
- Some remediation choices (package boundaries vs helpers on `internal/cmd`) depend on team preferences for layering and packaging; I noted this in remediation directions.
- I did not refactor code yet; this is a discovery pass and recommendation list.

---

## Suggested next scout / follow-up
- Implement a small refactor PR that introduces `internal/ui/format` with `PrintJSON/PrintYAML` and update `plan_list`, `plan_show`, `snapshot`, `job_run (dry-run)`, and `validate` to use it — run tests and update integration expectations.
- Next scout: a focused pass that updates Docker client creation to a shared factory and updates call sites, then runs and updates integration tests as needed.

## META feedback (delete once stable)
- What happened: The project has many small repeated CLI patterns; a standardized set of helper functions would reduce duplication.
- Why it’s a problem: Copy/paste increases maintenance burden and inconsistency risk.
- Proposed instruction change: Add a short guideline in the repo's contributing docs: "CLI commands should use `internal/ui/format` helpers for output and `internal/dockerclient` for Docker wiring".
