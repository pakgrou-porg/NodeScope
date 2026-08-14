#!/usr/bin/env bash
# Run the locally available release-quality checks without any production
# credentials, database mutation, or external runtime access.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

printf '%s\n' '==> Checking repository whitespace'
git diff --check

printf '%s\n' '==> Checking license metadata'
pnpm license:check

printf '%s\n' '==> Checking signed-tag release workflow policy'
./scripts/check-release-workflow-contract.sh
./scripts/test-release-workflow-contract.sh

printf '%s\n' '==> Checking continuous-integration workflow policy'
./scripts/check-ci-workflow-contract.sh
./scripts/test-ci-workflow-contract.sh

printf '%s\n' '==> Checking installation guidance policy'
./scripts/check-installation-guidance.sh
./scripts/test-installation-guidance.sh

printf '%s\n' '==> Checking Linux installer provenance contract'
./scripts/check-install-linux-contract.sh
./scripts/test-install-linux-contract.sh
./scripts/test-install-linux-runtime.sh

printf '%s\n' '==> Checking release-evidence verifier contract'
./scripts/test-verify-agent-release-evidence.sh

printf '%s\n' '==> Checking migration application path-containment contract'
./scripts/test-apply-nodescope-migration-contract.sh

printf '%s\n' '==> Checking shared-Supabase disposable fixture contract'
./scripts/test-verify-shared-supabase-fixture-contract.sh

printf '%s\n' '==> Checking shared-Supabase isolation verifier contract'
./scripts/test-verify-shared-isolation-contract.sh

printf '%s\n' '==> Checking cloud control-plane canary contract'
./scripts/test-e2e-cloud-control-plane-canary-contract.sh

printf '%s\n' '==> Checking local resilience rehearsal contract'
./scripts/test-rehearse-resilience-local-contract.sh

printf '%s\n' '==> Checking machine-readable release evidence contract'
./scripts/test-release-evidence-contract.sh

printf '%s\n' '==> Checking console RBAC readiness contract'
./scripts/test-rehearse-console-rbac-local-contract.sh

printf '%s\n' '==> Checking inference privacy rehearsal contract'
./scripts/test-rehearse-inference-privacy-local-contract.sh

printf '%s\n' '==> Checking Windows baseline readiness contract'
./scripts/test-rehearse-windows-baseline-local-contract.sh

printf '%s\n' '==> Checking CI quality contract'
./scripts/test-ci-quality-contract.sh

printf '%s\n' '==> Checking workflow runtime contract'
./scripts/test-workflow-runtime-contract.sh

printf '%s\n' '==> Checking bounded server logging contract'
./scripts/test-server-logging-contract.sh

printf '%s\n' '==> Checking cloud replica compose preflight contract'
./scripts/test-preflight-cloud-replica-compose-contract.sh

printf '%s\n' '==> Checking Framework Portainer stack contract'
./scripts/test-framework-portainer-stack-contract.sh

printf '%s\n' '==> Running Go static analysis'
go vet ./...

printf '%s\n' '==> Testing Go packages'
go test ./...

printf '%s\n' '==> Cross-building Linux native targets'
for goarch in amd64 arm64; do
  GOOS=linux GOARCH="$goarch" go build ./...
done

printf '%s\n' '==> Cross-compiling Windows agent baselines'
for goarch in amd64 arm64; do
  GOOS=windows GOARCH="$goarch" go build -o "$(mktemp --suffix=.exe)" ./cmd/nodescope-agent
  GOOS=windows GOARCH="$goarch" go test -c -o "$(mktemp --suffix=.test.exe)" ./internal/agent
done

printf '%s\n' '==> Checking TypeScript and browser tests'
pnpm check
pnpm test

printf '%s\n' '==> Checking telemetry and OpenAPI contracts'
pnpm contracts:check

printf '%s\n' '==> Checking generated-contract drift'
generated_dir="$(mktemp -d)"
tooling_dir="$(mktemp -d)"
trap 'rm -rf "$generated_dir" "$tooling_dir"' EXIT
cp telemetry/v1/envelope.pb.go "$generated_dir/envelope.pb.go"
cp api/openapi/generated.ts "$generated_dir/generated.ts"
GOBIN="$tooling_dir" go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.32.0
PATH="$tooling_dir:$PATH" pnpm proto:generate
pnpm api:generate
cmp -s "$generated_dir/envelope.pb.go" telemetry/v1/envelope.pb.go
cmp -s "$generated_dir/generated.ts" api/openapi/generated.ts

printf '%s\n' '==> Building production browser and server bundles'
pnpm build

printf '%s\n' 'NodeScope release-readiness checks passed.'
