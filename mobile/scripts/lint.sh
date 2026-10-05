#!/usr/bin/env bash
# oxlint, the Prettier check, Expo's version check, then route types and tsc; writes only gitignored files.
set -euo pipefail

cd "$(dirname "$0")/.."

echo "mobile: oxlint"
oxlint

# Prettier reads .prettierignore and .gitignore only from its working directory.
echo "mobile: prettier --check"
(cd .. && prettier --check mobile)

# CI=1 makes a mismatch fail instead of prompting; offline so lint needs no network.
echo "mobile: expo install --check"
CI=1 EXPO_OFFLINE=1 expo install --check

# Writes expo-env.d.ts and the typed-route declarations without starting Metro.
echo "mobile: expo customize tsconfig.json"
expo customize tsconfig.json

echo "mobile: tsc --noEmit"
tsc --noEmit
