.PHONY: check test test-web test-go build-web build-go format

check:
	pnpm check
	go vet ./...

test: test-web test-go

test-web:
	pnpm test

test-go:
	go test ./...

build-web:
	pnpm build

build-go:
	go build ./...

format:
	pnpm format
	gofmt -w $$(find cmd internal -name '*.go' -type f)
