#!/usr/bin/env bash
# gofmt, go vet, oapi-codegen regenerate-and-diff, then migration lint; writes only under .scratch/.
set -euo pipefail

cd "$(dirname "$0")/.."
root=$(git rev-parse --show-toplevel)
mkdir -p "$root/.scratch"
tmp=$(mktemp -d "$root/.scratch/backend-lint.XXXXXX")
trap 'rm -rf "$tmp"' EXIT

echo "backend: gofmt"
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
  echo "backend: gofmt needed on:" >&2
  echo "$unformatted" >&2
  exit 1
fi

echo "backend: go vet"
go vet ./...

echo "backend: regenerate and diff"
go tool oapi-codegen -config api/oapi-codegen.yaml -o "$tmp/openapi/openapi.gen.go" api/openapi.yaml
diff -r "$tmp/openapi" gen/openapi

echo "backend: migration lint"
go run ./scripts/migrationlint
