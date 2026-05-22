.DEFAULT_GOAL:=help
-include .makerc

# --- Config -----------------------------------------------------------------

GOMODS=$(shell find . -type f -name go.mod)
# Newline hack for error output
define br


endef

# --- Targets -----------------------------------------------------------------

# This allows us to accept extra arguments
%: .mise .lefthook go.work
	@:

.PHONY: .mise
# Install dependencies
.mise:
ifeq (, $(shell command -v mise))
	$(error $(br)$(br)Please ensure you have 'mise' installed and activated!$(br)$(br)  $$ brew update$(br)  $$ brew install mise$(br)$(br)See the documentation: https://mise.jdx.dev/getting-started.html)
endif
	@mise install

.PHONY: .lefthook
# Configure git hooks for lefthook
.lefthook:
	@lefthook install --reset-hooks-path

# Ensure go.work file
go.work:
	@echo "〉initializing go work"
	@go work init && go work use -r . && go work sync

### Tasks

.PHONY: check
## Run tidy, generate, schema, lint & tests
check: tidy generate schema lint.fix test audit

.PHONY: lint
## Run linter
lint:
	@echo "〉golangci-lint run"
	@$(foreach mod,$(GOMODS), (cd $(dir $(mod)) && echo "📂 $(dir $(mod))" && golangci-lint run) &&) true

.PHONY: lint.fix
## Fix lint violations
lint.fix:
	@echo "〉golangci-lint run fix"
	@$(foreach mod,$(GOMODS), (cd $(dir $(mod)) && echo "📂 $(dir $(mod))" && golangci-lint run --fix) &&) true

.PHONY: generate
## Run go generate
generate:
	@echo "〉go generate"
	@go generate work

.PHONY: schema
## Generate JSON schema
schema:
	@echo "〉generating schema"
	@yq eval-all '. as $$item ireduce ({}; . *+ $$item)' base.schema.json \
		$(shell find . -name config.base.json -print | tr '\n' ' ') \
		> merged.schema.json
	@-jsonschema bundle merged.schema.json \
			$(shell find . -name config.schema.json -print | sed 's/^/--resolve /' | tr '\n' ' ') \
			--without-id \
			--http \
			> posh.schema.json
	@rm merged.schema.json

.PHONY: test
## Run tests
test:
	@echo "〉go test"
	@GO_TEST_TAGS=-skip go test -coverprofile=coverage.out -tags=safe work

.PHONY: test.race
## Run tests
test.race:
	@echo "〉go test with -race"
	@GO_TEST_TAGS=-skip go test -coverprofile=coverage.out -tags=safe -race work

.PHONY: test.nocache
## Run tests with -count=1
test.nocache:
	@echo "〉go test -count=1"
	@GO_TEST_TAGS=-skip go test -coverprofile=coverage.out -tags=safe -count=1 work

### Security

.PHONY: audit
## Run security audit
audit:
	@echo "〉security audit"
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@$(foreach mod,$(GOMODS), (cd $(dir $(mod)) && echo "📂 $(dir $(mod))" && govulncheck ./...) &&) true

### Dependencies

.PHONY: tidy
## Run go mod tidy
tidy:
	@echo "〉go mod tidy"
	@$(foreach mod,$(GOMODS), (cd $(dir $(mod)) && echo "📂 $(dir $(mod))" && go mod tidy) &&) true
	@go work use -r . && go work sync

.PHONY: outdated
## Show outdated direct dependencies
outdated:
	@echo "〉go mod outdated"
	@go list -u -m -json all | go-mod-outdated -update -direct

.PHONY: upgrade
## Upgrade dependencies
upgrade:
	@echo "〉go mod upgrade"
	@rm go.work go.work.sum
	@$(foreach mod,$(GOMODS), (cd $(dir $(mod)) && echo "📂 $(dir $(mod))" && go list -u -m -f '{{if and (not .Indirect) .Update}}{{.Path}}{{end}}' all | xargs -n1 -I{} go get {}@latest) &&) true
	@$(MAKE) tidy

### Release

.PHONY: tag.submodules
## Create tags for submodules TAG=1.0.0
tag.submodules:
	@echo "$(TAG)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$$' || { echo "❌ TAG must be vX.Y.Z format"; exit 1; }
	@git rev-parse "$(TAG)" >/dev/null 2>&1 || { echo "❌ Tag $(TAG) does not exist"; exit 1; }
	@echo "🔖 Creating submodule tags..."
	@find . -type f -name 'go.mod' -mindepth 2 -not -path './vendor/*' -exec sh -c 'dir=$$(dirname {} | sed "s|^\./||"); tag="$$dir/$(TAG)"; git rev-parse "$$tag" >/dev/null 2>&1 || { echo "🔖 $$tag"; git tag "$$tag"; }' \;
	@echo "🔖 Pushing tags..."
	@git push origin --tags

### Documentation

.PHONY: docs
## Open docs
docs:
	@echo "〉starting docs"
	@cd docs && bun install && bun run dev

.PHONY: docs.build
## Open docs
docs.build:
	@echo "〉building docs"
	@cd docs && bun install && bun run build

.PHONY: godocs
## Open go docs
godocs:
	@echo "〉starting go docs"
	@go doc -http

### Utils

.PHONY: help
# https://patorjk.com/software/taag/#p=display&f=Tmplr&t=posh+providers&x=none&v=4&h=4&w=80&we=false
help: g=\033[0;32m
help: b=\033[0;34m
help: w=\033[0;90m
help: e=\033[0m
## Show help text
help:
	@echo "$(g)"
	@echo "     ┓           • ┓"
	@echo "┏┓┏┓┏┣┓  ┏┓┏┓┏┓┓┏┓┏┫┏┓┏┓┏"
	@echo "┣┛┗┛┛┛┗  ┣┛┛ ┗┛┗┛┗┗┻┗ ┛ ┛"
	@echo "┛        ┛"
	@echo "with ❤ foomo by bestbytes"
	@echo "$(e)"
	@echo "$(b)Usage:$(e)\n  make [task]"
	@awk '{ \
		if($$0 ~ /^### /){ \
			if(help) printf "  %-21s $(w)%s$(e)\n\n", cmd, help; help=""; \
			printf "$(b)\n%s:$(e)\n", substr($$0,5); \
		} else if($$0 ~ /^[a-zA-Z0-9._-]+:/){ \
			cmd = substr($$0, 1, index($$0, ":")-1); \
			if(help) printf "  %-21s $(w)%s$(e)\n", cmd, help; help=""; \
		} else if($$0 ~ /^##/){ \
			help = help ? help "\n                        " substr($$0,3) : substr($$0,3); \
		} else if(help){ \
			print "\n                        $(w)" help "$(e)\n"; help=""; \
		} \
	}' $(MAKEFILE_LIST)
	@echo ""
