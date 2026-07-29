# Day-to-day commands for this project. Run `make` with no arguments to list them.
#
# Recipe lines must be indented with a TAB, not spaces. Make will refuse the file
# with "missing separator" otherwise — the most common Makefile error.
#
# Every target is listed in .PHONY at the bottom. Without that, Make treats a
# target name as a file it should build and skips the recipe whenever a file or
# directory of that name already exists: `make test` would silently do nothing the
# day someone adds a `test/` directory.

BINARY := bin/server

# The first target in the file is what a bare `make` runs, so help goes first.
help:
	@echo "Commands:"
	@echo "  make run      Start the server (reads .env, default port 8080)"
	@echo "  make build    Compile the server to $(BINARY)"
	@echo "  make templ    Regenerate Go code from .templ templates"
	@echo "  make sqlc     Regenerate Go code from sql/queries"
	@echo "  make migrate-up    Apply pending migrations (also runs on startup)"
	@echo "  make migrate-down  Roll back the most recent migration"
	@echo "  make assets   Build the Vite bundle into public/build"
	@echo "  make watch    Rebuild the bundle on every save"
	@echo "  make test     Run every test"
	@echo "  make vet      Report suspicious code that still compiles"
	@echo "  make fmt      Format every Go file in place"
	@echo "  make check    fmt + vet + test — run before committing"
	@echo "  make tidy     Sync go.mod and go.sum with the code's imports"
	@echo "  make clean    Delete build output"

# Templates compile to Go, so anything that compiles or tests the project must
# regenerate first. Otherwise an edited .templ file is silently ignored and you
# spend ten minutes wondering why the page has not changed.
templ:
	templ generate

# Regenerates internal/db from sql/queries and sql/migrations. Not a prerequisite
# of build or test: the generated code is committed, and regenerating needs the
# sqlc CLI. Run it after editing a query or a migration.
sqlc:
	sqlc generate

# Vite lives in web/, so its commands run there.
assets:
	cd web && bun run build

watch:
	cd web && bun run watch

# The application runs pending migrations itself on startup, so these targets are
# only for doing it deliberately — checking a new migration, or rolling one back.
#
# `set -a; . ./.env` loads .env the way the shell already knows how, so
# DATABASE_URL reaches the migrate CLI without the Makefile parsing the file.
migrate-up:
	@set -a; . ./.env; set +a; migrate -path sql/migrations -database "$$DATABASE_URL" up

migrate-down:
	@set -a; . ./.env; set +a; migrate -path sql/migrations -database "$$DATABASE_URL" down 1

run: templ
	go run .

# build depends on assets because the server refuses to start without the Vite
# manifest — a binary with no stylesheet or JavaScript is not deployable.
build: templ assets
	@mkdir -p $(dir $(BINARY))
	go build -o $(BINARY) .

# Loads .env so DATABASE_URL reaches the tests that need a real database. Without
# it those tests skip themselves, which would quietly hide a broken query.
test: templ
	@set -a; [ -f .env ] && . ./.env || true; set +a; go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

# A target with prerequisites and no recipe of its own: Make runs the three in
# order and stops at the first failure.
check: fmt vet test

tidy:
	go mod tidy

clean:
	rm -rf $(dir $(BINARY)) public/build

.PHONY: help templ sqlc migrate-up migrate-down assets watch run build test vet fmt check tidy clean
