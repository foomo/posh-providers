.DEFAULT_GOAL:=help

## === Tasks ===

.PHONY: check
## Run tests and linters
check: tidy lint test

.PHONY: tidy
## Run go mod tidy
tidy: wd=$(shell pwd)
tidy: files=$(shell find . -type f -name go.mod)
tidy: dirs=$(foreach file,$(files),$(dir $(file)))
tidy:
	@for dir in $(dirs); do cd ${wd} && cd $$dir && go mod tidy; done

.PHONY: outdated
## Show outdated direct dependencies
outdated: wd=$(shell pwd)
outdated: files=$(shell find . -type f -name go.mod)
outdated: dirs=$(foreach file,$(files),$(dir $(file)) )
outdated:
	@for dir in $(dirs); do cd ${wd} && cd $$dir && go list -u -m -json all | go-mod-outdated -update -direct; done

.PHONY: test
## Run tests
test: wd=$(shell pwd)
test: files=$(shell find . -type f -name go.mod)
test: dirs=$(foreach file,$(files),$(dir $(file)) )
test:
	@for dir in $(dirs); do cd ${wd} && cd $$dir && go test -v ./...; done

.PHONY: lint
## Run linter
lint: wd=$(shell pwd)
lint: files=$(shell find . -type f -name go.mod)
lint: dirs=$(foreach file,$(files),$(dir $(file)) )
lint:
	@for dir in $(dirs); do cd ${wd} && cd $$dir && golangci-lint run; done

.PHONY: lint.fix
## Fix lint violations
lint.fix: wd=$(shell pwd)
lint.fix: files=$(shell find . -type f -name go.mod)
lint.fix: dirs=$(foreach file,$(files),$(dir $(file)) )
lint.fix:
	@for dir in $(dirs); do cd ${wd} && cd $$dir && golangci-lint run --fix; done

## === Utils ===

## Show help text
help:
	@awk '{ \
			if ($$0 ~ /^.PHONY: [a-zA-Z\-\_0-9]+$$/) { \
				helpCommand = substr($$0, index($$0, ":") + 2); \
				if (helpMessage) { \
					printf "\033[36m%-23s\033[0m %s\n", \
						helpCommand, helpMessage; \
					helpMessage = ""; \
				} \
			} else if ($$0 ~ /^[a-zA-Z\-\_0-9.]+:/) { \
				helpCommand = substr($$0, 0, index($$0, ":")); \
				if (helpMessage) { \
					printf "\033[36m%-23s\033[0m %s\n", \
						helpCommand, helpMessage"\n"; \
					helpMessage = ""; \
				} \
			} else if ($$0 ~ /^##/) { \
				if (helpMessage) { \
					helpMessage = helpMessage"\n                        "substr($$0, 3); \
				} else { \
					helpMessage = substr($$0, 3); \
				} \
			} else { \
				if (helpMessage) { \
					print "\n                        "helpMessage"\n" \
				} \
				helpMessage = ""; \
			} \
		}' \
		$(MAKEFILE_LIST)
