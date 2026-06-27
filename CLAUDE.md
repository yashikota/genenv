# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

genenv is a Go CLI tool that generates `.env` files from template files (e.g., `.env.example`), replacing `${placeholder}` syntax with cryptographically secure random values. Same-named placeholders share the same generated value. Escaped placeholders (`\${...}`) are preserved literally.

## Commands

```bash
# Build
go build -o genenv .

# Run tests (CI uses these exact flags)
go test -shuffle=on -race -v ./...

# Format check
gofmt -l .

# Run the tool
go run . .env.example
go run . -o .env.production -l 32 -c numeric .env.example
```

## Architecture

Single-binary CLI with two packages:

- `main.go` — CLI entry point: flag parsing, argument reordering (flags work before or after positional args), user prompts
- `internal/generator/` — Core logic: template parsing, placeholder detection/replacement, secure value generation, incremental .env updates

Key design: when `.env` already exists, it is the primary source — existing values are always preserved, only missing keys from the template are appended (with their comment groups). `--force` only regenerates values for keys that have `${...}` placeholders in the template.

## Testing

Tests in `main_test.go` are integration tests that build the binary and invoke it as a subprocess. Tests in `internal/generator/generator_test.go` are unit tests for the generator package. Both use `t.TempDir()` for isolation.

## Release

GoReleaser via GitHub Actions on tag push (`.goreleaser.yml`). Version constant lives in `main.go`.
