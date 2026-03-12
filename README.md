# Gruyere

Age your APIs gracefully. Interactive TUI for generating API changelogs.

## Install

```sh
go install github.com/CheeziCrew/gruyere@latest
```

Or grab a binary from [Releases](https://github.com/CheeziCrew/gruyere/releases).

## Usage

### TUI (default)

```sh
gruyere
```

### CLI

```sh
gruyere <base-branch> <feature-branch>
gruyere -o changelog.md main feature/UF-123
gruyere --prepend main feature/UF-123
```

## What it does

1. Finds `openapi.yaml` in your repo
2. Fetches the spec from both branches via `git show`
3. Compares endpoints (added/removed) and schemas (added/removed/modified fields)
4. Generates a markdown changelog

## Part of the Swiss Tools family

Built on [curd](https://github.com/CheeziCrew/Curd) — the shared TUI component library.
