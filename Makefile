APP := bosun
PKG := github.com/simone-viozzi/bosun

.PHONY: build run test tidy fmt vet
build:
	go build -o bin/$(APP) ./cmd/$(APP)

run:
	go run ./cmd/$(APP)

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...
