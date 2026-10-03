#!/usr/bin/env bash
# Lock check, Ruff lint and format, strict mypy, then import-linter; never rewrites uv.lock.
set -euo pipefail

cd "$(dirname "$0")/.."

echo "ai: uv lock --check"
uv lock --check

echo "ai: ruff check"
uv run --locked ruff check

echo "ai: ruff format --check"
uv run --locked ruff format --check

echo "ai: mypy"
uv run --locked mypy

# --no-cache: the default cache dir is not gitignored and would block worktree removal.
echo "ai: lint-imports"
uv run --locked lint-imports --no-cache
