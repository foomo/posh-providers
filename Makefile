.DEFAULT_GOAL:=help
-include .makerc

# --- Config -----------------------------------------------------------------

# Newline hack for error output
define br


endef

# --- Targets -----------------------------------------------------------------

# This allows us to accept extra arguments
%: .mise .husky
	@:

.PHONY: .mise
# Install dependencies
.mise: msg := $(br)$(br)Please ensure you have 'mise' installed and activated!$(br)$(br)$$ brew update$(br)$$ brew install mise$(br)$(br)See the documentation: https://mise.jdx.dev/getting-started.html$(br)$(br)
.mise:
ifeq (, $(shell command -v mise))
	$(error ${msg})
endif
	@mise install

.PHONY: .husky
# Configure git hooks for husky
.husky:
	@git config core.hooksPath .husky

### Tasks

.PHONY: check
## Run tests and linters
check: tidy schema lint test

.PHONY: tidy
## Run go mod tidy
tidy:
	@go mod tidy

.PHONY: lint
## Run linter
lint:
	@golangci-lint run

.PHONY: lint.fix
## Fix lint violations
lint.fix:
	@golangci-lint run --fix

.PHONY: test
## Run tests
test:
	@GO_TEST_TAGS=-skip go test -coverprofile=coverage.out --tags=safe -race ./...

.PHONY: schema
## Run linter
schema:
	@jsonschema bundle config.schema.base.json \
  		$(shell find . -name config.schema.json -printf '--resolve %p ') \
  		--without-id \
  		> posh.schema.json

.PHONY: foo
## Run linter
foo:
	@yq eval-all '. as $$item ireduce ({}; . *+ $$item)' base.schema.json \
		$(shell find . -name config.base.json -print | tr '\n' ' ') \
		> merged.schema.json
	@jsonschema bundle merged.schema.json \
			$(shell find . -name config.schema.json -print | sed 's/^/--resolve /' | tr '\n' ' ') \
			--without-id \
			--http \
			> posh.schema.json
	@rm merged.schema.json

### Utils

.PHONY: help
## Show help text
help:
	@echo "Project Oriented SHELL (posh)\n"
	@echo "Usage:\n  make [task]"
	@awk '{ \
		if($$0 ~ /^### /){ \
			if(help) printf "%-23s %s\n\n", cmd, help; help=""; \
			printf "\n%s:\n", substr($$0,5); \
		} else if($$0 ~ /^[a-zA-Z0-9._-]+:/){ \
			cmd = substr($$0, 1, index($$0, ":")-1); \
			if(help) printf "  %-23s %s\n", cmd, help; help=""; \
		} else if($$0 ~ /^##/){ \
			help = help ? help "\n                        " substr($$0,3) : substr($$0,3); \
		} else if(help){ \
			print "\n                        " help "\n"; help=""; \
		} \
	}' $(MAKEFILE_LIST)

