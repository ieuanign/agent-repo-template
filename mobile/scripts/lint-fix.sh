#!/usr/bin/env bash
# oxlint's autofix, then Prettier's write mode; fixes only what lint.sh checks.
set -euo pipefail

cd "$(dirname "$0")/.."

# oxlint exits 1 on leftovers and has no exit-zero flag; those must not skip Prettier, the lint gate reports them.
echo "mobile: oxlint --fix"
oxlint --fix || [ $? -eq 1 ]

# Prettier reads .prettierignore and .gitignore only from its working directory.
echo "mobile: prettier --write"
(cd .. && prettier --write mobile)
