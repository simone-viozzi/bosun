APP := bosun
PKG := github.com/simone-viozzi/bosun

.PHONY: build run test coverage coverage-html it itv it-cover coverage-integration coverage-all tidy fmt vet docs
build:
	go build -o bin/$(APP) ./cmd/$(APP)

run:
	go run ./cmd/$(APP)

test:
	go test ./...

coverage:
	@mkdir -p covdata-unit
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

it-cover:
	@echo "Running integration tests with coverage..."
	@mkdir -p covdata-integration
	@rm -rf covdata-integration/*
	BOSUN_TEST_COVERAGE=1 GOCOVERDIR=$(PWD)/covdata-integration go test -tags=integration -parallel 6 -timeout=20m ./integration/...
	@$(MAKE) coverage-integration

coverage-integration:
	@echo "Converting integration coverage data..."
	@if [ -d "covdata-integration" ] && [ -n "$$(ls -A covdata-integration 2>/dev/null)" ]; then \
		go tool covdata textfmt -i=covdata-integration -o coverage.integration.out; \
		go tool cover -func coverage.integration.out | tail -n 1; \
	else \
		echo "No integration coverage data found (covdata-integration directory missing or empty)"; \
		exit 1; \
	fi

coverage-all: coverage it-cover
	@echo "Merging coverage profiles using Go native tooling..."
	@mkdir -p covdata-unit covdata-merged
	@rm -rf covdata-unit/* covdata-merged/*
	@# Convert unit test profile to covdata format
	go tool covdata textfmt -i=covdata-unit -o /dev/null 2>/dev/null || \
		(echo "Converting unit coverage to covdata format..." && \
		 go test ./... -coverpkg=./... -test.gocoverdir=$(PWD)/covdata-unit > /dev/null 2>&1)
	@# Merge both covdata directories
	go tool covdata merge -i=covdata-unit,covdata-integration -o=covdata-merged
	@# Convert merged result to profile
	go tool covdata textfmt -i=covdata-merged -o coverage.all.out
	@echo ""
	@echo "Combined Coverage Summary:"
	go tool cover -func coverage.all.out | tee coverage.all.txt
	go tool cover -html coverage.all.out -o coverage.all.html
	@echo "Generated coverage.all.html"

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
