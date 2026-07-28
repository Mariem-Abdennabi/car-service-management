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

# Vite lives in web/, so its commands run there.
assets:
	cd web && bun run build

watch:
	cd web && bun run watch

run: templ
	go run .

# build depends on assets because the server refuses to start without the Vite
# manifest — a binary with no stylesheet or JavaScript is not deployable.
build: templ assets
	@mkdir -p $(dir $(BINARY))
	go build -o $(BINARY) .

test: templ
	go test ./...

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

.PHONY: help templ assets watch run build test vet fmt check tidy clean
