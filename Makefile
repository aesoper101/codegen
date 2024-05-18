.PHONY: \
	help \
	coverage \
	vet \
	lint \
	fmt \
	version

all:  fmt lint vet

help:
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@echo '    help               Show this help screen.'
	@echo '    coverage           Report code tests coverage.'
	@echo '    vet                Run go vet.'
	@echo '    lint               Run golint.'
	@echo '    fmt                Run go fmt.'
	@echo '    version            Display Go version.'
	@echo ''
	@echo 'Targets run by default are: lint, vet.'
	@echo ''

print-%:
	@echo $* = $($*)

deps:
#	go install honnef.co/go/tools/cmd/staticcheck@latest

coverage:
	go test $(go list ./... | grep -v examples) -coverprofile coverage.txt ./...

vet:
	go vet ./...

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./... -c ./.golangci.yml

fmt:
	go install mvdan.cc/gofumpt@latest
	gofumpt -l -w -extra .


version:
	@go version