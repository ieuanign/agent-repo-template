#!/usr/bin/env bash
# buf lint, buf breaking against local main, then regenerate-and-diff; writes only under .scratch/.
set -euo pipefail

cd "$(dirname "$0")/.."
root=$(git rev-parse --show-toplevel)
mkdir -p "$root/.scratch"
tmp=$(mktemp -d "$root/.scratch/proto-lint.XXXXXX")
trap 'rm -rf "$tmp"' EXIT

echo "proto: buf lint"
buf lint

echo "proto: buf breaking"
if git cat-file -e main:proto/buf.yaml 2>/dev/null; then
  mkdir "$tmp/main"
  git -C "$root" archive main:proto | tar -x -C "$tmp/main"
  buf breaking --against "$tmp/main"
else
  echo "proto: main has no proto/buf.yaml yet, no breaking baseline; skipping"
fi

echo "proto: regenerate and diff"
buf generate -o "$tmp/proto"
# openapi is oapi-codegen's output, checked by backend#lint; pyproject.toml and py.typed are
# hand-written (packaging, PEP 561 marker); __pycache__ is gitignored interpreter output.
diff -r -x openapi "$tmp/backend/gen" "$root/backend/gen"
diff -r -x pyproject.toml -x py.typed -x __pycache__ "$tmp/ai/gen" "$root/ai/gen"
