.DEFAULT_GOAL := help

.PHONY: help
# A new target adds its help line here in the same change.
help:
	@echo "make help                             Lists every target (also what bare make runs)"
	@echo "make bootstrap                        One-time setup of a repo made from this template: labels, gh stack, SOPS; see README"
	@echo "make check                            Lints and tests every workspace, then runs smoke (starts and stops the stack)"
	@echo "make secrets-decrypt                  Writes secrets/*.env from the .sops.env files; needs age access, see docs/secrets.md"
	@echo "make dev                              Decrypts secrets, then builds and starts the stack; needs Docker"
	@echo "make down                             Stops the stack"
	@echo "make smoke                            Builds and starts the stack, probes its health endpoints, always stops it"
	@echo "make e2e                              Runs the browser tests against the stack make dev started; not part of check"
	@echo "make migrate-create name=             Creates a new up/down migration pair named <module>_<desc>"
	@echo "make migrate-up                       Applies pending migrations; starts the Compose PostgreSQL if down"
	@echo "make migrate-version                  Shows the current version and whether the database is dirty"
	@echo "make migrate-down                     Reverts the last migration, dev only"
	@echo "make migrate-redo                     Reverts and reapplies the last migration, dev only"
	@echo "make migrate-force version= confirm=  Clears the dirty flag after a failed migration; confirm= names the database"
	@echo "make seed                             Re-runs the seed data against the Compose PostgreSQL, starting it if down"
	@echo "make db-reset                         Drops the dev database, then migrates and seeds; destructive"

.PHONY: bootstrap
bootstrap:
	./scripts/bootstrap.sh

.PHONY: check
check:
	./node_modules/.bin/turbo run lint test
	$(MAKE) smoke

.PHONY: secrets-decrypt
# Decrypt via a temp file so a failed run never leaves a partial or empty .env behind.
secrets-decrypt:
	@for f in secrets/*.sops.env; do \
		out="$${f%.sops.env}.env"; \
		sops decrypt "$$f" > "$$out.tmp" || { rm -f "$$out.tmp"; echo "secrets-decrypt: cannot decrypt $$f; see docs/secrets.md to get access" >&2; exit 1; }; \
		mv "$$out.tmp" "$$out"; \
	done

.PHONY: dev
dev: secrets-decrypt
	docker compose up -d --build --renew-anon-volumes --wait --wait-timeout 60

.PHONY: down
down:
	docker compose rm --stop --force --volumes web
	docker compose down

.PHONY: smoke
# Poll: Traefik routes to a service only once it is healthy, and next dev compiles on first request. Each probe is path=required JSON field paths; down always runs.
smoke: dev
	@status=0; deadline=$$(( $$(date +%s) + 120 )); \
	for probe in /api/health=payload.version,payload.environment /api/health/ready= /health=version,environment /health/ready= /=; do \
		url="http://localhost:4008$${probe%%=*}"; fields="$${probe#*=}"; \
		until resp=$$(curl -s --max-time 5 -w ' %{http_code}' "$$url"); [ "$${resp##* }" = 200 ] || [ $$(date +%s) -ge $$deadline ]; do sleep 1; done; \
		code="$${resp##* }"; body="$${resp% *}"; \
		if [ "$$code" != 200 ]; then echo "smoke: $$url -> $$code, expected 200" >&2; status=1; \
		elif [ -z "$$fields" ] || BODY="$$body" FIELDS="$$fields" node -e 'const b = JSON.parse(process.env.BODY); if (process.env.FIELDS.split(",").some((f) => !f.split(".").reduce((o, k) => o?.[k], b))) process.exit(1)'; then echo "smoke: $$url -> $$code"; \
		else echo "smoke: $$url -> $$code, body lacks one of $$fields: $$body" >&2; status=1; fi; \
	done; \
	if [ $$status != 0 ]; then docker compose logs backend web >&2; fi; \
	$(MAKE) down; \
	exit $$status

.PHONY: e2e
# No dev prerequisite: runs against the stack make dev started, and stays out of check because browsers are slow.
e2e:
	docker compose --profile e2e run --rm e2e

.PHONY: migrate-create
migrate-create:
	cd backend && go tool migrate create -ext sql -dir migrations $(name)

.PHONY: migrate-up
migrate-up: secrets-decrypt
	docker compose run --rm migrate up

.PHONY: migrate-version
migrate-version: secrets-decrypt
	docker compose run --rm migrate version

.PHONY: migrate-down
migrate-down: secrets-decrypt
	docker compose run --rm migrate down

.PHONY: migrate-redo
migrate-redo: secrets-decrypt
	docker compose run --rm migrate redo

.PHONY: migrate-force
# The binary validates both flags, so an empty one is refused there.
migrate-force: secrets-decrypt
	docker compose run --rm migrate force -version=$(version) -confirm=$(confirm)

.PHONY: seed
# --build: worktrees share one image tag, so a stale image could run another branch's seed.
seed: secrets-decrypt
	docker compose run --rm --build seed

.PHONY: db-reset
# --no-deps so a dirty database or failing migrate cannot block the reset meant to clear it.
db-reset: secrets-decrypt
	docker compose up -d --wait postgres
	docker compose run --rm --build --no-deps seed -reset
	$(MAKE) migrate-up
	$(MAKE) seed
