# TODO
## internal/adapters/docker/worker/runner.go
* [internal/adapters/docker/worker/runner.go:135](internal/adapters/docker/worker/runner.go#L135): BUG - Magic 137 exit code when ContainerInspect fails is misleading.
* [internal/adapters/docker/worker/runner.go:141](internal/adapters/docker/worker/runner.go#L141): SMELL - ContainerStop error is silently ignored.
* [internal/adapters/docker/worker/runner.go:151](internal/adapters/docker/worker/runner.go#L151): BUG - Returning 137 here is wrong. We don't know if container
* [internal/adapters/docker/worker/runner.go:162](internal/adapters/docker/worker/runner.go#L162): BUG - Missing stdcopy.StdCopy demultiplexing!
* [internal/adapters/docker/worker/runner.go:174](internal/adapters/docker/worker/runner.go#L174): SMELL - Error is silently swallowed. At minimum, log it.
* [internal/adapters/docker/worker/runner.go:179](internal/adapters/docker/worker/runner.go#L179): BUG - This produces corrupted output! Docker logs have 8-byte headers.

## internal/app/executor/executor.go
* [internal/app/executor/executor.go:316](internal/app/executor/executor.go#L316): Add worker env vars from labels (bosun.job.worker.env.*)

## internal/app/planner/planner.go
* [internal/app/planner/planner.go:49](internal/app/planner/planner.go#L49): DESIGN ISSUE - Current logic is simplistic: assumes useCompose=true if

## internal/cmd/job_run.go
* [internal/cmd/job_run.go:419](internal/cmd/job_run.go#L419): DESIGN ISSUE - Plan rendering is in CLI instead of planner/app layer.
* [internal/cmd/job_run.go:438](internal/cmd/job_run.go#L438): why we have printDryRunText here instead of in the planner itself? what happens if the schema of the planner changes? do we need to also change this function here?

## internal/domain/labels/types.go
* [internal/domain/labels/types.go:9](internal/domain/labels/types.go#L9): Consider a better way of handling shared label constants.

## internal/testutil/docker.go
* [internal/testutil/docker.go:92](internal/testutil/docker.go#L92): BUG - Same as runner.go: missing stdcopy.StdCopy demultiplexing!

## internal/testutil/harness.go
* [internal/testutil/harness.go:23](internal/testutil/harness.go#L23): use testcontainers-go to up / down compose stacks. do not rely on CLI.

## internal/testutil/worker/Dockerfile
* [internal/testutil/worker/Dockerfile:11](internal/testutil/worker/Dockerfile#L11): Document test worker in README or docs/testing.md
* [internal/testutil/worker/Dockerfile:12](internal/testutil/worker/Dockerfile#L12): Add failure mode variants (exit 1, timeout, etc.) for error path testing
* [internal/testutil/worker/Dockerfile:13](internal/testutil/worker/Dockerfile#L13): Consider publishing to a test registry for CI reproducibility
