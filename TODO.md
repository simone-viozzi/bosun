# TODO
## .golangci.yml
* [.golangci.yml:23](.golangci.yml#L23): Enable these linters after fixing existing issues
* [.golangci.yml:27](.golangci.yml#L27): Re-enable after addressing existing violations

## internal/app/planner/planner.go
* [internal/app/planner/planner.go:57](internal/app/planner/planner.go#L57): In future, verify all target containers are in this stack

## internal/domain/labels/types.go
* [internal/domain/labels/types.go:8](internal/domain/labels/types.go#L8): this cannot be here, we need a better way of handling this

## internal/testutil/harness.go
* [internal/testutil/harness.go:23](internal/testutil/harness.go#L23): use testcontainers-go to up / down compose stacks. do not rely on CLI.
