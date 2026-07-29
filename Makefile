# Formatting variables
ifneq ($(TERM),)
YELLOW := $(shell tput setaf 3 2>/dev/null)
RESET  := $(shell tput sgr0 2>/dev/null)
endif

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: require/test-db
require/test-db:
	@test -n "${RELOHELPER_TEST_DB_DSN}" || (echo 'RELOHELPER_TEST_DB_DSN is required; load .envrc with direnv or run source .envrc' && exit 1)

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the cmd/api application with authentication disabled
.PHONY: run/api
run/api:
	@RELOHELPER_AUTH_ENABLED=false go run ./cmd/api

## run/api/auth: run the cmd/api application with authentication enabled
.PHONY: run/api/auth
run/api/auth:
	@RELOHELPER_AUTH_ENABLED=true go run ./cmd/api

## run/api/load: run the cmd/api application for load testing
.PHONY: run/api/load
run/api/load:
	@RELOHELPER_AUTH_ENABLED=false RELOHELPER_LIMITER_ENABLED=false go run ./cmd/api

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	@psql ${RELOHELPER_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	@migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	@migrate -path ./migrations -database ${RELOHELPER_DB_DSN} up

## db/test/prepare: apply migrations and load deterministic test fixtures
.PHONY: db/test/prepare
db/test/prepare:
	@test -n "${RELOHELPER_TEST_DB_DSN}" || (echo 'RELOHELPER_TEST_DB_DSN is required' && exit 1)
	@dsn="${RELOHELPER_TEST_DB_DSN}"; \
		db_url="$${dsn%%\?*}"; \
		db_name="$${db_url##*/}"; \
		case "$$(printf '%s' "$${db_name}" | tr '[:upper:]' '[:lower:]')" in \
			*test*) ;; \
			*) echo "Refusing to prepare database '$${db_name}': its name must contain 'test'"; exit 1 ;; \
		esac
	@echo 'Preparing test database...'
	@migrate -path ./migrations -database "${RELOHELPER_TEST_DB_DSN}" up
	@dsn="${RELOHELPER_TEST_DB_DSN}"; \
		case "$${dsn}" in \
			*\?*) fixture_dsn="$${dsn}&x-migrations-table=relohelper_test_fixtures" ;; \
			*) fixture_dsn="$${dsn}?x-migrations-table=relohelper_test_fixtures" ;; \
		esac; \
		migrate -path ./tests/fixtures -database "$${fixture_dsn}" force 0; \
		migrate -path ./tests/fixtures -database "$${fixture_dsn}" up

## db/sqlc/generate: generate type-safe database query code
.PHONY: db/sqlc/generate
db/sqlc/generate:
	@echo 'Generating sqlc database code...'
	@sqlc generate

## db/sqlc/check: verify generated database query code is up to date
.PHONY: db/sqlc/check
db/sqlc/check:
	@echo 'Checking sqlc generated database code...'
	@before=$$(mktemp); after=$$(mktemp); \
		git diff -- internal/db/*.go > $${before}; \
		sqlc generate; \
		git diff -- internal/db/*.go > $${after}; \
		diff -u $${before} $${after}; \
		status=$$?; \
		rm -f $${before} $${after}; \
		exit $${status}

## db/sqlc/vet: check SQL queries for correctness
.PHONY: db/sqlc/vet
db/sqlc/vet:
	@echo 'Checking sqlc queries...'
	@sqlc vet

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format code and tidy module dependencies
.PHONY: tidy
tidy:
	@echo '${YELLOW}===> Formatting code${RESET}'
	-@goimports -w .
	@echo '${YELLOW}===> Running linter fixes${RESET}'
	-@golangci-lint run --fix
	@echo '${YELLOW}===> Tidying module dependencies${RESET}'
	@go mod tidy
	@echo '${YELLOW}===> Verifying and vendoring dependencies${RESET}'
	-@go mod verify

## audit: run quality control checks (no changes to code)
.PHONY: audit
audit: require/test-db
	@echo '${YELLOW}===> Running code quality checks...${RESET}'
	@go mod tidy -diff
	@go mod verify
	@echo '${YELLOW}===> Running modernize...${RESET}'
	@go list ./... | grep -v '/internal/db$$' | xargs modernize -test
	@echo '${YELLOW}===> Running linter...${RESET}'
	@golangci-lint run
	@echo '${YELLOW}===> Checking generated database code...${RESET}'
	@$(MAKE) db/sqlc/check
	@$(MAKE) db/sqlc/vet
	@echo '${YELLOW}===> Running full test suite...${RESET}'
	@go test -count=1 ./...

# ==================================================================================== #
# TESTING
# ==================================================================================== #

## test: run all tests (fast)
.PHONY: test
test: require/test-db
	@echo 'Running tests...'
	@go test -count=1 ./...

## test/v: run all tests with verbose output and logs at debug level
.PHONY: test/v
test/v: require/test-db
	@echo 'Running tests (verbose)...'
	@RELOHELPER_TEST_LOGS=true go test -v -count=1 ./...

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags='-s' -o=./bin/api ./cmd/api
    GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api

# ==================================================================================== #
# PRODUCTION
# ==================================================================================== #

## production/connect: connect to the production server
.PHONY: production/connect
production/connect:
	ssh relohelper@${RELOHELPER_PROD_HOST}

## production/deploy/api: deploy the api to production
.PHONY: production/deploy/api
production/deploy/api:
	rsync -P ./bin/api relohelper@${RELOHELPER_PROD_HOST}:~
	rsync -rP --delete ./migrations relohelper@${RELOHELPER_PROD_HOST}:~
	rsync -P ./remote/production/api.service relohelper@${RELOHELPER_PROD_HOST}:~
	rsync -P ./remote/production/Caddyfile relohelper@${RELOHELPER_PROD_HOST}:~
	ssh -t relohelper@${RELOHELPER_PROD_HOST} '\
		migrate -path ~/migrations -database $$RELOHELPER_DB_DSN up \
		&& sudo mv ~/api.service /etc/systemd/system/ \
		&& sudo systemctl enable api \
		&& sudo systemctl restart api \
		&& sudo mv ~/Caddyfile /etc/caddy/ \
		&& sudo systemctl reload caddy \
	'
