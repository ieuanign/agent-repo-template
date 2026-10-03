#!/usr/bin/env bash
# oxlint, the Prettier check, then typegen and tsc; writes only gitignored files.
set -euo pipefail

cd "$(dirname "$0")/.."

echo "web: oxlint"
oxlint

# Prettier reads .prettierignore and .gitignore only from its working directory.
echo "web: prettier --check"
(cd .. && prettier --check web)

# next dev in the container points next-env.d.ts at dev types the host never has; typegen rewrites it.
echo "web: next typegen"
next typegen

echo "web: tsc --noEmit"
tsc --noEmit
