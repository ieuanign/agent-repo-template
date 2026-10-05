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
	@echo "make ios-build device=                Generates the iOS project, builds the dev client and installs it on a Simulator or the named device"
	@echo "make ios                              Starts Metro for the dev client already installed on iOS"
	@echo "make android-build device=            Generates the Android project, builds the dev client and installs it on the Emulator or the named device"
	@echo "make android                          Starts Metro for the dev client already installed on Android"
	@echo "make e2e-mobile platform= device=     Runs the Maestro suite on a booted Simulator or Emulator against the running stack and Metro; not part of check"

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

.PHONY: ios-build
# Prebuild always runs first: run:ios generates the project only when ios/ is missing. --clean is passed by hand.
ios-build:
	cp mobile/environments/dev.env mobile/.env
	cd mobile && ../scripts/xcode-env.sh npx expo prebuild --no-install -p ios
	cd mobile && ../scripts/xcode-env.sh npx expo run:ios$(if $(device), --device "$(device)")

.PHONY: ios
ios:
	cp mobile/environments/dev.env mobile/.env
	npm run ios -w mobile

.PHONY: android-build
# adb reverse lets the device reach the stack on localhost:4008; ignored with no device attached so expo's own message shows.
android-build:
	cp mobile/environments/dev.env mobile/.env
	-adb reverse tcp:4008 tcp:4008
	cd mobile && npx expo prebuild --no-install -p android
	cd mobile && npx expo run:android$(if $(device), --device "$(device)")

.PHONY: android
android:
	cp mobile/environments/dev.env mobile/.env
	-adb reverse tcp:4008 tcp:4008
	npm run android -w mobile

.PHONY: e2e-mobile
# Checks and never sets up: each missing prerequisite fails naming its fix. iOS runs xcrun and maestro through xcode-env.sh, as nix's xcrun breaks Maestro's iOS driver.
e2e-mobile:
	@case "$(platform)" in ios|android) ;; *) echo "e2e-mobile: platform= must be ios or android" >&2; exit 1;; esac; \
	app=com.agent.template; wrap=; device="$(device)"; \
	if [ "$(platform)" = ios ]; then wrap=scripts/xcode-env.sh; fi; \
	if [ -z "$$device" ]; then \
		if [ "$(platform)" = ios ]; then \
			booted=$$(scripts/xcode-env.sh xcrun simctl list devices booted -j | node -e 'const d = JSON.parse(require("fs").readFileSync(0, "utf8")).devices; for (const x of Object.values(d).flat()) if (x.state === "Booted") console.log(`$${x.udid} $${x.name}`)'); \
		else \
			booted=$$(adb devices | awk 'NR > 1 && $$2 == "device" { print $$1 }'); \
		fi; \
		n=$$(printf '%s\n' "$$booted" | grep -c .); \
		if [ "$$n" != 1 ]; then echo "e2e-mobile: $$n booted $(platform) devices, need exactly one; boot a Simulator or Emulator, or pass device=<id>" >&2; [ -z "$$booted" ] || echo "$$booted" >&2; exit 1; fi; \
		device="$${booted%% *}"; \
	fi; \
	resp=$$(curl -s --max-time 5 -w ' %{http_code}' http://localhost:4008/api/health); body="$${resp% *}"; \
	if [ "$${resp##* }" != 200 ]; then echo "e2e-mobile: http://localhost:4008/api/health is not up; run make dev" >&2; exit 1; fi; \
	if [ "$$(curl -s -o /dev/null --max-time 5 -w '%{http_code}' http://localhost:8081/status)" != 200 ]; then echo "e2e-mobile: Metro is not running on localhost:8081; run make ios or make android" >&2; exit 1; fi; \
	if [ "$(platform)" = ios ]; then scripts/xcode-env.sh xcrun simctl get_app_container "$$device" $$app >/dev/null 2>&1; \
	else adb -s "$$device" shell pm path $$app 2>/dev/null | grep -q '^package:'; fi \
		|| { echo "e2e-mobile: $$app is not installed on $$device; run make ios-build or make android-build" >&2; exit 1; }; \
	if [ "$(platform)" = android ]; then \
		adb -s "$$device" reverse tcp:4008 tcp:4008 >/dev/null && adb -s "$$device" reverse tcp:8081 tcp:8081 >/dev/null \
			|| { echo "e2e-mobile: adb reverse failed on $$device" >&2; exit 1; }; \
	fi; \
	version=$$(BODY="$$body" node -e 'process.stdout.write(String(JSON.parse(process.env.BODY).payload?.version ?? ""))'); \
	if [ -z "$$version" ]; then echo "e2e-mobile: /api/health body lacks payload.version: $$body" >&2; exit 1; fi; \
	rm -rf e2e-mobile/.output/$(platform); \
	$$wrap maestro --device "$$device" test -e BACKEND_VERSION="$$version" --test-output-dir e2e-mobile/.output/$(platform) e2e-mobile
