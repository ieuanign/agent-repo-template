#!/usr/bin/env bash
# Ruff's autofix then formatter; --locked so nothing here rewrites uv.lock.
set -euo pipefail

cd "$(dirname "$0")/.."

# --exit-zero: leftovers must not skip the formatter; the lint gate reports them.
echo "ai: ruff check --fix"
uv run --locked ruff check --fix --exit-zero

echo "ai: ruff format"
uv run --locked ruff format
