# TODO
## internal/app/executor/executor.go
* [internal/app/executor/executor.go:257](internal/app/executor/executor.go#L257): Add worker env vars from labels (bosun.job.worker.env.*)

## internal/app/planner/planner.go
* [internal/app/planner/planner.go:57](internal/app/planner/planner.go#L57): In future, verify all target containers are in this stack

## internal/domain/labels/types.go
* [internal/domain/labels/types.go:9](internal/domain/labels/types.go#L9): Consider a better way of handling shared label constants.

## internal/testutil/harness.go
* [internal/testutil/harness.go:23](internal/testutil/harness.go#L23): use testcontainers-go to up / down compose stacks. do not rely on CLI.

## internal/testutil/worker/Dockerfile
* [internal/testutil/worker/Dockerfile:11](internal/testutil/worker/Dockerfile#L11): Document test worker in README or docs/testing.md
* [internal/testutil/worker/Dockerfile:12](internal/testutil/worker/Dockerfile#L12): Add failure mode variants (exit 1, timeout, etc.) for error path testing
* [internal/testutil/worker/Dockerfile:13](internal/testutil/worker/Dockerfile#L13): Consider publishing to a test registry for CI reproducibility
