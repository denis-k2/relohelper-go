include .envrc

# Formatting variables
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput sgr0)

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

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run ./cmd/api -db-dsn=${RELOHELPER_DB_DSN}

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
audit:
	@echo '${YELLOW}===> Running code quality checks...${RESET}'
	-@go mod tidy -diff
	-@go mod verify
	@echo '${YELLOW}===> Running modernize...${RESET}'
	-@modernize -test ./...
	@echo '${YELLOW}===> Running linter...${RESET}'
	-@golangci-lint run
	@echo '${YELLOW}===> Running full test suite...${RESET}'
	-@go test -count=1 ./... -args -db-dsn=${RELOHELPER_TEST_DB_DSN}

# ==================================================================================== #
# TESTING
# ==================================================================================== #

## test: run all tests (fast)
.PHONY: test
test:
	@echo 'Running tests...'
	@go test -count=1 ./... -args -db-dsn=${RELOHELPER_TEST_DB_DSN}

## test/v: run all tests with verbose output and logs at debug level
.PHONY: test/v
test/v:
	@echo 'Running tests (verbose)...'
	@go test -v -count=1 ./... -args -db-dsn=${RELOHELPER_TEST_DB_DSN} -env testLogs

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
