.DEFAULT_GOAL:=help
-include .makerc

# --- Targets -----------------------------------------------------------------

# This allows us to accept extra arguments
%: .husky
	@:

.PHONY: .husky
# Configure git hooks for husky
.husky:
	@if ! command -v husky &> /dev/null; then \
		echo "ERROR: missing executeable 'husky', please run:"; \
		echo "\n$ go install github.com/go-courier/husky/cmd/husky@latest\n"; \
	fi
	@git config core.hooksPath .husky

## === Tasks ===

.PHONY: tidy
## Run go mod tidy
tidy:
	@go mod tidy

.PHONY: lint
## Run linter
lint:
	@golangci-lint run

.PHONY: schema
## Run linter
schema:
	@jsonschema bundle config.schema.base.json \
		--resolve ./arbitrary/open/config.schema.json \
		--resolve ./arbitrary/task/config.schema.json \
		--resolve ./arbitrary/zip/config.schema.json \
		--resolve ./azure/az/config.schema.json \
		--resolve ./cloudflare/cloudflared/config.schema.json \
		--resolve ./digitalocean/doctl/config.schema.json \
		--resolve ./etcd-io/etcd/config.schema.json \
		--resolve ./facebook/docusaurus/config.schema.json \
		--resolve ./filosottile/mkcert/config.schema.json \
		--resolve ./foomo/beam/config.schema.json \
		--resolve ./foomo/gocontentful/config.schema.json \
		--resolve ./foomo/sesamy/config.schema.json \
		--resolve ./foomo/squadron/config.schema.json \
		--resolve ./goharbor/harbor/config.schema.json \
		--resolve ./golang-migrate/migrate/config.schema.json \
		--resolve ./google/gcloud/config.schema.json \
		--resolve ./grafana/k6/config.schema.json \
		--resolve ./gravitational/teleport/config.schema.json \
		--resolve ./gruntwork-io/terragrunt/config.schema.json \
		--resolve ./hashicorp/cdktf/config.schema.json \
		--resolve ./jondot/hygen/config.schema.json \
		--resolve ./k3d-io/k3d/config.schema.json \
		--resolve ./kubernets/kubectl/config.schema.json \
		--resolve ./kubernets/kubeforward/config.schema.json \
		--resolve ./onepassword/config.schema.json \
		--resolve ./pivotal/licensefinder/config.schema.json \
		--resolve ./pulumi/pulumi/azure/config.schema.json \
		--resolve ./pulumi/pulumi/gcloud/config.schema.json \
		--resolve ./rclone/rclone/config.schema.json \
		--resolve ./slack-go/slack/config.schema.json \
		--resolve ./sqlc-dev/sqlc/config.schema.json \
		--resolve ./stackitcloud/stackit/config.schema.json \
		--resolve ./stern/stern/config.schema.json \
		--resolve ./usebruno/bruno/config.schema.json \
		--resolve ./webdriverio/webdriverio/config.schema.json \
		--resolve ./sshtunnel/config.schema.json \
		--without-id \
		> config.schema.json

.PHONY: outdated
## Show outdated direct dependencies
outdated:
	@go list -u -m -json all | go-mod-outdated -update -direct

.PHONY: test
## Run tests
test:
	@GO_TEST_TAGS=-skip go test -coverprofile=coverage.out --tags=safe -race ./...
	#@GO_TEST_TAGS=-skip go test -coverprofile=coverage.out --tags=safe -race -json ./... 2>&1 | tee /tmp/gotest.log | gotestfmt

.PHONY: lint.fix
## Fix lint violations
lint.fix:
	@golangci-lint run --fix

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
