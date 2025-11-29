# TODO
## internal/app/app.go
* [internal/app/app.go:10](internal/app/app.go#L10): wire ports/adapters, parse config/flags, start services

## internal/config/loader/parse.go
* [internal/config/loader/parse.go:12](internal/config/loader/parse.go#L12): we really need this? can't we use a lib or rely on json.Unmarshal or yaml parsing?

## internal/domain/labels/types.go
* [internal/domain/labels/types.go:8](internal/domain/labels/types.go#L8): this cannot be here, we need a better way of handling this
