APP := bosun
PKG := github.com/simone-viozzi/bosun
COVERAGE_DIR := coverage

.PHONY: build run test coverage coverage-html it itv it-cover coverage-integration coverage-all tidy fmt vet docs
build:
	go build -o bin/$(APP) ./cmd/$(APP)

run:
	go run ./cmd/$(APP)

test:
	go test ./...

coverage:
	@mkdir -p $(COVERAGE_DIR)
	go test ./... -coverprofile=$(COVERAGE_DIR)/coverage.out -coverpkg=./...
	@echo ""
	@echo "Coverage Summary:"
	go tool cover -func $(COVERAGE_DIR)/coverage.out | tee $(COVERAGE_DIR)/coverage.txt

coverage-html: coverage
	go tool cover -html $(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Generated $(COVERAGE_DIR)/coverage.html"

it:
	go test -tags=integration -parallel 6 -timeout=20m ./integration/...

itv:
	go test -tags=integration -parallel 6 -timeout=20m -v ./integration/...

it-cover:
	@echo "Running integration tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)/covdata-integration
	@rm -rf $(COVERAGE_DIR)/covdata-integration/*
	BOSUN_TEST_COVERAGE=1 GOCOVERDIR=$(PWD)/$(COVERAGE_DIR)/covdata-integration go test -tags=integration -parallel 6 -timeout=20m ./integration/...
	@$(MAKE) coverage-integration

coverage-integration:
	@echo "Converting integration coverage data..."
	@if [ -d "$(COVERAGE_DIR)/covdata-integration" ] && [ -n "$$(ls -A $(COVERAGE_DIR)/covdata-integration 2>/dev/null)" ]; then \
		go tool covdata textfmt -i=$(COVERAGE_DIR)/covdata-integration -o $(COVERAGE_DIR)/coverage.integration.out; \
		go tool cover -func $(COVERAGE_DIR)/coverage.integration.out | tail -n 1; \
	else \
		echo "No integration coverage data found ($(COVERAGE_DIR)/covdata-integration directory missing or empty)"; \
		exit 1; \
	fi

coverage-all: coverage it-cover
	@echo "Merging coverage profiles using Go native tooling..."
	@mkdir -p $(COVERAGE_DIR)/covdata-unit $(COVERAGE_DIR)/covdata-merged
	@rm -rf $(COVERAGE_DIR)/covdata-unit/* $(COVERAGE_DIR)/covdata-merged/*
	@# Convert unit test profile to covdata format
	go tool covdata textfmt -i=$(COVERAGE_DIR)/covdata-unit -o /dev/null 2>/dev/null || \
		(echo "Converting unit coverage to covdata format..." && \
		 go test ./... -coverpkg=./... -test.gocoverdir=$(PWD)/$(COVERAGE_DIR)/covdata-unit > /dev/null 2>&1)
	@# Merge both covdata directories
	go tool covdata merge -i=$(COVERAGE_DIR)/covdata-unit,$(COVERAGE_DIR)/covdata-integration -o=$(COVERAGE_DIR)/covdata-merged
	@# Convert merged result to profile
	go tool covdata textfmt -i=$(COVERAGE_DIR)/covdata-merged -o $(COVERAGE_DIR)/coverage.all.out
	@echo ""
	@echo "Combined Coverage Summary:"
	go tool cover -func $(COVERAGE_DIR)/coverage.all.out | tee $(COVERAGE_DIR)/coverage.all.txt
	go tool cover -html $(COVERAGE_DIR)/coverage.all.out -o $(COVERAGE_DIR)/coverage.all.html
	@echo "Generated $(COVERAGE_DIR)/coverage.all.html"

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
