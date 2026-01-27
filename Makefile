APP := bosun
PKG := github.com/simone-viozzi/bosun

.PHONY: build run test coverage coverage-html it itv tidy fmt vet docs
build:
	go build -o bin/$(APP) ./cmd/$(APP)

run:
	go run ./cmd/$(APP)

test:
	go test ./...

coverage:
	go test ./... -coverprofile=coverage.out -coverpkg=./...
	@echo ""
	@echo "Coverage Summary:"
	go tool cover -func coverage.out | tee coverage.txt

coverage-html: coverage
	go tool cover -html coverage.out -o coverage.html
	@echo "Generated coverage.html"

it:
	go test -tags=integration -parallel 6 -timeout=20m ./integration/...

itv:
	go test -tags=integration -parallel 6 -timeout=20m -v ./integration/...

tidy:
	go mod tidy

fmt:
	go fmt ./... ./integration/
	goimports -w -local github.com/simone-viozzi/bosun ./cmd ./internal ./integration

vet:
	go vet ./...

lint:
	golangci-lint run

docs:
	go generate ./internal/config/schema/...
