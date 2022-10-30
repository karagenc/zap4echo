GO = go

all: test

.PHONY: test
test:
	$(GO) test -v -race -short ./...

.PHONY: test-coverage
test-coverage:
	$(GO) test -short -coverprofile coverage.out -covermode=atomic ./...
