# TODO
## internal/app/executor/executor.go
* [internal/app/executor/executor.go:322](internal/app/executor/executor.go#L322): Add worker env vars from labels (bosun.job.worker.env.*)

## internal/app/planner/planner.go
* [internal/app/planner/planner.go:49](internal/app/planner/planner.go#L49): DESIGN ISSUE - Current logic is simplistic: assumes useCompose=true if

## internal/cmd/job_run.go
* [internal/cmd/job_run.go:424](internal/cmd/job_run.go#L424): DESIGN ISSUE - Plan rendering is in CLI instead of planner/app layer.
* [internal/cmd/job_run.go:443](internal/cmd/job_run.go#L443): why we have printDryRunText here instead of in the planner itself? what happens if the schema of the planner changes? do we need to also change this function here?

## internal/domain/labels/types.go
* [internal/domain/labels/types.go:9](internal/domain/labels/types.go#L9): Consider a better way of handling shared label constants.

## internal/testutil/harness.go
* [internal/testutil/harness.go:23](internal/testutil/harness.go#L23): use testcontainers-go to up / down compose stacks. do not rely on CLI.

## internal/testutil/worker/Dockerfile
* [internal/testutil/worker/Dockerfile:11](internal/testutil/worker/Dockerfile#L11): Document test worker in README or docs/testing.md
* [internal/testutil/worker/Dockerfile:12](internal/testutil/worker/Dockerfile#L12): Add failure mode variants (exit 1, timeout, etc.) for error path testing
* [internal/testutil/worker/Dockerfile:13](internal/testutil/worker/Dockerfile#L13): Consider publishing to a test registry for CI reproducibility
