.DEFAULT_GOAL := default

.PHONY: test
test:
	gotestsum ./...

.PHONY: test-verbose
test-verbose:
	gotestsum --format standard-verbose ./...

.PHONY: coverage
coverage:
	go test -coverpkg=./... -coverprofile=coverage.out ./... && go tool cover -func coverage.out && rm coverage.out

.PHONY: coverage-persist
coverage-persist:
	go test -coverpkg=./... -coverprofile=coverage.out ./... && go tool cover -func coverage.out

.PHONY: install-gotestsum
install-gotestsum:
	go get github.com/gotestyourself/gotestsum

.PHONY: install-linter
install-linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.41.1

.PHONY: lint
lint:
	golangci-lint run

