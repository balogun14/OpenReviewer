.PHONY: fmt test build check commit-each-file

fmt:
	gofmt -w cmd internal

test:
	go test ./...

build:
	go build ./cmd/openreview

check: fmt test build

commit-each-file:
	powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/commit-each-file.ps1
