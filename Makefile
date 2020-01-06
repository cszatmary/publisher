.DEFAULT_GOAL = build
NAME = publisher

# Get all dependencies
setup:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	go mod download
.PHONY: setup

# Build publisher
build:
	go build
.PHONY: build

# Clean all build & test artifacts
clean:
	rm $(NAME)
	rm -rf coverage
.PHONY: clean

# Run the linter
lint:
	./bin/golangci-lint run ./...
.PHONY: lint

# Remove version of publisher installed with go install
go-uninstall:
	rm $(shell go env GOPATH)/bin/$(NAME)
.PHONY: go-uninstall

# Run tests and collect coverage data
test:
	mkdir -p coverage
	go test -coverprofile=coverage/coverage.txt ./...
	go tool cover -html=coverage/coverage.txt -o coverage/coverage.html
.PHONY: test
