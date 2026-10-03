#!/usr/bin/env bash
# One-time setup for a repository created from this template. Safe to re-run: each step skips what is done.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

# devbox gives gh its own config dir, so the first run inside devbox needs its own login.
if ! gh auth status >/dev/null 2>&1; then
  echo "bootstrap: gh is not signed in inside devbox; run: gh auth login" >&2
  exit 1
fi

# The strings docs/agents/triage-labels.md maps the skills' roles to; a template copies no labels.
while IFS='|' read -r name color description; do
  gh label create "$name" --color "$color" --description "$description" --force >/dev/null
done <<'LABELS'
needs-triage|fbca04|Maintainer needs to evaluate this issue
needs-info|d876e3|Waiting on reporter for more information
ready-for-agent|0e8a16|Fully specified, ready for an AFK agent
ready-for-human|5319e7|Requires human implementation
wontfix|ffffff|Will not be actioned
in-progress|1d76db|An unattended /dev-loop run is working this issue
awaiting-human|d93f0b|The run reached a conclusion someone must act on
failed|b60205|A stage broke — a crash, not a verdict; a retry may work
wayfinder:map|1d76db|Wayfinder map
wayfinder:research|1d76db|Wayfinder research
wayfinder:grilling|1d76db|Wayfinder grilling
wayfinder:prototype|1d76db|Wayfinder prototype
wayfinder:task|1d76db|Wayfinder task
LABELS
echo "bootstrap: labels ready"

if ! gh extension list | grep -q 'github/gh-stack'; then
  gh extension install github/gh-stack
fi
echo "bootstrap: gh stack ready"

# The key is never generated here: docs/secrets.md keeps that a step only its owner takes.
if grep -qxF '    age: []' .sops.yaml; then
  key=${SOPS_AGE_KEY_FILE:-}
  if [ -z "$key" ]; then
    if [ -n "${XDG_CONFIG_HOME:-}" ]; then key="$XDG_CONFIG_HOME/sops/age/keys.txt"
    elif [ "$(uname -s)" = Darwin ]; then key="$HOME/Library/Application Support/sops/age/keys.txt"
    else key="$HOME/.config/sops/age/keys.txt"; fi
  fi
  if [ ! -f "$key" ]; then
    echo "bootstrap: no age key at $key; create it as docs/secrets.md's Getting access step 1 says, then re-run" >&2
    exit 1
  fi
  recipient=$(age-keygen -y "$key" | head -n 1)
  login=$(gh api user --jq .login)
  awk -v r="$recipient" -v l="$login" \
    '$0 == "    age: []" { print "    age:"; print "      - " r " # " l; next } { print }' \
    .sops.yaml >.sops.yaml.tmp
  mv .sops.yaml.tmp .sops.yaml
fi
echo "bootstrap: SOPS recipient ready"

# Piped in, so the plaintext never lands on disk.
if [ ! -f secrets/postgres.sops.env ]; then
  mkdir -p secrets
  password=$(head -c 16 /dev/urandom | od -An -tx1 | tr -d ' \n')
  printf 'POSTGRES_USER=app\nPOSTGRES_PASSWORD=%s\nPOSTGRES_DB=app\n' "$password" |
    sops encrypt --filename-override secrets/postgres.sops.env >secrets/postgres.sops.env.tmp
  mv secrets/postgres.sops.env.tmp secrets/postgres.sops.env
fi
echo "bootstrap: secrets ready"

if [ -n "$(git status --porcelain -- .sops.yaml secrets)" ]; then
  echo "bootstrap: commit .sops.yaml and secrets/postgres.sops.env"
fi
