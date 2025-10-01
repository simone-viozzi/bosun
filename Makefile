APP := bosun
PKG := github.com/simone-viozzi/bosun

.PHONY: build run test it itv tidy fmt vet
build:
	go build -o bin/$(APP) ./cmd/$(APP)

run:
	go run ./cmd/$(APP)

test:
	go test ./...

it:
	go test -tags=integration -parallel 6 -timeout=20m ./integration/...

itv:
	go test -tags=integration -parallel 6 -timeout=20m -v ./integration/...

tidy:
	go mod tidy

fmt:
	go fmt ./... ./integration/

vet:
	go vet ./...
