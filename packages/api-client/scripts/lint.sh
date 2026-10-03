#!/usr/bin/env bash
# tsc, then openapi-typescript regenerate-and-diff; writes only under .scratch/.
set -euo pipefail

cd "$(dirname "$0")/.."
root=$(git rev-parse --show-toplevel)
mkdir -p "$root/.scratch"
tmp=$(mktemp -d "$root/.scratch/api-client-lint.XXXXXX")
trap 'rm -rf "$tmp"' EXIT

echo "api-client: tsc"
tsc --noEmit

echo "api-client: regenerate and diff"
openapi-typescript ../../backend/api/openapi.yaml -o "$tmp/schema.d.ts"
diff "$tmp/schema.d.ts" src/schema.d.ts
