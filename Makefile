APP := bosun
PKG := github.com/simone-viozzi/bosun

.PHONY: build run test test-integration tidy fmt vet
build:
	go build -o bin/$(APP) ./cmd/$(APP)

run:
	go run ./cmd/$(APP)

test:
	go test ./...

test-integration:
	go test ./integration -tags=integration

tidy:
	go mod tidy

fmt:
	go fmt ./... ./integration/

vet:
	go vet ./...
